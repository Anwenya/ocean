package binaryheap

// Heap 二叉堆
type Heap[T any] struct {
	// 堆元素
	nodes []T
	// 当前元素数
	size int
	// 堆容量
	capacity int
	// 比较堆元素大小的函数
	// 返回 1 left > right
	// 返回 0 left == right
	// 返回 -1 left < right
	comparator func(left T, right T) int
}

func NewHeap[T any](capacity int, comparator func(T, T) int) *Heap[T] {
	return &Heap[T]{
		capacity:   capacity,
		nodes:      make([]T, capacity, capacity),
		comparator: comparator,
	}
}

// Push 入堆
func (heap *Heap[T]) Push(values ...T) {
	if len(values) == 1 {
		// 堆满
		if heap.size == heap.capacity {
			return
		}

		// 单插入 插入到末尾并向上堆化
		heap.nodes[heap.size] = values[0]
		heap.size++
		heap.up()
	} else {
		// 多插入 相当于构建堆
		for _, value := range values {
			// 堆满
			if heap.size == heap.capacity {
				continue
			}

			heap.nodes[heap.size] = value
			heap.size++
		}

		// 从最后一个非叶子节点开始 逐个向下堆化
		size := heap.size>>1 - 1
		for i := size; i >= 0; i-- {
			heap.down(i)
		}

	}
}

// Pop 出堆
func (heap *Heap[T]) Pop() (value T, ok bool) {
	if heap.size <= 0 {
		return
	}
	ok = true
	// 交换第一个和最后一个
	value = heap.nodes[0]
	lastIndex := heap.size - 1
	heap.swap(0, lastIndex)
	heap.size--
	// 从头向下堆化
	heap.down(0)
	return
}

// Values 按pop顺序返回当前堆中的值
func (heap *Heap[T]) Values() []T {
	if heap.size <= 0 {
		return []T{}
	}

	// 任何大小都是0的2倍
	// 不会触发扩容
	var backNodes = make([]T, 0)
	backNodes = append(backNodes, heap.nodes...)
	backSize := heap.size

	values := make([]T, heap.size, heap.size)
	for i := 0; i < backSize; i++ {
		values[i], _ = heap.Pop()
	}

	heap.nodes = backNodes
	heap.size = backSize
	return values
}

// Peek 查看堆顶元素
func (heap *Heap[T]) Peek() (T, bool) {
	return heap.nodes[0], heap.size >= 1
}

// Empty 堆是否为空
func (heap *Heap[T]) Empty() bool {
	return heap.size <= 0
}

// Size 当前堆中元素的个数
func (heap *Heap[T]) Size() int {
	return heap.size
}

// Full 堆是否已满
func (heap *Heap[T]) Full() bool {
	return heap.size == heap.capacity
}

// Clear 清除堆
func (heap *Heap[T]) Clear() {
	heap.size = 0
	heap.nodes = make([]T, heap.capacity, heap.capacity)
}

// 从最后一个非叶子节点向上堆化
func (heap *Heap[T]) up() {
	// 拿到最后一个非空元素的索引
	index := heap.size - 1
	// 子节点是 n 父节点就是 n/2
	for parentIndex := (index - 1) >> 1; index > 0; parentIndex = (index - 1) >> 1 {
		// 取出新插入的元素 和 它的父元素
		indexValue := heap.nodes[index]
		parentValue := heap.nodes[parentIndex]

		// parent <= index 的情况 退出循环
		// 正常情况这是小顶堆 实际看comparator中是如何比较的
		if heap.comparator(parentValue, indexValue) <= 0 {
			break
		}

		// parent > index 时 交换
		heap.swap(index, parentIndex)

		// 继续向上堆化
		index = parentIndex
	}
}

// 从指定节点向下堆化
func (heap *Heap[T]) down(parentIndex int) {
	size := heap.size

	// 父节点是n 左子节点是 2n+1 右子节点是 2n+2
	// 前提是使用下标0
	for leftIndex := parentIndex<<1 + 1; leftIndex < size; leftIndex = parentIndex<<1 + 1 {
		// 右子节点
		rightIndex := parentIndex<<1 + 2

		// 按照小顶堆的规则 需要找出最小的那个
		smallerIndex := leftIndex

		// todo 是否要检查越界 目前开放的操作不会触发越界
		leftValue := heap.nodes[leftIndex]
		rightValue := heap.nodes[rightIndex]

		// 先要判断右子节点是否存在
		// 之后比较大小
		if rightIndex < size && heap.comparator(leftValue, rightValue) > 0 {
			smallerIndex = rightIndex
		}

		// 比较父节点和子节点
		indexValue := heap.nodes[parentIndex]
		smallerValue := heap.nodes[smallerIndex]

		if heap.comparator(indexValue, smallerValue) > 0 {
			heap.swap(parentIndex, smallerIndex)
		} else {
			break
		}

		// 继续向下堆化
		parentIndex = smallerIndex
	}
}

// 交换元素
func (heap *Heap[T]) swap(i1 int, i2 int) {
	heap.nodes[i1], heap.nodes[i2] = heap.nodes[i2], heap.nodes[i1]
}
