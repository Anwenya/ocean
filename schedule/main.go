package main

import (
	"fmt"
	"sync"
	"time"
)

type (
	subscriber chan Event
	topicFunc  func(e Event) bool
)

type Publisher struct {
	rwMutex     sync.RWMutex             // 读写锁
	buffer      int                      // 订阅队列的缓存大小
	timeout     time.Duration            // 发布超时时间
	subscribers map[subscriber]topicFunc // 订阅者信息
}

func NewPublisher(buffer int) *Publisher {
	return &Publisher{
		buffer:      buffer,
		subscribers: make(map[subscriber]topicFunc),
	}
}

func (p *Publisher) Subscribe() chan Event {
	return p.SubscribeTopic(nil)
}

func (p *Publisher) SubscribeTopic(topic topicFunc) chan Event {
	ch := make(chan Event, p.buffer)
	p.rwMutex.Lock()
	p.subscribers[ch] = topic
	p.rwMutex.Unlock()
	return ch
}

func (p *Publisher) Evict(sub chan Event) {
	p.rwMutex.Lock()
	defer p.rwMutex.Unlock()

	delete(p.subscribers, sub)
	close(sub)
}

func (p *Publisher) Publish(e Event) {
	p.rwMutex.RLock()
	defer p.rwMutex.RUnlock()

	var wg sync.WaitGroup
	for sub, topic := range p.subscribers {
		wg.Add(1)
		go p.sendTopic(sub, topic, e, &wg)
	}
	wg.Wait()
}

func (p *Publisher) Close() {
	p.rwMutex.Lock()
	defer p.rwMutex.Unlock()

	for sub := range p.subscribers {
		delete(p.subscribers, sub)
		close(sub)
	}
}

// 发送主题，可以容忍一定的超时
func (p *Publisher) sendTopic(
	sub subscriber,
	topic topicFunc,
	v Event,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	if topic != nil && !topic(v) {
		return
	}
	select {
	case sub <- v:
	default:
		go func() {
			sub <- v
		}()
	}
}

func main() {
	p := NewPublisher(10)
	defer p.Close()

	relation := p.SubscribeTopic(func(v Event) bool {
		switch v.GetTopic() {
		case EventTaskEnd, EventImMsg:
			return true
		}
		return false
	})

	word := p.SubscribeTopic(func(v Event) bool {
		switch v.GetTopic() {
		case EventTaskEnd, EventImMsg:
			return true
		}
		return false
	})

	topic := p.SubscribeTopic(func(v Event) bool {
		switch v.GetTopic() {
		case EventTaskEnd, EventImMsg:
			return true
		}
		return false
	})

	relationHandler := NewRelationHandler()
	relationHandler.Start(relation)

	wordHandler := NewRelationHandler()
	wordHandler.Start(word)

	topicHandler := NewRelationHandler()
	topicHandler.Start(topic)

	for i := 0; i < 10; i++ {
		e := NewImMsgEvent("data")
		p.Publish(e)
	}

	endEvent := NewTaskEndEvent(3)
	p.Publish(endEvent)
	endEvent.WaitForCompleted()
	fmt.Println("完成")
}
