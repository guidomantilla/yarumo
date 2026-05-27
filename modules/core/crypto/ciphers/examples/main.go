package main

import (
	"bytes"
	"fmt"
	"log"

	caead "github.com/guidomantilla/yarumo/core/crypto/ciphers/aead"
	crsaoaep "github.com/guidomantilla/yarumo/core/crypto/ciphers/rsaoaep"
	ctypes "github.com/guidomantilla/yarumo/core/common/types"
)

func main() {
	aeadExample()
	rsaOaepExample()
	rsaOaepPEMExample()
	registryExample()
	aeadStreamExample()
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

// rsaOaepPEMExample demonstrates loading an RSA private key from a PKCS#8 PEM blob
// (the realistic deployment case where keys are mounted from a secrets manager)
// and using it to decrypt a payload encrypted under the matching public key,
// itself distributed via PKIX/SubjectPublicKeyInfo PEM.
func rsaOaepPEMExample() {
	fmt.Println("=== RSA-OAEP PEM marshal/parse ===")

	priv, err := crsaoaep.RSA_OAEP_SHA256.GenerateKey(2048)
	if err != nil {
		log.Fatalf("RSA-OAEP key generation failed: %v", err)
	}

	privPEM, err := crsaoaep.MarshalPrivateKeyPEM(priv)
	if err != nil {
		log.Fatalf("MarshalPrivateKeyPEM failed: %v", err)
	}

	fmt.Printf("Marshalled private key PEM (%d bytes)\n", len(privPEM))

	pubPEM, err := crsaoaep.MarshalPublicKeyPEM(&priv.PublicKey)
	if err != nil {
		log.Fatalf("MarshalPublicKeyPEM failed: %v", err)
	}

	fmt.Printf("Marshalled public key PEM (%d bytes)\n", len(pubPEM))

	loadedPriv, err := crsaoaep.ParsePrivateKeyPEM(privPEM)
	if err != nil {
		log.Fatalf("ParsePrivateKeyPEM failed: %v", err)
	}

	loadedPub, err := crsaoaep.ParsePublicKeyPEM(pubPEM)
	if err != nil {
		log.Fatalf("ParsePublicKeyPEM failed: %v", err)
	}

	plaintext := ctypes.Bytes("short secret")

	ciphered, err := crsaoaep.RSA_OAEP_SHA256.Encrypt(loadedPub, plaintext, nil)
	if err != nil {
		log.Fatalf("Encrypt with parsed public key failed: %v", err)
	}

	plain, err := crsaoaep.RSA_OAEP_SHA256.Decrypt(loadedPriv, ciphered, nil)
	if err != nil {
		log.Fatalf("Decrypt with parsed private key failed: %v", err)
	}

	fmt.Printf("PEM round-trip decrypted=%q\n\n", string(plain))
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

// aeadStreamExample demonstrates Method.EncryptStream / Method.DecryptStream.
// Each plaintext chunk of at most caead.StreamFrameSize bytes is sealed
// independently with the underlying AEAD primitive and emitted with a
// 4-byte big-endian uint32 length prefix; a zero-length frame closes the
// stream. The frame counter is appended to the caller-supplied AAD so any
// reordering or truncation fails authentication.
func aeadStreamExample() {
	fmt.Println("=== AEAD Streaming (EncryptStream / DecryptStream) ===")

	// Use a payload that spans multiple frames to make the framing visible.
	plaintext := bytes.Repeat([]byte("streaming-aead-payload "), 6000) // ~138 KiB

	method := caead.AES_256_GCM
	aad := ctypes.Bytes("stream-context")

	key, err := method.GenerateKey()
	if err != nil {
		log.Fatalf("GenerateKey failed: %v", err)
	}

	// Encrypt: bytes.Reader -> bytes.Buffer.
	var encrypted bytes.Buffer
	err = method.EncryptStream(key, bytes.NewReader(plaintext), &encrypted, aad)
	if err != nil {
		log.Fatalf("EncryptStream failed: %v", err)
	}

	fmt.Printf("Plaintext: %d bytes (~%d frames of %d bytes)\n",
		len(plaintext),
		(len(plaintext)+caead.StreamFrameSize-1)/caead.StreamFrameSize,
		caead.StreamFrameSize)
	fmt.Printf("Encrypted stream: %d bytes\n", encrypted.Len())

	// Decrypt: bytes.Buffer -> bytes.Buffer.
	var decrypted bytes.Buffer
	err = method.DecryptStream(key, &encrypted, &decrypted, aad)
	if err != nil {
		log.Fatalf("DecryptStream failed: %v", err)
	}

	if !bytes.Equal(decrypted.Bytes(), plaintext) {
		log.Fatalf("plaintext mismatch after stream round-trip")
	}

	fmt.Printf("Round-trip recovered %d bytes (match: %t)\n",
		decrypted.Len(), bytes.Equal(decrypted.Bytes(), plaintext))
	fmt.Println()
}
