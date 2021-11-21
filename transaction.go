package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"log"
)

const subsidy = 10

// Coin transaction
type Transaction struct {
	ID		[]byte
	Txin	[]TXInput
	Txout	[]TXOutput
}

// Transaction input
type TXInput struct {
	Txid      []byte
	Txout     int
	ScriptSig string // Unlock script
}

// Transaction output
type TXOutput struct {
	Value        int
	ScriptPubKey string // Lock script
}

// Sets ID of a transaction
func (tx Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// Creates a new coinbase transaction
func NewCoinbaseTX(to, data string) *Transaction {
	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}

	return &tx
}

// Creates a new transaction
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	
	balance, validOutputs := bc.FindUTXOs(from, amount)

	if balance < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// Build a list of inputs
	for txid, outs := range validOutputs {
		for _, out := range outs {
			txID, err := hex.DecodeString(txid)
			if err != nil {
				log.Panic(err)
			}

			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	outputs = append(outputs, TXOutput{amount, to})
	if balance > amount {
		outputs = append(outputs, TXOutput{balance - amount, from}) // a change
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}

// Checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Txin) == 1 && len(tx.Txin[0].Txid) == 0 && tx.Txin[0].Txout == -1
}