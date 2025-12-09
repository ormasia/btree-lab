package btree

// key 就在当前节点的 items 中，返回索引及 true
// key 不在当前节点的 items 中，返回应该插入的位置索引及 false
func (t *BTree[K, V]) findIndex(n *node[K, V], key K) (int, bool) {
	i := 0
	for i < len(n.items) && t.lessThan(n.items[i].key, key) {
		i++
	}
	if i < len(n.items) && t.equal(n.items[i].key, key) {
		return i, true
	}
	return i, false
}
