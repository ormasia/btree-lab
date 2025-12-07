package btree

type item[K any, V any] struct {
	key   K
	value V
}

type node[K any, V any] struct {
	isLeaf   bool
	items    []item[K, V]
	children []*node[K, V]
}

func newLeafNode[K any, V any]() *node[K, V] {
	return &node[K, V]{
		isLeaf: true,
		// items:    make([]item[K, V], 0),
		// children: nil,
	}
}

func newInternalNodeWithChild[K any, V any](child *node[K, V]) *node[K, V] {
	return &node[K, V]{
		isLeaf:   false,
		items:    make([]item[K, V], 0),
		children: []*node[K, V]{child},
	}
}
