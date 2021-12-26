package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

const networkProtocol = "tcp"

// Starts a node
func StartServer(nodeID string) {
	// Creates servers
	ln, err := net.Listen(networkProtocol, fmt.Sprintf(":%d", nodeID))
	if err != nil {
		log.Panic(err)
	}

	// Close Listener
	defer ln.Close()

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
