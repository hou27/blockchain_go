package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	TimeStamp	int32 `validate:"required"`
	Hash		[]byte `validate:"required"`
	PrevHash	[]byte `validate:"required"`
	Data		[]byte `validate:"required"`
	Nonce		int `validate:"min=0"`
}

// Generate genesis block
func GenerateGenesis() *Block {
	return NewBlock("Genesis Block", []byte{})
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

// Serialize before sending
func (b *Block) Serialize() []byte {
	var value bytes.Buffer

	encoder := gob.NewEncoder(&value)
	err := encoder.Encode(b)
	if err != nil {
		log.Fatal("Encode Error:", err)
	}

	return value.Bytes()
}

// Deserialize block(not a method)
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Fatal("Decode Error:", err)
	}

	return &block
}