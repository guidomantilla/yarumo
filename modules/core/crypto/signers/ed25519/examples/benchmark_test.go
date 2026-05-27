package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	ced25519 "github.com/guidomantilla/yarumo/core/crypto/signers/ed25519"
)

// ed25519InputSize is the payload size driven through Sign / Verify benchmarks.
const ed25519InputSize = 1024

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

// BenchmarkGenerateKey measures Ed25519 GenerateKey throughput.
func BenchmarkGenerateKey(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		_, _ = ced25519.Ed25519.GenerateKey()
	}
}

// BenchmarkSign measures Ed25519 Sign throughput at a 1 KiB payload.
func BenchmarkSign(b *testing.B) {
	data := randomBytes(b, ed25519InputSize)

	key, err := ced25519.Ed25519.GenerateKey()
	if err != nil {
		b.Fatalf("GenerateKey failed: %v", err)
	}

	b.ReportAllocs()

	for b.Loop() {
		_, _ = ced25519.Ed25519.Sign(&key, data)
	}
}

// BenchmarkVerify measures Ed25519 Verify throughput at a 1 KiB payload.
func BenchmarkVerify(b *testing.B) {
	data := randomBytes(b, ed25519InputSize)

	key, err := ced25519.Ed25519.GenerateKey()
	if err != nil {
		b.Fatalf("GenerateKey failed: %v", err)
	}

	sig, err := ced25519.Ed25519.Sign(&key, data)
	if err != nil {
		b.Fatalf("Sign failed: %v", err)
	}

	pub, ok := key.Public().(ed25519.PublicKey)
	if !ok {
		b.Fatal("ed25519 public key type assertion failed")
	}

	b.ReportAllocs()

	for b.Loop() {
		_, _ = ced25519.Ed25519.Verify(&pub, sig, data)
	}
}
