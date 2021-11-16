package blockchain

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-playground/validator"
)

type Block struct {
	TimeStamp		int64 `validate:"required"`
	Hash			string `validate:"required"`
	PrevHash		string `validate:"required"`
	Data			string `validate:"required"`
}

type Blockchain struct {
	blocks []*Block
}

var bc *Blockchain
var once sync.Once
var errNotValid = errors.New("Can't add this block")

func (bc *Blockchain) validateStructure(newBlock Block) error {
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
		bc = &Blockchain{}
		bc.AddBlock("Genesis Block")
	})
}

// Get All Blockchains
func GetBlockchain() *Blockchain {
	if bc == nil {
		generateGenesis()
	}
	return bc
}

func (bc Blockchain) getPrevHash() string {
	if len(GetBlockchain().blocks) > 0 {
		return GetBlockchain().blocks[len(GetBlockchain().blocks)-1].Hash
	}
	return "First Block"
}

func (b *Block) calculateHash() {
	hash := sha256.Sum256([]byte(b.Data + b.PrevHash)) // func sha256.Sum256(data []byte) [32]byte
	b.Hash = fmt.Sprintf("%x", hash)
}

func NewBlock(data string, prevHash string) *Block {
	newblock := &Block{time.Now().Unix(), "", prevHash, data}
	newblock.calculateHash()
	return newblock
}

// Add Blockchain
func (bc *Blockchain) AddBlock(data string) {
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
func (bc Blockchain) ShowBlocks() {
	for _, block := range GetBlockchain().blocks {
		fmt.Println("TimeStamp: ", block.TimeStamp)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %s\n", block.Hash)
		fmt.Printf("Prev Hash: %s\n", block.PrevHash)
	}
}