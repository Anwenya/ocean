package redislock

import (
	"context"
	"github.com/Anwenya/ocean/distributedlock/redislock/client"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type TestSuite struct {
	suite.Suite
	client *client.Client
}

func (ts *TestSuite) SetupSuite() {
	ts.client = client.NewClient("tcp", "192.168.2.130:6379", "")
}

func (ts *TestSuite) TestRedisLock_Lock() {
	t := ts.T()

	// 非阻塞锁
	lock1 := NewRedisLock("TEST", ts.client)

	// 总链路超时10秒
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// 正常拿到锁 默认超时时间是30秒 刷新间隔是10秒
	err := lock1.Lock(ctx)
	require.NoError(t, err)

	// 抢锁失败
	lock2 := NewRedisLock("TEST", ts.client)
	err = lock2.Lock(context.Background())
	require.Equal(t, ErrLockAcquiredByOthers, err)

	// 解锁
	err = lock1.Unlock(context.Background())
	require.NoError(t, err)

	// 再次抢锁成功
	err = lock2.Lock(context.Background())
	require.NoError(t, err)
}

func (ts *TestSuite) TestRedisLock_Lease() {
	t := ts.T()
	// 过期时间是3秒 续约间隔是1秒
	lock1 := NewRedisLock("TEST", ts.client, WithExpireSeconds(time.Second*3))
	// 加锁
	err := lock1.Lock(context.Background())
	require.NoError(t, err)
	// 6秒后还能正常解锁 说明续约成功
	time.Sleep(time.Second * 6)
	// 解锁
	err = lock1.Unlock(context.Background())
	require.NoError(t, err)
}

func TestRedisLock(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
