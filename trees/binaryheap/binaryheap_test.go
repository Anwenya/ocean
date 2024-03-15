package binaryheap

import (
	"testing"
)

type Element struct {
	score int
	name  string
}

func TestBinaryHeap(t *testing.T) {
	heap := NewHeap[Element](
		10,
		func(l Element, r Element) int {
			if l.score > r.score {
				return 1
			} else if l.score == r.score {
				return 0
			} else {
				return -1
			}
		},
	)

	// 这时堆为空
	values := heap.Values()
	t.Log(values)

	value, ok := heap.Pop()
	t.Log(value, ok)

	ok = heap.Empty()
	t.Log("堆是否为空:", ok)

	// 入
	t.Log("插入数据")
	heap.Push(Element{score: 1, name: "小明"})
	values = heap.Values()
	t.Log(values)
	heap.Push(Element{score: 5, name: "小红"})
	values = heap.Values()
	t.Log(values)
	heap.Push(Element{score: 2, name: "小强"})
	values = heap.Values()
	t.Log(values)

	t.Log("取数据")
	value, ok = heap.Pop()
	t.Log(value)
	values = heap.Values()
	t.Log(values)
	value, ok = heap.Pop()
	t.Log(value)
	values = heap.Values()
	t.Log(values)
	value, ok = heap.Pop()
	t.Log(value)
	values = heap.Values()
	t.Log(values)

	// 此时堆为空
	value, ok = heap.Pop()
	t.Log(value, ok)
	values = heap.Values()
	t.Log(values)
	ok = heap.Empty()
	t.Log("堆是否为空:", ok)
}
