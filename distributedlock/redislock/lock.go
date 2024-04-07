package redislock

import (
	"context"
	"errors"
	"fmt"
	"github.com/Anwenya/ocean/distributedlock/redislock/client"
	"github.com/Anwenya/ocean/distributedlock/redislock/lua"
	"github.com/Anwenya/ocean/os"
	"github.com/gomodule/redigo/redis"
	"strings"
	"sync/atomic"
	"time"
)

// RedisLockKeyPrefix 前缀
const RedisLockKeyPrefix = "RLP:"

// ErrLockAcquiredByOthers 锁已经被其他人持有
var ErrLockAcquiredByOthers = errors.New("lock is acquired by others")

var ErrNil = redis.ErrNil

func IsRetryableErr(err error) bool {
	return errors.Is(err, ErrLockAcquiredByOthers)
}

// RedisLock 基于redis的分布式锁
type RedisLock struct {
	LockOptions
	// 锁的key
	key string
	// 当前身份标识
	token  string
	client client.LockClient

	// 正在续约标识
	leasing int32
	// 停止续约
	stopLease context.CancelFunc
}

func NewRedisLock(key string, client client.LockClient, opts ...LockOption) *RedisLock {
	rl := RedisLock{
		key:    key,
		token:  os.GetCurrentProcessAndGoroutineIDStr(),
		client: client,
	}

	for _, opt := range opts {
		opt(&rl.LockOptions)
	}

	repairLock(&rl.LockOptions)

	return &rl
}

// Lock 加锁
func (rl *RedisLock) Lock(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			return
		}
		// 抢锁成功 启动续约机制
		rl.startLease(ctx)
	}()

	// 尝试抢锁
	err = rl.tryLock(ctx)
	// 抢到了锁
	if err == nil {
		return nil
	}

	// 非阻塞模式 直接返回
	if !rl.isBlock {
		return err
	}

	// 是否可以重试
	// 如果错误是[锁已被其他人持有]则可以重试
	if !IsRetryableErr(err) {
		return err
	}

	// 阻塞模式 轮询抢锁
	err = rl.blockingLock(ctx)
	return
}

// 尝试抢锁
func (rl *RedisLock) tryLock(ctx context.Context) error {
	// 抢锁
	keysAndArgs := []interface{}{rl.getLockKey(), rl.token, int(rl.expire.Seconds())}

	// 执行抢锁逻辑
	reply, err := rl.client.Eval(ctx, lua.LuaLock, 1, keysAndArgs)
	if err != nil {
		return err
	}

	// 成功
	if respStr, ok := reply.(string); ok && strings.ToLower(respStr) == "ok" {
		return nil
	}
	// 其他情况视为失败
	return ErrLockAcquiredByOthers
}

// 阻塞循环抢锁
func (rl *RedisLock) blockingLock(ctx context.Context) error {
	// 超时时间
	timeout := time.After(rl.blockWaiting)
	// 轮询抢锁
	ticker := time.NewTicker(time.Millisecond * 50)
	defer ticker.Stop()

	for range ticker.C {
		select {
		// 链路终止
		case <-ctx.Done():
			return fmt.Errorf("lock failed, ctx timeout, err: %w", ctx.Err())
		// 超时
		case <-timeout:
			return fmt.Errorf("block waiting time out")
		// 尝试取锁
		default:
			err := rl.tryLock(ctx)
			if err == nil {
				return nil
			}

			if !IsRetryableErr(err) {
				return err
			}
		}
	}
	return nil
}

// 启动续约
func (rl *RedisLock) startLease(ctx context.Context) {
	// 在释放后立刻又抢到锁时等待前一次退出
	for !atomic.CompareAndSwapInt32(&rl.leasing, 0, 1) {
	}

	ctx, rl.stopLease = context.WithCancel(ctx)
	go func() {
		defer func() {
			atomic.StoreInt32(&rl.leasing, 0)
		}()
		rl.runLeasing(ctx)
	}()
}

// 续约
func (rl *RedisLock) runLeasing(ctx context.Context) {
	// 过期时间的 1/3
	ticker := time.NewTicker(rl.expire / 3)
	count := 0

	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
			err := rl.DelayExpire(ctx, rl.expire)
			if err != nil {
				count++
				// 连续续约失败5次
				if count >= 5 {
					return
				}
			}
			count = 0
		}
	}
}

// DelayExpire 更新锁的过期时间
func (rl *RedisLock) DelayExpire(ctx context.Context, expire time.Duration) error {
	keysAndArgs := []interface{}{rl.getLockKey(), rl.token, int(expire.Seconds())}
	reply, err := rl.client.Eval(ctx, lua.LuaLease, 1, keysAndArgs)
	if err != nil {
		return err
	}

	if ret, _ := reply.(int64); ret != 1 {
		return errors.New("can not expire lock without ownership of lock")
	}

	return nil
}

// Unlock 解锁
func (rl *RedisLock) Unlock(ctx context.Context) error {
	defer func() {
		// 停止续约
		if rl.stopLease != nil {
			rl.stopLease()
		}
	}()

	keysAndArgs := []interface{}{rl.getLockKey(), rl.token}
	reply, err := rl.client.Eval(ctx, lua.LuaUnlock, 1, keysAndArgs)
	if err != nil {
		return err
	}

	// 未持有锁
	if ret, _ := reply.(int64); ret != 1 {
		return errors.New("can not unlock without ownership of lock")
	}

	return nil
}

func (rl *RedisLock) getLockKey() string {
	return RedisLockKeyPrefix + rl.key
}

// TODO:可重入锁
