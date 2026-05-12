package tokens

import (
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	cassert "github.com/guidomantilla/yarumo/common/assert"
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
)

// Method represents a token signing algorithm with its configuration.
//
// signingKey and verifyingKey are stored as any so a single Method type
// can carry the byte-slice secrets used by HMAC variants alongside the
// *rsa.PrivateKey / *ecdsa.PrivateKey / ed25519.PrivateKey (and matching
// public-key) values required by the asymmetric variants. The underlying
// golang-jwt/v5 SignedString call uses reflection internally to match
// the signing method to the key type.
type Method struct {
	name          string
	signingMethod jwt.SigningMethod
	signingKey    any
	verifyingKey  any
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
