# B 树（B-Tree）设计文档

目标 Go 版本：`go1.25`

本文档描述了在本仓库中实现 B 树（B-Tree）时的设计要点：数据结构（Go）、对外接口、主要不变式与性质、操作概述以及实现建议。

## 1. 目标与假设
- 键类型：默认应支持泛型（Go 1.18+）。实现中将使用类型参数 `K` 并配合比较器函数以支持任意可比较类型。
- 最小度（minimum degree）：记为 `t`（`t >= 2`）。实现中将以 `t` 作为树的参数。

## 2. 核心不变式与性质
- 每个节点最多有 `2t-1` 个键；至少（非根节点）有 `t-1` 个键。
- 节点的子指针数量 = 键数量 + 1（内部节点）。
- 所有叶子具有相同高度（树平衡）。
- 键在节点内保持有序。
- 搜索、插入、删除的时间复杂度为 O(t * log_t N)，当 `t` 为常数时可视为 O(log N)。

## 3. Go 数据结构设计（泛型）
建议建立 `btree` 包并使用类型参数 `K`（键类型），通过比较器函数进行键比较。核心结构示例：

```go
package btree

// CompareFunc 比较两个键的顺序：返回负值表示 a<b，0 表示相等，正值表示 a>b。
type CompareFunc[K any] func(a, b K) int

type BTreeNode[K any] struct {
  keys     []K
  children []*BTreeNode[K]
  leaf     bool
}

type BTree[K any] struct {
  root *BTreeNode[K]
  t    int // minimum degree
  cmp  CompareFunc[K]
}
```

说明：
- 推荐使用比较器 `CompareFunc[K]`，因为它能支持任意类型（包括自定义类型）而无需依赖语言内置的 `Ordered` 约束。
- 如果只需支持内置可比较类型（如数字、字符串），也可以使用 `constraints.Ordered` 作为类型约束，但比较器更灵活并且能支持自定义比较逻辑（例如按结构体字段排序）。
- 使用切片而非固定数组能使实现更简洁。`len(keys)` 即节点当前键数；为性能可预分配容量 `2*t-1`。

## 4. 对外接口（API 设计，泛型）
建议的公共函数与方法签名（使用泛型 `K`）：

```go
// NewBTree 创建并返回一个最小度为 t 的 B 树实例，必须传入比较器 cmp。
func NewBTree[K any](t int, cmp CompareFunc[K]) *BTree[K]

// Search 返回是否找到键 k。
func (tr *BTree[K]) Search(k K) bool

// Insert 在 B 树中插入键 k（重复策略由实现决定）。
func (tr *BTree[K]) Insert(k K)

// Delete 从 B 树中删除键 k（可选）。
func (tr *BTree[K]) Delete(k K) error

// Traverse 返回按键升序排列的切片（用于测试与验证）。
func (tr *BTree[K]) Traverse() []K

// Dump 或 String 用于可视化打印树结构（调试）。
func (tr *BTree[K]) Dump() string
```

可选：支持 `Contains(k K) bool`、`Len() int`（键总数）等。

## 5. 主要操作概述

- Search: 在节点内使用二分或线性查找键；若存在则返回；否则在对应的子节点继续查找，直到叶子。

- Insert: 基于 CLRS 的插入算法：
  1. 若根节点满（2t-1 键），则新建节点 s 为新根，令 s.children[0]=oldRoot，然后对 oldRoot 执行 `splitChild(s,0)`，最后在非满节点上执行 `insertNonFull(s,k)`。
  2. `insertNonFull(x,k)`：若 `x` 是叶子，则把 `k` 插入 `x.keys` 的合适位置；否则找到合适的子 `i`，若该子满则先 `splitChild(x,i)`，然后根据 `k` 与中间键比较选择正确的子，递归调用 `insertNonFull`。

- splitChild(x,i): 将 `y = x.children[i]`（满节点）分裂为 `y` 与新节点 `z`。把 `y` 的中间键上移到 `x`，把后半部分键与孩子移动到 `z`。

- Delete（要点）：
  - 若在叶子中直接删除即可（保证不违反最小键数约束）。
  - 若在内部节点删除，需要用前驱或后继替换并在相应子树中递归删除。
  - 在递归进入某子节点之前，要确保该子节点至少有 `t` 个键；否则需要从相邻兄弟借键或与兄弟合并以补齐。
  - 删除实现较复杂，建议作为可选扩展。

## 6. 边界条件与实现注意事项
- 根节点是特殊的：根可以有少于 `t-1` 个键（甚至 0 个，如果树为空）。
- 键的重复策略：决定是否允许重复键。一种策略是允许重复并将其作为多重集合，另一种是忽略重复插入或返回错误。
- 切片插入/删除会分配或移动元素；为了效率，可使用预分配并显式管理长度。
- 使用二分查找可以把单节点内查找从 O(t) 降为 O(log t)。

## 7. 复杂度
- 高度：O(log_t N)。
- 搜索：O(t * log_t N)（节点内支配因子 t）；若 t 为常数，则 O(log N)。
- 插入：同搜索量级（含可能的节点分裂成本，摊还仍为 O(log N)）。

## 8. 测试建议
- 单元测试：
  - 插入若干随机键并检查 `Traverse()` 返回排序后的结果。
  - 边界测试：插入导致根分裂，插入到不同深度、插入大量键。
  - 搜索测试：存在/不存在键的查找。
  - （可选）删除测试：删除叶、删除内部节点、合并/借用场景测试。

## 9. 使用示例（`cmd/main.go` 演示）
```go
package main

import (
    "fmt"
    "btree"
)

func main() {
    tr := btree.NewBTree(2) // t=2
    for _, v := range []int{10,20,5,6,12,30,7,17} {
        tr.Insert(v)
    }
    fmt.Println(tr.Traverse())
    fmt.Println(tr.Dump())
}
```

## 10. 扩展与优化
- 泛型支持：使用 Go 泛型使 B 树支持任意可比较类型。
- 持久化/磁盘格式：将节点序列化到文件，适合数据库或索引实现。
- 并发访问：通过读写锁保护根和节点，或使用更复杂的并发 B 树 结构（如 B-link tree）。

---

如需我把示例 `btree` 包骨架（含 `Search`、`Insert`、`splitChild`、`Traverse`）写入仓库并添加示例 `cmd/main.go`，请回复“实现代码”。
