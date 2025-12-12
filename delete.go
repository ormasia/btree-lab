package btree

// Delete 删除给定 key。
// 如果 key 存在，删除并返回旧值和 true；
// 如果 key 不存在，返回零值和 false。
func (t *BTree[K, V]) Delete(key K) (old V, deleted bool) {
	if t == nil || t.root == nil {
		var zero V
		return zero, false
	}

	old, deleted = t.deleteFromNode(t.root, key)
	if !deleted {
		var zero V
		return zero, false
	}

	t.size--

	// 根节点缩高逻辑：
	// 1. 如果根是内部节点且没有 key，但有一个 child，那么提升 child 为新的根
	// 2. 如果根是叶子且 key 数为 0，则整棵树为空，root = nil
	if t.root != nil && len(t.root.items) == 0 {
		if t.root.isLeaf {
			t.root = nil
		} else if len(t.root.children) == 1 {
			t.root = t.root.children[0]
		}
	}

	return old, true
}

// deleteFromNode 在以 n 为根的子树中删除 key。
// 返回：old, deleted 表示是否删除成功以及被删除的旧值。
func (t *BTree[K, V]) deleteFromNode(n *node[K, V], key K) (old V, deleted bool) {
	// 在当前结点内查找 key（或其应在位置）
	i, found := t.findIndex(n, key)

	if found {
		// Case 1 & Case 2: key 在当前节点中
		if n.isLeaf {
			return t.deleteFromLeaf(n, i)
		}
		return t.deleteFromInternal(n, i)
	}

	// key 不在当前节点
	if n.isLeaf {
		// 到叶子还没找到，说明整棵子树都没有这个 key
		var zero V
		return zero, false
	}

	// 需要沿着子树继续下沉：
	// key 应当落在 children[i] 对应的区间
	return t.deleteFromChild(n, i, key)
}

// Case 1：key 在叶子结点中，直接删除
func (t *BTree[K, V]) deleteFromLeaf(n *node[K, V], idx int) (old V, deleted bool) {
	old = n.items[idx].value

	// 删除 n.items[idx]：经典切片删除
	copy(n.items[idx:], n.items[idx+1:])
	n.items = n.items[:len(n.items)-1]

	return old, true
}

// Case 2：key 在内部结点中，使用前驱 / 后继 / 合并策略
func (t *BTree[K, V]) deleteFromInternal(n *node[K, V], idx int) (old V, deleted bool) {
	degree := t.options.Degree

	// 记录当前要删除的 key 和旧值
	targetKey := n.items[idx].key
	old = n.items[idx].value

	leftChild := n.children[idx]
	rightChild := n.children[idx+1]

	// Case 2A：左子树至少有 degree 个 key，用前驱替换
	if len(leftChild.items) >= degree {
		// 找左子树中的最大 key（前驱）
		predNode := leftChild
		for !predNode.isLeaf {
			predNode = predNode.children[len(predNode.children)-1]
		}
		predItem := predNode.items[len(predNode.items)-1]

		// 用前驱覆盖当前节点的 items[idx]
		n.items[idx] = predItem

		// 然后在左子树中删除前驱 key
		_, deleted = t.deleteFromNode(leftChild, predItem.key)

		return old, deleted
	}

	// Case 2B：右子树至少有 degree 个 key，用后继替换
	if len(rightChild.items) >= degree {
		// 找右子树中的最小 key（后继）
		succNode := rightChild
		for !succNode.isLeaf {
			succNode = succNode.children[0]
		}
		succItem := succNode.items[0]

		// 用后继覆盖当前节点的 items[idx]
		n.items[idx] = succItem

		// 然后在右子树中删除后继 key
		_, deleted = t.deleteFromNode(rightChild, succItem.key)

		return old, deleted
	}

	// Case 2C：左右子树都只有 degree-1 个 key，需要合并
	t.mergeChildren(n, idx)
	// 合并后：
	// - 原来的 n.items[idx] 已经下沉到 leftChild 里面
	// - parent.items[idx] 被删掉
	// - children[idx] 仍指向 merge 后的那个大节点（就是 leftChild 本身）
	//
	// 现在这棵 merged 子树里一定包含 targetKey，
	// 所以在 merged 子树里递归删除 targetKey 即可。
	return t.deleteFromNode(leftChild, targetKey)
}

// Case 3：key 不在当前节点，需要沿某个子节点继续下沉。
// 在下沉前要保证该子节点至少有 degree 个 key（不然删一下就会 < degree-1）。
func (t *BTree[K, V]) deleteFromChild(parent *node[K, V], childIndex int, key K) (old V, deleted bool) {
	degree := t.options.Degree

	child := parent.children[childIndex]

	// 如果 child 的 key 数已经是最小值 degree-1，则下沉前需要修补
	if len(child.items) == degree-1 {
		// 优先尝试从左兄弟借
		if childIndex > 0 {
			leftSibling := parent.children[childIndex-1]
			if len(leftSibling.items) >= degree {
				t.borrowFromLeft(parent, childIndex)
				child = parent.children[childIndex] // 重新取 child 引用
			} else if childIndex+1 < len(parent.children) {
				// 左兄弟也不够，从右兄弟借或合并
				rightSibling := parent.children[childIndex+1]
				if len(rightSibling.items) >= degree {
					t.borrowFromRight(parent, childIndex)
					child = parent.children[childIndex]
				} else {
					// 左右兄弟都只有 degree-1，合并 child 和右兄弟
					t.mergeChildren(parent, childIndex)
					child = parent.children[childIndex] // merge 后 child 位置不变
				}
			} else {
				// 没有右兄弟，只能和左兄弟合并
				t.mergeChildren(parent, childIndex-1)
				child = parent.children[childIndex-1]
			}
		} else { // childIndex == 0，没有左兄弟，只能考虑右兄弟
			if childIndex+1 < len(parent.children) {
				rightSibling := parent.children[childIndex+1]
				if len(rightSibling.items) >= degree {
					t.borrowFromRight(parent, childIndex)
					child = parent.children[childIndex]
				} else {
					// 右兄弟也不够，只能合并 0 和 1
					t.mergeChildren(parent, 0)
					child = parent.children[0]
				}
			}
		}
	}

	// 至此 child 至少有 degree 个 key，可以安全递归
	return t.deleteFromNode(child, key)
}

// mergeChildren 将 parent 的 children[idx] 和 children[idx+1] 以及中间的 items[idx]
// 合并为一个节点，保存在 children[idx] 中。
func (t *BTree[K, V]) mergeChildren(parent *node[K, V], idx int) {
	left := parent.children[idx]
	right := parent.children[idx+1]

	// 中间的 key 下沉到左子节点
	midItem := parent.items[idx]

	// 合并 items：left.items + midItem + right.items
	left.items = append(left.items, midItem)
	left.items = append(left.items, right.items...)

	// 合并 children（如果不是叶子）
	if !left.isLeaf {
		left.children = append(left.children, right.children...)
	}

	// 从 parent 中移除 items[idx] 和 children[idx+1]
	copy(parent.items[idx:], parent.items[idx+1:])
	parent.items = parent.items[:len(parent.items)-1]

	copy(parent.children[idx+1:], parent.children[idx+2:])
	parent.children = parent.children[:len(parent.children)-1]
}

// borrowFromLeft 从左兄弟借一个 key 给 parent.children[idx]。
func (t *BTree[K, V]) borrowFromLeft(parent *node[K, V], idx int) {
	child := parent.children[idx]
	leftSibling := parent.children[idx-1]

	// 左兄弟最后一个 key 上移到父节点
	// 父节点的 items[idx-1] 下移到 child 的最前面
	// 1）child.items 前面空出一个位置
	child.items = append(child.items, item[K, V]{}) // append 占位
	copy(child.items[1:], child.items[:len(child.items)-1])
	child.items[0] = parent.items[idx-1]

	// 2）父节点更新 items[idx-1]
	parent.items[idx-1] = leftSibling.items[len(leftSibling.items)-1]
	leftSibling.items = leftSibling.items[:len(leftSibling.items)-1]

	// 3）如果有 children，同样移动一个 child 指针
	if !child.isLeaf {
		child.children = append(child.children, (*node[K, V])(nil))
		copy(child.children[1:], child.children[:len(child.children)-1])
		child.children[0] = leftSibling.children[len(leftSibling.children)-1]
		leftSibling.children = leftSibling.children[:len(leftSibling.children)-1]
	}
}

// borrowFromRight 从右兄弟借一个 key 给 parent.children[idx]。
func (t *BTree[K, V]) borrowFromRight(parent *node[K, V], idx int) {
	child := parent.children[idx]
	rightSibling := parent.children[idx+1]

	// 父节点的 items[idx] 下移到 child 的末尾
	// 右兄弟的第一个 key 上移到父节点
	child.items = append(child.items, parent.items[idx])
	parent.items[idx] = rightSibling.items[0]

	// 右兄弟 items 左移
	copy(rightSibling.items[0:], rightSibling.items[1:])
	rightSibling.items = rightSibling.items[:len(rightSibling.items)-1]

	// children 同理
	if !child.isLeaf {
		child.children = append(child.children, rightSibling.children[0])
		copy(rightSibling.children[0:], rightSibling.children[1:])
		rightSibling.children = rightSibling.children[:len(rightSibling.children)-1]
	}
}
