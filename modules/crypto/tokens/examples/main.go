package main

import (
	"fmt"
	"log"
	"time"

	caead "github.com/guidomantilla/yarumo/crypto/ciphers/aead"
	rsassas "github.com/guidomantilla/yarumo/crypto/signers/rsassas"
	ctokens "github.com/guidomantilla/yarumo/crypto/tokens"
)

func main() {
	predefinedMethods()
	customMethod()
	asymmetricMethod()
	registryLookup()
	listSupported()
	opaqueRoundTrip()
}

// predefinedMethods demonstrates generating and validating tokens with all predefined methods.
//
// As of YA-0008, the predefined JWT_HS256/384/512 are templates without keys.
// To use them for Generate/Validate, build a fresh Method with WithGeneratedKey()
// (or WithKey for a deterministic secret).
func predefinedMethods() {
	fmt.Println("=== Predefined Methods ===")

	subject := "user-123"
	payload := ctokens.Payload{
		"role":  "admin",
		"scope": "read:write",
	}

	methods := []*ctokens.Method{
		ctokens.NewMethod(ctokens.JWT_HS256.Name(), ctokens.AlgorithmHS256, ctokens.WithGeneratedKey()),
		ctokens.NewMethod(ctokens.JWT_HS384.Name(), ctokens.AlgorithmHS384, ctokens.WithGeneratedKey()),
		ctokens.NewMethod(ctokens.JWT_HS512.Name(), ctokens.AlgorithmHS512, ctokens.WithGeneratedKey()),
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

	custom := ctokens.NewMethod("CustomHS256", ctokens.AlgorithmHS256,
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

// asymmetricMethod demonstrates an end-to-end RS256 round-trip using a fresh
// RSA key generated through the signers/rsassas helper. The same shape works
// for the other predefined asymmetric variants (RS384/512, PS256/384/512,
// ES256/384/512, EdDSA) — only the key generator and Algorithm constant change.
func asymmetricMethod() {
	fmt.Println("=== Asymmetric Method (RS256) ===")

	priv, err := rsassas.RSASSA_PKCS1v15_using_SHA256.GenerateKey(2048)
	if err != nil {
		log.Fatalf("rsa GenerateKey failed: %v", err)
	}

	m := ctokens.NewMethod("JWT_RS256_demo", ctokens.AlgorithmRS256,
		ctokens.WithSigningKey(priv),
		ctokens.WithVerifyingKey(&priv.PublicKey),
		ctokens.WithIssuer("yarumo-example"),
	)

	token, err := m.Generate("user-789", ctokens.Payload{
		"role":  "issuer",
		"scope": "rsa:sign",
	})
	if err != nil {
		log.Fatalf("RS256 generate failed: %v", err)
	}

	fmt.Printf("RS256 Token: %.60s...\n", token)

	recovered, err := m.Validate(token)
	if err != nil {
		log.Fatalf("RS256 validate failed: %v", err)
	}

	fmt.Printf("RS256 Payload: role=%v, scope=%v\n\n", recovered["role"], recovered["scope"])
}

// registryLookup demonstrates retrieving a method by name and handling unknown algorithms.
//
// The registry stores templates without keys; callers register a method with a
// key (or WithGeneratedKey()) before they can Generate/Validate via the lookup.
func registryLookup() {
	fmt.Println("=== Registry Lookup ===")

	ctokens.Register(*ctokens.NewMethod("JWT_HS256_keyed", ctokens.AlgorithmHS256, ctokens.WithGeneratedKey()))

	method, err := ctokens.Get("JWT_HS256_keyed")
	if err != nil {
		log.Fatalf("registry lookup failed: %v", err)
	}

	token, err := method.Generate("registry-user", ctokens.Payload{"via": "registry"})
	if err != nil {
		log.Fatalf("generate via registry failed: %v", err)
	}

	fmt.Printf("ctokens.Get(\"JWT_HS256_keyed\"): %s, token %d bytes\n", method.Name(), len(token))

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

// opaqueRoundTrip demonstrates the YA-0019 opaque (AEAD-encrypted) flavor.
// The entire claims payload is encrypted under a symmetric key, so the
// emitted token is base64url ciphertext — nothing leaks to the client.
func opaqueRoundTrip() {
	fmt.Println("=== Opaque (AEAD) Round Trip ===")

	// AES-256-GCM expects a 32-byte key.
	key, err := caead.AES_256_GCM.GenerateKey()
	if err != nil {
		log.Fatalf("opaque key generation failed: %v", err)
	}

	opaque := ctokens.NewMethod("opaque-demo", ctokens.AlgorithmOpaqueAESGCM,
		ctokens.WithKey(key),
		ctokens.WithIssuer("yarumo-example"),
		ctokens.WithTimeout(1*time.Hour),
	)

	subject := "user-789"
	payload := ctokens.Payload{
		"role":   "admin",
		"tenant": "acme",
		"scope":  "read:write",
	}

	token, err := opaque.Generate(subject, payload)
	if err != nil {
		log.Fatalf("opaque generate failed: %v", err)
	}

	fmt.Printf("Opaque Token: %.60s...\n", token)

	recovered, err := opaque.Validate(token)
	if err != nil {
		log.Fatalf("opaque validate failed: %v", err)
	}

	fmt.Printf("Opaque Payload: role=%v, tenant=%v, scope=%v\n\n",
		recovered["role"], recovered["tenant"], recovered["scope"])
}
