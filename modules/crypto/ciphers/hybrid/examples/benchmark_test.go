package main

import (
	"crypto/rand"
	"testing"

	chybrid "github.com/guidomantilla/yarumo/crypto/ciphers/hybrid"
)

// hybridInputSize is the payload size driven through HPKE Encrypt / Decrypt
// benchmarks. 16 KiB is large enough to exercise the AEAD path without
// dominating suite wall time.
const hybridInputSize = 16 * 1024

// randomBytes returns size bytes drawn from crypto/rand for benchmark inputs.
func randomBytes(b *testing.B, size int) []byte {
	b.Helper()

	buf := make([]byte, size)

	_, err := rand.Read(buf)
	if err != nil {
		b.Fatalf("rand.Read failed: %v", err)
	}

	return buf
}

// BenchmarkEncrypt measures Method.Encrypt for the HPKE suite registered as
// HPKE_X25519_HKDF_SHA256_AES_256_GCM.
func BenchmarkEncrypt(b *testing.B) {
	data := randomBytes(b, hybridInputSize)
	info := []byte("benchmark-info")

	pub, _, err := chybrid.HPKE_X25519_HKDF_SHA256_AES_256_GCM.GenerateKey()
	if err != nil {
		b.Fatalf("GenerateKey failed: %v", err)
	}

	b.ReportAllocs()
	b.SetBytes(int64(len(data)))

	for b.Loop() {
		_, _ = chybrid.HPKE_X25519_HKDF_SHA256_AES_256_GCM.Encrypt(pub, data, info)
	}
}

// BenchmarkDecrypt measures Method.Decrypt for the HPKE suite registered as
// HPKE_X25519_HKDF_SHA256_AES_256_GCM. Each loop iteration decrypts a
// pre-sealed ciphertext bound to the original info.
func BenchmarkDecrypt(b *testing.B) {
	data := randomBytes(b, hybridInputSize)
	info := []byte("benchmark-info")

	pub, priv, err := chybrid.HPKE_X25519_HKDF_SHA256_AES_256_GCM.GenerateKey()
	if err != nil {
		b.Fatalf("GenerateKey failed: %v", err)
	}

	ciphered, err := chybrid.HPKE_X25519_HKDF_SHA256_AES_256_GCM.Encrypt(pub, data, info)
	if err != nil {
		b.Fatalf("Encrypt failed: %v", err)
	}

	b.ReportAllocs()
	b.SetBytes(int64(len(data)))

	for b.Loop() {
		_, _ = chybrid.HPKE_X25519_HKDF_SHA256_AES_256_GCM.Decrypt(priv, ciphered, info)
	}
}
