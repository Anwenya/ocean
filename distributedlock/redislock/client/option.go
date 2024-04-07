package client

import "time"

const (
	// DefaultIdleTimeout 默认连接池超过 10 s 释放连接
	DefaultIdleTimeout = time.Second * 10
	// DefaultMaxActive 默认最大激活连接数
	DefaultMaxActive = 100
	// DefaultMaxIdle 默认最大空闲连接数
	DefaultMaxIdle = 20
)

type Options struct {
	maxIdle     int
	idleTimeout time.Duration
	maxActive   int
	wait        bool
	// 必填参数
	network  string
	address  string
	password string
}

type Option func(c *Options)

func WithMaxIdle(maxIdle int) Option {
	return func(c *Options) {
		c.maxIdle = maxIdle
	}
}

func WithIdleTimeoutSeconds(idleTimeout time.Duration) Option {
	return func(c *Options) {
		c.idleTimeout = idleTimeout
	}
}

func WithMaxActive(maxActive int) Option {
	return func(c *Options) {
		c.maxActive = maxActive
	}
}

func WithWaitMode() Option {
	return func(c *Options) {
		c.wait = true
	}
}

func repair(c *Options) {
	if c.maxIdle < 0 {
		c.maxIdle = DefaultMaxIdle
	}

	if c.idleTimeout < 0 {
		c.idleTimeout = DefaultIdleTimeout
	}

	if c.maxActive < 0 {
		c.maxActive = DefaultMaxActive
	}
}
