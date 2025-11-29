package main

import (
	"crypto"

	"github.com/guidomantilla/yarumo/security/hashes"
)

func main() {

	data := []byte("test data")

	hash := hashes.BLAKE2b_256.Hash(data)
	println(hash.ToHex())
	println(hash.ToBase64Std())

	hash = hashes.Hash(crypto.BLAKE2b_256, data)
	println(hash.ToHex())
	println(hash.ToBase64Std())

}
