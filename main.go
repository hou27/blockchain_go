package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
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
	blocks []*block
}

var bc *blockchain
var once sync.Once
var errNotValid = errors.New("Can't add this block")

func (bc *blockchain) validateStructure(newBlock block) error {
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

func (bc *blockchain) generateGenesis() {
	once.Do(func() {
        bc.addBlock("Genesis Block")
    })
}

func (bc *blockchain) getblockchain() *blockchain {
	if bc == nil {
		bc.generateGenesis()
	}
	return bc
}

func (bc *blockchain) getPrevHash() string {
	if len(bc.blocks) > 0 {
		return bc.blocks[len(bc.blocks)-1].hash
	}
	return "First Block"
}

func (bc *blockchain) getIndex() int {
	if len(bc.blocks) > 0 {
		return bc.blocks[len(bc.blocks)-1].index + 1
	}
	return 0
}

func (bc *blockchain) addBlock(data string) {
	index := bc.getIndex()
	newBlock := &block{index, "", bc.getPrevHash(), data, time.Now()}
	hash := sha256.Sum256([]byte(newBlock.data + newBlock.previousHash)) // func sha256.Sum256(data []byte) [32]byte
	newBlock.hash = fmt.Sprintf("%x", hash)

	isValidated := bc.validateStructure(*newBlock)

	if isValidated != nil {
		fmt.Println(isValidated)
	} else {
		bc.blocks = append(bc.blocks, newBlock)
	}
}

func (bc *blockchain) showBlocks() {
	for _, block := range bc.blocks {
		bT := block.timeStamp.Format("Mon Jan _2 15:04:05 2006")
		fmt.Printf("Index: %d\n", block.index)
		fmt.Printf("Data: %s\n", block.data)
		fmt.Printf("Hash: %s\n", block.hash)
		fmt.Printf("Prev Hash: %s\n", block.previousHash)
		fmt.Println("TimeStamp: ", bT)
	}
}

func main() {
	chain := &blockchain{}
	chain.addBlock("Genesis Block")
	chain.addBlock("Second Block")
	chain.addBlock("Third Block")
	chain.showBlocks()
}