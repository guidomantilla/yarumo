package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	caead "github.com/guidomantilla/yarumo/common/crypto/ciphers/aead"
	rsassas "github.com/guidomantilla/yarumo/common/crypto/signers/rsassas"
	ctokens "github.com/guidomantilla/yarumo/common/crypto/tokens"
)

// TestRoundtrip_Tokens exercises a cross-process flow for tokens: issue a token
// (JWT or opaque), write the serialized token string to disk, reload it in a
// fresh reader, and validate.
//
// Encoding choices:
//   - Token (HS / RS / opaque): written as UTF-8 text, no wrapping. JWTs are
//     already a compact base64url-encoded string (header.payload.signature);
//     opaque tokens (YA-0019) are a single base64url-encoded ciphertext blob.
//     Both are intrinsically file-safe and need no additional encoding.
//   - HMAC keys / opaque AEAD keys: not persisted in this test. The realistic
//     deployment keeps the verifier key in a secrets manager. We mirror that by
//     keeping the Method (which holds the key) in memory and only persisting
//     the emitted token to disk.
//   - RS256 private/public keys: not persisted here — the signer roundtrip
//     test in `signers/examples` already covers PEM persistence end-to-end.
//     This test focuses specifically on the token-string round-trip.
func TestRoundtrip_Tokens(t *testing.T) {
	t.Parallel()

	t.Run("JWT_HS256", func(t *testing.T) {
		t.Parallel()

		method := ctokens.NewMethod("roundtrip-hs256", ctokens.AlgorithmHS256,
			ctokens.WithGeneratedKey(),
			ctokens.WithIssuer("yarumo-roundtrip"),
			ctokens.WithTimeout(1*time.Hour),
		)
		runTokenRoundtrip(t, method)
	})

	t.Run("JWT_HS384", func(t *testing.T) {
		t.Parallel()

		method := ctokens.NewMethod("roundtrip-hs384", ctokens.AlgorithmHS384,
			ctokens.WithGeneratedKey(),
			ctokens.WithIssuer("yarumo-roundtrip"),
			ctokens.WithTimeout(1*time.Hour),
		)
		runTokenRoundtrip(t, method)
	})

	t.Run("JWT_HS512", func(t *testing.T) {
		t.Parallel()

		method := ctokens.NewMethod("roundtrip-hs512", ctokens.AlgorithmHS512,
			ctokens.WithGeneratedKey(),
			ctokens.WithIssuer("yarumo-roundtrip"),
			ctokens.WithTimeout(1*time.Hour),
		)
		runTokenRoundtrip(t, method)
	})

	t.Run("JWT_RS256", func(t *testing.T) {
		t.Parallel()

		priv, err := rsassas.RSASSA_PKCS1v15_using_SHA256.GenerateKey(2048)
		if err != nil {
			t.Fatalf("RSA GenerateKey: %v", err)
		}

		method := ctokens.NewMethod("roundtrip-rs256", ctokens.AlgorithmRS256,
			ctokens.WithSigningKey(priv),
			ctokens.WithVerifyingKey(&priv.PublicKey),
			ctokens.WithIssuer("yarumo-roundtrip"),
			ctokens.WithTimeout(1*time.Hour),
		)
		runTokenRoundtrip(t, method)
	})

	t.Run("Opaque_AESGCM", func(t *testing.T) {
		t.Parallel()

		key, err := caead.AES_256_GCM.GenerateKey()
		if err != nil {
			t.Fatalf("AEAD GenerateKey: %v", err)
		}

		// opaqueAEADKey expects a plain []byte; ctypes.Bytes is a named type
		// and would fail the key.([]byte) assertion. Convert explicitly.
		method := ctokens.NewMethod("roundtrip-opaque", ctokens.AlgorithmOpaqueAESGCM,
			ctokens.WithKey([]byte(key)),
			ctokens.WithIssuer("yarumo-roundtrip"),
			ctokens.WithTimeout(1*time.Hour),
		)
		runTokenRoundtrip(t, method)
	})
}

// runTokenRoundtrip generates a token with the given method, writes the
// serialized string to disk, reads it back from a fresh handle, and validates.
// The claims are checked end-to-end so any silent corruption during the
// write/read trip would be caught by the assertion.
func runTokenRoundtrip(t *testing.T, method *ctokens.Method) {
	t.Helper()

	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "token.txt")

	subject := "user-roundtrip"
	payload := ctokens.Payload{
		"role":   "auditor",
		"tenant": "yarumo",
		"scope":  "read",
	}

	// Issuer side: generate, persist token string.
	token, err := method.Generate(subject, payload)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	writeErr := os.WriteFile(tokenPath, []byte(token), 0o600)
	if writeErr != nil {
		t.Fatalf("WriteFile: %v", writeErr)
	}

	// Verifier side: read fresh from disk and validate.
	loaded, readErr := os.ReadFile(tokenPath) //nolint:gosec // examples write to t.TempDir paths derived from the test, not user input.
	if readErr != nil {
		t.Fatalf("ReadFile: %v", readErr)
	}

	recovered, err := method.Validate(string(loaded))
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}

	if recovered["role"] != "auditor" {
		t.Fatalf("role mismatch: got %v, want auditor", recovered["role"])
	}

	if recovered["tenant"] != "yarumo" {
		t.Fatalf("tenant mismatch: got %v, want yarumo", recovered["tenant"])
	}

	if recovered["scope"] != "read" {
		t.Fatalf("scope mismatch: got %v, want read", recovered["scope"])
	}
}
