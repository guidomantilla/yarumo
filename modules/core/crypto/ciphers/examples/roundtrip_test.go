package main

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	caead "github.com/guidomantilla/yarumo/core/crypto/ciphers/aead"
	crsaoaep "github.com/guidomantilla/yarumo/core/crypto/ciphers/rsaoaep"
	ctypes "github.com/guidomantilla/yarumo/core/common/types"
)

// TestRoundtrip_Ciphers exercises a cross-process flow for every cipher
// package: encrypt with one process, persist the ciphertext (and key material)
// to disk, then reopen and decrypt in a "fresh" reader.
//
// Encoding choices:
//   - AEAD ciphertext: raw bytes. The Encrypt output already embeds nonce ||
//     tag || ciphertext, so it round-trips through `os.WriteFile` / `os.ReadFile`
//     without any further framing. Writing raw avoids inflating the file size
//     and matches how an `at-rest` storage layer would persist the blob.
//   - AEAD symmetric key: hex (lowercase). Raw symmetric keys must travel
//     out-of-band; hex was chosen because it is the canonical encoding for
//     short binary secrets and is greppable in audit logs.
//   - AEAD additional authenticated data (AAD): raw bytes. Same rationale as
//     the ciphertext.
//   - RSA-OAEP keys: PEM (PKCS#8 for private, PKIX/SPKI for public). PEM is
//     the universal interchange format for asymmetric keys.
//   - RSA-OAEP ciphertext: raw bytes. Same rationale as AEAD ciphertext.
func TestRoundtrip_Ciphers(t *testing.T) {
	t.Parallel()

	t.Run("AES_128_GCM", func(t *testing.T) {
		t.Parallel()
		runAeadRoundtrip(t, caead.AES_128_GCM)
	})

	t.Run("AES_256_GCM", func(t *testing.T) {
		t.Parallel()
		runAeadRoundtrip(t, caead.AES_256_GCM)
	})

	t.Run("ChaCha20_Poly1305", func(t *testing.T) {
		t.Parallel()
		runAeadRoundtrip(t, caead.CHACHA20_POLY1305)
	})

	t.Run("XChaCha20_Poly1305", func(t *testing.T) {
		t.Parallel()
		runAeadRoundtrip(t, caead.XCHACHA20_POLY1305)
	})

	t.Run("RSA_OAEP_SHA256", func(t *testing.T) {
		t.Parallel()
		runRsaOaepRoundtrip(t, crsaoaep.RSA_OAEP_SHA256, 2048)
	})

	t.Run("RSA_OAEP_SHA384", func(t *testing.T) {
		t.Parallel()
		runRsaOaepRoundtrip(t, crsaoaep.RSA_OAEP_SHA384, 3072)
	})

	t.Run("RSA_OAEP_SHA512", func(t *testing.T) {
		t.Parallel()
		runRsaOaepRoundtrip(t, crsaoaep.RSA_OAEP_SHA512, 3072)
	})
}

// runAeadRoundtrip writes the symmetric key (hex), the AAD (raw) and the
// ciphertext (raw) to disk, then reads them back in a fresh reader and
// decrypts. Plaintext recovery is asserted byte-for-byte.
func runAeadRoundtrip(t *testing.T, method *caead.Method) {
	t.Helper()

	dir := t.TempDir()
	plaintext := ctypes.Bytes("aead-roundtrip-plaintext-payload-with-some-length")
	aad := ctypes.Bytes("on-disk-aead-context")

	// Encryptor side: generate key, encrypt, persist ciphertext + key.
	key, err := method.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	ciphertext, err := method.Encrypt(key, plaintext, aad)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	keyPath := filepath.Join(dir, "key.hex")
	aadPath := filepath.Join(dir, "aad.bin")
	cipherPath := filepath.Join(dir, "ciphertext.bin")

	writeFile(t, keyPath, []byte(hex.EncodeToString(key)))
	writeFile(t, aadPath, aad)
	writeFile(t, cipherPath, ciphertext)

	// Decryptor side: reload everything fresh from disk.
	keyHex := readFile(t, keyPath)

	loadedKey, err := hex.DecodeString(string(keyHex))
	if err != nil {
		t.Fatalf("hex.DecodeString: %v", err)
	}

	loadedAAD := readFile(t, aadPath)
	loadedCipher := readFile(t, cipherPath)

	recovered, err := method.Decrypt(loadedKey, loadedCipher, loadedAAD)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(recovered, plaintext) {
		t.Fatalf("plaintext mismatch: got %q, want %q", string(recovered), string(plaintext))
	}
}

// runRsaOaepRoundtrip writes the public key PEM, encrypts with it, persists
// the ciphertext to disk, then reloads the private key PEM and decrypts. The
// private key is intentionally only read on the decryptor side to model a
// production split where the encryptor never holds the private half.
func runRsaOaepRoundtrip(t *testing.T, method *crsaoaep.Method, keySize int) {
	t.Helper()

	dir := t.TempDir()
	plaintext := ctypes.Bytes("rsa-oaep-roundtrip-payload")
	label := ctypes.Bytes("on-disk-rsaoaep-context")

	// Out-of-band setup: produce a key pair and write both PEMs to disk.
	priv, err := method.GenerateKey(keySize)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	privPEM, err := crsaoaep.MarshalPrivateKeyPEM(priv)
	if err != nil {
		t.Fatalf("MarshalPrivateKeyPEM: %v", err)
	}

	pubPEM, err := crsaoaep.MarshalPublicKeyPEM(&priv.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPublicKeyPEM: %v", err)
	}

	privPath := filepath.Join(dir, "rsa-priv.pem")
	pubPath := filepath.Join(dir, "rsa-pub.pem")
	labelPath := filepath.Join(dir, "label.bin")
	cipherPath := filepath.Join(dir, "ciphertext.bin")

	writeFile(t, privPath, privPEM)
	writeFile(t, pubPath, pubPEM)
	writeFile(t, labelPath, label)

	// Encryptor side: only has the public key PEM on disk.
	loadedPub, err := crsaoaep.ParsePublicKeyPEM(readFile(t, pubPath))
	if err != nil {
		t.Fatalf("ParsePublicKeyPEM: %v", err)
	}

	ciphertext, err := method.Encrypt(loadedPub, plaintext, readFile(t, labelPath))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	writeFile(t, cipherPath, ciphertext)

	// Decryptor side: only has the private key PEM on disk; reloads everything.
	loadedPriv, err := crsaoaep.ParsePrivateKeyPEM(readFile(t, privPath))
	if err != nil {
		t.Fatalf("ParsePrivateKeyPEM: %v", err)
	}

	recovered, err := method.Decrypt(loadedPriv, readFile(t, cipherPath), readFile(t, labelPath))
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(recovered, plaintext) {
		t.Fatalf("plaintext mismatch: got %q, want %q", string(recovered), string(plaintext))
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
