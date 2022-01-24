package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

const maxNonce = math.MaxInt64
var targetBits = 6

type ProofOfWork struct {
	block *Block
	target *big.Int
}

// Build a new ProofOfWork and return
func NewProofOfWork(b *Block) *ProofOfWork {
	// targetBits += len(Bc.blocks)
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	pow := &ProofOfWork{b, target}

	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			[]byte(pow.block.PrevHash),
			[]byte(pow.block.Data),
			IntToHex(int64(pow.block.TimeStamp)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}

// Mining
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)

		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			fmt.Println()
			break
		} else {
			nonce++
		}
	}
	return nonce, hash[:]
}

// Validate hash
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	hash := sha256.Sum256(
		pow.prepareData(pow.block.Nonce),
	)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}
