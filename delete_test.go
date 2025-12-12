package btree

import (
	"slices"
	"testing"
)

func leaf(keys ...int) *node[int, int] {
	items := make([]item[int, int], len(keys))
	for i, k := range keys {
		items[i] = item[int, int]{key: k, value: k}
	}
	return &node[int, int]{isLeaf: true, items: items}
}

func internalNode(keys []int, children ...*node[int, int]) *node[int, int] {
	items := make([]item[int, int], len(keys))
	for i, k := range keys {
		items[i] = item[int, int]{key: k, value: k}
	}
	return &node[int, int]{
		isLeaf:   false,
		items:    items,
		children: children,
	}
}

func countKeys(n *node[int, int]) int {
	if n == nil {
		return 0
	}
	total := len(n.items)
	for _, c := range n.children {
		total += countKeys(c)
	}
	return total
}

func treeFromRoot(root *node[int, int], degree int) *BTree[int, int] {
	return &BTree[int, int]{
		root:    root,
		options: OptionsWithDegree(degree, intLess),
		size:    countKeys(root),
	}
}

func buildTree(degree int, keys ...int) *BTree[int, int] {
	t := NewWithOptions[int, int](OptionsWithDegree(degree, intLess))
	for _, k := range keys {
		t.Set(k, k)
	}
	return t
}

func keysInOrder(tree *BTree[int, int]) []int {
	var keys []int
	tree.Ascend(func(k, v int) bool {
		keys = append(keys, k)
		return true
	})
	return keys
}

func nodeKeys(n *node[int, int]) []int {
	out := make([]int, len(n.items))
	for i, it := range n.items {
		out[i] = it.key
	}
	return out
}

func assertVerify(t *testing.T, tree *BTree[int, int]) {
	t.Helper()
	if err := tree.Verify(); err != nil {
		t.Fatalf("Verify() = %v, want nil", err)
	}
}

func assertKeys(t *testing.T, tree *BTree[int, int], want []int) {
	t.Helper()
	got := keysInOrder(tree)
	if !slices.Equal(got, want) {
		t.Fatalf("keys = %v, want %v", got, want)
	}
}

func TestDeleteOnNilAndEmpty(t *testing.T) {
	var nilTree *BTree[int, int]
	if old, ok := nilTree.Delete(1); ok || old != 0 {
		t.Fatalf("Delete on nil tree = (%d,%v), want (0,false)", old, ok)
	}

	empty := NewWithOptions[int, int](DefaultOptions(intLess))
	if old, ok := empty.Delete(10); ok || old != 0 {
		t.Fatalf("Delete on empty tree = (%d,%v), want (0,false)", old, ok)
	}
	if empty.Len() != 0 {
		t.Fatalf("Len() on empty tree = %d, want 0", empty.Len())
	}

	empty.Set(1, 1)
	if old, ok := empty.Delete(99); ok || old != 0 {
		t.Fatalf("Delete missing key returned (%d,%v), want (0,false)", old, ok)
	}
	if empty.Len() != 1 {
		t.Fatalf("Len() changed after deleting missing key, got %d want 1", empty.Len())
	}
}

func TestDeleteSingleKeyShrinksRoot(t *testing.T) {
	tree := buildTree(2, 7)

	old, deleted := tree.Delete(7)
	if !deleted || old != 7 {
		t.Fatalf("Delete(7) = (%d,%v), want (7,true)", old, deleted)
	}
	if tree.Len() != 0 {
		t.Fatalf("Len() after delete = %d, want 0", tree.Len())
	}
	if tree.root != nil {
		t.Fatalf("root should be nil after deleting last key, got %+v", tree.root)
	}
}

func TestDeleteLeafNoBorrow(t *testing.T) {
	tree := buildTree(2, 10, 20, 5, 6)
	assertVerify(t, tree)

	old, deleted := tree.Delete(6)
	if !deleted || old != 6 {
		t.Fatalf("Delete(6) = (%d,%v), want (6,true)", old, deleted)
	}
	if tree.Len() != 3 {
		t.Fatalf("Len() after delete = %d, want 3", tree.Len())
	}
	if _, ok := tree.Get(6); ok {
		t.Fatalf("Get(6) returned ok=true, want false after delete")
	}

	assertVerify(t, tree)
	assertKeys(t, tree, []int{5, 10, 20})
}

func TestDeleteInternalUsesPredecessor(t *testing.T) {
	root := internalNode([]int{30}, leaf(10, 20), leaf(40, 50))
	tree := treeFromRoot(root, 2)

	old, deleted := tree.Delete(30)
	if !deleted || old != 30 {
		t.Fatalf("Delete(30) = (%d,%v), want (30,true)", old, deleted)
	}

	assertVerify(t, tree)
	assertKeys(t, tree, []int{10, 20, 40, 50})

	if tree.root == nil || tree.root.isLeaf {
		t.Fatalf("root should remain internal, got %+v", tree.root)
	}
	if !slices.Equal(nodeKeys(tree.root), []int{20}) {
		t.Fatalf("root keys = %v, want [20]", nodeKeys(tree.root))
	}
}

func TestDeleteInternalUsesSuccessor(t *testing.T) {
	root := internalNode([]int{30}, leaf(10), leaf(40, 50))
	tree := treeFromRoot(root, 2)

	old, deleted := tree.Delete(30)
	if !deleted || old != 30 {
		t.Fatalf("Delete(30) = (%d,%v), want (30,true)", old, deleted)
	}

	assertVerify(t, tree)
	assertKeys(t, tree, []int{10, 40, 50})

	if tree.root == nil || tree.root.isLeaf {
		t.Fatalf("root should remain internal, got %+v", tree.root)
	}
	if !slices.Equal(nodeKeys(tree.root), []int{40}) {
		t.Fatalf("root keys = %v, want [40]", nodeKeys(tree.root))
	}
}

func TestDeleteInternalMergeAndShrink(t *testing.T) {
	root := internalNode([]int{30}, leaf(10), leaf(40))
	tree := treeFromRoot(root, 2)

	old, deleted := tree.Delete(30)
	if !deleted || old != 30 {
		t.Fatalf("Delete(30) = (%d,%v), want (30,true)", old, deleted)
	}
	if tree.Len() != 2 {
		t.Fatalf("Len() after delete = %d, want 2", tree.Len())
	}

	assertVerify(t, tree)
	assertKeys(t, tree, []int{10, 40})

	if tree.root == nil || !tree.root.isLeaf {
		t.Fatalf("root should shrink to merged leaf, got %+v", tree.root)
	}
}

func TestDeleteBorrowFromRightWhileDescending(t *testing.T) {
	root := internalNode([]int{20, 40}, leaf(10), leaf(30), leaf(50, 60))
	tree := treeFromRoot(root, 2)

	old, deleted := tree.Delete(30)
	if !deleted || old != 30 {
		t.Fatalf("Delete(30) = (%d,%v), want (30,true)", old, deleted)
	}

	assertVerify(t, tree)
	assertKeys(t, tree, []int{10, 20, 40, 50, 60})

	if !slices.Equal(nodeKeys(tree.root), []int{20, 50}) {
		t.Fatalf("root keys = %v, want [20 50]", nodeKeys(tree.root))
	}
}

func TestDeleteBorrowFromLeftWhileDescending(t *testing.T) {
	root := internalNode([]int{20, 40}, leaf(10, 15), leaf(30), leaf(50))
	tree := treeFromRoot(root, 2)

	old, deleted := tree.Delete(30)
	if !deleted || old != 30 {
		t.Fatalf("Delete(30) = (%d,%v), want (30,true)", old, deleted)
	}

	assertVerify(t, tree)
	assertKeys(t, tree, []int{10, 15, 20, 40, 50})

	if !slices.Equal(nodeKeys(tree.root), []int{15, 40}) {
		t.Fatalf("root keys = %v, want [15 40]", nodeKeys(tree.root))
	}
}

func TestDeleteBulk(t *testing.T) {
	const degree = 3
	const N = 200

	var keys []int
	for i := 0; i < N; i++ {
		keys = append(keys, i)
	}
	tree := buildTree(degree, keys...)
	assertVerify(t, tree)

	for i := 0; i < N; i += 2 {
		old, deleted := tree.Delete(i)
		if !deleted || old != i {
			t.Fatalf("Delete(%d) = (%d,%v), want (%d,true)", i, old, deleted, i)
		}
	}

	if tree.Len() != N/2 {
		t.Fatalf("Len() after deleting evens = %d, want %d", tree.Len(), N/2)
	}

	var want []int
	for i := 1; i < N; i += 2 {
		want = append(want, i)
	}

	assertVerify(t, tree)
	assertKeys(t, tree, want)
}
