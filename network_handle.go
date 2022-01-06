package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net"
)

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
	mempool[hex.EncodeToString(tx.ID)] = *tx

	if nodeAddr == nodesOnline[0] {
		for _, node := range nodesOnline {
			if node != nodeAddr && node != payload.From {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(mempool) >= 2 && len(mineNode) > 0 {
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
                    sendInv(node, "blocks", [][]byte{newBlock.Hash})
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
