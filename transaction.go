package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"log"

	"github.com/btcsuite/btcutil/base58"
)

const subsidy = 10

// Coin transaction
type Transaction struct {
	ID		[]byte
	Vin		[]TXInput
	Vout	[]TXOutput
}

// Transaction input
type TXInput struct {
	Txid		[]byte
	TxoutIdx	int
	ScriptSig	[]byte // Unlock script
}

// Transaction output
type TXOutput struct {
	Value        int
	ScriptPubKey []byte // Lock script
}

// Sets ID of a transaction
func (tx *Transaction) SetID() {
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

// Unlock Tx
func (tI TXInput) Unlock(publicKeyHash []byte) bool {
	lockingHash := HashPublicKey(tI.ScriptSig)

	return bytes.Compare(lockingHash, publicKeyHash) == 0
}

// Check key
func (tO TXOutput) IsLockedWithKey(publicKeyHash []byte) bool {
	return bytes.Compare(tO.ScriptPubKey, publicKeyHash) == 0
}

// Lock with publicKey
func (tO *TXOutput) Lock(address string) {
	publicKeyHash, _, err := base58.CheckDecode(address)
	if err != nil {
		log.Panic(err)
	}
	tO.ScriptPubKey = publicKeyHash
}

// Create a new TXOutput
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock(address)

	return txo
}

// Creates a new coinbase transaction
func NewCoinbaseTX(to, data string) *Transaction {
	txin := TXInput{[]byte{}, -1, []byte(data)}
	txout := *NewTXOutput(subsidy, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}

	return &tx
}

// Creates a new transaction
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)
	publicKeyHash := HashPublicKey(wallet.PublicKey)
	balance, validOutputs := bc.FindUTXOs(publicKeyHash, amount)

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
			input := TXInput{txID, out, wallet.PublicKey}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	outputs = append(outputs, *NewTXOutput(amount, to))
	if balance > amount {
		outputs = append(outputs, *NewTXOutput(balance - amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	
	return &tx
}

// Checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].TxoutIdx == -1
}