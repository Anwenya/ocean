package main

import (
	"context"
	"fmt"
)

type Task interface {
	Start(ctx context.Context)
}

type QueryTask struct {
	// 任务类型 仅用于区分是属于哪个根任务的
	TaskType string
}

func NewQueryTask(taskType string) Task {
	return &QueryTask{TaskType: taskType}
}

func (task *QueryTask) Start(ctx context.Context) {
	fmt.Println("执行查询任务")
}

type WorldCloudTask struct {
	// 任务类型 仅用于区分是属于哪个根任务的
	TaskType string
}

func NewWorldCloudTask(taskType string) Task {
	return &WorldCloudTask{TaskType: taskType}
}

func (task *WorldCloudTask) Start(ctx context.Context) {
	fmt.Println("执行词云任务")
}

type RelationTask struct {
	// 任务类型 仅用于区分是属于哪个根任务的
	TaskType string
}

func NewRelationTask(taskType string) Task {
	return &RelationTask{TaskType: taskType}
}

func (task *RelationTask) Start(ctx context.Context) {
	fmt.Println("执行关系任务")
}
