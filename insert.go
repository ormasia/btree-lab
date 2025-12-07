package btree

// grow: 当根满时，分裂根并增加树高
func (t *BTree[K, V]) grow() {
	oldRoot := t.root
	newRoot := newInternalNodeWithChild(oldRoot)

	// children[0] 是原来的根，如果它是满节点，splitChild 会把它拆成两半，
	// 中间的 key 上浮到 newRoot.items[0]
	t.splitChild(newRoot, 0)
	t.root = newRoot

}

func (t *BTree[K, V]) insertNonFull(n *node[K, V], key K, value V) (old V, replaced bool) {
	// 叶子节点：直接插入/更新
	if n.isLeaf {
		i, found := t.findIndex(n, key)
		if found {
			old = n.items[i].value
			n.items[i].value = value
			return old, true
		}

		// 在切片中间插入
		n.items = append(n.items, item[K, V]{})
		copy(n.items[i+1:], n.items[i:])
		n.items[i] = item[K, V]{key: key, value: value}

		return // old = zero value, replaced = false
	}

	// 内部节点：先找到要下沉的 child
	i, found := t.findIndex(n, key)
	if found {
		// 当前节点已包含 key，直接更新
		old = n.items[i].value
		n.items[i].value = value
		return old, true
	}

	child := n.children[i]

	// 下沉前：如果 child 是满的，先分裂它
	if t.isFull(child) {
		t.splitChild(n, i)

		// splitChild 之后，n.items[i] 是从 child 提升上来的中间 key
		// 判断 key 应该去左孩子还是右孩子
		if t.greaterThan(key, n.items[i].key) {
			i++
		}
	}

	// 此时 n.children[i] 一定是不满节点，可以安全递归
	return t.insertNonFull(n.children[i], key, value)
}

// @param parent: 父节点
// @param index: parent.children 中要被 split 的子节点索引,即插入已经满了的节点
func (t *BTree[K, V]) splitChild(parent *node[K, V], index int) {
	degree := t.options.Degree
	child := parent.children[index]
	mid := degree - 1 // 中间节点索引

	// right 节点存储 child 右半部分的 items 和 children
	right := &node[K, V]{
		isLeaf: child.isLeaf,
	}

	// 保留中间节点
	midItem := child.items[mid]

	// 把 child 右半部分的 items 移动到 right
	right.items = append(right.items, child.items[mid+1:]...)
	child.items = child.items[:mid] // 保留左半部分,不包括中间节点(左开右闭)

	// 如果 child 不是叶节点，还要移动 children
	if !child.isLeaf {
		right.children = append(right.children, child.children[mid+1:]...)
		child.children = child.children[:mid+1]
	}

	// 把 child 的中间节点上浮到 parent
	parent.items = append(parent.items, item[K, V]{})  // 扩容
	copy(parent.items[index+1:], parent.items[index:]) // 后移
	parent.items[index] = midItem                      // 中间节点上浮

	// 把 right 作为 parent 的新子节点插入
	parent.children = append(parent.children, nil) // 扩容
	copy(parent.children[index+2:], parent.children[index+1:])
	parent.children[index+1] = right

}
