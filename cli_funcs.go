package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/btcsuite/btcutil/base58"
)

func (cli *Cli) send(from, to string, amount int, nodeID string) {
	bc := GetBlockchain(nodeID)
	defer bc.db.Close()

	UTXOSet := UTXOSet{bc}
	tx := NewUTXOTransaction(from, to, amount, &UTXOSet, nodeID)
	rewardTx := NewCoinbaseTX(from, "Mining reward")
	newBlock := bc.AddBlock([]*Transaction{rewardTx, tx})
	UTXOSet.Update(newBlock)
	fmt.Println("Send Complete!!")
}

func (cli *Cli) createBlockchain(address, nodeID string) {
	if !IsValidWallet(address) {
		fmt.Println("Use correct wallet")
		os.Exit(1)
	}
	newBc := CreateBlockchain(address, nodeID)
	defer newBc.db.Close()

	UTXOSet := UTXOSet{newBc}
	UTXOSet.Build()
	fmt.Println("Successfully done with create blockchain!")
}

// Show Blockchains
func (cli *Cli) showBlocks(nodeID string) {
	bc := GetBlockchain(nodeID)
	defer bc.db.Close()
	bcI := bc.Iterator()
	for {
		block := bcI.getNextBlock()
		pow := NewProofOfWork(block)

		fmt.Println("\nTimeStamp:", block.TimeStamp)
		for index := range block.Transactions {
			fmt.Println("Transactions: ")
			fmt.Printf(" ID: %v\n", block.Transactions[index].ID)
			fmt.Printf(" Vin: %v\n", block.Transactions[index].Vin[0])
			fmt.Printf("    .ScriptSig: %v\n", block.Transactions[index].Vin[0].ScriptSig)
			fmt.Printf(" Vout: %v\n", block.Transactions[index].Vout)
		}
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Prev Hash: %x\n", block.PrevHash)
		fmt.Printf("Nonce: %d\n", block.Nonce)
		fmt.Printf("is Validated: %s\n", strconv.FormatBool(pow.Validate()))

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *Cli) getBalance(address, nodeID string) {
	bc := GetBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()

	balance := 0

	publicKeyHash, _, err := base58.CheckDecode(address)
	if err != nil {
		log.Panic(err)
	}
	utxos := UTXOSet.FindUTXOs(publicKeyHash)

	for _, out := range utxos {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *Cli) createWallet(nodeID string) {
	wallets, _ := NewWallets(nodeID)
	address := wallets.CreateWallet(nodeID)
	wallets.SaveToFile(nodeID)

	fmt.Printf("Your new address: %s\n", address)
}

func (cli *Cli) showAddresses(nodeID string) {
	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *Cli) printUsage() {
	fmt.Printf("How to use:\n\n")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - send AMOUNT of coins from FROM address to TO")
	fmt.Println("  createblockchain -address ADDRESS - create new blockchain")
	fmt.Println("  showblocks - print all the blocks of the blockchain")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createwallet - Create your Wallet")
	fmt.Println("  showaddresses - Show all addresses")
	fmt.Println("  startnode - Start a node with NODE_ID")
}

func (cli *Cli) startNode(nodeID string) {
	fmt.Printf("Node ID ::: %s - started\n", nodeID)
	StartServer(nodeID)
}