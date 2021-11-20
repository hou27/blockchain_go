package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type Cli struct {
	bc *Blockchain
}

func (cli *Cli) Active() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	showBlocksCmd := flag.NewFlagSet("showblocks", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block data")

	switch os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "showblocks":
		err := showBlocksCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		fmt.Println("addblock is being implemented.")
		// cli.bc.AddBlock(*addBlockData)
	}

	if showBlocksCmd.Parsed() {
		cli.bc.ShowBlocks()
	}
}

func (cli *Cli) printUsage() {
	fmt.Printf("How to use:\n\n")
	fmt.Println("  addblock DATA - add a block to the blockchain")
	fmt.Println("  showblocks - print all the blocks of the blockchain")
}