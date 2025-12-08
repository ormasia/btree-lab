package btree

import "testing"

func TestAscend_OrderAndCount(t *testing.T) {
	tree := NewWithOptions[int, int](DefaultOptions(intLess))

	const N = 200
	// 倒序插入，确保不是“天然有序数组”的假象
	for i := N - 1; i >= 0; i-- {
		tree.Set(i, i)
	}

	var (
		lastKey int
		hasLast bool // 这是闭包变量，不需要显式捕获
		count   int
	)

	tree.Ascend(func(k, v int) bool {
		// 检查严格递增
		if hasLast && k <= lastKey {
			t.Fatalf("keys not strictly increasing: prev=%d, cur=%d", lastKey, k)
		}
		hasLast = true
		lastKey = k

		if v != k {
			t.Fatalf("value mismatch: key=%d, value=%d", k, v)
		}

		count++
		return true
	})

	if count != N {
		t.Fatalf("Ascend visited %d items, want %d", count, N)
	}
}
