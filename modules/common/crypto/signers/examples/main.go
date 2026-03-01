package main

import (
	"crypto/ed25519"
	"fmt"
	"log"

	cecdsas "github.com/guidomantilla/yarumo/common/crypto/signers/ecdsas"
	ced25519 "github.com/guidomantilla/yarumo/common/crypto/signers/ed25519"
	chmacs "github.com/guidomantilla/yarumo/common/crypto/signers/hmacs"
	crsapss "github.com/guidomantilla/yarumo/common/crypto/signers/rsapss"
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func main() {
	data := []byte("message to sign")

	hmacExample(data)
	ecdsaExample(data)
	ed25519Example(data)
	rsaPssExample(data)
	registryExample()
}

// hmacExample demonstrates HMAC key generation, digest computation, and validation.
func hmacExample(data []byte) {
	fmt.Println("=== HMAC ===")

	// SHA-256 variant with a generated key
	key256, err := chmacs.HMAC_with_SHA256.GenerateKey()
	if err != nil {
		log.Fatalf("HMAC SHA256 key generation failed: %v", err)
	}

	digest, err := chmacs.HMAC_with_SHA256.Digest(key256, data)
	if err != nil {
		log.Fatalf("HMAC SHA256 digest failed: %v", err)
	}

	fmt.Printf("HMAC_with_SHA256 Digest Hex:    %s\n", digest.ToHex())
	fmt.Printf("HMAC_with_SHA256 Digest Base64: %s\n", digest.ToBase64Std())

	ok, err := chmacs.HMAC_with_SHA256.Validate(key256, digest, data)
	if err != nil {
		log.Fatalf("HMAC SHA256 validate failed: %v", err)
	}

	fmt.Printf("HMAC_with_SHA256 Valid: %v\n\n", ok)

	// SHA-512 variant with a user-provided key
	userKey := ctypes.Bytes("my-secret-key")

	digest, err = chmacs.HMAC_with_SHA512.Digest(userKey, data)
	if err != nil {
		log.Fatalf("HMAC SHA512 digest failed: %v", err)
	}

	fmt.Printf("HMAC_with_SHA512 Digest Hex:    %s\n", digest.ToHex())

	ok, err = chmacs.HMAC_with_SHA512.Validate(userKey, digest, data)
	if err != nil {
		log.Fatalf("HMAC SHA512 validate failed: %v", err)
	}

	fmt.Printf("HMAC_with_SHA512 Valid: %v\n\n", ok)
}

// ecdsaExample demonstrates ECDSA key generation, signing, and verification
// using both ASN1 and RS signature formats.
func ecdsaExample(data []byte) {
	fmt.Println("=== ECDSA ===")

	// P256 / SHA-256 with ASN1 encoding
	p256Key, err := cecdsas.ECDSA_with_SHA256_over_P256.GenerateKey()
	if err != nil {
		log.Fatalf("ECDSA P256 key generation failed: %v", err)
	}

	sig, err := cecdsas.ECDSA_with_SHA256_over_P256.Sign(p256Key, data, cecdsas.ASN1)
	if err != nil {
		log.Fatalf("ECDSA P256 sign failed: %v", err)
	}

	fmt.Printf("P256/SHA256 Signature (ASN1) Hex: %s\n", sig.ToHex())

	ok, err := cecdsas.ECDSA_with_SHA256_over_P256.Verify(&p256Key.PublicKey, sig, data, cecdsas.ASN1)
	if err != nil {
		log.Fatalf("ECDSA P256 verify failed: %v", err)
	}

	fmt.Printf("P256/SHA256 Verify (ASN1): %v\n\n", ok)

	// P521 / SHA-512 with RS encoding
	p521Key, err := cecdsas.ECDSA_with_SHA512_over_P521.GenerateKey()
	if err != nil {
		log.Fatalf("ECDSA P521 key generation failed: %v", err)
	}

	sig, err = cecdsas.ECDSA_with_SHA512_over_P521.Sign(p521Key, data, cecdsas.RS)
	if err != nil {
		log.Fatalf("ECDSA P521 sign failed: %v", err)
	}

	fmt.Printf("P521/SHA512 Signature (RS) Hex: %s\n", sig.ToHex())

	ok, err = cecdsas.ECDSA_with_SHA512_over_P521.Verify(&p521Key.PublicKey, sig, data, cecdsas.RS)
	if err != nil {
		log.Fatalf("ECDSA P521 verify failed: %v", err)
	}

	fmt.Printf("P521/SHA512 Verify (RS): %v\n\n", ok)
}

// ed25519Example demonstrates Ed25519 key generation, signing, and verification.
func ed25519Example(data []byte) {
	fmt.Println("=== Ed25519 ===")

	key, err := ced25519.Ed25519.GenerateKey()
	if err != nil {
		log.Fatalf("Ed25519 key generation failed: %v", err)
	}

	sig, err := ced25519.Ed25519.Sign(&key, data)
	if err != nil {
		log.Fatalf("Ed25519 sign failed: %v", err)
	}

	fmt.Printf("Ed25519 Signature Hex: %s\n", sig.ToHex())

	pubKey, ok := key.Public().(ed25519.PublicKey)
	if !ok {
		log.Fatal("Ed25519 public key type assertion failed")
	}

	ok, err = ced25519.Ed25519.Verify(&pubKey, sig, data)
	if err != nil {
		log.Fatalf("Ed25519 verify failed: %v", err)
	}

	fmt.Printf("Ed25519 Verify: %v\n\n", ok)
}

// rsaPssExample demonstrates RSA-PSS key generation, signing, and verification
// with different key sizes and hash functions.
func rsaPssExample(data []byte) {
	fmt.Println("=== RSA-PSS ===")

	// SHA-256 with 2048-bit key
	key256, err := crsapss.RSASSA_PSS_using_SHA256.GenerateKey(2048)
	if err != nil {
		log.Fatalf("RSA-PSS SHA256 key generation failed: %v", err)
	}

	sig, err := crsapss.RSASSA_PSS_using_SHA256.Sign(key256, data)
	if err != nil {
		log.Fatalf("RSA-PSS SHA256 sign failed: %v", err)
	}

	fmt.Printf("RSA-PSS/SHA256 (2048) Signature (%d bytes)\n", len(sig))

	ok, err := crsapss.RSASSA_PSS_using_SHA256.Verify(&key256.PublicKey, sig, data)
	if err != nil {
		log.Fatalf("RSA-PSS SHA256 verify failed: %v", err)
	}

	fmt.Printf("RSA-PSS/SHA256 Verify: %v\n\n", ok)

	// SHA-512 with 4096-bit key
	key512, err := crsapss.RSASSA_PSS_using_SHA512.GenerateKey(4096)
	if err != nil {
		log.Fatalf("RSA-PSS SHA512 key generation failed: %v", err)
	}

	sig, err = crsapss.RSASSA_PSS_using_SHA512.Sign(key512, data)
	if err != nil {
		log.Fatalf("RSA-PSS SHA512 sign failed: %v", err)
	}

	fmt.Printf("RSA-PSS/SHA512 (4096) Signature (%d bytes)\n", len(sig))

	ok, err = crsapss.RSASSA_PSS_using_SHA512.Verify(&key512.PublicKey, sig, data)
	if err != nil {
		log.Fatalf("RSA-PSS SHA512 verify failed: %v", err)
	}

	fmt.Printf("RSA-PSS/SHA512 Verify: %v\n\n", ok)
}

// registryExample demonstrates the Get and Supported functions across all signer packages.
func registryExample() {
	fmt.Println("=== Registry ===")

	// Look up an HMAC method by name
	hmac, err := chmacs.Get("HMAC_with_SHA256")
	if err != nil {
		log.Fatalf("hmac registry lookup failed: %v", err)
	}

	fmt.Printf("chmacs.Get: %s\n", hmac.Name())

	// List all registered HMAC methods
	fmt.Println("chmacs.Supported:")

	for _, m := range chmacs.Supported() {
		fmt.Printf("  - %s\n", m.Name())
	}

	// Look up an ECDSA method by name
	ecdsa, err := cecdsas.Get("ECDSA_with_SHA256_over_P256")
	if err != nil {
		log.Fatalf("ecdsa registry lookup failed: %v", err)
	}

	fmt.Printf("cecdsas.Get: %s\n", ecdsa.Name())

	// List all registered ECDSA methods
	fmt.Println("cecdsas.Supported:")

	for _, m := range cecdsas.Supported() {
		fmt.Printf("  - %s\n", m.Name())
	}

	// Look up an Ed25519 method by name
	ed, err := ced25519.Get("Ed25519")
	if err != nil {
		log.Fatalf("ed25519 registry lookup failed: %v", err)
	}

	fmt.Printf("ced25519.Get: %s\n", ed.Name())

	// List all registered Ed25519 methods
	fmt.Println("ced25519.Supported:")

	for _, m := range ced25519.Supported() {
		fmt.Printf("  - %s\n", m.Name())
	}

	// Look up an RSA-PSS method by name
	rsa, err := crsapss.Get("RSASSA_PSS_using_SHA256")
	if err != nil {
		log.Fatalf("rsapss registry lookup failed: %v", err)
	}

	fmt.Printf("crsapss.Get: %s\n", rsa.Name())

	// List all registered RSA-PSS methods
	fmt.Println("crsapss.Supported:")

	for _, m := range crsapss.Supported() {
		fmt.Printf("  - %s\n", m.Name())
	}

	// Non-existent algorithm (hmacs)
	_, err = chmacs.Get("UNKNOWN")
	fmt.Printf("chmacs.Get(\"UNKNOWN\") error: %v\n", err)

	// Non-existent algorithm (ecdsas)
	_, err = cecdsas.Get("UNKNOWN")
	fmt.Printf("cecdsas.Get(\"UNKNOWN\") error: %v\n", err)

	// Non-existent algorithm (ed25519)
	_, err = ced25519.Get("UNKNOWN")
	fmt.Printf("ced25519.Get(\"UNKNOWN\") error: %v\n", err)

	// Non-existent algorithm (rsapss)
	_, err = crsapss.Get("UNKNOWN")
	fmt.Printf("crsapss.Get(\"UNKNOWN\") error: %v\n", err)

	fmt.Println()
}
