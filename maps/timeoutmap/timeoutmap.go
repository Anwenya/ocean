package timeoutmap

import (
	"context"
	"sync"
	"time"
)

type TimeoutMap[K comparable, V any] struct {
	sync.Mutex
	innerMap    map[K]V
	waitChanMap map[K]chan struct{}
}

func NewTimeoutMap[K comparable, V any]() *TimeoutMap[K, V] {
	return &TimeoutMap[K, V]{
		innerMap:    make(map[K]V),
		waitChanMap: make(map[K]chan struct{}),
	}
}

func (mp *TimeoutMap[K, V]) Put(key K, value V) {
	mp.Lock()
	defer mp.Unlock()

	mp.innerMap[key] = value

	ch, ok := mp.waitChanMap[key]
	if !ok {
		return
	}

	// 有g在等待
	select {
	// 如果对应的ch已经关闭了就不要重复关闭
	case <-ch:
		return
	// 否则就关闭ch唤醒等待的g
	default:
		close(ch)
	}

}

func (mp *TimeoutMap[K, V]) Get(key K, waitTimeout time.Duration) (value V, err error) {
	mp.Lock()

	value, ok := mp.innerMap[key]
	// 如果存在key直接返回
	if ok {
		mp.Unlock()
		return
	}

	// key不存在要等待
	// 1.判断是否已经存在对应的waitChan
	//   1.1 存在则等待
	//   1.2 不存在则创建一个并等待
	ch, ok := mp.waitChanMap[key]
	if !ok {
		ch = make(chan struct{})
		mp.waitChanMap[key] = ch
	}

	// 解锁后开始超时等待
	mp.Unlock()

	// 初始化超时控制
	tCtx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	defer cancel()
	select {
	// 超时返回零值及对应异常
	case <-tCtx.Done():
		err = tCtx.Err()
	// 等待到了key正常返回
	// 需要再次加锁 避免map读和写有并发
	case <-ch:
		mp.Lock()
		value = mp.innerMap[key]
		mp.Unlock()
	}

	// 返回
	return
}

func (mp *TimeoutMap[K, V]) Delete(key K) {
	mp.Lock()
	defer mp.Unlock()
	// todo 支持删除操作

}
