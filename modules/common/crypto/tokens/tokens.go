package tokens

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// Predefined methods with default parameters.
var (
	JWT_HS256 = NewMethod("JWT_HS256", jwt.SigningMethodHS256)
	JWT_HS384 = NewMethod("JWT_HS384", jwt.SigningMethodHS384)
	JWT_HS512 = NewMethod("JWT_HS512", jwt.SigningMethodHS512)
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

// NewMethod creates a new token method with the given name and signing method.
func NewMethod(name string, signingMethod jwt.SigningMethod, options ...Option) *Method {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(signingMethod, "signing method is nil")

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
