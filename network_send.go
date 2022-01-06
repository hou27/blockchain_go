package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func sendData(dest string, data []byte) {
	conn, err := net.Dial(networkProtocol, dest)
	if err != nil {
		fmt.Printf("%s is not online\n", dest)
		var offlineNode int

		for idx, node := range nodesOnline {
			if node == dest {
				offlineNode = idx
			}
		}

		nodesOnline = append(nodesOnline[:offlineNode], nodesOnline[offlineNode+1:]...)
		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func sendVersion(dest string, bc *Blockchain) {
	bestHeight := bc.getBestHeight()
	payload := GobEncode(Version{nodeVersion, bestHeight, nodeAddr})

	request := append(commandToBytes("version"), payload...)
	sendData(dest, request)
}

func sendInv(dest, kind string, items [][]byte) {
	inven := Inv{kind, items, nodeAddr}
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

	fmt.Println("Send Block ::: ", b.Hash)
	sendData(dest, request)
}

func sendGetBlocks(dest string, myBestHeight int) {
	payload := GobEncode(getblocks{nodeAddr, myBestHeight})
	request := append(commandToBytes("getblocks"), payload...)

	sendData(dest, request)
}

func sendGetData(dest, kind string, id []byte) {
	payload := GobEncode(getdata{nodeAddr, kind, id})
	request := append(commandToBytes("getdata"), payload...)

	sendData(dest, request)
}