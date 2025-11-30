package main

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/guidomantilla/yarumo/common/types"
	"github.com/guidomantilla/yarumo/security/signatures/ecdsas"
	"github.com/guidomantilla/yarumo/security/signatures/hmacs"
)

func main() {
	var key any
	data := []byte("test data")

	key = hmacs.HMAC_with_SHA256.GenerateKey()
	digest := hmacs.HMAC_with_SHA256.Digest(key.(types.Bytes), data)
	println(fmt.Sprintf("HMAC_with_SHA256 Digest - Hex: %s, Base64: %s", digest.ToHex(), digest.ToBase64Std()))
	println(fmt.Sprintf("HMAC_with_SHA256 Validate: %v", hmacs.HMAC_with_SHA256.Validate(key.(types.Bytes), digest, data)))

	key, err := ecdsas.ECDSA_with_SHA256_over_P256.GenerateKey()
	println(err)

	signature, err := ecdsas.ECDSA_with_SHA256_over_P256.Sign(key.(*ecdsa.PrivateKey), data, ecdsas.ASN1DER)
	println(err)
	println(fmt.Sprintf("ECDSA_with_SHA256_over_P256 Signature - Hex: %s, Base64: %s", signature.ToHex(), signature.ToBase64Std()))
	verify, err := ecdsas.ECDSA_with_SHA256_over_P256.Verify(&key.(*ecdsa.PrivateKey).PublicKey, signature, data, ecdsas.ASN1DER)
	println(err)
	println(fmt.Sprintf("ECDSA_with_SHA256_over_P256 Verify: %v", verify))

	key, err = ecdsas.ECDSA_with_SHA512_over_P521.GenerateKey()
	println(err)

	signature, err = ecdsas.ECDSA_with_SHA512_over_P521.Sign(key.(*ecdsa.PrivateKey), data, ecdsas.RS)
	println(err)
	println(fmt.Sprintf("ECDSA_with_SHA512_over_P521 Signature - Hex: %s, Base64: %s", signature.ToHex(), signature.ToBase64Std()))
	verify, err = ecdsas.ECDSA_with_SHA512_over_P521.Verify(&key.(*ecdsa.PrivateKey).PublicKey, signature, data, ecdsas.RS)
	println(err)
	println(fmt.Sprintf("ECDSA_with_SHA512_over_P521 Verify: %v", verify))

}
