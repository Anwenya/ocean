package redislock

import (
	"time"
)

const (
	// DefaultLockExpire 默认的分布式锁过期时间
	DefaultLockExpire = time.Second * 30
)

type LockOptions struct {
	// 是否会阻塞抢锁
	isBlock bool
	// 阻塞抢锁的超时时间
	blockWaiting time.Duration
	// 锁的过期时间
	expire time.Duration
}

type LockOption func(*LockOptions)

func WithBlock() LockOption {
	return func(o *LockOptions) {
		o.isBlock = true
	}
}

func WithBlockWaitingSeconds(waiting time.Duration) LockOption {
	return func(o *LockOptions) {
		o.blockWaiting = waiting
	}
}

func WithExpireSeconds(expire time.Duration) LockOption {
	return func(o *LockOptions) {
		o.expire = expire
	}
}

func repairLock(o *LockOptions) {
	if o.isBlock && o.blockWaiting <= 0 {
		// 默认阻塞等待时间上限为 5 秒
		o.blockWaiting = time.Second * 5
	}

	// 倘若未设置分布式锁的过期时间，则会启动 watchdog
	if o.expire > 0 {
		return
	}

	// 不指定锁的过期时间使用默认值
	o.expire = DefaultLockExpire
}
