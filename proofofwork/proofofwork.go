package proofofwork

import (
	"bytes"
	"encoding/binary"
	"math/big"

	"github.com/hou27/blockchain_go/blockchain"
)

const targetBits = 6

type Block blockchain.Block

type ProofOfWork struct {
	block *Block
	target *big.Int
}

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	bs := make([]byte, 4)
    binary.LittleEndian.PutUint32(bs, 31415926)

	return bs
}

func NewProofOfWork(b *Block) *ProofOfWork {
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
				IntToHex(pow.block.TimeStamp),
				IntToHex(int64(targetBits)),
				IntToHex(int64(nonce)),
			},
			[]byte{},
	)
	return data
}