package btree

import "testing"

// 构造一个正常的树，Verify 应该通过
func TestVerify_ValidTree(t *testing.T) {
	tree := NewWithOptions[int, int](DefaultOptions(intLess))

	const N = 200
	for i := 0; i < N; i++ {
		tree.Set(i, i)
	}

	if err := tree.Verify(); err != nil {
		t.Fatalf("expected Verify() == nil for valid tree, got %v", err)
	}
}

// 故意破坏有序性，Verify 必须报错
func TestVerify_DetectUnorderedKeys(t *testing.T) {
	tree := NewWithOptions[int, int](DefaultOptions(intLess))

	for i := 0; i < 10; i++ {
		tree.Set(i, i)
	}

	// 强行破坏 root 的有序性（因为测试在同一个 package，可以直接访问内部结构）
	if len(tree.root.items) >= 2 {
		// 把第一个 key 改成一个比第二个还大的值
		tree.root.items[0].key = tree.root.items[1].key + 100
	}

	if err := tree.Verify(); err == nil {
		t.Fatalf("expected Verify() to fail on unordered keys, got nil")
	}
}

// 故意破坏 children 个数与 key 数量关系
func TestVerify_DetectBadChildrenCount(t *testing.T) {
	tree := NewWithOptions[int, int](DefaultOptions(intLess))

	for i := 0; i < 50; i++ {
		tree.Set(i, i)
	}

	// 找一个内部节点（root 通常会是内部节点）
	n := tree.root
	if n.isLeaf {
		t.Skip("root is leaf, not splitting enough keys in this test run")
	}

	// 强行删掉一个 child
	if len(n.children) > 0 {
		n.children = n.children[:len(n.children)-1]
	}

	if err := tree.Verify(); err == nil {
		t.Fatalf("expected Verify() to fail on bad children count, got nil")
	}
}
