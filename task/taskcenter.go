package main

import (
	"fmt"
	"sync/atomic"
)

var (
	queryTaskChannel      chan *Task
	worldCloudTaskChannel chan *Task
	relationTaskChannel   chan *Task
)

type TaskCenter struct {
	TaskStatus map[string]atomic.Uint64
}

// SubmitTask 提交根任务
func (tc *TaskCenter) SubmitTask(taskList []Task) {
	for _, task := range taskList {
		switch v := task.(type) {
		case *QueryTask:
			tc.doSubmit(queryTaskChannel, &task)
		case *WorldCloudTask:
			tc.doSubmit(worldCloudTaskChannel, &task)
		case *RelationTask:
			tc.doSubmit(relationTaskChannel, &task)
		default:
			fmt.Println("未知任务类型:", v)
		}
	}
}

// SubmitSubTask 提交子任务
func (tc *TaskCenter) doSubmit(channel chan *Task, task *Task) {
	select {
	case channel <- task:
	default:
		// 异步
		go func() {
			channel <- task
		}()
	}
}

// GetTask 从多个任务channel中随机获得一个任务
func (tc *TaskCenter) GetTask() *Task {
	var task *Task
	select {
	case task = <-queryTaskChannel:
	case task = <-worldCloudTaskChannel:
	case task = <-relationTaskChannel:
	}
	return task
}
