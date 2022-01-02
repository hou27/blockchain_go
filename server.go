package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

const (
	networkProtocol = "tcp"
	nodeVersion = 1
	commandLength = 12
)

var (
	nodesOnline = []string{":3000"}
	nodeAddr string
	mineNode string
	mempool = make(map[string]Transaction)
)

// Information about program version and block count.
// Exchanged when first connecting.
type Version struct {
	Version 	int
	BlockHeight int
	From		string
	To			string
}

// "I have these blocks/transactions: ..."
// Normally sent only when a new block or transaction is being relayed.
// This is only a list, not the actual data.
type Inv struct {
	Type	string
	Items	[][]byte
	From	string
	To		string
}

type block struct {
	From	string
	Block	[]byte
}

// Send a transaction. This is sent only in response to a getdata request.
type tx struct {
	From	string
	Tx		[]byte
}

// Request an inv of all blocks in a range.
// It isn't bringing all the blocks, but requesting a hash list of blocks.
type getblocks struct {
	From	string
}

// Request a single block or transaction by hash.
type getdata struct {
	From	string
	Type	string
	ID		[]byte
}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for idx, c := range command {
		bytes[idx] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return string(command[:])
}

func sendData(dest string, data []byte) {
	conn, err := net.Dial(networkProtocol, dest)
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	fmt.Printf("%x\n", data)
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func sendVersion(dest string, bc *Blockchain) {
	bestHeight := bc.getBestHeight()
	payload := GobEncode(Version{nodeVersion, bestHeight, nodeAddr, dest})

	request := append(commandToBytes("version"), payload...)
	sendData(dest, request)
}

func sendInv(dest, kind string, items [][]byte) {
	inven := Inv{kind, items, nodeAddr, dest}
	payload := GobEncode(inven)
	request := append(commandToBytes("inv"), payload...)

	sendData(dest, request)
}

func sendTx(dest string, transaction *Transaction) {
	data := tx{nodeAddr, transaction.Serialize()}
	payload := GobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	sendData(dest, request)
}

func sendBlock(dest string, b *Block) {
	data := block{nodeAddr, b.Serialize()}
	payload := GobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(dest, request)
}

func sendGetBlocks(dest string) {
	payload := GobEncode(getblocks{nodeAddr})
	request := append(commandToBytes("getblocks"), payload...)

	sendData(dest, request)
}

func sendGetData(dest, kind string, id []byte) {
	payload := GobEncode(getdata{nodeAddr, kind, id})
	request := append(commandToBytes("getdata"), payload...)

	sendData(dest, request)
}

func handleInv(request []byte) {
	var payload Inv

	dec := returnDecoder(request)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Recevied %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "blocks" {
		for _, blockHash := range payload.Items {
			sendGetData(payload.From, "block", blockHash)
		}
	} else if payload.Type == "tx" {
		txID := payload.Items[0]

		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.From, "tx", txID)
		}
	}
}

func handleTx(request []byte, bc *Blockchain) {
	var payload tx

	dec := returnDecoder(request)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Tx
	tx := DeserializeTx(txData)
	fmt.Println("after ::: ", tx)
	if !bc.VerifyTransaction(tx) {
		fmt.Println("why not")
	} else {
		fmt.Println("Verified.")
	}
	mempool[hex.EncodeToString(tx.ID)] = *tx

	fmt.Println(nodesOnline, len(mineNode), len(mempool))
	if nodeAddr == nodesOnline[0] {
		for _, node := range nodesOnline {
			if node != nodeAddr && node != payload.From {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(mempool) >= 1 && len(mineNode) > 0 {
		MineTxs:
			var txs []*Transaction

			for id := range mempool {
                tx := mempool[id]
                if bc.VerifyTransaction(&tx) {
                    txs = append(txs, &tx)
                }
            }

            if len(txs) == 0 {
                fmt.Println("There's no valid Transaction...")
                return
            }

			rewardTx := NewCoinbaseTX(mineNode, "Mining reward")
			txs = append(txs, rewardTx)
			newBlock := bc.MineBlock(txs)
			UTXOSet := UTXOSet{bc}
			UTXOSet.Update(newBlock)

			for _, tx := range txs {
                txID := hex.EncodeToString(tx.ID)
                delete(mempool, txID)
            }

            for _, node := range nodesOnline {
                if node != nodeAddr {
                    sendInv(node, "block", [][]byte{newBlock.Hash})
                }
            }

			if len(mempool) > 0 {
                goto MineTxs
            }
		}
	}
}

func handleBlock(request []byte, bc *Blockchain) {
	var payload block

	dec := returnDecoder(request)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := DeserializeBlock(blockData)

	fmt.Println("Recevied a new block")
	fmt.Println(block)

	// newBlock := bc.MineBlock(block.Transactions)
	bc.AddBlock(block)
	UTXOSet := UTXOSet{bc}
	UTXOSet.Update(block)
}

func handleGetBlocks(request []byte, bc *Blockchain) {
	var payload getblocks

	dec := returnDecoder(request)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := bc.GetBlockHashes()
	sendInv(payload.From, "blocks", blocks)
}

func handleGetData(request []byte, bc *Blockchain) {
	var payload getdata

	dec := returnDecoder(request)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			fmt.Println("Err on getblock ::: ", err)
			return
		}

		sendBlock(payload.From, &block)
	} else if payload.Type == "tx" {
		tx := mempool[hex.EncodeToString(payload.ID)]

		sendTx(payload.From, &tx)
	}
}

func handleVersion(request []byte, bc *Blockchain) {
	payload := Version{}

	dec := returnDecoder(request)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if nodeAddr == nodesOnline[0] {
		chkFlag := false
		for _, node := range nodesOnline {
			if node == payload.From {
				chkFlag = !chkFlag
			}
		}
		if !chkFlag {
			nodesOnline = append(nodesOnline, payload.From)
		}		
	}

	myBestHeight := bc.getBestHeight()
	foreignerBestHeight := payload.BlockHeight

	if myBestHeight > foreignerBestHeight {
		sendVersion(payload.From, bc)
	} else if myBestHeight < foreignerBestHeight {
		sendGetBlocks(payload.From)
	}
}

func handleConnection(conn net.Conn, bc *Blockchain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}

	command := bytesToCommand(request[:commandLength])
	fmt.Printf("Received ::: %s\n", command)

	switch command {
	case "version":
		handleVersion(request, bc)
	case "inv":
		handleInv(request)
	case "getblocks":
		handleGetBlocks(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "block":
		handleBlock(request, bc)
	case "tx":
		handleTx(request, bc)
	default:
		fmt.Println("Command unknown.")
	}

	conn.Close()
}

// Starts a node
func StartServer(nodeID, minenode string) {
	nodeAddr = fmt.Sprintf(":%s", nodeID)

	if len(minenode) > 0 {
		mineNode = minenode
		fmt.Println("Now mining is on. Address ::: ", mineNode)
	}

	// Creates servers
	ln, err := net.Listen(networkProtocol, nodeAddr)
	if err != nil {
		log.Panic(err)
	}

	// Close Listener
	defer ln.Close()

	bc := GetBlockchain(nodeID)

	if nodeID != nodesOnline[0] {
		sendVersion(nodesOnline[0], bc)
	}

	for {
		// Wait for connection
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc)
	}
}

func returnDecoder(request []byte) *gob.Decoder {
	var buf bytes.Buffer

	buf.Write(request[commandLength:])
	dec := gob.NewDecoder(&buf)
	
	return dec
}