package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	TimeStamp		int32 `validate:"required"`
	Hash			[]byte `validate:"required"`
	PrevHash		[]byte `validate:"required"`
	Transactions	[]*Transaction `validate:"required"`
	Nonce			int `validate:"min=0"`
}

// Generate genesis block
func GenerateGenesis(tx *Transaction) *Block {
	return NewBlock([]*Transaction{tx}, []byte{})
}

// Prepare new block
func NewBlock(transactions []*Transaction, prevHash []byte) *Block {
	newblock := &Block{int32(time.Now().Unix()), nil, prevHash, transactions, 0}
	pow := NewProofOfWork(newblock)
	nonce, hash := pow.Run()

	newblock.Hash = hash[:]
	newblock.Nonce = nonce
	return newblock
}

// Hash transactions
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
			transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)
	
	return mTree.merkleRoot
}

// Serialize before sending
func (b *Block) Serialize() []byte {
	var writer bytes.Buffer

	encoder := gob.NewEncoder(&writer)
	err := encoder.Encode(b)
	if err != nil {
		log.Fatal("Encode Error:", err)
	}

	return writer.Bytes()
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