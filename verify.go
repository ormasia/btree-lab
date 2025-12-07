package btree

import "fmt"

// Verify 检查 B-Tree 的结构不变式是否成立。
// 返回 nil 表示通过检查，非 nil 表示检测到结构错误。
func (t *BTree[K, V]) Verify() error {
	if t == nil || t.root == nil {
		return nil
	}
	leafDepth := -1
	return t.verifyNode(t.root, true, nil, nil, 0, &leafDepth)
}

// verifyNode 递归检查以 n 为根的子树是否满足 B-Tree 不变式。
// isRoot: 是否是根节点
// minKey/maxKey: 该子树允许的 key 开区间边界 (min, max)，nil 表示无界
// depth: 当前节点深度（根为 0）
// leafDepth: 首次遇到叶子的深度，之后所有叶子都必须与之相同
func (t *BTree[K, V]) verifyNode(n *node[K, V], isRoot bool, minKey *K, maxKey *K, depth int, leafDepth *int) error {
	// 检查节点是否为 nil
	if n == nil {
		return fmt.Errorf("btree: encountered nil node at depth %d", depth)
	}

	degree := t.options.Degree
	itemCount := len(n.items)
	// 1. 检查 key 数量范围
	if isRoot {
		if itemCount > 2*degree-1 {
			return fmt.Errorf("btree: root has %d keys, max allowed %d", itemCount, 2*degree-1)
		}
	} else { // 非根节点
		if itemCount < degree-1 || itemCount > 2*degree-1 {
			return fmt.Errorf("btree: non-root node at depth %d has %d keys, expect in [%d,%d]", depth, itemCount, degree-1, 2*degree-1)
		}
	}

	// 2. 检查 keys 有序 且在 (minKey, maxKey) 范围内
	for i := range itemCount {
		key := n.items[i].key
		if minKey != nil && t.lessThan(key, *minKey) {
			return fmt.Errorf("btree: node at depth %d has key %v <= minKey %v", depth, key, *minKey)
		}
		if maxKey != nil && t.greaterThan(key, *maxKey) {
			return fmt.Errorf("btree: node at depth %d has key %v >= maxKey %v", depth, key, *maxKey)
		}
		// 检查有序性 从第二个 key 开始检查
		if i > 0 && t.greaterThan(n.items[i-1].key, key) {
			return fmt.Errorf("btree: node at depth %d has unordered keys: %v > %v", depth, n.items[i-1].key, key)
		}
	}

	// 3. 叶子节点检查
	if n.isLeaf {
		// 没有孩子节点
		if len(n.children) != 0 {
			return fmt.Errorf("btree: leaf node at depth %d has children", depth)
		}
		// 检查所有叶子节点深度相同
		if *leafDepth == -1 {
			*leafDepth = depth
		} else if *leafDepth != depth {
			return fmt.Errorf("btree: leaf nodes have different depths: %d and %d", *leafDepth, depth)
		}
		return nil
	}

	// 4. 内部节点检查 孩子节点 = key 数量 + 1
	childCount := len(n.children)
	if childCount != itemCount+1 {
		return fmt.Errorf("btree: internal node at depth %d has %d keys but %d children", depth, itemCount, childCount)
	}

	// 5. 内部节点--递归检查子节点
	for i := range childCount { // 注意：i < childCount，因为有 childCount(items + 1) 个孩子节点
		var childMinKey, childMaxKey *K
		switch i {
		case 0: // 第一个孩子节点
			childMinKey = minKey
			if itemCount > 0 {
				childMaxKey = &n.items[0].key
			}
		case itemCount: // 最后一个孩子节点
			childMaxKey = maxKey
			childMinKey = &n.items[itemCount-1].key
		default: // 中间节点
			childMaxKey = &n.items[i].key
			childMinKey = &n.items[i-1].key
		}
		if err := t.verifyNode(n.children[i], false, childMinKey, childMaxKey, depth+1, leafDepth); err != nil {
			return err
		}
	}
	return nil
}
