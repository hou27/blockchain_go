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
	timeStamp		int64 `validate:"required"`
	hash			string `validate:"required"`
	prevHash		string `validate:"required"`
	data			string `validate:"required"`
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

func (b *block) calculateHash() {
	hash := sha256.Sum256([]byte(b.data + b.prevHash)) // func sha256.Sum256(data []byte) [32]byte
	b.hash = fmt.Sprintf("%x", hash)
}

func NewBlock(data string, prevHash string) *block {
	newblock := &block{time.Now().Unix(), "", prevHash, data}
	newblock.calculateHash()
	return newblock
}

// Add Blockchain
func (bc *blockchain) AddBlock(data string) {
	prevHash := bc.getPrevHash()
	newBlock := NewBlock(data, prevHash)

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
		fmt.Println("TimeStamp: ", block.timeStamp)
		fmt.Printf("Data: %s\n", block.data)
		fmt.Printf("Hash: %s\n", block.hash)
		fmt.Printf("Prev Hash: %s\n", block.prevHash)
	}
}