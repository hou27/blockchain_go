package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/go-playground/validator"
)

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
			cb := NewCoinbaseTX("init", "init base")
			genesis := GenerateGenesis(cb)
			b, err := tx.CreateBucket([]byte("blocks"))
			if err != nil {
				log.Fatal(err)
			}
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Fatal(err)
			}
			err = b.Put([]byte("last"), genesis.Hash)
			if err != nil {
				log.Fatal(err)
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

// Add Blockchain
func (bc *Blockchain) AddBlock(transactions []*Transaction) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		lastHash = b.Get([]byte("last"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash)

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
		block := bcT.getNextBlock()
		pow := NewProofOfWork(block)

		fmt.Println("\nTimeStamp:", block.TimeStamp)
		fmt.Println("Data: ", block.Transactions)
        fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Prev Hash: %x\n", block.PrevHash)
		fmt.Printf("Nonce: %d\n", block.Nonce)

		fmt.Printf("is Validated: %s\n", strconv.FormatBool(pow.Validate()))

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

// Blockchain iterator
func (bc *Blockchain) Iterator() *BlockchainTmp {
	bcT := &BlockchainTmp{bc.db, bc.last}
 
	return bcT
}

func (bct *BlockchainTmp) getNextBlock() *Block {
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
