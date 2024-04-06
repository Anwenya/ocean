package redislock

import (
	"github.com/Anwenya/ocean/distributedlock/redislock/client"
	"time"
)

const (
	// DefaultLockExpire 默认的分布式锁过期时间
	DefaultLockExpire = time.Second * 30
	// WatchDogWorkStep 看门狗工作时间间隙/续约间隔
	WatchDogWorkStep = time.Second * 10
)

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

	// 显式指定锁的过期时间 则此时会启动看门狗
	o.expire = DefaultLockExpire
	o.watchDogMode = true
}

type LockOptions struct {
	isBlock      bool
	blockWaiting time.Duration
	expire       time.Duration
	watchDogMode bool
}

type RedLockOption func(*RedLockOptions)

type RedLockOptions struct {
	singleNodesTimeout time.Duration
	expireDuration     time.Duration
}

func WithSingleNodesTimeout(singleNodesTimeout time.Duration) RedLockOption {
	return func(o *RedLockOptions) {
		o.singleNodesTimeout = singleNodesTimeout
	}
}

func WithRedLockExpireDuration(expireDuration time.Duration) RedLockOption {
	return func(o *RedLockOptions) {
		o.expireDuration = expireDuration
	}
}

type SingleNodeConf struct {
	Network  string
	Address  string
	Password string
	Opts     []client.Option
}

func repairRedLock(o *RedLockOptions) {
	if o.singleNodesTimeout <= 0 {
		o.singleNodesTimeout = DefaultSingleLockTimeout
	}
}
