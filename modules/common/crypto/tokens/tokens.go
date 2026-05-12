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
	OPAQUE_AES_256_GCM        = NewOpaqueMethod("OPAQUE_AES_256_GCM", caead.AES_256_GCM)
	OPAQUE_XCHACHA20_POLY1305 = NewOpaqueMethod("OPAQUE_XCHACHA20_POLY1305", caead.XCHACHA20_POLY1305)
)

// Method represents a token signing or encryption configuration.
//
// The cipher field discriminates the flavor:
//
//   - cipher == nil  →  JWT signed token (HS256/HS384/HS512). signingMethod
//     carries the jwt.SigningMethod used by golang-jwt/v5.
//   - cipher != nil  →  Opaque AEAD-encrypted token (YA-0019). signingMethod
//     is unused; signingKey doubles as the AEAD symmetric key.
type Method struct {
	name          string
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
// algorithm selects the signing primitive via an opaque enum so callers do
// not import golang-jwt/v5. Passing an unknown Algorithm asserts via
// cassert.NotNil, with ErrAlgorithmInvalid as the underlying cause.
func NewMethod(name string, algorithm Algorithm, options ...Option) *Method {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotEmpty(string(algorithm), "algorithm is empty")

	signingMethod := signingMethodFor(algorithm)
	cassert.NotNil(signingMethod, fmt.Sprintf("algorithm %q is invalid", string(algorithm)))

	opts := NewOptions(options...)

	return &Method{
		name:          name,
		signingMethod: signingMethod,
		signingKey:    opts.signingKey,
		verifyingKey:  opts.verifyingKey,
		issuer:        opts.issuer,
		timeout:       opts.timeout,
		generateFn:    opts.generateFn,
		validateFn:    opts.validateFn,
	}
}

// signingMethodFor maps an Algorithm enum value to the concrete
// jwt.SigningMethod used by the golang-jwt/v5 backend. It returns nil for
// unrecognized values; callers must convert nil into ErrAlgorithmInvalid.
//
// Opaque algorithm values (AlgorithmOpaqueAESGCM,
// AlgorithmOpaqueXChaCha20Poly1305) intentionally map to nil since AEAD has
// no jwt.SigningMethod analogue — opaque construction must go through
// NewOpaqueMethod, not NewMethod.
func signingMethodFor(algorithm Algorithm) jwt.SigningMethod {
	switch algorithm {
	case AlgorithmHS256:
		return jwt.SigningMethodHS256
	case AlgorithmHS384:
		return jwt.SigningMethodHS384
	case AlgorithmHS512:
		return jwt.SigningMethodHS512
	case AlgorithmOpaqueAESGCM, AlgorithmOpaqueXChaCha20Poly1305:
		return nil
	default:
		return nil
	}
}

// NewOpaqueMethod creates a new opaque (AEAD-encrypted) token method.
//
// The entire claims payload is encrypted via cipher.Encrypt(signingKey,
// jsonBytes, nil), then base64url-encoded. Nothing leaks to the client; only
// holders of the symmetric key can decrypt and validate.
//
// The cipher is required as a positional argument — there is no WithCipher
// option, because cipher choice is structural (it determines the wire
// format) and must be visible at the call site. signingKey doubles as the
// AEAD key; for symmetric AEAD, signingKey and verifyingKey are the same
// secret, so callers typically use WithKey or WithGeneratedKey.
//
// Passing a nil cipher trips the cassert.NotNil invariant.
func NewOpaqueMethod(name string, cipher *caead.Method, options ...Option) *Method {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(cipher, "cipher is nil")

	opts := NewOptions(options...)

	return &Method{
		name:         name,
		cipher:       cipher,
		signingKey:   opts.signingKey,
		verifyingKey: opts.verifyingKey,
		issuer:       opts.issuer,
		timeout:      opts.timeout,
		// generate/validate dispatch on method.cipher != nil — see
		// functions.go. The Options pipeline still allows callers to
		// override via WithGenerateFn / WithValidateFn for testing.
		generateFn: opts.generateFn,
		validateFn: opts.validateFn,
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
