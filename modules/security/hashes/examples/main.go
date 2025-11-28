package main

import (
	"github.com/guidomantilla/yarumo/security/hashes"
)

func main() {

	hash := hashes.BLAKE2b_256([]byte("test data"))
	println(hash.ToHex())
	println(hash.ToBase64Std())

}
