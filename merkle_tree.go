package main

// MerkleTree is the container for the tree. It holds a pointer to the root of the tree,
// a list of pointers to the leaf nodes, and the merkle root.
type MerkleTree struct {
	Root         *Node
	merkleRoot   []byte
	Leafs        []*Node
}

// Node represents a node, root, or leaf in the tree. It stores pointers to its immediate
// relationships, a hash, the content stored if it is a leaf, and other metadata.
type Node struct {
	Tree		*MerkleTree
	Parent		*Node
	Left		*Node
	Right		*Node
	child		bool
	duplicate	bool
	Hash		[]byte
}