package main

import (
	"fmt"
	"time"
)

type RelationHandler struct {
}

func NewRelationHandler() *RelationHandler {
	return &RelationHandler{}
}

func (handler *RelationHandler) Handle(e Event) {
	switch e.(type) {
	case *TaskEndEvent:
		event := e.(*TaskEndEvent)
		event.Done()
		fmt.Println("完成Task", e.GetTopic())
	case *ImMsgEvent:
		time.Sleep(time.Second * 1)
		fmt.Println("处理事件:", e.GetTopic())
	default:
		fmt.Println("未订阅的事件")
	}

}

func (handler *RelationHandler) Start(channel chan Event) {
	go func() {
		for e := range channel {
			handler.Handle(e)
		}
	}()
}
