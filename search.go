package btree

func (t *BTree[K, V]) get(n *node[K, V], key K) (V, bool) {
	var zero V

	i, found := t.findIndex(n, key)

	if found {
		return n.items[i].value, true
	}

	if n.isLeaf {
		return zero, false
	}
	return t.get(n.children[i], key)
}
