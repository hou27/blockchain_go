package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

type Block struct {
	TimeStamp	int32 `validate:"required"`
	Hash		[]byte `validate:"required"`
	PrevHash	[]byte `validate:"required"`
	Data		[]byte `validate:"required"`
	Nonce		int `validate:"min=0"`
}

type Blockchain struct {
	db		*bolt.DB
	last	[]byte
}

type BlockchainTmp struct {
	db          *bolt.DB
	currentHash []byte
}

const dbFile = "houchain_%s.db"

var Bc *Blockchain

func generateGenesis() *Block {
	return NewBlock("Genesis Block", []byte{})
}

// Get All Blockchains
func GetBlockchain() *Blockchain {
	var last []byte

	dbFile := fmt.Sprintf(dbFile, "0600")
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		bc := tx.Bucket([]byte("blocks"))
		if bc == nil {
			genesis := generateGenesis()
			b, err := tx.CreateBucket([]byte("blocks"))
			if err != nil {
				return err
			}
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				return err
			}
			err = b.Put([]byte("last"), genesis.Hash)
			if err != nil {
				return err
			}
			last = genesis.Hash
		} else {
			last = bc.Get([]byte("last"))
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	bc := Blockchain{db, last}
    return &bc
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
// add to boltDB
// Add Blockchain
func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		lastHash = b.Get([]byte("last"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(data, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("last"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.last = newBlock.Hash

		fmt.Println("Successfully Added")

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// Show Blockchains
func (bc Blockchain) ShowBlocks() {
	bcT := bc.Iterator()
	
	for {
		block := bcT.GetNextBlock()
		pow := NewProofOfWork(block)

		fmt.Println("\nTimeStamp:", block.TimeStamp)
		fmt.Printf("Data: %s\n", block.Data)
        fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Prev Hash: %x\n", block.PrevHash)
		fmt.Printf("Nonce: %d\n", block.Nonce)

		fmt.Printf("is Validated: %s\n", strconv.FormatBool(pow.Validate()))

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (bc *Blockchain) Iterator() *BlockchainTmp {
	bcT := &BlockchainTmp{bc.db, bc.last}
 
	return bcT
}

func (bct *BlockchainTmp) GetNextBlock() *Block {
	var block *Block

	err := bct.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		encodedBlock := b.Get(bct.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bct.currentHash = block.PrevHash
	return block
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
