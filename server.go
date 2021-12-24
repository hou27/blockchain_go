package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

// Starts a node
func StartServer(nodeID int) {
	// Creates servers
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", nodeID))
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
