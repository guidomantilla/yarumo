package main

import (
	"crypto/rand"
	"testing"

	caead "github.com/guidomantilla/yarumo/crypto/ciphers/aead"
)

// AEAD benchmark input sizes spanning a header-frame (1 KiB), a typical TLS
// record (64 KiB), and a bulk-transfer chunk (1 MiB).
const (
	aeadInputSmall  = 1024
	aeadInputMedium = 64 * 1024
	aeadInputLarge  = 1024 * 1024
)

// aeadMethods is the predefined AEAD registry exercised by the benchmark suite.
var aeadMethods = []struct {
	name   string
	method *caead.Method
}{
	{"AES_128_GCM", caead.AES_128_GCM},
	{"AES_256_GCM", caead.AES_256_GCM},
	{"ChaCha20_Poly1305", caead.CHACHA20_POLY1305},
	{"XChaCha20_Poly1305", caead.XCHACHA20_POLY1305},
}

// aeadSizes are the per-row input sizes shared by Encrypt and Decrypt benches.
var aeadSizes = []struct {
	name string
	size int
}{
	{"1KiB", aeadInputSmall},
	{"64KiB", aeadInputMedium},
	{"1MiB", aeadInputLarge},
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

// BenchmarkEncrypt measures Method.Encrypt across all predefined AEAD
// algorithms at 1 KiB / 64 KiB / 1 MiB plaintext sizes.
func BenchmarkEncrypt(b *testing.B) {
	aad := []byte("benchmark-aad")

	for _, m := range aeadMethods {
		key, err := m.method.GenerateKey()
		if err != nil {
			b.Fatalf("GenerateKey(%s) failed: %v", m.name, err)
		}

		for _, s := range aeadSizes {
			data := randomBytes(b, s.size)

			b.Run(m.name+"/"+s.name, func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(s.size))

				for b.Loop() {
					_, _ = m.method.Encrypt(key, data, aad)
				}
			})
		}
	}
}

// BenchmarkDecrypt measures Method.Decrypt across all predefined AEAD
// algorithms at 1 KiB / 64 KiB / 1 MiB plaintext sizes.
func BenchmarkDecrypt(b *testing.B) {
	aad := []byte("benchmark-aad")

	for _, m := range aeadMethods {
		key, err := m.method.GenerateKey()
		if err != nil {
			b.Fatalf("GenerateKey(%s) failed: %v", m.name, err)
		}

		for _, s := range aeadSizes {
			data := randomBytes(b, s.size)

			ciphered, err := m.method.Encrypt(key, data, aad)
			if err != nil {
				b.Fatalf("Encrypt(%s/%s) failed: %v", m.name, s.name, err)
			}

			b.Run(m.name+"/"+s.name, func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(s.size))

				for b.Loop() {
					_, _ = m.method.Decrypt(key, ciphered, aad)
				}
			})
		}
	}
}
