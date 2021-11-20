package main

import (
	"fmt"
	"os"
)

func main() {
	chain := GetBlockchain()
	defer chain.db.Close()

	if len(os.Args) < 2 {
		fmt.Println("Not Enough Args")
		os.Exit(1)
	}

	if os.Args[1] == "showblocks" {
		chain.ShowBlocks()
	}
}