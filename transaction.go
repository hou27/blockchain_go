package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"log"
	"math/big"

	"github.com/btcsuite/btcutil/base58"
)

const subsidy = 10

// Coin transaction
type Transaction struct {
	ID		[]byte
	Vin		[]TXInput
	Vout	[]TXOutput
}

type ScriptSig struct {
	Signature	[]byte
    PublicKey	[]byte
}

// Transaction input
type TXInput struct {
	Txid		[]byte
	TxoutIdx	int
	ScriptSig	*ScriptSig // Unlock script 다른 사람으로부터 받은 UTXO를 사용하기 위해 잠금 해제하는 스크립트
}

// Transaction output
type TXOutput struct {
	Value        int
	ScriptPubKey []byte // Lock script 수신자 이외에 그 누구도 열지 못하도록 잠그는 스크립트
}

// Returns the hash of the Transaction to use it as an ID.
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
	var buf bytes.Buffer

	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(tI.ScriptSig)
	if err != nil {
		log.Fatal("Encode Error:", err)
	}

	lockingHash := HashPublicKey(buf.Bytes())

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
	txin := TXInput{[]byte{}, -1, &ScriptSig{nil, []byte(data)}}
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
			input := TXInput{txID, out, &ScriptSig{nil, wallet.PublicKey}}
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
	bc.SignTransaction(&tx, wallet.PrivateKey)
	
	return &tx
}

// Checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].TxoutIdx == -1
}

// Creates a abbreviated copy of Transaction to use in sign
func (tx *Transaction) AbbreviatedCopy() Transaction {
	var inputs []TXInput

	// The public key stored in the input doesn't need to be signed.
	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.TxoutIdx, nil})
	}

	abbreviatedTx := Transaction{tx.ID, inputs, tx.Vout}

	return abbreviatedTx
}

// Signs each input of a Transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTx *Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, vin := range tx.Vin {
		if bytes.Compare(vin.Txid, prevTx.ID) != 0 {
			log.Panic("It's not the previous Transaction.")
		}
	}

	abbreviatedTx := tx.AbbreviatedCopy()

	for inId, vin := range abbreviatedTx.Vin {
		abbreviatedTx.Vin[inId].ScriptSig = &ScriptSig{}
		abbreviatedTx.Vin[inId].ScriptSig.PublicKey = prevTx.Vout[vin.TxoutIdx].ScriptPubKey
		abbreviatedTx.SetID()
		abbreviatedTx.Vin[inId].ScriptSig.PublicKey = nil

		// Use ECDSA(not RSA)
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, abbreviatedTx.ID)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inId].ScriptSig.Signature = signature
	}
}

// Verifies signatures of Transaction inputs
func (tx *Transaction) Verify(prevTx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, vin := range tx.Vin {
		if bytes.Compare(vin.Txid, prevTx.ID) != 0 {
			log.Panic("It's not the previous Transaction.")
		}
	}

	abbreviatedTx := tx.AbbreviatedCopy()
	curve := elliptic.P256() // 키 쌍을 생성할 때 사용된 것과 동일한 곡선

	for inId, vin := range tx.Vin {
		abbreviatedTx.Vin[inId].ScriptSig = &ScriptSig{}
		abbreviatedTx.Vin[inId].ScriptSig.PublicKey = prevTx.Vout[vin.TxoutIdx].ScriptPubKey
		abbreviatedTx.SetID()
		abbreviatedTx.Vin[inId].ScriptSig.PublicKey = nil

		sigLen := len(vin.ScriptSig.Signature)
		keyLen := len(vin.ScriptSig.PublicKey)

		var r, s big.Int
		var x, y big.Int

		// Signature is a pair of numbers.
		r.SetBytes(vin.ScriptSig.Signature[:(sigLen / 2)])
		s.SetBytes(vin.ScriptSig.Signature[(sigLen / 2):])
		
		// PublicKey is a pair of coordinates.
		x.SetBytes(vin.ScriptSig.PublicKey[:(keyLen / 2)])
		y.SetBytes(vin.ScriptSig.PublicKey[(keyLen / 2):])

		rawPublicKey := &ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(rawPublicKey, abbreviatedTx.ID, &r, &s) == false {
			return false
		}
	}
	return true
}