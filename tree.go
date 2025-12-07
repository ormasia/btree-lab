package btree

type BTree[K any, V any] struct {
	root    *node[K, V]
	options Options[K]
	size    int
}

func NewWithOptions[K any, V any](options Options[K]) *BTree[K, V] {
	if options.Degree == 0 {
		options.Degree = defaultDegree
	}
	if options.Less == nil {
		panic("btree: LessFunc must not be nil")
	}
	if options.Degree < minDegree {
		panic("btree: degree must be >= MinDegree")
	}
	return &BTree[K, V]{
		root:    nil,
		options: options,
		size:    0,
	}
}

func (t *BTree[K, V]) Len() int {
	if t == nil {
		return 0
	}
	return t.size
}

func (t *BTree[K, V]) Clear() {
	if t == nil {
		return
	}
	t.root = nil // allow GC to reclaim nodes 属于go GC特性一旦失去外部联系，自动回收
	t.size = 0
}

// some helpers for cmparing keys
func (t *BTree[K, V]) cmp(a, b K) int {
	return t.options.Less(a, b)
}

func (t *BTree[K, V]) lessThan(a, b K) bool {
	return t.options.Less(a, b) < 0
}

func (t *BTree[K, V]) equal(a, b K) bool {
	return t.options.Less(a, b) == 0
}

func (t *BTree[K, V]) greaterThan(a, b K) bool {
	return t.options.Less(a, b) > 0
}

// 满节点：items 数量达到 2*degree - 1
func (t *BTree[K, V]) isFull(n *node[K, V]) bool {
	return len(n.items) >= 2*t.options.Degree-1
}

// Get
func (t *BTree[K, V]) Get(key K) (V, bool) {
	var value V
	if t == nil || t.root == nil {
		return value, false
	}
	return t.get(t.root, key)
}

// Set
func (t *BTree[K, V]) Set(key K, value V) (old V, replaced bool) {
	if t == nil {
		return old, false
	}
	if t.root == nil {
		t.root = newLeafNode[K, V]()
	}
	// 如果根满了，增长树高
	if t.isFull(t.root) {
		t.grow()
	}
	old, replaced = t.insertNonFull(t.root, key, value)
	if !replaced {
		t.size++
	}
	return old, replaced
}
