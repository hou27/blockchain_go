package main

import "encoding/binary"

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, uint64(num))

	return bs
}