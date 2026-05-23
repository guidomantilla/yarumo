package main

import (
	"testing"

	cpasswords "github.com/guidomantilla/yarumo/crypto/passwords"
)

// rawPassword is the input password used by every password benchmark.
const rawPassword = "benchmark-password-2026"

// passwordMethods is the predefined password registry exercised by the
// benchmark suite. Each method runs at its default parameter set.
var passwordMethods = []struct {
	name   string
	method *cpasswords.Method
}{
	{"Argon2id", cpasswords.Argon2id},
	{"Argon2i", cpasswords.Argon2i},
	{"Bcrypt", cpasswords.Bcrypt},
	{"Pbkdf2", cpasswords.Pbkdf2},
	{"Scrypt", cpasswords.Scrypt},
}

// BenchmarkEncode measures Method.Encode for each predefined password
// algorithm at its default parameter set. Default parameters are tuned for
// human-perceptible login latency (~100 ms), so a single Encode loop body
// is the dominant cost — there is no point pre-computing inputs.
func BenchmarkEncode(b *testing.B) {
	for _, m := range passwordMethods {
		b.Run(m.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				_, _ = m.method.Encode(rawPassword)
			}
		})
	}
}

// BenchmarkVerify measures Method.Verify for each predefined password
// algorithm using a single pre-encoded hash per row.
func BenchmarkVerify(b *testing.B) {
	for _, m := range passwordMethods {
		encoded, err := m.method.Encode(rawPassword)
		if err != nil {
			b.Fatalf("Encode(%s) failed: %v", m.name, err)
		}

		b.Run(m.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				_, _ = m.method.Verify(encoded, rawPassword)
			}
		})
	}
}
