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
		if addBlockCmd.Arg(0) == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.bc.AddBlock(addBlockCmd.Arg(0))
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