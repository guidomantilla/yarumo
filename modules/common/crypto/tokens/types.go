// Package tokens provides JWT token generation and validation using HMAC signing methods.
package tokens

import (
	jwt "github.com/golang-jwt/jwt/v5"
)

var (
	_ GenerateFn = generate
	_ ValidateFn = validate
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
