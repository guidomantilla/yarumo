package passwords

import (
	"testing"
)

// BenchmarkBcryptEncode_DefaultCost measures Method.Encode latency for the
// predefined Bcrypt method at the OWASP-2024-aligned default cost (12).
func BenchmarkBcryptEncode_DefaultCost(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		_, err := Bcrypt.Encode("bench-password")
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkScryptEncode_DefaultN measures Method.Encode latency for the
// predefined Scrypt method at the OWASP-2024-aligned default N (2^17 = 131072).
func BenchmarkScryptEncode_DefaultN(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		_, err := Scrypt.Encode("bench-password")
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkArgon2Encode_DefaultParams measures Method.Encode latency for the
// predefined Argon2 method at default parameters (t=1, m=64 MiB, p=2).
func BenchmarkArgon2Encode_DefaultParams(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		_, err := Argon2.Encode("bench-password")
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkPbkdf2Encode_DefaultIterations measures Method.Encode latency for the
// predefined Pbkdf2 method at OWASP-2024 default iterations (600,000) with SHA-512.
func BenchmarkPbkdf2Encode_DefaultIterations(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		_, err := Pbkdf2.Encode("bench-password")
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
