package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"log"
)

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, uint64(num))

	return bs
}

// Generic Encoding
func GobEncode(data interface{}) []byte {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buf.Bytes()
}