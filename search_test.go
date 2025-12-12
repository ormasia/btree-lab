package btree

import "testing"

func intLess(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func TestGetManualTree(t *testing.T) {
	// 我们手工创建一个节点用于测试

	root := &node[int, string]{
		isLeaf: true,
		items: []item[int, string]{
			{key: 10, value: "ten"},
			{key: 20, value: "twenty"},
			{key: 30, value: "thirty"},
		},
	}

	tree := &BTree[int, string]{
		root: root,
		options: Options[int]{
			Degree: 4,
			Less:   intLess,
		},
		size: 3,
	}

	tests := []struct {
		key      int
		wantVal  string
		wantFind bool
	}{
		{10, "ten", true},
		{20, "twenty", true},
		{25, "", false},
		{30, "thirty", true},
		{5, "", false},
		{40, "", false},
	}

	for _, tc := range tests {
		gotVal, ok := tree.Get(tc.key)
		if ok != tc.wantFind || (ok && gotVal != tc.wantVal) {
			t.Fatalf("Get(%d) = (%v,%v), want (%v,%v)",
				tc.key, gotVal, ok, tc.wantVal, tc.wantFind)
		}
	}
}

func TestFindIndex(t *testing.T) {
	tree := NewWithOptions[int, string](DefaultOptions(intLess))

	n := &node[int, string]{
		isLeaf: true,
		items: []item[int, string]{
			{10, "a"},
			{20, "b"},
			{30, "c"},
		},
	}

	tests := []struct {
		key   int
		i     int
		found bool
	}{
		{5, 0, false},
		{10, 0, true},
		{15, 1, false},
		{20, 1, true},
		{40, 3, false},
	}

	for _, tc := range tests {
		i, ok := tree.findIndex(n, tc.key)
		if i != tc.i || ok != tc.found {
			t.Fatalf("findIndex(%d) = (%d,%v), want (%d,%v)", tc.key, i, ok, tc.i, tc.found)
		}
	}
}

func TestSetInsertAndUpdate(t *testing.T) {
	tree := NewWithOptions[int, string](OptionsWithDegree(2, intLess))

	inserts := []struct {
		key int
		val string
	}{
		{10, "ten"},
		{20, "twenty"},
		{5, "five"},
		{6, "six"}, // 触发根分裂
	}

	for i, kv := range inserts {
		old, replaced := tree.Set(kv.key, kv.val)
		if replaced || old != "" {
			t.Fatalf("Set(%d,%q) insert got (old=%q,replaced=%v), want (old=\"\",replaced=false)", kv.key, kv.val, old, replaced)
		}
		if tree.Len() != i+1 {
			t.Fatalf("Len after inserting %d = %d, want %d", kv.key, tree.Len(), i+1)
		}
		got, ok := tree.Get(kv.key)
		if !ok || got != kv.val {
			t.Fatalf("Get(%d) after insert = (%q,%v), want (%q,true)", kv.key, got, ok, kv.val)
		}
	}

	old, replaced := tree.Set(20, "TWENTY")
	if !replaced || old != "twenty" {
		t.Fatalf("Set update returned (old=%q,replaced=%v), want (old=%q,replaced=true)", old, replaced, "twenty")
	}
	if tree.Len() != len(inserts) {
		t.Fatalf("Len changed on update, got %d want %d", tree.Len(), len(inserts))
	}
	got, ok := tree.Get(20)
	if !ok || got != "TWENTY" {
		t.Fatalf("Get(20) after update = (%q,%v), want (%q,true)", got, ok, "TWENTY")
	}
}

// TestSetGetBasic 基础插入和查询
func TestSetGetBasic(t *testing.T) {
	tree := NewWithOptions[int, string](DefaultOptions(intLess))

	kvs := map[int]string{
		1:  "one",
		2:  "two",
		10: "ten",
		5:  "five",
	}

	for k, v := range kvs {
		old, replaced := tree.Set(k, v)
		if replaced {
			t.Fatalf("Set on new key %d returned replaced=true, old=%v", k, old)
		}
	}

	if tree.Len() != len(kvs) {
		t.Fatalf("Len() = %d, want %d", tree.Len(), len(kvs))
	}

	for k, want := range kvs {
		got, ok := tree.Get(k)
		if !ok {
			t.Fatalf("Get(%d) = (_, false), want (_, true)", k)
		}
		if got != want {
			t.Fatalf("Get(%d) = (%q, true), want (%q, true)", k, got, want)
		}
	}

	// 不存在的 key
	if _, ok := tree.Get(999); ok {
		t.Fatalf("Get(999) = (_, true), want (_, false)")
	}
}

// TestSetOverwrite 测试覆盖已有 key 的行为
func TestSetOverwrite(t *testing.T) {
	tree := NewWithOptions[int, string](DefaultOptions(intLess))

	old, replaced := tree.Set(1, "one")
	if replaced {
		t.Fatalf("first Set should not replace, got replaced=true, old=%v", old)
	}

	old, replaced = tree.Set(1, "uno")
	if !replaced {
		t.Fatalf("second Set should replace, got replaced=false")
	}
	if old != "one" {
		t.Fatalf("old value = %q, want %q", old, "one")
	}

	if tree.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", tree.Len())
	}

	v, ok := tree.Get(1)
	if !ok || v != "uno" {
		t.Fatalf("Get(1) = (%q, %v), want (%q, true)", v, ok, "uno")
	}
}

// TestSetManyWithSplit 用小 degree 触发多次分裂，验证结构没乱
func TestSetManyWithSplit(t *testing.T) {
	// 用一个小一点的 Degree（比如 3），更容易触发 split
	opts := OptionsWithDegree(3, intLess)
	tree := NewWithOptions[int, int](opts)

	const N = 200

	// 插入 0..N-1
	for i := 0; i < N; i++ {
		old, replaced := tree.Set(i, i)
		if replaced {
			t.Fatalf("Set new key %d returned replaced=true, old=%v", i, old)
		}
	}

	if tree.Len() != N {
		t.Fatalf("Len() = %d, want %d", tree.Len(), N)
	}

	// 再查一遍，确保都在
	for i := 0; i < N; i++ {
		v, ok := tree.Get(i)
		if !ok {
			t.Fatalf("Get(%d) = (_, false), want (_, true)", i)
		}
		if v != i {
			t.Fatalf("Get(%d) = (%d, true), want (%d, true)", i, v, i)
		}
	}

	// 再覆盖一遍，确保 replaced=true，Len 不变
	for i := 0; i < N; i++ {
		old, replaced := tree.Set(i, i*10)
		if !replaced {
			t.Fatalf("overwrite Set(%d) returned replaced=false", i)
		}
		if old != i {
			t.Fatalf("old value for key %d = %d, want %d", i, old, i)
		}
	}

	if tree.Len() != N {
		t.Fatalf("Len() after overwrite = %d, want %d", tree.Len(), N)
	}
}
