package tokens

import (
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	caead "github.com/guidomantilla/yarumo/crypto/ciphers/aead"
)

// Predefined methods with default parameters.
//
// Each entry is a template only: it carries the registered name and the
// chosen Algorithm but no key material. Callers must add a key at
// construction time via WithKey / WithSigningKey / WithVerifyingKey /
// WithGeneratedKey (HMAC variants only — asymmetric methods always need
// an explicit key pair generated outside this package).
var (
	JWT_HS256 = NewMethod("JWT_HS256", AlgorithmHS256)
	JWT_HS384 = NewMethod("JWT_HS384", AlgorithmHS384)
	JWT_HS512 = NewMethod("JWT_HS512", AlgorithmHS512)

	JWT_RS256 = NewMethod("JWT_RS256", AlgorithmRS256)
	JWT_RS384 = NewMethod("JWT_RS384", AlgorithmRS384)
	JWT_RS512 = NewMethod("JWT_RS512", AlgorithmRS512)

	JWT_PS256 = NewMethod("JWT_PS256", AlgorithmPS256)
	JWT_PS384 = NewMethod("JWT_PS384", AlgorithmPS384)
	JWT_PS512 = NewMethod("JWT_PS512", AlgorithmPS512)

	JWT_ES256 = NewMethod("JWT_ES256", AlgorithmES256)
	JWT_ES384 = NewMethod("JWT_ES384", AlgorithmES384)
	JWT_ES512 = NewMethod("JWT_ES512", AlgorithmES512)

	JWT_EdDSA = NewMethod("JWT_EdDSA", AlgorithmEdDSA)

	OPAQUE_AES_256_GCM        = NewMethod("OPAQUE_AES_256_GCM", AlgorithmOpaqueAESGCM)
	OPAQUE_XCHACHA20_POLY1305 = NewMethod("OPAQUE_XCHACHA20_POLY1305", AlgorithmOpaqueXChaCha20Poly1305)
)

// Method represents a token signing or encryption configuration. The
// algorithm field is the single source of truth for the flavor:
//
//   - algorithm.isOpaque() == false →  JWT signed token. signingMethod carries
//     the jwt.SigningMethod used by golang-jwt/v5; cipher is nil.
//   - algorithm.isOpaque() == true  →  Opaque AEAD-encrypted token. cipher
//     carries the *caead.Method used to seal/open the claims envelope;
//     signingMethod is nil. signingKey doubles as the AEAD symmetric key.
//
// signingMethod and cipher are pre-computed at construction by signingMethodFor
// and cipherFor; callers do not set them directly.
//
// signingKey and verifyingKey are stored as any so a single Method type can
// carry the byte-slice secrets used by HMAC and opaque-AEAD variants alongside
// the *rsa.PrivateKey / *ecdsa.PrivateKey / ed25519.PrivateKey (and matching
// public-key) values required by the asymmetric JWT variants. The underlying
// golang-jwt/v5 SignedString call uses reflection internally to match the
// signing method to the key type.
type Method struct {
	name          string
	algorithm     Algorithm
	signingMethod jwt.SigningMethod
	cipher        *caead.Method
	signingKey    any
	verifyingKey  any
	issuer        string
	timeout       time.Duration
	generateFn    GenerateFn
	validateFn    ValidateFn
}

// NewMethod creates a new token method with the given name and algorithm.
//
// algorithm selects the underlying primitive — JWT signing method for the
// HS/RS/PS/ES/EdDSA families, AEAD cipher for the opaque family — via an
// opaque enum so callers do not import golang-jwt/v5 or the AEAD package
// directly. Both families flow through the same constructor; the Algorithm
// value alone is enough to decide.
//
// Passing an unknown Algorithm asserts via cassert.True with a message that
// names the invalid value.
func NewMethod(name string, algorithm Algorithm, options ...Option) *Method {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotEmpty(string(algorithm), "algorithm is empty")

	signingMethod := signingMethodFor(algorithm)
	cipher := cipherFor(algorithm)

	cassert.True(signingMethod != nil || cipher != nil, fmt.Sprintf("algorithm %q is invalid", string(algorithm)))

	opts := NewOptions(options...)

	return &Method{
		name:          name,
		algorithm:     algorithm,
		signingMethod: signingMethod,
		cipher:        cipher,
		signingKey:    opts.signingKey,
		verifyingKey:  opts.verifyingKey,
		issuer:        opts.issuer,
		timeout:       opts.timeout,
		generateFn:    opts.generateFn,
		validateFn:    opts.validateFn,
	}
}

// signingMethodFor maps a JWT-family Algorithm to its concrete
// jwt.SigningMethod from golang-jwt/v5. Returns nil for opaque-family values
// or unrecognized algorithms — the caller (NewMethod) cross-checks with
// cipherFor to decide whether the algorithm is invalid.
func signingMethodFor(algorithm Algorithm) jwt.SigningMethod {
	switch algorithm {
	case AlgorithmHS256:
		return jwt.SigningMethodHS256
	case AlgorithmHS384:
		return jwt.SigningMethodHS384
	case AlgorithmHS512:
		return jwt.SigningMethodHS512
	case AlgorithmRS256:
		return jwt.SigningMethodRS256
	case AlgorithmRS384:
		return jwt.SigningMethodRS384
	case AlgorithmRS512:
		return jwt.SigningMethodRS512
	case AlgorithmPS256:
		return jwt.SigningMethodPS256
	case AlgorithmPS384:
		return jwt.SigningMethodPS384
	case AlgorithmPS512:
		return jwt.SigningMethodPS512
	case AlgorithmES256:
		return jwt.SigningMethodES256
	case AlgorithmES384:
		return jwt.SigningMethodES384
	case AlgorithmES512:
		return jwt.SigningMethodES512
	case AlgorithmEdDSA:
		return jwt.SigningMethodEdDSA
	default:
		return nil
	}
}

// cipherFor maps an opaque-family Algorithm to its concrete *caead.Method.
// Returns nil for JWT-family values or unrecognized algorithms.
func cipherFor(algorithm Algorithm) *caead.Method {
	switch algorithm {
	case AlgorithmOpaqueAESGCM:
		return caead.AES_256_GCM
	case AlgorithmOpaqueXChaCha20Poly1305:
		return caead.XCHACHA20_POLY1305
	default:
		return nil
	}
}

// Name returns the method name.
func (m *Method) Name() string {
	cassert.NotNil(m, "method is nil")
	return m.name
}

// Generate creates a signed token for the given subject and payload.
func (m *Method) Generate(subject string, payload Payload) (string, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.generateFn, "method generateFn is nil")

	token, err := m.generateFn(m, subject, payload)
	if err != nil {
		return "", ErrGeneration(err)
	}
	return token, nil
}

// Validate parses and validates a token, returning its payload.
func (m *Method) Validate(tokenString string) (Payload, error) {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.validateFn, "method validateFn is nil")

	payload, err := m.validateFn(m, tokenString)
	if err != nil {
		return nil, ErrValidation(err)
	}
	return payload, nil
}

// DecodeUnsafe returns the token claims WITHOUT verifying the signature or
// expiration. The caller MUST NOT trust the returned payload for authorization
// decisions. Intended for diagnostic use only (logging, routing pre-verification).
//
// For opaque (AEAD-encrypted) tokens the payload is unreadable without the key,
// so DecodeUnsafe on an opaque Method returns ErrCipherRequired — opaque tokens
// have no unsafe-peek path by design.
func (m *Method) DecodeUnsafe(tokenString string) (Payload, error) {
	cassert.NotNil(m, "method is nil")

	parser := jwt.NewParser()
	jwtToken, _, err := parser.ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, ErrValidation(err)
	}

	claims, ok := jwtToken.Claims.(*Claims)
	if !ok {
		return nil, ErrValidation(ErrTokenParseFailed)
	}

	return claims.Payload, nil
}

// With returns a clone of m with the given options applied on top of
// the original configuration. The receiver m is never mutated, so the
// predefined JWT_*/OPAQUE_* singletons remain reusable as templates.
//
// Typical use:
//
//	method := tokens.JWT_HS256.With(tokens.WithKey([]byte(secret)))
//	token, err := method.Generate("alice", payload)
//
// Or chained for asymmetric variants:
//
//	method := tokens.JWT_RS256.With(
//	    tokens.WithSigningKey(privKey),
//	    tokens.WithVerifyingKey(pubKey),
//	    tokens.WithIssuer("ltk"),
//	)
func (m *Method) With(options ...Option) *Method {
	cassert.NotNil(m, "method is nil")

	opts := &Options{
		signingKey:   m.signingKey,
		verifyingKey: m.verifyingKey,
		issuer:       m.issuer,
		timeout:      m.timeout,
		generateFn:   m.generateFn,
		validateFn:   m.validateFn,
	}

	for _, opt := range options {
		opt(opts)
	}

	return &Method{
		name:          m.name,
		algorithm:     m.algorithm,
		signingMethod: m.signingMethod,
		cipher:        m.cipher,
		signingKey:    opts.signingKey,
		verifyingKey:  opts.verifyingKey,
		issuer:        opts.issuer,
		timeout:       opts.timeout,
		generateFn:    opts.generateFn,
		validateFn:    opts.validateFn,
	}
}
