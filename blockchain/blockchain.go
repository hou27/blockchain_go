package blockchain

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

func generateGenesis() {
	once.Do(func() {
		bc = &blockchain{}
		bc.AddBlock("Genesis Block")
	})
}

// Get All Blockchains
func GetBlockchain() *blockchain {
	if bc == nil {
		generateGenesis()
	}
	return bc
}

func (bc blockchain) getPrevHash() string {
	if len(GetBlockchain().blocks) > 0 {
		return GetBlockchain().blocks[len(GetBlockchain().blocks)-1].hash
	}
	return "First Block"
}

func (bc blockchain) getIndex() int {
	if len(GetBlockchain().blocks) > 0 {
		return GetBlockchain().blocks[len(GetBlockchain().blocks)-1].index + 1
	}
	return 0
}

func (bc *blockchain) calculateHash(newBlock *block) {
	hash := sha256.Sum256([]byte(newBlock.data + newBlock.previousHash)) // func sha256.Sum256(data []byte) [32]byte
	newBlock.hash = fmt.Sprintf("%x", hash)
}

// Add Blockchain
func (bc *blockchain) AddBlock(data string) {
	index := bc.getIndex()
	newBlock := &block{index, "", bc.getPrevHash(), data, time.Now()}
	bc.calculateHash(newBlock)

	isValidated := bc.validateStructure(*newBlock)

	if isValidated != nil {
		fmt.Println(isValidated)
	} else {
		bc.blocks = append(GetBlockchain().blocks, newBlock)
	}
}

// Show Blockchains
func (bc blockchain) ShowBlocks() {
	for _, block := range GetBlockchain().blocks {
		bT := block.timeStamp.Format("Mon Jan _2 15:04:05 2006")
		fmt.Printf("Index: %d\n", block.index)
		fmt.Printf("Data: %s\n", block.data)
		fmt.Printf("Hash: %s\n", block.hash)
		fmt.Printf("Prev Hash: %s\n", block.previousHash)
		fmt.Println("TimeStamp: ", bT)
	}
}