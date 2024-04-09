package etcdlock

import (
	"context"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"testing"
	"time"
)

// etcd客户端提供了分布式锁的实现 这里记录如何使用

func TestEtcdLock(t *testing.T) {

	client, err := clientv3.New(
		clientv3.Config{
			Endpoints:   []string{"192.168.2.130:12379"},
			DialTimeout: time.Second * 5,
		},
	)
	require.NoError(t, err)

	session1, err := concurrency.NewSession(client, concurrency.WithTTL(10))
	require.NoError(t, err)
	mutex1 := concurrency.NewMutex(session1, "/lock")
	// 加锁成功
	err = mutex1.Lock(context.Background())
	require.NoError(t, err)

	session2, err := concurrency.NewSession(client, concurrency.WithTTL(10))
	require.NoError(t, err)
	mutex2 := concurrency.NewMutex(session2, "/lock")
	// 加锁失败
	err = mutex2.TryLock(context.Background())
	t.Log(err)
}
