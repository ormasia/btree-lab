package btree

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
