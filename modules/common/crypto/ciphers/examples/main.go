package main

import (
	"fmt"
	"log"

	caead "github.com/guidomantilla/yarumo/common/crypto/ciphers/aead"
	crsaoaep "github.com/guidomantilla/yarumo/common/crypto/ciphers/rsaoaep"
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func main() {
	aeadExample()
	rsaOaepExample()
	registryExample()
}

// aeadExample demonstrates all four predefined AEAD methods:
// AES-128-GCM, AES-256-GCM, ChaCha20-Poly1305, and XChaCha20-Poly1305.
func aeadExample() {
	fmt.Println("=== AEAD ===")

	plaintext := ctypes.Bytes("confidential payload")
	additionalData := ctypes.Bytes("authenticated metadata")

	methods := []struct {
		name   string
		method *caead.Method
	}{
		{"AES-128-GCM", caead.AES_128_GCM},
		{"AES-256-GCM", caead.AES_256_GCM},
		{"ChaCha20-Poly1305", caead.CHACHA20_POLY1305},
		{"XChaCha20-Poly1305", caead.XCHACHA20_POLY1305},
	}

	for _, m := range methods {
		key, err := m.method.GenerateKey()
		if err != nil {
			log.Fatalf("%s key generation failed: %v", m.name, err)
		}

		ciphered, err := m.method.Encrypt(key, plaintext, additionalData)
		if err != nil {
			log.Fatalf("%s encryption failed: %v", m.name, err)
		}

		decrypted, err := m.method.Decrypt(key, ciphered, additionalData)
		if err != nil {
			log.Fatalf("%s decryption failed: %v", m.name, err)
		}

		fmt.Printf("%-20s key=%d bytes, ciphertext=%d bytes, decrypted=%q\n",
			m.name, len(key), len(ciphered), string(decrypted))
	}

	fmt.Println()
}

// rsaOaepExample demonstrates RSA-OAEP encryption and decryption
// with different key sizes and hash functions.
func rsaOaepExample() {
	fmt.Println("=== RSA-OAEP ===")

	plaintext := ctypes.Bytes("short secret")
	label := ctypes.Bytes("example-context")

	// SHA-256 with 2048-bit key
	key256, err := crsaoaep.RSA_OAEP_SHA256.GenerateKey(2048)
	if err != nil {
		log.Fatalf("RSA-OAEP SHA256 key generation failed: %v", err)
	}

	ciphered, err := crsaoaep.RSA_OAEP_SHA256.Encrypt(&key256.PublicKey, plaintext, label)
	if err != nil {
		log.Fatalf("RSA-OAEP SHA256 encryption failed: %v", err)
	}

	decrypted, err := crsaoaep.RSA_OAEP_SHA256.Decrypt(key256, ciphered, label)
	if err != nil {
		log.Fatalf("RSA-OAEP SHA256 decryption failed: %v", err)
	}

	fmt.Printf("SHA256/2048 ciphertext=%d bytes, decrypted=%q\n", len(ciphered), string(decrypted))

	// SHA-512 with 4096-bit key
	key512, err := crsaoaep.RSA_OAEP_SHA512.GenerateKey(4096)
	if err != nil {
		log.Fatalf("RSA-OAEP SHA512 key generation failed: %v", err)
	}

	ciphered, err = crsaoaep.RSA_OAEP_SHA512.Encrypt(&key512.PublicKey, plaintext, label)
	if err != nil {
		log.Fatalf("RSA-OAEP SHA512 encryption failed: %v", err)
	}

	decrypted, err = crsaoaep.RSA_OAEP_SHA512.Decrypt(key512, ciphered, label)
	if err != nil {
		log.Fatalf("RSA-OAEP SHA512 decryption failed: %v", err)
	}

	fmt.Printf("SHA512/4096 ciphertext=%d bytes, decrypted=%q\n", len(ciphered), string(decrypted))

	// Encryption with nil label (label is optional)
	ciphered, err = crsaoaep.RSA_OAEP_SHA256.Encrypt(&key256.PublicKey, plaintext, nil)
	if err != nil {
		log.Fatalf("RSA-OAEP SHA256 encryption (nil label) failed: %v", err)
	}

	decrypted, err = crsaoaep.RSA_OAEP_SHA256.Decrypt(key256, ciphered, nil)
	if err != nil {
		log.Fatalf("RSA-OAEP SHA256 decryption (nil label) failed: %v", err)
	}

	fmt.Printf("SHA256/2048 (no label) decrypted=%q\n\n", string(decrypted))
}

// registryExample demonstrates the Get and Supported functions for both cipher packages.
func registryExample() {
	fmt.Println("=== Registry ===")

	// Look up an AEAD method by name
	method, err := caead.Get("AES_256_GCM")
	if err != nil {
		log.Fatalf("aead registry lookup failed: %v", err)
	}

	fmt.Printf("caead.Get: %s\n", method.Name())

	// List all registered AEAD methods
	fmt.Println("caead.Supported:")

	for _, m := range caead.Supported() {
		fmt.Printf("  - %s\n", m.Name())
	}

	// Look up an RSA-OAEP method by name
	rsaMethod, err := crsaoaep.Get("RSA-OAEP-SHA256")
	if err != nil {
		log.Fatalf("rsaoaep registry lookup failed: %v", err)
	}

	fmt.Printf("crsaoaep.Get: %s\n", rsaMethod.Name())

	// List all registered RSA-OAEP methods
	fmt.Println("crsaoaep.Supported:")

	for _, m := range crsaoaep.Supported() {
		fmt.Printf("  - %s\n", m.Name())
	}

	// Non-existent algorithm (aead)
	_, err = caead.Get("UNKNOWN")
	fmt.Printf("caead.Get(\"UNKNOWN\") error: %v\n", err)

	// Non-existent algorithm (rsaoaep)
	_, err = crsaoaep.Get("UNKNOWN")
	fmt.Printf("crsaoaep.Get(\"UNKNOWN\") error: %v\n", err)

	fmt.Println()
}
