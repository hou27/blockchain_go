package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

type Cli struct {
	bc *Blockchain
}

// Run CLI
func (cli *Cli) Active() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	showBlocksCmd := flag.NewFlagSet("showblocks", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)

	sendFrom := sendCmd.String("from", "", "Source address")
	sendTo := sendCmd.String("to", "", "Destination address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")

	switch os.Args[1] {
	case "send":
		err := sendCmd.Parse(os.Args[2:])
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
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if showBlocksCmd.Parsed() {
		cli.showBlocks()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}
}

func (cli *Cli) send(from, to string, amount int) {
	bc := GetBlockchain(from)
	defer bc.db.Close()
	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.AddBlock([]*Transaction{tx})
	fmt.Println("Send Complete!!")
}

func (cli *Cli) getBalance(address string) {
	bc := GetBlockchain(address)
	defer bc.db.Close()

	balance := 0

	utxs := bc.FindUnspentTxs(address)

	for _, tx := range utxs {
		for _, out := range tx.Txout {
			if out.ScriptPubKey == address {
				balance += out.Value
			}
		}
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

// Show Blockchains
func (cli *Cli) showBlocks() {
	bc := GetBlockchain("")
	defer bc.db.Close()
	bcI := bc.Iterator()
	for {
		block := bcI.getNextBlock()
		pow := NewProofOfWork(block)

		fmt.Println("\nTimeStamp:", block.TimeStamp)
		fmt.Println("Data: ", block.Transactions)
        fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Prev Hash: %x\n", block.PrevHash)
		fmt.Printf("Nonce: %d\n", block.Nonce)
		fmt.Printf("is Validated: %s\n", strconv.FormatBool(pow.Validate()))

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *Cli) printUsage() {
	fmt.Printf("How to use:\n\n")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - send AMOUNT of coins from FROM address to TO")
	fmt.Println("  showblocks - print all the blocks of the blockchain")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
}