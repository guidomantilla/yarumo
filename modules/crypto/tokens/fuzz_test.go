package tokens

import "testing"

// FuzzValidate exercises Method.Validate with attacker-controlled token strings
// to ensure the parser never panics. JWT parsing dispatches through golang-jwt/v5
// plus this package's wrapping logic (key resolution, issuer match, claim type
// assertion). The contract: any input, regardless of how malformed, must produce
// an error and never a panic.
//
// The fuzz uses a fully-keyed HS256 method so the input actually reaches the
// underlying jwt.ParseWithClaims call. The predefined tokens.JWT_HS256 lacks a
// key and would short-circuit on ErrVerifyingKeyNil before any parsing happens.
func FuzzValidate(f *testing.F) {
	key := []byte("fuzz-hs256-key-with-enough-length-1234567890abcdef")
	method := NewMethod("FUZZ_HS256", AlgorithmHS256, WithKey(key))

	f.Add("malformed.token.string")
	f.Add("a.b.c")
	f.Add("")
	f.Add(".")
	f.Add("..")
	f.Add("...")

	f.Fuzz(func(t *testing.T, tokenString string) {
		t.Parallel()
		_, _ = method.Validate(tokenString)
	})
}

// FuzzDecodeUnsafe exercises Method.DecodeUnsafe which calls jwt.NewParser
// directly without key verification. The unverified path tends to have a
// different attack surface than the verified path: signature stripping,
// alg=none acceptance, oversized claim handling.
func FuzzDecodeUnsafe(f *testing.F) {
	method := NewMethod("FUZZ_DECODE_UNSAFE", AlgorithmHS256)

	f.Add("malformed.token.string")
	f.Add("a.b.c")
	f.Add("")

	f.Fuzz(func(t *testing.T, tokenString string) {
		t.Parallel()
		_, _ = method.DecodeUnsafe(tokenString)
	})
}
