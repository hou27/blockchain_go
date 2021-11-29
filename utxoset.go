package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
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
func (u UTXOSet) Build(bc *Blockchain) {
	db := u.Blockchain.db

	bucketName := []byte(utxoBucket)
	err := u.init(db, bucketName)
	if err != nil {
		log.Panic(err)
	}

	UTXO := bc.FindAllUTXOs()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txId, outs := range UTXO {
			key, err := hex.DecodeString(txId)
			if err != nil {
				log.Panic(err)
			}

			var writer bytes.Buffer

			enc := gob.NewEncoder(&writer)
			err = enc.Encode(outs)
			if err != nil {
				log.Panic(err)
			}
			
			err = b.Put(key, writer.Bytes())
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