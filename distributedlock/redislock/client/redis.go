package client

import (
	"context"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strings"
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

// Get the value of key.
// If the key does not exist the special value nil is returned.
// An error is returned if the value stored at key is not a string,
// because GET only handles string values.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("get:invalid key:%s", key)
	}

	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return "", err
	}

	// 对于pool来说 只是将该连接入池
	defer func() {
		_ = conn.Close()
	}()

	// 执行get操作
	return redis.String(conn.Do("GET", key))
}

// SetNEX
// Set key to hold the string value and set key to timeout after a given number of seconds.
// When key already holds a value, no operation is performed.
// SETNX is short for "SET if Not eXists".
func (c *Client) SetNEX(ctx context.Context, key, value string, expire time.Duration) (int64, error) {
	if key == "" || value == "" {
		return -1, fmt.Errorf("setNEX:invalid key or value:%s, %s", key, value)
	}

	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return -1, err
	}

	defer func() {
		_ = conn.Close()
	}()

	reply, err := conn.Do("SET", key, value, "EX", expire.Seconds(), "NX")
	if err != nil {
		return -1, err
	}

	// 设置成功
	if respStr, ok := reply.(string); ok && strings.ToLower(respStr) == "ok" {
		return 1, nil
	}

	return redis.Int64(reply, err)
}

// Del
// Removes the specified keys.
// A key is ignored if it does not exist.
// key 可以是复数 中间用空格隔开
func (c *Client) Del(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("del:invalid key:%s", key)
	}

	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = conn.Close()
	}()

	_, err = conn.Do("DEL", key)
	return err
}

// Incr
// Increments the number stored at key by one.
// If the key does not exist, it is set to 0
// 值的合法范围是 有符号64位整数
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	if key == "" {
		return -1, fmt.Errorf("incr:invalid key:%s", key)
	}

	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return -1, err
	}

	defer func() {
		_ = conn.Close()
	}()

	return redis.Int64(conn.Do("INCR", key))
}

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
