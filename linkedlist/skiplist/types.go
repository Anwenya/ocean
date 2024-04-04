package skiplist

import "golang.org/x/exp/constraints"

// MaxLevel 跳表的最大高度
const MaxLevel = 16

// key的泛型使用compareable不支持比较

// As per the documentation comparable only means it supports == and !=.
// (Probably it should have been called equatable instead.)
// You are looking for golang.org/x/exp/constraints.Ordered.
// https://go.dev/play/p/zRP649RJU6t

// Skiplist 跳表
type Skiplist[K constraints.Ordered, V any] struct {
	// 跳表的头节点 该节点不存储具体数据
	// 仅用于记录各层的起始指针
	head *node[K, V]
}

func NewSkiplist[K constraints.Ordered, V any]() *Skiplist[K, V] {
	return &Skiplist[K, V]{
		head: &node[K, V]{
			nexts: make([]*node[K, V], MaxLevel),
			key:   *new(K),
			val:   *new(V),
		},
	}
}

// 跳表中的节点 每个节点存储一对 k-v 对
type node[K constraints.Ordered, V any] struct {
	// 记录该节点各层的情况
	// 其元素长度对应该节点有几层
	nexts []*node[K, V]
	// 该节点存储的key
	key K
	// 该节点存储的val
	val V
}

type ResNode[K constraints.Ordered, V any] struct {
	key K
	val V
}
