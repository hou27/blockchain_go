package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"
)

type Block struct {
	TimeStamp		int32 `validate:"required"`
	Hash			[]byte `validate:"required"`
	PrevHash		[]byte `validate:"required"`
	Transactions	[]*Transaction `validate:"required"`
	Nonce			int `validate:"min=0"`
	Height			int `validate:"min=0"`
}

// Generate genesis block
func GenerateGenesis(tx *Transaction) *Block {
	// Genesis block's height is 0.
	return NewBlock([]*Transaction{tx}, []byte{}, 0)
}

// Prepare new block
func NewBlock(transactions []*Transaction, prevHash []byte, height int) *Block {
	newblock := &Block{int32(time.Now().Unix()), nil, prevHash, transactions, 0, height}
	pow := NewProofOfWork(newblock)
	nonce, hash := pow.Run()

	newblock.Hash = hash
	newblock.Nonce = nonce

	return newblock
}

// Hash transactions
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte
	
	for _, tx := range b.Transactions {
			transactions = append(transactions, tx.Serialize())
			fmt.Printf("tx.serialize in hash Tx %x\n", tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)
	fmt.Println("merkleRoot ::: ", mTree.merkleRoot)
	
	return mTree.merkleRoot
}

// Serialize before sending
func (b *Block) Serialize() []byte {
	var writer bytes.Buffer

	encoder := gob.NewEncoder(&writer)
	err := encoder.Encode(b)
	if err != nil {
		log.Panic("Encode Error:", err)
	}

	return writer.Bytes()
}

// Deserialize block(not a method)
func DeserializeBlock(d []byte) *Block {
	var block Block

	dec := gob.NewDecoder(bytes.NewReader(d))
	err := dec.Decode(&block)
	if err != nil {
		log.Panic("Decode Error:", err)
	}

	return &block
}