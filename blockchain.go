package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
	"github.com/go-playground/validator"
)

type Blockchain struct {
	db   *bolt.DB
	last []byte
}

type BlockchainIterator struct {
	db          *bolt.DB
	currentHash []byte
}

const dbFile = "houchain_%s.db"

var Bc *Blockchain
var errNotValid = errors.New("can't add this block")

func (bc *Blockchain) validateStructure(newBlock *Block) error {
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

func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// Creates a new blockchain
func CreateBlockchain(address, nodeID string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)

	if dbExists(dbFile) {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var last []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cb := NewCoinbaseTX(address, "init base")
		genesis := GenerateGenesis(cb)
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
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	bc := Blockchain{db, last}
	return &bc
}

// Get All Blockchains
func GetBlockchain(nodeID string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)

	if !dbExists(dbFile) {
		fmt.Println("There's no blockchain yet. Create one first.")
		os.Exit(1)
	}
	var last []byte
	
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bc := tx.Bucket([]byte("blocks"))
		last = bc.Get([]byte("last"))

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	bc := Blockchain{db, last}
	return &bc
}

// Add Blockchain
func (bc *Blockchain) AddBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		if !bc.VerifyTransaction(tx) {
			log.Panic("!!Invalid transaction!!")
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		lastHash = b.Get([]byte("last"))

		block := DeserializeBlock(b.Get(lastHash))
		lastHeight = block.Height

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash, lastHeight + 1)
	err = bc.validateStructure(newBlock)
	if err != nil {
		log.Panic(err)
	}

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

	return newBlock
}

// Blockchain iterator
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bcI := &BlockchainIterator{bc.db, bc.last}

	return bcI
}

func (bcI *BlockchainIterator) getNextBlock() *Block {
	var block *Block

	err := bcI.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		encodedBlock := b.Get(bcI.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bcI.currentHash = block.PrevHash
	return block
}

func (bc Blockchain) getBestHeight() int {
	var lastHeight int

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		lastHash := b.Get([]byte("last"))

		block := DeserializeBlock(b.Get(lastHash))
		lastHeight = block.Height

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastHeight
}

// Finds all unspent transaction outputs
func (bc *Blockchain) FindAllUTXOs() map[string][]TXOutput {
	UTXO := make(map[string][]TXOutput)
	spentTXOs := make(map[string][]int)
	bcI := bc.Iterator()

	for {
		block := bcI.getNextBlock()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIndex, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIndex {
							continue Outputs
						}
					}
				}
				UTXO[txID] = append(UTXO[txID], out)
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.TxoutIdx)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return UTXO
}

// Get transaction
func (bc *Blockchain) GetTransaction(id []byte) (Transaction, error) {
	bcI := bc.Iterator()
	for {
		block := bcI.getNextBlock()
		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, id) {
				return *tx, nil
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("Transaction not found")
}

// Signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.GetTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// Verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.GetTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}
