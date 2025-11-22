package tokens

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/guidomantilla/yarumo/common/utils"
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

func (generator *jwtGenerator) Generate(subject string, principal Principal) (*string, error) {
	if utils.Empty(subject) {
		return nil, ErrTokenGeneration(ErrSubjectCannotBeEmpty)
	}
	if utils.Empty(principal) {
		return nil, ErrTokenGeneration(ErrPrincipalCannotBeNil)
	}

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    generator.issuer,
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(generator.timeout)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Principal: principal,
	}

	token := jwt.NewWithClaims(generator.signingMethod, claims)

	tokenString, err := token.SignedString(generator.signingKey)
	if err != nil {
		return nil, ErrTokenGeneration(err)
	}

	return &tokenString, nil
}

func (generator *jwtGenerator) Validate(tokenString string) (Principal, error) {
	if utils.Empty(tokenString) {
		return nil, ErrTokenValidation(ErrTokenCannotBeEmpty)
	}

	getKeyFunc := func(token *jwt.Token) (any, error) {
		return generator.verifyingKey, nil
	}

	parserOptions := []jwt.ParserOption{
		jwt.WithIssuer(generator.issuer),
		jwt.WithValidMethods([]string{generator.signingMethod.Alg()}),
	}

	token, err := jwt.ParseWithClaims(tokenString, Claims{}, getKeyFunc, parserOptions...)
	if err != nil {
		return nil, ErrTokenValidation(ErrTokenFailedParsing, err)
	}

	if !token.Valid {
		return nil, ErrTokenValidation(ErrTokenInvalid)
	}

	claims, ok := token.Claims.(Claims)
	if !ok {
		return nil, ErrTokenValidation(ErrTokenEmptyClaims)
	}

	now := time.Now()
	if utils.NotNil(claims.RegisteredClaims.ExpiresAt) && now.After(claims.RegisteredClaims.ExpiresAt.Time) {
		return nil, ErrTokenValidation(ErrTokenExpired)
	}

	if utils.Nil(claims.Principal) {
		return nil, ErrTokenValidation(ErrTokenEmptyPrincipal)
	}

	return claims.Principal, nil
}
