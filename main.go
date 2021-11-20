package main

import "strconv"

func main() {
	chain := GetBlockchain()
	defer chain.db.Close()

	for i := 1; i < 10; i++ {
		chain.AddBlock("Test Data " + strconv.Itoa(i))
	}
	chain.ShowBlocks()
}