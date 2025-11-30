package main

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/guidomantilla/yarumo/common/crypto/signatures/ecdsas"
	"github.com/guidomantilla/yarumo/common/crypto/signatures/hmacs"
	"github.com/guidomantilla/yarumo/common/types"
)

func main() {
	var key any
	data := []byte("test data")

	/*
	 * HMAC with SHA256
	 */

	key, _ = hmacs.HMAC_with_SHA256.GenerateKey()
	digest, _ := hmacs.HMAC_with_SHA256.Digest(key.(types.Bytes), data)
	println(fmt.Sprintf("HMAC_with_SHA256 Digest - Hex: %s, Base64: %s", digest.ToHex(), digest.ToBase64Std()))
	verify, _ := hmacs.HMAC_with_SHA256.Validate(key.(types.Bytes), digest, data)
	println(fmt.Sprintf("HMAC_with_SHA256 Validate: %v", verify))

	/*
	 * HMAC with SHA512
	 */
	digest, _ = hmacs.HMAC_with_SHA512.Digest([]byte("abc123"), []byte("Guido Mauricio Mantilla Tarazona"))
	println(fmt.Sprintf("HMAC_with_SHA512 Digest - Hex: %s, Base64: %s", digest.ToHex(), digest.ToBase64Std()))
	verify, _ = hmacs.HMAC_with_SHA512.Validate([]byte("abc123"), digest, []byte("Guido Mauricio Mantilla Tarazona"))
	println(fmt.Sprintf("HMAC_with_SHA512 Validate: %v", verify))

	/*
	 * ECDSA with SHA256 over P256
	 */
	key, err := ecdsas.ECDSA_with_SHA256_over_P256.GenerateKey()
	println(err)

	signature, err := ecdsas.ECDSA_with_SHA256_over_P256.Sign(key.(*ecdsa.PrivateKey), data, ecdsas.ASN1)
	println(err)
	println(fmt.Sprintf("ECDSA_with_SHA256_over_P256 Signature - Hex: %s, Base64: %s", signature.ToHex(), signature.ToBase64Std()))
	verify, err = ecdsas.ECDSA_with_SHA256_over_P256.Verify(&key.(*ecdsa.PrivateKey).PublicKey, signature, data, ecdsas.ASN1)
	println(err)
	println(fmt.Sprintf("ECDSA_with_SHA256_over_P256 Verify: %v", verify))

	/*
	 * ECDSA with SHA512 over P521
	 */
	key, err = ecdsas.ECDSA_with_SHA512_over_P521.GenerateKey()
	println(err)

	signature, err = ecdsas.ECDSA_with_SHA512_over_P521.Sign(key.(*ecdsa.PrivateKey), data, ecdsas.RS)
	println(err)
	println(fmt.Sprintf("ECDSA_with_SHA512_over_P521 Signature - Hex: %s, Base64: %s", signature.ToHex(), signature.ToBase64Std()))
	verify, err = ecdsas.ECDSA_with_SHA512_over_P521.Verify(&key.(*ecdsa.PrivateKey).PublicKey, signature, data, ecdsas.RS)
	println(err)
	println(fmt.Sprintf("ECDSA_with_SHA512_over_P521 Verify: %v", verify))

}
