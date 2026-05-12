package main

import (
	"crypto/rand"
	"testing"

	chashes "github.com/guidomantilla/yarumo/common/crypto/hashes"
)

// benchInput sizes used by the hash benchmark suite.
const (
	hashInputSmall  = 64
	hashInputMedium = 4 * 1024
	hashInputLarge  = 1024 * 1024
)

// hashMethods is the predefined hash algorithm registry exercised by the
// benchmark suite.
var hashMethods = []struct {
	name   string
	method *chashes.Method
}{
	// SHA1 is included for legacy interop coverage only.
	{"SHA1", chashes.SHA1},
	{"SHA224", chashes.SHA224},
	{"SHA256", chashes.SHA256},
	{"SHA384", chashes.SHA384},
	{"SHA512", chashes.SHA512},
	{"SHA3_256", chashes.SHA3_256},
	{"SHA3_384", chashes.SHA3_384},
	{"SHA3_512", chashes.SHA3_512},
	{"BLAKE2b_256", chashes.BLAKE2b_256},
	{"BLAKE2b_512", chashes.BLAKE2b_512},
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

// BenchmarkHash measures Method.Hash across all predefined hash algorithms
// at three representative input sizes (64 B / 4 KiB / 1 MiB).
func BenchmarkHash(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"64B", hashInputSmall},
		{"4KiB", hashInputMedium},
		{"1MiB", hashInputLarge},
	}

	for _, m := range hashMethods {
		for _, s := range sizes {
			data := randomBytes(b, s.size)

			b.Run(m.name+"/"+s.name, func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(s.size))

				for b.Loop() {
					_, _ = m.method.Hash(data)
				}
			})
		}
	}
}
