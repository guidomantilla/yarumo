package main

import (
	"crypto/rand"
	"testing"

	chmacs "github.com/guidomantilla/yarumo/crypto/signers/hmacs"
)

// hmacInputSize is the payload size driven through HMAC Digest / Validate
// benchmarks. 1 KiB is representative of authentication tags computed over
// API request bodies.
const hmacInputSize = 1024

// hmacMethods is the predefined HMAC registry exercised by the benchmark suite.
var hmacMethods = []struct {
	name   string
	method *chmacs.Method
}{
	{"HMAC_SHA256", chmacs.HMAC_with_SHA256},
	{"HMAC_SHA384", chmacs.HMAC_with_SHA384},
	{"HMAC_SHA512", chmacs.HMAC_with_SHA512},
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

// BenchmarkDigest measures Method.Digest across each predefined HMAC method.
func BenchmarkDigest(b *testing.B) {
	data := randomBytes(b, hmacInputSize)

	for _, m := range hmacMethods {
		key, err := m.method.GenerateKey()
		if err != nil {
			b.Fatalf("GenerateKey(%s) failed: %v", m.name, err)
		}

		b.Run(m.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))

			for b.Loop() {
				_, _ = m.method.Digest(key, data)
			}
		})
	}
}

// BenchmarkValidate measures Method.Validate across each predefined HMAC method.
func BenchmarkValidate(b *testing.B) {
	data := randomBytes(b, hmacInputSize)

	for _, m := range hmacMethods {
		key, err := m.method.GenerateKey()
		if err != nil {
			b.Fatalf("GenerateKey(%s) failed: %v", m.name, err)
		}

		digest, err := m.method.Digest(key, data)
		if err != nil {
			b.Fatalf("Digest(%s) failed: %v", m.name, err)
		}

		b.Run(m.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))

			for b.Loop() {
				_, _ = m.method.Validate(key, digest, data)
			}
		})
	}
}
