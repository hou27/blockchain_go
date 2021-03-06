package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const walletFile = "gowallet_%s.dat"

// Wallet
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// Wallets stores bunch of wallets
type Wallets struct {
	Wallets map[string]*Wallet
}

// Get wallet address
func (w Wallet) GetAddress() string {
	publicKeyHash := HashPublicKey(w.PublicKey)

	return base58.CheckEncode(publicKeyHash, version)
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
func NewWallet() *Wallet {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)

	return &Wallet{*privateKey, publicKey}
}

// CreateWallet adds a Wallet to Wallets
func (ws *Wallets) CreateWallet(nodeID string) string {
	wallet := NewWallet()
	address := wallet.GetAddress()

	ws.Wallets[address] = wallet

	return address
}

// Creates wallets and fills it from a file if it exists
func NewWallets(nodeID string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadSpecificWallet(nodeID)

	return &wallets, err
}

// Saves the wallets to a file
func (ws Wallets) SaveToFile(nodeID string) {
	var content bytes.Buffer

	walletFile := fmt.Sprintf(walletFile, nodeID)

	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

// Returns addresses stored at wallet file
func (ws *Wallets) GetAddresses() []string {
	var addrs []string

	for address := range ws.Wallets {
		addrs = append(addrs, address)
	}

	return addrs
}

// Returns a Wallet by address
func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

func IsValidWallet(address string) bool {
	_, _, err := base58.CheckDecode(address)

	return err == nil
}

// Loads wallets from the file
func (ws *Wallets) LoadSpecificWallet(nodeID string) error {
	walletFile := fmt.Sprintf(walletFile, nodeID)

	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	content, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var wallets Wallets

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(content))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	ws.Wallets = wallets.Wallets

	return nil
}