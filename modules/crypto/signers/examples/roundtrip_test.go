package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	cecdsas "github.com/guidomantilla/yarumo/crypto/signers/ecdsas"
	ced25519 "github.com/guidomantilla/yarumo/crypto/signers/ed25519"
	crsassas "github.com/guidomantilla/yarumo/crypto/signers/rsassas"
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// TestRoundtrip_Signers exercises a cross-process flow for every asymmetric
// signer package: generate key, marshal to PEM, persist to disk, reload from
// disk, sign a payload, write the signature out (hex-encoded), and verify it
// using only the on-disk artefacts.
//
// Encoding choices:
//   - Private/public keys: PEM (PKCS#8 for private, PKIX/SPKI for public). PEM
//     is the universal interchange format for asymmetric keys — it is what
//     secret managers, cert tooling, and TLS stacks all consume.
//   - Signatures: hex (lowercase). Signatures are raw bytes; we want a
//     newline-safe text encoding so the on-disk artefact is greppable and
//     diff-able. Hex was chosen over base64 because it is unambiguous and the
//     payloads here are tiny (no size pressure).
//   - Payload: raw bytes. The payload is the canonical message and gets no
//     transformation — signers operate on bytes, not encodings.
func TestRoundtrip_Signers(t *testing.T) {
	t.Parallel()

	t.Run("ECDSA_P256_ASN1", func(t *testing.T) {
		t.Parallel()
		runEcdsaRoundtrip(t, cecdsas.ECDSA_with_SHA256_over_P256, cecdsas.ASN1)
	})

	t.Run("ECDSA_P384_ASN1", func(t *testing.T) {
		t.Parallel()
		runEcdsaRoundtrip(t, cecdsas.ECDSA_with_SHA384_over_P384, cecdsas.ASN1)
	})

	t.Run("ECDSA_P521_RS", func(t *testing.T) {
		t.Parallel()
		runEcdsaRoundtrip(t, cecdsas.ECDSA_with_SHA512_over_P521, cecdsas.RS)
	})

	t.Run("Ed25519", func(t *testing.T) {
		t.Parallel()
		runEd25519Roundtrip(t)
	})

	t.Run("RSASSA_PSS_SHA256", func(t *testing.T) {
		t.Parallel()
		runRsassaRoundtrip(t, crsassas.RSASSA_PSS_using_SHA256)
	})

	t.Run("RSASSA_PKCS1v15_SHA256", func(t *testing.T) {
		t.Parallel()
		runRsassaRoundtrip(t, crsassas.RSASSA_PKCS1v15_using_SHA256)
	})
}

// runEcdsaRoundtrip writes the key pair as PEM, the payload as raw bytes, and
// the signature as hex; then reloads everything from disk and verifies.
func runEcdsaRoundtrip(t *testing.T, method *cecdsas.Method, format cecdsas.Format) {
	t.Helper()

	dir := t.TempDir()
	payload := ctypes.Bytes("ecdsa-roundtrip-payload")

	// Issuer side: generate, marshal, persist.
	priv, err := method.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	privPEM, err := cecdsas.MarshalPrivateKeyPEM(priv)
	if err != nil {
		t.Fatalf("MarshalPrivateKeyPEM: %v", err)
	}

	pubPEM, err := cecdsas.MarshalPublicKeyPEM(&priv.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPublicKeyPEM: %v", err)
	}

	privPath := filepath.Join(dir, "ecdsa-priv.pem")
	pubPath := filepath.Join(dir, "ecdsa-pub.pem")
	payloadPath := filepath.Join(dir, "payload.bin")
	sigPath := filepath.Join(dir, "signature.hex")

	writeFile(t, privPath, privPEM)
	writeFile(t, pubPath, pubPEM)
	writeFile(t, payloadPath, payload)

	// Reload private key from disk, sign, persist signature.
	privBytes := readFile(t, privPath)

	loadedPriv, err := cecdsas.ParsePrivateKeyPEM(privBytes)
	if err != nil {
		t.Fatalf("ParsePrivateKeyPEM: %v", err)
	}

	loadedPayload := readFile(t, payloadPath)

	sig, err := method.Sign(loadedPriv, loadedPayload, format)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	writeFile(t, sigPath, []byte(hex.EncodeToString(sig)))

	// Verifier side: reload public key, payload and signature from disk.
	pubBytes := readFile(t, pubPath)

	loadedPub, err := cecdsas.ParsePublicKeyPEM(pubBytes)
	if err != nil {
		t.Fatalf("ParsePublicKeyPEM: %v", err)
	}

	verifyPayload := readFile(t, payloadPath)

	sigHex := readFile(t, sigPath)

	verifySig, err := hex.DecodeString(string(sigHex))
	if err != nil {
		t.Fatalf("hex.DecodeString: %v", err)
	}

	ok, err := method.Verify(loadedPub, verifySig, verifyPayload, format)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}

	if !ok {
		t.Fatal("Verify returned false after on-disk round-trip")
	}
}

// runEd25519Roundtrip writes the key pair as PEM, the payload as raw bytes, and
// the signature as hex; then reloads everything from disk and verifies.
func runEd25519Roundtrip(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	payload := ctypes.Bytes("ed25519-roundtrip-payload")

	priv, err := ced25519.Ed25519.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	pub, ok := priv.Public().(ed25519.PublicKey)
	if !ok {
		t.Fatal("public key type assertion failed")
	}

	privPEM, err := ced25519.MarshalPrivateKeyPEM(priv)
	if err != nil {
		t.Fatalf("MarshalPrivateKeyPEM: %v", err)
	}

	pubPEM, err := ced25519.MarshalPublicKeyPEM(pub)
	if err != nil {
		t.Fatalf("MarshalPublicKeyPEM: %v", err)
	}

	privPath := filepath.Join(dir, "ed25519-priv.pem")
	pubPath := filepath.Join(dir, "ed25519-pub.pem")
	payloadPath := filepath.Join(dir, "payload.bin")
	sigPath := filepath.Join(dir, "signature.hex")

	writeFile(t, privPath, privPEM)
	writeFile(t, pubPath, pubPEM)
	writeFile(t, payloadPath, payload)

	// Reload, sign, persist signature.
	loadedPriv, err := ced25519.ParsePrivateKeyPEM(readFile(t, privPath))
	if err != nil {
		t.Fatalf("ParsePrivateKeyPEM: %v", err)
	}

	sig, err := ced25519.Ed25519.Sign(&loadedPriv, readFile(t, payloadPath))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	writeFile(t, sigPath, []byte(hex.EncodeToString(sig)))

	// Verifier side.
	loadedPub, err := ced25519.ParsePublicKeyPEM(readFile(t, pubPath))
	if err != nil {
		t.Fatalf("ParsePublicKeyPEM: %v", err)
	}

	verifySig, err := hex.DecodeString(string(readFile(t, sigPath)))
	if err != nil {
		t.Fatalf("hex.DecodeString: %v", err)
	}

	verified, err := ced25519.Ed25519.Verify(&loadedPub, verifySig, readFile(t, payloadPath))
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}

	if !verified {
		t.Fatal("Verify returned false after on-disk round-trip")
	}
}

// runRsassaRoundtrip writes the key pair as PEM, the payload as raw bytes, and
// the signature as hex; then reloads everything from disk and verifies.
func runRsassaRoundtrip(t *testing.T, method *crsassas.Method) {
	t.Helper()

	dir := t.TempDir()
	payload := ctypes.Bytes("rsassa-roundtrip-payload")

	priv, err := method.GenerateKey(2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	privPEM, err := crsassas.MarshalPrivateKeyPEM(priv)
	if err != nil {
		t.Fatalf("MarshalPrivateKeyPEM: %v", err)
	}

	pubPEM, err := crsassas.MarshalPublicKeyPEM(&priv.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPublicKeyPEM: %v", err)
	}

	privPath := filepath.Join(dir, "rsa-priv.pem")
	pubPath := filepath.Join(dir, "rsa-pub.pem")
	payloadPath := filepath.Join(dir, "payload.bin")
	sigPath := filepath.Join(dir, "signature.hex")

	writeFile(t, privPath, privPEM)
	writeFile(t, pubPath, pubPEM)
	writeFile(t, payloadPath, payload)

	// Reload, sign, persist.
	loadedPriv, err := crsassas.ParsePrivateKeyPEM(readFile(t, privPath))
	if err != nil {
		t.Fatalf("ParsePrivateKeyPEM: %v", err)
	}

	sig, err := method.Sign(loadedPriv, readFile(t, payloadPath))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	writeFile(t, sigPath, []byte(hex.EncodeToString(sig)))

	// Verifier side.
	loadedPub, err := crsassas.ParsePublicKeyPEM(readFile(t, pubPath))
	if err != nil {
		t.Fatalf("ParsePublicKeyPEM: %v", err)
	}

	verifySig, err := hex.DecodeString(string(readFile(t, sigPath)))
	if err != nil {
		t.Fatalf("hex.DecodeString: %v", err)
	}

	ok, err := method.Verify(loadedPub, verifySig, readFile(t, payloadPath))
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}

	if !ok {
		t.Fatal("Verify returned false after on-disk round-trip")
	}
}

// writeFile is a tiny helper that writes data with 0o600 perms and fails the
// test on error.
func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()

	err := os.WriteFile(path, data, 0o600)
	if err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
}

// readFile is a tiny helper that reads a file and fails the test on error.
func readFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path) //nolint:gosec // examples write to t.TempDir paths derived from the test, not user input.
	if err != nil {
		t.Fatalf("ReadFile %s: %v", path, err)
	}

	return data
}
