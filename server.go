package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
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

	if nodeAddr != nodesOnline[0] {
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