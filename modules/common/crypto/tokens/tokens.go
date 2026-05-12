package tokens

import (
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	caead "github.com/guidomantilla/yarumo/common/crypto/ciphers/aead"
)

// Predefined methods with default parameters.
var (
	JWT_HS256                 = NewMethod("JWT_HS256", AlgorithmHS256)
	JWT_HS384                 = NewMethod("JWT_HS384", AlgorithmHS384)
	JWT_HS512                 = NewMethod("JWT_HS512", AlgorithmHS512)
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
type Method struct {
	name          string
	algorithm     Algorithm
	signingMethod jwt.SigningMethod
	cipher        *caead.Method
	signingKey    []byte
	verifyingKey  []byte
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
