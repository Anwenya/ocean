package client

import (
	"context"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

// redis相关的操作

// Client redis客户端
type Client struct {
	Options
	pool *redis.Pool
}

func NewClient(network, address, password string, opts ...Option) *Client {
	clientOptions := Options{
		network:  network,
		address:  address,
		password: password,
	}

	// 设置配置
	for _, opt := range opts {
		opt(&clientOptions)
	}

	// 过滤非法配置
	repair(&clientOptions)

	client := Client{
		Options: clientOptions,
	}

	// 创建连接池
	client.initRedisPool()

	return &client
}

// 初始化redis连接池
func (c *Client) initRedisPool() {
	c.pool = &redis.Pool{
		MaxIdle:     c.maxIdle,
		IdleTimeout: c.idleTimeout,
		Dial: func() (redis.Conn, error) {
			c, err := c.getRedisConn()
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		MaxActive: c.maxActive,
		Wait:      c.wait,
		TestOnBorrow: func(c redis.Conn, lastUsed time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// 建立redis连接
func (c *Client) getRedisConn() (redis.Conn, error) {
	if c.address == "" {
		panic(any(fmt.Sprintf("invalid address:%s", c.address)))
	}

	var dialOpts []redis.DialOption
	if len(c.password) > 0 {
		dialOpts = append(dialOpts, redis.DialPassword(c.password))
	}

	conn, err := redis.DialContext(context.Background(), c.network, c.address, dialOpts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// GetConn 从连接池中获得一个连接
func (c *Client) GetConn(ctx context.Context) (redis.Conn, error) {
	return c.pool.GetContext(ctx)
}

// Eval
// 执行lua脚本
func (c *Client) Eval(ctx context.Context, src string, keyCount int, keysAndArgs []interface{}) (interface{}, error) {
	args := make([]interface{}, 2+len(keysAndArgs))
	args[0] = src
	args[1] = keyCount

	copy(args[2:], keysAndArgs)

	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return -1, err
	}

	defer func() {
		_ = conn.Close()
	}()

	return conn.Do("EVAL", args...)
}
