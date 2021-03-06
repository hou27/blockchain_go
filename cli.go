package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type Cli struct{}

// Run CLI
func (cli *Cli) Active() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Println("NODE_ID is not set yet.")
		os.Exit(1)
	}

	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	showBlocksCmd := flag.NewFlagSet("showblocks", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	showAddrsCmd := flag.NewFlagSet("showaddresses", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	sendFrom := sendCmd.String("from", "", "Source address")
	sendTo := sendCmd.String("to", "", "Destination address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	createBlockchainAddr := createBlockchainCmd.String("address", "", "First Miner's address")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	startMinerNode := startNodeCmd.String("minenode", "", "Turn on the mining mode")
	sendMineOneself := sendCmd.Bool("mine", false, "Mine by itself")

	switch os.Args[1] {
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "showblocks":
		err := showBlocksCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "showaddresses":
		err := showAddrsCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, *sendAmount, nodeID, *sendMineOneself)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddr == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddr, nodeID)
	}

	if showBlocksCmd.Parsed() {
		cli.showBlocks(nodeID)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress, nodeID)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}

	if showAddrsCmd.Parsed() {
		cli.showAddresses(nodeID)
	}

	if startNodeCmd.Parsed() {
		if nodeID == "" {
			startNodeCmd.Usage()
			os.Exit(1)
		}
		cli.startNode(nodeID, *startMinerNode)
	}
}
