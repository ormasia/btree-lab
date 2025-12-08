package btree

func (t *BTree[K, V]) Ascend(fn func(k K, v V) bool) {
	if t == nil || t.root == nil {
		return
	}
	t.ascend(t.root, fn)
}

func (t *BTree[K, V]) ascend(n *node[K, V], fn func(k K, v V) bool) bool {
	if n == nil {
		return true
	}
	if n.isLeaf {
		for _, it := range n.items {
			if !fn(it.key, it.value) {
				return false
			}
		}
		return true
	}
	for i, it := range n.items { // 从左开始遍历
		if !t.ascend(n.children[i], fn) {
			return false
		}
		if !fn(it.key, it.value) {
			return false
		}
	}
	// 处理最右侧的子节点
	return t.ascend(n.children[len(n.children)-1], fn)
}
