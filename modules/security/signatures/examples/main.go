package main

import (
	"github.com/guidomantilla/yarumo/security/signatures/macs"
)

func main() {

	data := []byte("test data")
	key := macs.HMAC_with_SHA256.Key()
	signature := macs.HMAC_with_SHA256.Sign(key, data)
	println(signature.ToHex())
	println(signature.ToBase64Std())

}
