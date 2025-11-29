package main

import (
	"crypto"
	"crypto/elliptic"

	"github.com/guidomantilla/yarumo/security/signatures/ecdsa"
)

func main() {
	key, err := ecdsa.Key(ecdsa.ES256)
	if err != nil {
		panic(err)
	}

	data := []byte("hola guido")
	signature, err := ecdsa.ECDSA_P256_SHA256(key, data)
	if err != nil {
		panic(err)
	}

	println(signature.ToHex())
	println(signature.ToBase64Std())
	println(key.Curve == elliptic.P256())
	println(len(signature) == 64)

	hash := crypto.SHA256
	hash.Available()
}
