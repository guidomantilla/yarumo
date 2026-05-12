// Example demonstrating HPKE hybrid encryption for arbitrary-size payloads.
//
// RSA-OAEP cannot encrypt more than roughly 190 bytes for an RSA-2048 +
// SHA-256 key. Hybrid encryption (HPKE/RFC 9180) sidesteps that limit by
// using an asymmetric KEM to ship a per-message AEAD key and the AEAD to
// encrypt the actual payload. This program encrypts a multi-megabyte buffer
// to demonstrate the difference.
package main

import (
	"bytes"
	"fmt"
	"log"

	chybrid "github.com/guidomantilla/yarumo/common/crypto/ciphers/hybrid"
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func main() {
	hybridExample()
	registryExample()
}

// hybridExample demonstrates HPKE encryption/decryption of a multi-megabyte
// payload — explicitly the case RSA-OAEP cannot handle.
func hybridExample() {
	fmt.Println("=== HPKE (RFC 9180) ===")

	pub, priv, err := chybrid.HPKE_X25519_HKDF_SHA256_AES_256_GCM.GenerateKey()
	if err != nil {
		log.Fatalf("HPKE key generation failed: %v", err)
	}

	// 4 MiB payload — comfortably above any RSA-OAEP limit.
	payload := ctypes.Bytes(bytes.Repeat([]byte("yarumo-hybrid-encryption "), 4*1024*1024/24))
	info := ctypes.Bytes("example-context")

	fmt.Printf("plaintext size: %d bytes\n", len(payload))

	ciphered, err := chybrid.HPKE_X25519_HKDF_SHA256_AES_256_GCM.Encrypt(pub, payload, info)
	if err != nil {
		log.Fatalf("HPKE encryption failed: %v", err)
	}

	fmt.Printf("ciphertext size: %d bytes (overhead = %d bytes)\n",
		len(ciphered), len(ciphered)-len(payload))

	decrypted, err := chybrid.HPKE_X25519_HKDF_SHA256_AES_256_GCM.Decrypt(priv, ciphered, info)
	if err != nil {
		log.Fatalf("HPKE decryption failed: %v", err)
	}

	if !bytes.Equal(decrypted, payload) {
		log.Fatal("decrypted bytes do not match the original payload")
	}

	fmt.Printf("decrypted payload matches original (%d bytes)\n\n", len(decrypted))
}

// registryExample demonstrates Get and Supported for the hybrid package.
func registryExample() {
	fmt.Println("=== Registry ===")

	method, err := chybrid.Get("HPKE_X25519_HKDF_SHA256_AES_256_GCM")
	if err != nil {
		log.Fatalf("registry lookup failed: %v", err)
	}

	fmt.Printf("chybrid.Get: %s\n", method.Name())

	fmt.Println("chybrid.Supported:")

	for _, m := range chybrid.Supported() {
		fmt.Printf("  - %s\n", m.Name())
	}

	_, err = chybrid.Get("UNKNOWN")
	fmt.Printf("chybrid.Get(\"UNKNOWN\") error: %v\n", err)

	fmt.Println()
}
