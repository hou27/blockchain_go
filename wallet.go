package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const walletFile = "gowallet.dat"

// Wallet
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// Get wallet address
func (w Wallet) GetAddress() string {
	pubKeyHash := HashPublicKey(w.PublicKey)

	return base58.CheckEncode(pubKeyHash, version)
}

// Hash public key
func HashPublicKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}

	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)
	return publicRIPEMD160
}

// Generate New Wallet
func NewWallet() (*Wallet, error) {
	if _, err := os.Stat(walletFile); !os.IsNotExist(err) {
		return nil, errors.New("Wallet already exists")
	}
	
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)

	return &Wallet{*privateKey, publicKey}, nil
}

// SaveToFile saves the wallet to a file
func (w Wallet) SaveToFile() {
	var content bytes.Buffer

	// https://stackoverflow.com/questions/32676898/whats-the-purpose-of-gob-register-method
	// 인코더와 디코더에 대한 구체적인 유형을 등록한다. https://runebook.dev/ko/docs/go/encoding/gob/index
	gob.Register(elliptic.P256()) // https://pkg.go.dev/crypto/elliptic
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(w)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}