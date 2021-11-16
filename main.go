package main

import (
	"github.com/hou27/blockchain_go/blockchain"
)

func main() {
	chain := blockchain.GetBlockchain()
	// chain.AddBlock("Genesis Block?")
	chain.AddBlock("Second Block")
	// chain.AddBlock("Third Block")
	chain.ShowBlocks()
}