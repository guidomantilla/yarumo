package tokens

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/utils"
	"github.com/guidomantilla/yarumo/security/tokens/jsonwebtoken"
)

type jwtGenerator struct {
	issuer        string
	timeout       time.Duration
	signingKey    []byte
	verifyingKey  []byte
	signingMethod jwt.SigningMethod
}

func NewJwtGenerator(opts ...Option) Generator {
	options := NewOptions(opts...)
	return &jwtGenerator{
		issuer:        options.issuer,
		timeout:       options.timeout,
		signingKey:    append([]byte(nil), options.signingKey...),
		verifyingKey:  append([]byte(nil), options.verifyingKey...),
		signingMethod: options.signingMethod,
	}
}

func (g *jwtGenerator) Name() Name {
	assert.NotNil(g, "generator is nil")
	return Name(fmt.Sprintf("%s-%s", "JWT", g.signingMethod.Alg()))
}

func (g *jwtGenerator) Generate(subject string, principal Principal) (string, error) {
	assert.NotNil(g, "generator is nil")

	if utils.Empty(subject) {
		return "", ErrTokenGeneration(ErrSubjectCannotBeEmpty)
	}
	if utils.Nil(principal) {
		return "", ErrTokenGeneration(ErrPrincipalCannotBeNil)
	}

	claims := &jsonwebtoken.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    g.issuer,
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.timeout)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Payload: principal,
	}

	return jsonwebtoken.Generate(claims, g.signingKey, g.signingMethod)
}

func (g *jwtGenerator) Validate(token string) (Principal, error) {
	assert.NotNil(g, "generator is nil")

	if utils.Empty(token) {
		return nil, ErrTokenValidation(ErrTokenCannotBeEmpty)
	}

	claims, err := jsonwebtoken.Validate(token, g.verifyingKey, g.signingMethod, jwt.WithIssuer(g.issuer))
	if err != nil {
		return nil, ErrTokenValidation(err)
	}

	principal := claims.Payload.(Principal)

	if utils.Nil(principal) {
		return nil, ErrTokenValidation(ErrTokenEmptyPrincipal)
	}

	return principal, nil
}
