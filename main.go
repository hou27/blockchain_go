package main

import "strconv"

func main() {
	chain := GetBlockchain()
	for i := 1; i < 10; i++ {
		chain.AddBlock(strconv.Itoa(i))
	}
	chain.ShowBlocks()
}