package tokens

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// Predefined methods with default parameters.
var (
	JWT_HS256 = NewMethod("JWT_HS256", AlgorithmHS256)
	JWT_HS384 = NewMethod("JWT_HS384", AlgorithmHS384)
	JWT_HS512 = NewMethod("JWT_HS512", AlgorithmHS512)
)

// Method represents a token signing algorithm with its configuration.
type Method struct {
	name          string
	signingMethod jwt.SigningMethod
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
	cassert.NotNil(signingMethod, ErrAlgorithmInvalid(algorithm).Error())

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
