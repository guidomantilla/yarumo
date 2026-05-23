package main

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	crsaoaep "github.com/guidomantilla/yarumo/crypto/ciphers/rsaoaep"
)

// rsaoaepInputSize is the payload size driven through Encrypt / Decrypt
// benchmarks. RSA-OAEP imposes a strict ciphertext size ceiling: at 2048
// bits with SHA-256 the maximum plaintext is roughly 190 bytes. 128 bytes
// stays well inside that limit for every predefined combination.
const rsaoaepInputSize = 128

// rsaoaepMethods is the predefined RSA-OAEP registry exercised by Encrypt /
// Decrypt benchmarks. Each method is paired with its smallest allowed key
// size so suite wall time stays bounded.
var rsaoaepMethods = []struct {
	name    string
	method  *crsaoaep.Method
	keySize int
}{
	{"SHA256_2048", crsaoaep.RSA_OAEP_SHA256, 2048},
	{"SHA384_3072", crsaoaep.RSA_OAEP_SHA384, 3072},
	{"SHA512_3072", crsaoaep.RSA_OAEP_SHA512, 3072},
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

// BenchmarkEncrypt measures Method.Encrypt across the predefined RSA-OAEP
// variants using pre-generated keys.
func BenchmarkEncrypt(b *testing.B) {
	data := randomBytes(b, rsaoaepInputSize)
	label := []byte("benchmark-label")

	for _, m := range rsaoaepMethods {
		key, err := m.method.GenerateKey(m.keySize)
		if err != nil {
			b.Fatalf("GenerateKey(%s) failed: %v", m.name, err)
		}

		pub := publicKey(key)

		b.Run(m.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				_, _ = m.method.Encrypt(pub, data, label)
			}
		})
	}
}

// BenchmarkDecrypt measures Method.Decrypt across the predefined RSA-OAEP
// variants using pre-generated ciphertexts.
func BenchmarkDecrypt(b *testing.B) {
	data := randomBytes(b, rsaoaepInputSize)
	label := []byte("benchmark-label")

	for _, m := range rsaoaepMethods {
		key, err := m.method.GenerateKey(m.keySize)
		if err != nil {
			b.Fatalf("GenerateKey(%s) failed: %v", m.name, err)
		}

		ciphered, err := m.method.Encrypt(&key.PublicKey, data, label)
		if err != nil {
			b.Fatalf("Encrypt(%s) failed: %v", m.name, err)
		}

		b.Run(m.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				_, _ = m.method.Decrypt(key, ciphered, label)
			}
		})
	}
}

// publicKey extracts the *rsa.PublicKey from a generated private key.
func publicKey(key *rsa.PrivateKey) *rsa.PublicKey {
	return &key.PublicKey
}
