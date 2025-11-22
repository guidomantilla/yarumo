package tokens

import "github.com/golang-jwt/jwt/v5"

var (
	_ Generator = (*jwtGenerator)(nil)
)

type Claims struct {
	jwt.RegisteredClaims
	Principal Principal `json:"principal,omitempty"`
}

type Generator interface {
	Generate(subject string, principal Principal) (*string, error)
	Validate(tokenString string) (Principal, error)
}

type Principal map[string]any
