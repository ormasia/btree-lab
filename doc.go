// Package btree implements a B-tree data structure for efficient storage and retrieval
// of key-value pairs. B-trees are self-balancing tree data structures that maintain
// sorted data and allow for efficient insertion, deletion, and search operations.
//
// This implementation supports configurable degree (minimum degree) for the B-tree,
// allowing users to tune the performance characteristics based on their specific use case.
// The degree affects the number of keys and children each node can hold.
//
// Key features:
// - Insertion, deletion, and search operations in O(log n) time
// - Configurable degree for performance tuning
// - Memory-efficient node structure
// - Support for range queries and iteration
package btree

// B-Tree 不变式：
// 每个非根节点的 key 数量在 [degree-1, 2*degree-1] 之间；
//
// 根节点的 key 数量不超过 2*degree-1（可以为 0 或 >=1）；
//
// 对于内部节点：len(children) == len(items)+1；
//
// 对于叶子：len(children) == 0；
//
// 每个节点内的 key 严格升序；
//
// 每个子树的 key 都落在由父节点 key 决定的 (min, max) 区间内；
//
// 所有叶子深度相同（B-Tree 的高度一致性）。

// 数据结构

// items：

// 当前节点本身持有的一组有序 key（以及它们的 value）；

// 数量是 m，节点内部保证严格递增；

// 对非根节点，m 必须在 [degree-1, 2*degree-1] 内。

// children：

// 是一个长度为 m+1 的数组（对内部节点），每个元素是一个子节点指针；

// 第 i 个 child（children[i]）代表这个节点的 key 被分割成的第 i 段区间；

// 每个 child 维护自己的子树，子树里所有 key 必须落在该 child 对应的“区间”里。
