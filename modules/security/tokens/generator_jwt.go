package tokens

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/utils"
)

var (
	DefaultJwtGenerator = NewJwtGenerator()
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

func (g *jwtGenerator) Generate(subject string, principal Principal) (*string, error) {
	assert.NotNil(g, "generator is nil")

	if utils.Empty(subject) {
		return nil, ErrTokenGeneration(ErrSubjectCannotBeEmpty)
	}
	if utils.Empty(principal) {
		return nil, ErrTokenGeneration(ErrPrincipalCannotBeNil)
	}

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    g.issuer,
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.timeout)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Principal: principal,
	}

	token := jwt.NewWithClaims(g.signingMethod, claims)

	tokenString, err := token.SignedString(g.signingKey)
	if err != nil {
		return nil, ErrTokenGeneration(err)
	}

	return &tokenString, nil
}

func (g *jwtGenerator) Validate(tokenString string) (Principal, error) {
	assert.NotNil(g, "generator is nil")

	if utils.Empty(tokenString) {
		return nil, ErrTokenValidation(ErrTokenCannotBeEmpty)
	}

	getKeyFunc := func(token *jwt.Token) (any, error) {
		return g.verifyingKey, nil
	}

	parserOptions := []jwt.ParserOption{
		jwt.WithIssuer(g.issuer),
		jwt.WithValidMethods([]string{g.signingMethod.Alg()}),
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, getKeyFunc, parserOptions...)
	if err != nil {
		return nil, ErrTokenValidation(ErrTokenFailedParsing, err)
	}

	claims := token.Claims.(*Claims)
	if utils.Nil(claims.Principal) {
		return nil, ErrTokenValidation(ErrTokenEmptyPrincipal)
	}

	return claims.Principal, nil
}
