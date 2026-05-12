package main

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	crsassas "github.com/guidomantilla/yarumo/common/crypto/signers/rsassas"
)

// rsaInputSize is the payload size driven through Sign / Verify benchmarks.
const rsaInputSize = 1024

// rsaSignMethods is the predefined RSASSA registry used for Sign / Verify
// benchmarks. Each entry pairs a method with its allowed minimum key size so
// pre-generated keys remain valid across PSS and PKCS1v15 variants.
var rsaSignMethods = []struct {
	name    string
	method  *crsassas.Method
	keySize int
}{
	{"PSS_SHA256_2048", crsassas.RSASSA_PSS_using_SHA256, 2048},
	{"PSS_SHA384_2048", crsassas.RSASSA_PSS_using_SHA384, 2048},
	{"PSS_SHA512_3072", crsassas.RSASSA_PSS_using_SHA512, 3072},
	{"PKCS1v15_SHA256_2048", crsassas.RSASSA_PKCS1v15_using_SHA256, 2048},
	{"PKCS1v15_SHA384_2048", crsassas.RSASSA_PKCS1v15_using_SHA384, 2048},
	{"PKCS1v15_SHA512_3072", crsassas.RSASSA_PKCS1v15_using_SHA512, 3072},
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

// BenchmarkGenerateKey measures Method.GenerateKey for the 2048-bit and
// 3072-bit allowed key sizes. 4096-bit generation is intentionally skipped:
// at -benchtime=100ms a single 4096 keygen can dominate the entire suite
// wall time without changing the qualitative picture.
func BenchmarkGenerateKey(b *testing.B) {
	method := crsassas.RSASSA_PKCS1v15_using_SHA256

	b.Run("2048", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			_, _ = method.GenerateKey(2048)
		}
	})

	b.Run("3072", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			_, _ = method.GenerateKey(3072)
		}
	})
}

// BenchmarkSign measures Method.Sign across each predefined padding / hash
// variant using a single pre-generated private key per row.
func BenchmarkSign(b *testing.B) {
	data := randomBytes(b, rsaInputSize)

	for _, m := range rsaSignMethods {
		key, err := m.method.GenerateKey(m.keySize)
		if err != nil {
			b.Fatalf("GenerateKey(%s) failed: %v", m.name, err)
		}

		b.Run(m.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				_, _ = m.method.Sign(key, data)
			}
		})
	}
}

// BenchmarkVerify measures Method.Verify across each predefined padding /
// hash variant using a single pre-generated signature per row.
func BenchmarkVerify(b *testing.B) {
	data := randomBytes(b, rsaInputSize)

	for _, m := range rsaSignMethods {
		key, err := m.method.GenerateKey(m.keySize)
		if err != nil {
			b.Fatalf("GenerateKey(%s) failed: %v", m.name, err)
		}

		sig, err := m.method.Sign(key, data)
		if err != nil {
			b.Fatalf("Sign(%s) failed: %v", m.name, err)
		}

		pub := publicKey(key)

		b.Run(m.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				_, _ = m.method.Verify(pub, sig, data)
			}
		})
	}
}

// publicKey extracts the *rsa.PublicKey from a generated private key for use
// as the verifier-side input.
func publicKey(key *rsa.PrivateKey) *rsa.PublicKey {
	return &key.PublicKey
}
