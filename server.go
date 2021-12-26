package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

const networkProtocol = "tcp"
const dnsNode = "3000"
const nodeVersion = 1

type data struct {
	version 	int
	blockHeight int
	from		string
}

func sendData(nodeID string, bc *Blockchain, data []byte) {
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

// Starts a node
func StartServer(nodeID string) {
	// Creates servers
	ln, err := net.Listen(networkProtocol, fmt.Sprintf(":%s", nodeID))
	if err != nil {
		log.Panic(err)
	}

	// Close Listener
	defer ln.Close()

	bc := GetBlockchain()

	if nodeID != dnsNode {
		sendData(dnsNode, bc, []byte{})
	}

	for {
		// Wait for connection
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go func(c net.Conn) {
			// Echo data back to the client
			io.Copy(c, c)
			c.Close()
		}(conn)
	}
}
