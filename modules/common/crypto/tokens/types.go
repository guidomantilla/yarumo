// Package tokens provides JWT token generation and validation using HMAC
// signing methods.
//
// # Key management
//
// As of YA-0008, NewOptions no longer pre-generates a random signing/verifying
// key at construction time. The default Options has nil keys, and callers must
// choose explicitly between three paths:
//
//   - Bring your own key: NewMethod(name, alg, WithKey(secret)), or
//     NewMethod(name, alg, WithSigningKey(s), WithVerifyingKey(v)).
//   - Mint a random key: NewMethod(name, alg, WithGeneratedKey()). The 64-byte
//     entropy draw happens at option-apply time, never at package init.
//   - Use the predefined JWT_HS256 / JWT_HS384 / JWT_HS512 templates only as
//     algorithm anchors; they carry no key. To use them, copy via Register/Get
//     with a key, or just call NewMethod directly with one of the options
//     above.
//
// If Method.Generate is called without a signing key it returns
// ErrSigningKeyNil; Method.Validate without a verifying key returns
// ErrVerifyingKeyNil. There is no implicit lazy generation on first use —
// explicit beats magic, and construction stays free of side effects on the
// runtime entropy pool.
//
// Migration note (YA-0008): callers that relied on the previous auto-generated
// key must add WithGeneratedKey() to their NewMethod / NewOptions call.
// Callers already passing WithKey, WithSigningKey, or WithVerifyingKey are
// unaffected.
//
// # Algorithm selection
//
// As of YA-0009, NewMethod takes a tokens.Algorithm enum value instead of a
// jwt.SigningMethod. This stops the public API from leaking the underlying
// golang-jwt/v5 type so future opaque or AEAD methods can reuse the same enum
// without binding callers to a third-party signing-method interface.
//
// Migration note (YA-0009): replace jwt.SigningMethod arguments with the
// matching Algorithm constant.
//
//	// Before:
//	import jwt "github.com/golang-jwt/jwt/v5"
//	m := tokens.NewMethod("app", jwt.SigningMethodHS256, tokens.WithKey(key))
//
//	// After:
//	m := tokens.NewMethod("app", tokens.AlgorithmHS256, tokens.WithKey(key))
//
// The previously re-exported tokens.SigningMethodHS256/384/512 vars are
// removed; use AlgorithmHS256/384/512 instead. Passing an unknown Algorithm
// causes NewMethod to panic via the package's assertion path, surfacing
// ErrAlgorithmInvalid as the underlying cause.
package tokens

import (
	jwt "github.com/golang-jwt/jwt/v5"
)

var (
	_ GenerateFn = generate
	_ ValidateFn = validate
)

// Algorithm names a signing algorithm without leaking jwt.SigningMethod
// through the public API. Future opaque/AEAD methods reuse this enum even
// though they don't map to a jwt.SigningMethod.
type Algorithm string

// Supported Algorithm values. Keep names aligned with the JWS "alg" header
// strings (RFC 7518) so the enum value can double as the registered JWT
// algorithm identifier.
const (
	AlgorithmHS256 Algorithm = "HS256"
	AlgorithmHS384 Algorithm = "HS384"
	AlgorithmHS512 Algorithm = "HS512"
)

// Payload is a named type for token claims payload data.
type Payload map[string]any

// Claims extends JWT registered claims with a custom payload.
type Claims struct {
	jwt.RegisteredClaims

	Payload Payload `json:"payload,omitempty"`
}

// GenerateFn is the function type for generating a token.
type GenerateFn func(method *Method, subject string, payload Payload) (string, error)

// ValidateFn is the function type for validating a token and extracting its payload.
type ValidateFn func(method *Method, tokenString string) (Payload, error)
