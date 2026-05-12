package main

import (
	"crypto"
	"fmt"
	"log"

	chashes "github.com/guidomantilla/yarumo/common/crypto/hashes"
)

func main() {
	data := []byte("the quick brown fox jumps over the lazy dog")

	// Using predefined methods directly
	predefinedMethods(data)

	// Using the standalone Hash function with a crypto.Hash constant
	standaloneHash(data)

	// Using the registry to look up methods by name
	registryLookup(data)

	// Listing all supported methods
	listSupported()
}

// predefinedMethods demonstrates direct use of every predefined hash method.
func predefinedMethods(data []byte) {
	fmt.Println("=== Predefined Methods ===")

	methods := []struct {
		name   string
		method *chashes.Method
	}{
		// SHA-1 is included only for legacy interop (TLS 1.0/1.1, Git,
		// HMAC-SHA1). Do not use it for new collision-sensitive workloads.
		{"SHA1", chashes.SHA1},
		{"SHA224", chashes.SHA224},
		{"SHA256", chashes.SHA256},
		{"SHA512", chashes.SHA512},
		{"SHA3_256", chashes.SHA3_256},
		{"SHA3_384", chashes.SHA3_384},
		{"SHA3_512", chashes.SHA3_512},
		{"BLAKE2b_256", chashes.BLAKE2b_256},
		{"BLAKE2b_512", chashes.BLAKE2b_512},
	}

	for _, m := range methods {
		digest, err := m.method.Hash(data)
		if err != nil {
			fmt.Printf("%-12s error: %v\n", m.name, err)

			continue
		}

		fmt.Printf("%-12s Hex:    %s\n", m.name, digest.ToHex())
		fmt.Printf("%-12s Base64: %s\n", m.name, digest.ToBase64Std())
	}

	fmt.Println()
}

// standaloneHash demonstrates calling chashes.Hash with a crypto.Hash constant,
// bypassing the Method wrapper entirely.
func standaloneHash(data []byte) {
	fmt.Println("=== Standalone Hash Function ===")

	digest, err := chashes.Hash(crypto.SHA256, data)
	if err != nil {
		fmt.Printf("standalone hash error: %v\n", err)

		return
	}

	fmt.Printf("SHA256 (standalone) Hex:    %s\n", digest.ToHex())
	fmt.Printf("SHA256 (standalone) Base64: %s\n", digest.ToBase64Std())
	fmt.Println()
}

// registryLookup demonstrates retrieving a method from the global registry by name.
func registryLookup(data []byte) {
	fmt.Println("=== Registry Lookup ===")

	method, err := chashes.Get("BLAKE2b_256")
	if err != nil {
		log.Fatalf("registry lookup failed: %v", err)
	}

	digest, err := method.Hash(data)
	if err != nil {
		fmt.Printf("registry hash error: %v\n", err)

		return
	}

	fmt.Printf("BLAKE2b_256 (via Get) Hex:    %s\n", digest.ToHex())
	fmt.Printf("BLAKE2b_256 (via Get) Base64: %s\n", digest.ToBase64Std())

	// Attempting to get a non-existent algorithm returns an error.
	_, err = chashes.Get("UNKNOWN")
	fmt.Printf("Get(\"UNKNOWN\") error: %v\n", err)
	fmt.Println()
}

// listSupported demonstrates listing all registered hash methods.
func listSupported() {
	fmt.Println("=== Supported Methods ===")

	supported := chashes.Supported()
	for _, m := range supported {
		fmt.Printf("  - %s\n", m.Name())
	}

	fmt.Println()
}
