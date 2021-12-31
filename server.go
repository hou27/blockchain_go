package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

const (
	networkProtocol = "tcp"
	dnsNode = "3000"
	nodeVersion = 1
	commandLength = 12
)

var nodeAddr string

// Information about program version and block count.
// Exchanged when first connecting.
type Version struct {
	Version 	int
	BlockHeight int
	From		string
}

// "I have these blocks/transactions: ..."
// Normally sent only when a new block or transaction is being relayed.
// This is only a list, not the actual data.
type Inv struct {
	Type	string
	Items	[][]byte
	From	string
}

type block struct {
	From string
	Block    []byte
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

func sendData(nodeID string, data []byte) {
	conn, err := net.Dial(networkProtocol, nodeID)
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
	payload := GobEncode(Version{nodeVersion, bestHeight, dest})

	request := append(commandToBytes("version"), payload...)
	sendData(dest, request)
}

func sendInv(dest, kind string, items [][]byte) {
	inven := Inv{kind, items, dest}
	payload := GobEncode(inven)
	request := append(commandToBytes("inv"), payload...)

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

func sendGetData(address, kind string, id []byte) {
	payload := GobEncode(getdata{nodeAddr, kind, id})
	request := append(commandToBytes("getdata"), payload...)

	sendData(address, request)
}

func handleInv(request []byte) {
	var payload Inv

	dec := returnDecoder(request)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Recevied %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blockHash := payload.Items[0]
		sendGetData(payload.From, "block", blockHash)
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
			return
		}

		sendBlock(payload.From, &block)
	}
}

func handleVersion(request []byte, bc *Blockchain) {
	payload := Version{}

	dec := returnDecoder(request)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
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
	default:
		fmt.Println("Command unknown.")
	}

	conn.Close()
}

// Starts a node
func StartServer(nodeID string) {
	nodeAddr = fmt.Sprintf(":%s", nodeID)
	// Creates servers
	ln, err := net.Listen(networkProtocol, nodeAddr)
	if err != nil {
		log.Panic(err)
	}

	// Close Listener
	defer ln.Close()

	bc := GetBlockchain(nodeID)

	if nodeID != dnsNode {
		sendVersion(fmt.Sprintf(":%s", dnsNode), bc)
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