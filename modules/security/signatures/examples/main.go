package main

import (
	"github.com/guidomantilla/yarumo/security/signatures/hmacs"
)

func main() {

	data := []byte("test data")
	key := hmacs.HMAC_with_SHA256.Key()
	signature := hmacs.HMAC_with_SHA256.Sign(key, data)
	println(signature.ToHex())
	println(signature.ToBase64Std())

}
