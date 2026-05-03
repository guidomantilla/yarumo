package main

import (
	"fmt"
	"log"
	"time"

	ctokens "github.com/guidomantilla/yarumo/common/crypto/tokens"
)

func main() {
	predefinedMethods()
	customMethod()
	registryLookup()
	listSupported()
}

// predefinedMethods demonstrates generating and validating tokens with all predefined methods.
func predefinedMethods() {
	fmt.Println("=== Predefined Methods ===")

	subject := "user-123"
	payload := ctokens.Payload{
		"role":  "admin",
		"scope": "read:write",
	}

	methods := []*ctokens.Method{
		ctokens.JWT_HS256,
		ctokens.JWT_HS384,
		ctokens.JWT_HS512,
	}

	for _, m := range methods {
		token, err := m.Generate(subject, payload)
		if err != nil {
			log.Fatalf("%s generate failed: %v", m.Name(), err)
		}

		fmt.Printf("%-10s Token: %.60s...\n", m.Name(), token)

		recovered, err := m.Validate(token)
		if err != nil {
			log.Fatalf("%s validate failed: %v", m.Name(), err)
		}

		fmt.Printf("%-10s Payload: role=%v, scope=%v\n\n", m.Name(), recovered["role"], recovered["scope"])
	}
}

// customMethod demonstrates creating a method with custom options.
func customMethod() {
	fmt.Println("=== Custom Method ===")

	custom := ctokens.NewMethod("CustomHS256", ctokens.SigningMethodHS256,
		ctokens.WithKey([]byte("my-secret-key-at-least-32-bytes!")),
		ctokens.WithIssuer("yarumo-example"),
		ctokens.WithTimeout(1*time.Hour),
	)

	token, err := custom.Generate("user-456", ctokens.Payload{"action": "demo"})
	if err != nil {
		log.Fatalf("custom generate failed: %v", err)
	}

	fmt.Printf("Custom Token: %.60s...\n", token)

	recovered, err := custom.Validate(token)
	if err != nil {
		log.Fatalf("custom validate failed: %v", err)
	}

	fmt.Printf("Custom Payload: action=%v\n\n", recovered["action"])
}

// registryLookup demonstrates retrieving a method by name and handling unknown algorithms.
func registryLookup() {
	fmt.Println("=== Registry Lookup ===")

	method, err := ctokens.Get("JWT_HS256")
	if err != nil {
		log.Fatalf("registry lookup failed: %v", err)
	}

	token, err := method.Generate("registry-user", ctokens.Payload{"via": "registry"})
	if err != nil {
		log.Fatalf("generate via registry failed: %v", err)
	}

	fmt.Printf("ctokens.Get(\"JWT_HS256\"): %s, token %d bytes\n", method.Name(), len(token))

	// Non-existent algorithm
	_, err = ctokens.Get("UNKNOWN")
	fmt.Printf("ctokens.Get(\"UNKNOWN\") error: %v\n\n", err)
}

// listSupported demonstrates listing all registered token methods.
func listSupported() {
	fmt.Println("=== Supported Methods ===")

	for _, m := range ctokens.Supported() {
		fmt.Printf("  - %s\n", m.Name())
	}

	fmt.Println()
}
