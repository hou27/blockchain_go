package main

func main() {
	chain := GetBlockchain()
	chain.AddBlock("Second Block")
	chain.ShowBlocks()
}