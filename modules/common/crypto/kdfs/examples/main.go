package main

import (
	"fmt"
	"log"

	ckdfs "github.com/guidomantilla/yarumo/common/crypto/kdfs"
)

func main() {
	aeadKeyDerivation()
	predefinedMethods()
	registryLookup()
	listSupported()
}

// aeadKeyDerivation demonstrates the canonical use case for HKDF: deriving a
// 32-byte symmetric key suitable for AES-256-GCM or ChaCha20-Poly1305 from a
// high-entropy master secret (e.g. an ECDH shared secret or a TLS handshake
// secret). The info argument binds the derived key to a specific purpose so
// the same master secret can safely produce multiple keys.
func aeadKeyDerivation() {
	fmt.Println("=== HKDF: 32-byte AEAD Key Derivation ===")

	master := []byte("master-secret-from-ecdh-or-tls-handshake")
	salt := []byte("yarumo-app-v1")
	info := []byte("yarumo.aead.key.v1")

	key, err := ckdfs.HKDF_with_SHA256.Derive(master, salt, info, 32)
	if err != nil {
		log.Fatalf("HKDF derive failed: %v", err)
	}

	fmt.Printf("Derived AEAD key length: %d bytes\n", len(key))
	fmt.Printf("Derived AEAD key (hex):  %x\n\n", key)
}

// predefinedMethods demonstrates direct use of every predefined KDF method.
// PBKDF2 and Scrypt require a non-nil salt; HKDF accepts a nil/empty salt.
func predefinedMethods() {
	fmt.Println("=== Predefined Methods ===")

	type call struct {
		name   string
		method *ckdfs.Method
		secret []byte
		salt   []byte
		info   []byte
		length int
	}

	calls := []call{
		{"HKDF_with_SHA256", ckdfs.HKDF_with_SHA256, []byte("ikm-256"), []byte("salt-256"), []byte("info-256"), 32},
		{"HKDF_with_SHA384", ckdfs.HKDF_with_SHA384, []byte("ikm-384"), []byte("salt-384"), []byte("info-384"), 48},
		{"HKDF_with_SHA512", ckdfs.HKDF_with_SHA512, []byte("ikm-512"), []byte("salt-512"), []byte("info-512"), 64},
		// PBKDF2 and Scrypt use the iteration counts / cost parameters bundled
		// with the predefined Methods. With OWASP 2024 defaults this is slow;
		// use small custom params if you are running the example interactively.
		{"PBKDF2_with_SHA256", ckdfs.PBKDF2_with_SHA256, []byte("password"), []byte("salt-pbkdf2"), nil, 32},
		// Scrypt with N=2^17 can take seconds; demoed via a custom small-param
		// method below for responsiveness.
	}

	for _, c := range calls {
		out, err := c.method.Derive(c.secret, c.salt, c.info, c.length)
		if err != nil {
			log.Fatalf("%s derive failed: %v", c.name, err)
		}

		fmt.Printf("%-20s len=%d hex=%x\n", c.name, len(out), out[:8])
	}

	// Scrypt demo with small parameters so the example finishes quickly.
	smallScrypt := ckdfs.NewMethod("Scrypt_Demo", 0, ckdfs.WithScryptParams(1024, 8, 1))

	out, err := smallScrypt.Derive([]byte("password"), []byte("salt-scrypt"), nil, 32)
	if err != nil {
		log.Fatalf("Scrypt demo derive failed: %v", err)
	}

	fmt.Printf("%-20s len=%d hex=%x\n\n", "Scrypt_Demo(N=1024)", len(out), out[:8])
}

// registryLookup demonstrates retrieving a method from the global registry
// by name and handling unknown algorithms.
func registryLookup() {
	fmt.Println("=== Registry Lookup ===")

	method, err := ckdfs.Get("HKDF_with_SHA256")
	if err != nil {
		log.Fatalf("registry lookup failed: %v", err)
	}

	key, err := method.Derive([]byte("ikm"), []byte("salt"), []byte("info"), 32)
	if err != nil {
		log.Fatalf("derive via registry failed: %v", err)
	}

	fmt.Printf("ckdfs.Get(%q): derived %d bytes\n", method.Name(), len(key))

	_, err = ckdfs.Get("UNKNOWN_KDF")
	fmt.Printf("ckdfs.Get(\"UNKNOWN_KDF\") error: %v\n\n", err)
}

// listSupported demonstrates listing all registered KDF methods.
func listSupported() {
	fmt.Println("=== Supported Methods ===")

	for _, m := range ckdfs.Supported() {
		fmt.Printf("  - %s\n", m.Name())
	}

	fmt.Println()
}
