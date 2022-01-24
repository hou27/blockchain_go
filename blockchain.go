package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

type Blockchain struct {
	db		*bolt.DB
	last	[]byte
}

type BlockchainIterator struct {
	db          *bolt.DB
	currentHash []byte
}

const dbFile = "houchain_%s.db"

var Bc *Blockchain

// Get All Blockchains
func GetBlockchain(address string) *Blockchain {
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
			cb := NewCoinbaseTX(address, "init base")
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

// Returns a list of transactions containing unspent outputs
func (bc *Blockchain) FindUnspentTxs(address string) []*Transaction {
	var unspentTXs []*Transaction
	spentTXOs := make(map[string][]int)
	bcI := bc.Iterator()

	for {
		block := bcI.getNextBlock()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIndex, out := range tx.Txout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIndex {
							continue Outputs
						}
					}
				}

				if out.ScriptPubKey == address {
					unspentTXs = append(unspentTXs, tx)
					continue Outputs
				}
			}
			
			if !tx.IsCoinbase() {
				for _, in := range tx.Txin {
					if in.ScriptSig == address {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.TxoutIdx)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return unspentTXs
}

// Finds and returns unspend transaction outputs for the address
func (bc *Blockchain) FindUTXOs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTxs(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		
		for index, txout := range tx.Txout {
			if txout.ScriptPubKey == address && accumulated < amount {
				accumulated += txout.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], index)
				if accumulated >= amount {
					break Work
				}
			}
		}
	}
	
	return accumulated, unspentOutputs
}