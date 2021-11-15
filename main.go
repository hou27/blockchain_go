package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator"
)

type block struct {
	index			int `validate:"min=0,max=100"`
	hash			string `validate:"required"`
	previousHash	string `validate:"required"`
	data			string `validate:"required"`
	timeStamp		time.Time `validate:"required"`
}

type blockchain struct {
	blocks []block
}

var errNotValid = errors.New("Can't add this block")

func (b *blockchain) validateStructure(newBlock block) error {
	fmt.Println(newBlock)
	validate := validator.New()

	err := validate.Struct(newBlock)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			fmt.Println(err)
		}
		return errNotValid
	}
	return nil
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

	isValidated := b.validateStructure(newBlock)

	if isValidated != nil {
		fmt.Println(isValidated)
	} else {
		b.blocks = append(b.blocks, newBlock)
	}
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