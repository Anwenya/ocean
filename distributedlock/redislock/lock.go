package redislock

import (
	"context"
	"errors"
	"fmt"
	"github.com/Anwenya/ocean/distributedlock/redislock/client"
	"github.com/Anwenya/ocean/distributedlock/redislock/lua"
	"github.com/Anwenya/ocean/os"
	"github.com/gomodule/redigo/redis"
	"sync/atomic"
	"time"
)

const RedisLockKeyPrefix = "REDIS_LOCK_PREFIX_"

var ErrLockAcquiredByOthers = errors.New("lock is acquired by others")

var ErrNil = redis.ErrNil

func IsRetryableErr(err error) bool {
	return errors.Is(err, ErrLockAcquiredByOthers)
}

// RedisLock 基于redis的分布式锁
type RedisLock struct {
	LockOptions
	key    string
	token  string
	client client.LockClient

	// 看门狗运作标识
	runningDog int32
	// 停止看门狗
	stopDog context.CancelFunc
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

func (rl *RedisLock) Lock(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			return
		}
		// 抢锁成功 启动续约机制
		rl.watchDog(ctx)
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
	// 先查询锁是否属于自己
	reply, err := rl.client.SetNEX(ctx, rl.getLockKey(), rl.token, rl.expire)
	if err != nil {
		return err
	}

	// 其他非预期结果
	if reply != 1 {
		return fmt.Errorf("tryLock:reply: %d, err: %w", reply, ErrLockAcquiredByOthers)
	}

	return nil
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

// 启动看门狗 也就是 续约
func (rl *RedisLock) watchDog(ctx context.Context) {
	if !rl.watchDogMode {
		return
	}

	for !atomic.CompareAndSwapInt32(&rl.runningDog, 0, 1) {
	}

	ctx, rl.stopDog = context.WithCancel(ctx)
	go func() {
		defer func() {
			atomic.StoreInt32(&rl.runningDog, 0)
		}()
		rl.runWatchDog(ctx)
	}()
}

func (rl *RedisLock) runWatchDog(ctx context.Context) {
	ticker := time.NewTicker(WatchDogWorkStep * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 为避免因为网络延迟而导致锁被提前释放的问题 续约时需要把锁的过期时长额外增加 5 s
		_ = rl.DelayExpire(ctx, WatchDogWorkStep+time.Second*5)
	}
}

// DelayExpire 更新锁的过期时间
func (rl *RedisLock) DelayExpire(ctx context.Context, expire time.Duration) error {
	keysAndArgs := []interface{}{rl.getLockKey(), rl.token, expire.Seconds()}
	reply, err := rl.client.Eval(ctx, lua.LuaCheckAndExpireDistributionLock, 1, keysAndArgs)
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
		// 停止看门狗
		if rl.stopDog != nil {
			rl.stopDog()
		}
	}()

	keysAndArgs := []interface{}{rl.getLockKey(), rl.token}
	reply, err := rl.client.Eval(ctx, lua.LuaCheckAndDeleteDistributionLock, 1, keysAndArgs)
	if err != nil {
		return err
	}

	if ret, _ := reply.(int64); ret != 1 {
		return errors.New("can not unlock without ownership of lock")
	}

	return nil
}

func (rl *RedisLock) getLockKey() string {
	return RedisLockKeyPrefix + rl.key
}
