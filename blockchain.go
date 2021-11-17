package main

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-playground/validator"
)

type Block struct {
	TimeStamp	int32 `validate:"required"`
	Hash		[]byte `validate:"required"`
	PrevHash	[]byte `validate:"required"`
	Data		[]byte `validate:"required"`
	Nonce		int `validate:"min=0"`
}

type Blockchain struct {
	blocks []*Block
}

var Bc *Blockchain
var once sync.Once
var errNotValid = errors.New("Can't add this block")

func (bc *Blockchain) validateStructure(newBlock Block) error {
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
		Bc = &Blockchain{}
		Bc.AddBlock("Genesis Block")
	})
}

// Get All Blockchains
func GetBlockchain() *Blockchain {
	if Bc == nil {
		generateGenesis()
	}
	return Bc
}

func (bc Blockchain) getPrevHash() []byte {
	if len(GetBlockchain().blocks) > 0 {
		return GetBlockchain().blocks[len(GetBlockchain().blocks)-1].Hash
	}
	return nil
}

// Prepare new block
func NewBlock(data string, prevHash []byte) *Block {
	newblock := &Block{int32(time.Now().Unix()), nil, prevHash, []byte(data), 0}
	pow := NewProofOfWork(newblock)
	nonce, hash := pow.Run()

	newblock.Hash = hash[:]
	newblock.Nonce = nonce
	return newblock
}

// Add Blockchain
func (bc *Blockchain) AddBlock(data string) {
	prevHash := bc.getPrevHash()
	newBlock := NewBlock(data, prevHash)

	if bc.blocks != nil {
		isValidated := bc.validateStructure(*newBlock)
		if isValidated != nil {
			fmt.Println(isValidated)
		} else {
			bc.blocks = append(GetBlockchain().blocks, newBlock)
			fmt.Println("Added")
		}
	} else {
		bc.blocks = append(GetBlockchain().blocks, newBlock)
		fmt.Println("Added")
	}
}

// Show Blockchains
func (bc Blockchain) ShowBlocks() {
	for _, block := range GetBlockchain().blocks {
		pow := NewProofOfWork(block)
		fmt.Println("TimeStamp:", block.TimeStamp)
		fmt.Printf("Data: %s\n", block.Data)
        fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Prev Hash: %x\n", block.PrevHash)
		fmt.Printf("Prev Hash: %d\n", block.Nonce)
		fmt.Printf("is Validated: %s\n", strconv.FormatBool(pow.Validate()))
	}
}