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