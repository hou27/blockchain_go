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
	"github.com/btcsuite/btcutil/base58"
	"github.com/go-playground/validator"
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

func dbExists() bool {
	dbFile := fmt.Sprintf(dbFile, "0600")
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func isValidWallet(address string) bool {
	_, _, err := base58.CheckDecode(address)
	if err != nil {
		return false
	}
	return true
}

// Creates a new blockchain
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}
	if isValidWallet(address) == false {
		fmt.Println("Use correct wallet")
		os.Exit(1)
	}
	var last []byte
	dbFile := fmt.Sprintf(dbFile, "0600")
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
func GetBlockchain() *Blockchain {
	if dbExists() == false {
		fmt.Println("There's no blockchain yet. Create one first.")
		os.Exit(1)
	}
	var last []byte

	dbFile := fmt.Sprintf(dbFile, "0600")
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
func (bc *Blockchain) AddBlock(transactions []*Transaction) {
	var lastHash []byte

	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) != true {
			log.Panic("!!Invalid transaction!!")
		}
	}

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
func (bc *Blockchain) FindUnspentTxs(publicKeyHash []byte) []*Transaction {
	var unspentTXs []*Transaction
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

				if out.IsLockedWithKey(publicKeyHash) {
					unspentTXs = append(unspentTXs, tx)
					continue Outputs
				}
			}
			
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.Unlock(publicKeyHash) {
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

// Finds unspend transaction outputs for the address
func (bc *Blockchain) FindUTXOs(publicKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTxs(publicKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		
		for index, txout := range tx.Vout {
			if txout.IsLockedWithKey(publicKeyHash) && accumulated < amount {
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

			if tx.IsCoinbase() == false {
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
			if bytes.Compare(tx.ID, id) == 0 {
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
	for _, vin := range tx.Vin {
		prevTX, err := bc.GetTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		tx.Sign(privKey, &prevTX)
	}
}

// Verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	var isVerified bool

	for _, vin := range tx.Vin {
		prevTX, err := bc.GetTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		isVerified = tx.Verify(&prevTX)
	}

	return isVerified
}