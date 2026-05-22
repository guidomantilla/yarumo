// Package tokens provides token generation and validation. Two distinct
// flavors are supported:
//
//   - JWT signed tokens via the Algorithm enum, covering HMAC (HS256/384/512),
//     RSASSA-PKCS1-v1_5 (RS256/384/512), RSASSA-PSS (PS256/384/512), ECDSA
//     (ES256/384/512), and Ed25519 (EdDSA). The claims payload is transparent
//     — anyone can base64-decode the body; only the signature protects
//     integrity. Asymmetric variants require a real key pair via
//     WithSigningKey / WithVerifyingKey; the signers/{rsassas, ecdsas, ed25519}
//     subpackages provide GenerateKey helpers and PEM marshal/parse for the
//     key types involved.
//   - Opaque tokens via AEAD encryption (YA-0019). The entire claims payload
//     is encrypted under a symmetric key, so nothing leaks to the client.
//     The token is base64url(AEAD.Encrypt(key, json(claims), nil)) with the
//     AEAD nonce prepended internally by the configured cipher.
//
// Both JWT and opaque methods are constructed with a single entry point —
// NewMethod(name, Algorithm, options...). The Algorithm enum value is the
// discriminator: HMAC and asymmetric variants belong to the JWT family;
// AlgorithmOpaqueAESGCM and AlgorithmOpaqueXChaCha20Poly1305 belong to the
// opaque family. Both flavors share the same Method struct, Generate/Validate
// API, options pipeline, and registry.
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
//
// # Config-driven algorithm selection
//
// *Method implements encoding.TextMarshaler / encoding.TextUnmarshaler.
// MarshalText emits the registered algorithm name; UnmarshalText resolves a
// name against the package registry (via Get) and overwrites the receiver.
// This makes Method directly compatible with libraries that honor the
// encoding interfaces — including encoding/json, viper, kong, and koanf —
// so deployments can load token algorithm choice from YAML/JSON/TOML config.
//
// Caveat: UnmarshalText resolves against whatever the registry contains at
// the time of the call. Custom methods registered via Register after config
// load will not resolve here; callers that need late-bound lookup should
// call Get(name) directly.
//
// # Recommended entry point for string-named algorithms
//
// Generate(name, subject, payload) and Validate(name, tokenString) are the
// recommended top-level helpers for callers that load the algorithm name
// from config. They each perform a single Get and forward to the
// corresponding Method operation. The named Method must already have been
// registered with keys configured (see Key management above): the
// predefined templates carry no key material on purpose.
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
// strings (RFC 7518) where applicable so the enum value can double as the
// registered JWT algorithm identifier. Opaque algorithms use bespoke names
// since they have no JWS analogue.
const (
	AlgorithmHS256 Algorithm = "HS256"
	AlgorithmHS384 Algorithm = "HS384"
	AlgorithmHS512 Algorithm = "HS512"

	AlgorithmRS256 Algorithm = "RS256"
	AlgorithmRS384 Algorithm = "RS384"
	AlgorithmRS512 Algorithm = "RS512"

	AlgorithmPS256 Algorithm = "PS256"
	AlgorithmPS384 Algorithm = "PS384"
	AlgorithmPS512 Algorithm = "PS512"

	AlgorithmES256 Algorithm = "ES256"
	AlgorithmES384 Algorithm = "ES384"
	AlgorithmES512 Algorithm = "ES512"

	AlgorithmEdDSA Algorithm = "EdDSA"

	AlgorithmOpaqueAESGCM            Algorithm = "OPAQUE_AES_GCM"
	AlgorithmOpaqueXChaCha20Poly1305 Algorithm = "OPAQUE_XCHACHA20_POLY1305"
)

// isOpaque reports whether the algorithm belongs to the opaque (AEAD-encrypted)
// family. The JWT family (HS256/384/512 and future asymmetric variants) returns
// false. This is the package-internal discriminator used by generate and
// validate to dispatch between the two implementations.
func (a Algorithm) isOpaque() bool {
	switch a {
	case AlgorithmOpaqueAESGCM, AlgorithmOpaqueXChaCha20Poly1305:
		return true
	default:
		return false
	}
}

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
