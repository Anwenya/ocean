package skiplist

import (
	"math/rand"
)

// Get 根据key查找val
func (s *Skiplist[K, V]) Get(key K) (V, bool) {
	if _node := s.search(key); _node != nil {
		return _node.val, true
	}
	return *new(V), false
}

// Put 根据key插入或更新
func (s *Skiplist[K, V]) Put(key K, val V) {
	// 如果之前存在该节点 直接更新
	if _node := s.search(key); _node != nil {
		_node.val = val
	}

	// 插入新节点
	// 计算新节点的高度
	// 这里level起始值是0 代表第一层
	// 节点在存储时第一层的下标是0
	level := s.roll()

	// 如果新节点的高度超出跳表最大高度 则需要对高度进行补齐
	// 因为我们限制了跳表的最大高度 这段逻辑是不必要的
	for len(s.head.nexts) <= level {
		s.head.nexts = append(s.head.nexts, nil)
	}

	// 创建出新的节点
	// 层级需要 +1 因为起始值是0
	newNode := node[K, V]{
		key:   key,
		val:   val,
		nexts: make([]*node[K, V], level+1),
	}
	// 从头节点出发 构建索引
	move := s.head
	// 遍历新节点需要插入的层
	for ; level >= 0; level-- {
		// 遍历找到新节点插入的位置
		for move.nexts[level] != nil && move.nexts[level].key < key {
			move = move.nexts[level]
		}

		// 调整指针更新 完成新节点在该层的插入

		// 新节点指向第一个[比自己大的节点 或者 nil]
		newNode.nexts[level] = move.nexts[level]
		// 旧节点指向新节点
		move.nexts[level] = &newNode
	}
}

// Delete 根据Key删除节点
func (s *Skiplist[K, V]) Delete(key K) {
	// key不存在 直接返回
	if _node := s.search(key); _node == nil {
		return
	}

	// 从节点的高层出发找到要删除的节点
	// 找到该节点在每一层的位置 并调整索引
	// 因为删除节点需要调整索引 所以需要自己遍历
	move := s.head

	for level := len(s.head.nexts) - 1; level >= 0; level-- {
		for move.nexts[level] != nil && move.nexts[level].key < key {
			move = move.nexts[level]
		}

		// 当前层已经遍历完了 没有找到目标
		if move.nexts[level] == nil || move.nexts[level].key > key {
			continue
		}

		// 到这里意味着 右侧节点就是我们要找的目标 调整对应的指针
		move.nexts[level] = move.nexts[level].nexts[level]
	}

	// 更新跳表的最大高度
	var dif int
	// 从上往下逐层判断是否该层已为空 为空就标记并删除该层
	for level := len(s.head.nexts) - 1; level >= 0 && s.head.nexts[level] != nil; level-- {
		dif++
	}

	// 用高度减去空层的数量
	s.head.nexts = s.head.nexts[:len(s.head.nexts)-dif]
}

// Range 范围查询  [start, end]
func (s *Skiplist[K, V]) Range(start, end K) []ResNode[K, V] {
	// 首先通过 ceiling 方法 找到 第一个 key >= start 的节点 ceilNode
	ceilNode := s.ceiling(start)
	// 如果不存在 直接返回
	if ceilNode == nil {
		return []ResNode[K, V]{}
	}

	// 从 ceilNode 首层出发向右遍历 把所有位于 [start,end] 区间内的节点统统返回
	var res []ResNode[K, V]
	for move := ceilNode; move != nil && move.key <= end; move = move.nexts[0] {
		res = append(res, ResNode[K, V]{key: move.key, val: move.val})
	}
	return res
}

// Ceiling 找到第一个 key >= target 的 key-value 对
func (s *Skiplist[K, V]) Ceiling(target K) (ResNode[K, V], bool) {
	if ceilNode := s.ceiling(target); ceilNode != nil {
		return ResNode[K, V]{ceilNode.key, ceilNode.val}, true
	}
	return ResNode[K, V]{}, false
}

// Floor 最后一个 ley <= target 的 key-value 对
func (s *Skiplist[K, V]) Floor(target K) (ResNode[K, V], bool) {
	// 引用 floor 方法，取 floorNode 值进行返回
	if floorNode := s.floor(target); floorNode != nil {
		return ResNode[K, V]{floorNode.key, floorNode.val}, true
	}

	return ResNode[K, V]{}, false
}

// 随机当前节点的高度
func (s *Skiplist[K, V]) roll() int {
	// 默认是1层
	level := 0
	for i := 1; i < MaxLevel; i++ {
		if rand.Int31()%7 == 1 {
			level++
		}
	}
	return level
}

// 从跳表中检索对应的node
func (s *Skiplist[K, V]) search(key K) *node[K, V] {
	// 先拿到头节点
	move := s.head

	// 从高层往低层遍历 高层是索引
	for level := len(s.head.nexts) - 1; level >= 0; level-- {

		// 从前往后遍历该层

		// 如果该层的下一个节点!=nil 并且 < key 继续向后遍历
		// 从[下一个节点]为启点开始比较是因为头节点不存储数据
		if move.nexts[level] != nil && move.nexts[level].key < key {
			// 向后移动
			move = move.nexts[level]
		}

		// 到这里上面的循环有两种可能
		// 1.该层已经被遍历完了 下一个节点就是 nil 了
		// 2.找到 >=key 的节点了
		// 所以这里可以先判断一下该节点是不是我们要找的
		if move.nexts[level] != nil && move.nexts[level].key == key {
			// 找到了就直接返回
			return move.nexts[level]
		}

		// 到这里说明在当前层没有找到 下沉到下一层继续找
	}

	// 到这里说明所有层都遍历完了 还是没有找到
	return nil
}

// 从跳表中检索第一个 key >= target 的node
func (s *Skiplist[K, V]) ceiling(target K) *node[K, V] {
	move := s.head

	// 从高层往低层遍历 高层是索引
	for level := len(s.head.nexts) - 1; level >= 0; level-- {

		// 如果该层的下一个节点!=nil 并且 < target 继续向后遍历
		for move.nexts[level] != nil && move.nexts[level].key < target {
			// 向后移动
			move = move.nexts[level]
		}

		// 到这里上面的循环有两种可能
		// 1.该层已经被遍历完了 下一个节点就是 nil 了
		// 2.找到 >=key 的节点了
		// 所以这里可以先判断一下该节点是不是我们要找的
		if move.nexts[level] != nil && move.nexts[level].key == target {
			// 找到了就直接返回
			return move.nexts[level]
		}
	}

	// 到这里已经遍历完所有的层了 如果中途找到 key == target 的情况是不会走到这里的
	// 所以 有两种可能
	// 1.下一个节点就是 nil 了
	// 2.下一个节点就是 第一个 key > target 的node
	// 直接返回即可 在调用方再判断是否为nil
	return move.nexts[0]
}

// 从跳表中检索最后一个 key <= target 的node
func (s *Skiplist[K, V]) floor(target K) *node[K, V] {
	move := s.head

	// 从高层往低层遍历 高层是索引
	for level := len(s.head.nexts) - 1; level >= 0; level-- {

		// 如果该层的下一个节点!=nil 并且 < target 继续向后遍历
		for move.nexts[level] != nil && move.nexts[level].key < target {
			// 向后移动
			move = move.nexts[level]
		}

		// 到这里上面的循环有两种可能
		// 1.该层已经被遍历完了 下一个节点就是 nil 了
		// 2.找到 >=key 的节点了
		// 所以这里可以先判断一下该节点是不是我们要找的
		if move.nexts[level] != nil && move.nexts[level].key == target {
			// 找到了就直接返回
			return move.nexts[level]
		}
	}

	// 到这里已经遍历完所有的层了 如果中途找到 key == target 的情况是不会走到这里的
	// 所以 有两种可能
	// 1.下一个节点就是 nil 了
	// 2.下一个节点就是 第一个 key > target 的node
	// 直接返回即可 在调用方再判断是否为nil
	// 找最后一个 key <= target的node 应返回当前节点
	// 因为上面的第二个可能

	// 如果当前move就是头节点
	if move == s.head {
		return nil
	}

	return move
}
