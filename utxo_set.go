package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

const utxoBucket = "chainstate"

// UTXOSet represents UTXO set
type UTXOSet struct {
	Blockchain *Blockchain
}

func (u UTXOSet) init(db *bolt.DB, bucketName []byte) error {
	err := db.Update(func(tx *bolt.Tx) error {
		
		b := tx.Bucket(bucketName)

		if b != nil {
			err := tx.DeleteBucket(bucketName)
			if err != nil {
				return err
			}
		}
		_, err := tx.CreateBucket(bucketName)
		if err != nil {
			return err
		}

		return nil
	})
	return err
}

// Builds the UTXO set
func (u UTXOSet) Build() {
	db := u.Blockchain.db

	bucketName := []byte(utxoBucket)
	err := u.init(db, bucketName)
	if err != nil {
		log.Panic(err)
	}

	UTXO := u.Blockchain.FindAllUTXOs()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}
			err = b.Put(key, SerializeTxs(outs))
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
}

// Finds UTXO in chainstate
func (u UTXOSet) FindUTXOs(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := u.Blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		b.ForEach(func(k, v []byte) error {
			outs := DeserializeTxs(v)

			fmt.Printf("%v\n", outs)
			for _, out := range outs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
			return nil
		})

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return UTXOs
}

// Updates the UTXO set(Add new UTXO && Remove used UTXO)
func (u UTXOSet) Update(block *Block) {
	db := u.Blockchain.db

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				// Remove used UTXO
				for _, vin := range tx.Vin {
					var newOuts []TXOutput
					data := b.Get(vin.Txid)
					outs := DeserializeTxs(data)

					for outIdx, out := range outs {
						if outIdx != vin.TxoutIdx {
							newOuts = append(newOuts, out)
						}
					}

					if len(newOuts) == 0 {
						err := b.Delete(vin.Txid)
						if err != nil {
							log.Panic(err)
						}
					} else {
						// Save other UTXOs that still available
						err := b.Put(vin.Txid, SerializeTxs(newOuts))
						if err != nil {
							log.Panic(err)
						}
					}
				}
			}

			// Add new UTXO
			var newOuts []TXOutput
			for _, out := range tx.Vout {
				newOuts = append(newOuts, out)
			}

			err := b.Put(tx.ID, SerializeTxs(newOuts))
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
}

// Finds unspend transaction outputs for the address
func (u UTXOSet) FindMyUTXOs(publicKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	db := u.Blockchain.db
	accumulated := 0

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		
		b.ForEach(func(k, v []byte) error {
			txID := hex.EncodeToString(k)
			outs := DeserializeTxs(v)
		Work:
			for index, txout := range outs {
				if txout.IsLockedWithKey(publicKeyHash) && accumulated < amount {
					accumulated += txout.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], index)
				}
				if accumulated >= amount {
					break Work
				}
			}
			return nil
		})

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	
	return accumulated, unspentOutputs
}