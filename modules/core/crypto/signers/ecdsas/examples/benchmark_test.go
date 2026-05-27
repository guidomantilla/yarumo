package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"testing"

	cecdsas "github.com/guidomantilla/yarumo/core/crypto/signers/ecdsas"
)

// ecdsaInputSize is the payload size driven through Sign / Verify benchmarks.
const ecdsaInputSize = 1024

// ecdsaMethods is the predefined ECDSA registry exercised by the benchmark suite.
var ecdsaMethods = []struct {
	name   string
	method *cecdsas.Method
}{
	{"P256_SHA256", cecdsas.ECDSA_with_SHA256_over_P256},
	{"P384_SHA384", cecdsas.ECDSA_with_SHA384_over_P384},
	{"P521_SHA512", cecdsas.ECDSA_with_SHA512_over_P521},
}

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

// BenchmarkGenerateKey measures Method.GenerateKey across each predefined curve.
func BenchmarkGenerateKey(b *testing.B) {
	for _, m := range ecdsaMethods {
		b.Run(m.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				_, _ = m.method.GenerateKey()
			}
		})
	}
}

// BenchmarkSign measures Method.Sign across each predefined curve in ASN1 format.
func BenchmarkSign(b *testing.B) {
	data := randomBytes(b, ecdsaInputSize)

	for _, m := range ecdsaMethods {
		key, err := m.method.GenerateKey()
		if err != nil {
			b.Fatalf("GenerateKey(%s) failed: %v", m.name, err)
		}

		b.Run(m.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				_, _ = m.method.Sign(key, data, cecdsas.ASN1)
			}
		})
	}
}

// BenchmarkVerify measures Method.Verify across each predefined curve in ASN1 format.
func BenchmarkVerify(b *testing.B) {
	data := randomBytes(b, ecdsaInputSize)

	for _, m := range ecdsaMethods {
		key, err := m.method.GenerateKey()
		if err != nil {
			b.Fatalf("GenerateKey(%s) failed: %v", m.name, err)
		}

		sig, err := m.method.Sign(key, data, cecdsas.ASN1)
		if err != nil {
			b.Fatalf("Sign(%s) failed: %v", m.name, err)
		}

		pub := publicKey(key)

		b.Run(m.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				_, _ = m.method.Verify(pub, sig, data, cecdsas.ASN1)
			}
		})
	}
}

// publicKey extracts the *ecdsa.PublicKey from a generated private key for
// use as the verifier-side input.
func publicKey(key *ecdsa.PrivateKey) *ecdsa.PublicKey {
	return &key.PublicKey
}
