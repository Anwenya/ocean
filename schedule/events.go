package main

import "sync"

type Event interface {
	GetTopic() string
	GetData() interface{}
}

const (
	EventImMsg   = "imMsg"
	EventTaskEnd = "taskEnd"
)

type BaseEvent struct {
	eventType string
	data      interface{}
}

func (e *BaseEvent) GetTopic() string {
	return e.eventType
}

func (e *BaseEvent) GetData() interface{} {
	return e.data
}

type ImMsgEvent struct {
	BaseEvent
}

func NewImMsgEvent(data interface{}) Event {
	return &ImMsgEvent{
		BaseEvent{
			eventType: EventImMsg,
			data:      data,
		},
	}
}

type TaskEndEvent struct {
	BaseEvent
	number int
	wg     sync.WaitGroup
}

func NewTaskEndEvent(number int) *TaskEndEvent {
	e := &TaskEndEvent{
		BaseEvent: BaseEvent{
			eventType: EventTaskEnd,
			data:      nil,
		},
		wg: sync.WaitGroup{},
	}
	e.wg.Add(number)
	return e
}

func (e *TaskEndEvent) Done() {
	e.wg.Done()
}

func (e *TaskEndEvent) WaitForCompleted() {
	e.wg.Wait()
}
