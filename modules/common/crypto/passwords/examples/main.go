package main

import (
	"fmt"
	"log"

	cpasswords "github.com/guidomantilla/yarumo/common/crypto/passwords"
)

func main() {
	predefinedMethods()
	registryLookup()
	listSupported()
	byPrefixExample()
}

// predefinedMethods demonstrates encoding, verifying, and upgrade checking with predefined methods.
func predefinedMethods() {
	fmt.Println("=== Predefined Methods ===")

	rawPassword := "my-secret-password"

	methods := []*cpasswords.Method{
		cpasswords.Argon2,
		cpasswords.Bcrypt,
		cpasswords.Pbkdf2,
		cpasswords.Scrypt,
	}

	for _, m := range methods {
		encoded, err := m.Encode(rawPassword)
		if err != nil {
			log.Fatalf("%s encode failed: %v", m.Name(), err)
		}

		fmt.Printf("%-8s Encoded: %.60s...\n", m.Name(), encoded)

		ok, err := m.Verify(encoded, rawPassword)
		if err != nil {
			log.Fatalf("%s verify failed: %v", m.Name(), err)
		}

		fmt.Printf("%-8s Verify:  %v\n", m.Name(), ok)

		needed, err := m.UpgradeNeeded(encoded)
		if err != nil {
			log.Fatalf("%s upgrade check failed: %v", m.Name(), err)
		}

		fmt.Printf("%-8s Upgrade: %v\n\n", m.Name(), needed)
	}
}

// registryLookup demonstrates retrieving a method by name and handling unknown algorithms.
func registryLookup() {
	fmt.Println("=== Registry Lookup ===")

	method, err := cpasswords.Get("Argon2")
	if err != nil {
		log.Fatalf("registry lookup failed: %v", err)
	}

	encoded, err := method.Encode("registry-test")
	if err != nil {
		log.Fatalf("encode via registry failed: %v", err)
	}

	fmt.Printf("cpasswords.Get(\"Argon2\"): %s, encoded %d bytes\n", method.Name(), len(encoded))

	// Non-existent algorithm
	_, err = cpasswords.Get("UNKNOWN")
	fmt.Printf("cpasswords.Get(\"UNKNOWN\") error: %v\n\n", err)
}

// listSupported demonstrates listing all registered password methods.
func listSupported() {
	fmt.Println("=== Supported Methods ===")

	for _, m := range cpasswords.Supported() {
		fmt.Printf("  - %s\n", m.Name())
	}

	fmt.Println()
}

// byPrefixExample demonstrates looking up a method by encoded password prefix.
func byPrefixExample() {
	fmt.Println("=== ByPrefix Lookup ===")

	encoded, err := cpasswords.Bcrypt.Encode("prefix-test")
	if err != nil {
		log.Fatalf("encode failed: %v", err)
	}

	method, err := cpasswords.ByPrefix(encoded)
	if err != nil {
		log.Fatalf("ByPrefix failed: %v", err)
	}

	fmt.Printf("ByPrefix detected: %s\n", method.Name())

	ok, err := method.Verify(encoded, "prefix-test")
	if err != nil {
		log.Fatalf("verify via ByPrefix failed: %v", err)
	}

	fmt.Printf("Verify via ByPrefix: %v\n\n", ok)
}
