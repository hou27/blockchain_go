package main

import "crypto/sha256"

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
	Hash		[]byte
}

// Creates a new Merkle tree
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []*Node

	// Make it even
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	// Hash transactions
	for _, tx := range data {
		node := NewMerkleNode(nil, nil, tx)
		nodes = append(nodes, node)
	}

	for i := 0; i < len(data)/2; i++ {
		var newLevel []*Node

		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(nodes[j], nodes[j+1], nil)
			newLevel = append(newLevel, node)
		}

		nodes = newLevel
	}

	mTree := MerkleTree{nodes[0], nodes[0].Hash, nodes}
	return &mTree
}

// Creates a new Merkle tree node
func NewMerkleNode(left, right *Node, data []byte) *Node {
	mNode := Node{}

	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		mNode.Hash = hash[:]
	} else {
		hashes := append(left.Hash, right.Hash...)
		hash := sha256.Sum256(hashes)
		mNode.Hash = hash[:]
		left.Parent = &mNode
		right.Parent = &mNode
	}

	mNode.Left = left
	mNode.Right = right

	return &mNode
}