package main

import (
	"github.com/hou27/blockchain_go/blockchain"
)

func main() {
	chain := blockchain.GetBlockchain()
	chain.AddBlock("Second Block")
	chain.ShowBlocks()
}