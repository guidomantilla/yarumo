package main

import (
	"fmt"
	"log"

	cpasswords "github.com/guidomantilla/yarumo/core/crypto/passwords"
)

func main() {
	predefinedMethods()
	registryLookup()
	listSupported()
	byPrefixExample()
	delegatingMigration()
}

// predefinedMethods demonstrates encoding, verifying, and upgrade checking with predefined methods.
func predefinedMethods() {
	fmt.Println("=== Predefined Methods ===")

	rawPassword := "my-secret-password"

	methods := []*cpasswords.Method{
		cpasswords.Argon2id,
		cpasswords.Argon2i,
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

	method, err := cpasswords.Get("Argon2id")
	if err != nil {
		log.Fatalf("registry lookup failed: %v", err)
	}

	encoded, err := method.Encode("registry-test")
	if err != nil {
		log.Fatalf("encode via registry failed: %v", err)
	}

	fmt.Printf("cpasswords.Get(\"Argon2id\"): %s, encoded %d bytes\n", method.Name(), len(encoded))

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

// delegatingMigration demonstrates the Spring-Security-style login-time
// upgrade pattern: legacy hashes verify, UpgradeNeeded signals migration,
// and re-encoding under the new primary completes the upgrade.
func delegatingMigration() {
	fmt.Println("=== DelegatingEncoder Migration Flow ===")

	const raw = "migrate-me"

	// 1. Existing system encoded the password with Bcrypt.
	legacy, err := cpasswords.Bcrypt.Encode(raw)
	if err != nil {
		log.Fatalf("legacy encode failed: %v", err)
	}
	fmt.Printf("Legacy bcrypt hash : %.60s...\n", legacy)

	// 2. Operator rolls out Argon2id as the new primary.
	delegating := cpasswords.NewDelegatingEncoder(cpasswords.Argon2id)

	// 3. Login: legacy hash still verifies via prefix routing.
	ok, err := delegating.Verify(legacy, raw)
	if err != nil {
		log.Fatalf("verify legacy via delegating failed: %v", err)
	}
	fmt.Printf("Verify legacy hash : %v\n", ok)

	// 4. UpgradeNeeded reports true — algorithm mismatch with primary.
	needed, err := delegating.UpgradeNeeded(legacy)
	if err != nil {
		log.Fatalf("upgrade check failed: %v", err)
	}
	fmt.Printf("Upgrade needed     : %v\n", needed)

	// 5. Re-encode under the primary and persist the new hash.
	upgraded, err := delegating.Encode(raw)
	if err != nil {
		log.Fatalf("re-encode failed: %v", err)
	}
	fmt.Printf("Upgraded hash      : %.60s...\n", upgraded)

	// 6. Subsequent logins verify under the primary and no further upgrade is needed.
	ok, err = delegating.Verify(upgraded, raw)
	if err != nil {
		log.Fatalf("verify upgraded failed: %v", err)
	}
	needed, err = delegating.UpgradeNeeded(upgraded)
	if err != nil {
		log.Fatalf("upgrade check on upgraded failed: %v", err)
	}
	fmt.Printf("Verify upgraded    : %v\n", ok)
	fmt.Printf("Upgrade still need : %v\n\n", needed)
}
