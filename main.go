package main

import (
	"crypto/sha256"
	"fmt"
	"time"
)

type block struct {
	index			int
	hash			string
    previousHash	string
	data			string
    timeStamp		time.Time
}

type blockchain struct {
	blocks []block
}

func (b *blockchain) validateStructure() bool {
	return true
}

func (b *blockchain) getPrevHash() string {
	if len(b.blocks) > 0 {
		return b.blocks[len(b.blocks)-1].hash
	}
	return "First Block"
}

func (b *blockchain) getIndex() int {
	if len(b.blocks) > 0 {
		return b.blocks[len(b.blocks)-1].index + 1
	}
	return 0
}

func (b *blockchain) addBlock(data string) {
	index := b.getIndex()
	newBlock := block{index, "", b.getPrevHash(), data, time.Now()}
	hash := sha256.Sum256([]byte(newBlock.data + newBlock.previousHash)) // func sha256.Sum256(data []byte) [32]byte
	newBlock.hash = fmt.Sprintf("%x", hash)
	b.blocks = append(b.blocks, newBlock)
}

func (b *blockchain) showBlocks() {
	for _, block := range b.blocks {
		bT := block.timeStamp.Format("Mon Jan _2 15:04:05 2006")
		fmt.Printf("Index: %d\n", block.index)
		fmt.Printf("Data: %s\n", block.data)
		fmt.Printf("Hash: %s\n", block.hash)
		fmt.Printf("Prev Hash: %s\n", block.previousHash)
		fmt.Println("TimeStamp: ", bT)
	}
}

func main() {
	chain := blockchain{}
	chain.addBlock("Genesis Block")
	chain.addBlock("Second Block")
	chain.addBlock("Third Block")
	chain.showBlocks()
}