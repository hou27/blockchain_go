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

// Information about program version and block count.
// Exchanged when first connecting.
type Version struct {
	version 	int
	blockHeight int
	from		string
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
	conn, err := net.Dial(networkProtocol, fmt.Sprintf(":%s", nodeID))
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

func handleVersion(request []byte, bc *Blockchain) {
	buf := new(bytes.Buffer)
	payload := Version{}

	buf.Write(request[commandLength:])
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
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
	default:
		fmt.Println("Command unknown.")
	}

	conn.Close()
}

// Starts a node
func StartServer(nodeID string) {
	// Creates servers
	ln, err := net.Listen(networkProtocol, fmt.Sprintf(":%s", nodeID))
	if err != nil {
		log.Panic(err)
	}

	// Close Listener
	defer ln.Close()

	bc := GetBlockchain(nodeID)

	if nodeID != dnsNode {
		sendVersion(dnsNode, bc)
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
