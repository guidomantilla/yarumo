package tokens

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/guidomantilla/yarumo/pkg/common/assert"
	"github.com/guidomantilla/yarumo/pkg/common/pointer"
)

type Claims struct {
	jwt.RegisteredClaims
	Principal
}

type jwtGenerator struct {
	issuer        string
	timeout       time.Duration
	signingKey    any
	verifyingKey  any
	signingMethod jwt.SigningMethod
}

func NewJwtGenerator(opts ...JwtGeneratorOption) Generator {
	options := NewJwtGeneratorOptions(opts...)
	return &jwtGenerator{
		issuer:        options.issuer,
		timeout:       options.timeout,
		signingKey:    options.signingKey,
		verifyingKey:  options.verifyingKey,
		signingMethod: options.signingMethod,
	}
}

func (manager *jwtGenerator) Generate(subject string, principal Principal) (*string, error) {
	assert.NotNil(principal, "token generator - error generating token: principal is nil")

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    manager.issuer,
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(manager.timeout)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Principal: principal,
	}

	token := jwt.NewWithClaims(manager.signingMethod, claims)

	tokenString, err := token.SignedString(manager.signingKey)
	if err != nil {
		return nil, ErrTokenGenerationFailed(err)
	}

	return &tokenString, nil
}

func (manager *jwtGenerator) Validate(tokenString string) (Principal, error) {
	assert.NotEmpty(tokenString, "token manager - error validating token: token is empty")

	getKeyFunc := func(token *jwt.Token) (any, error) {
		return manager.verifyingKey, nil
	}

	parserOptions := []jwt.ParserOption{
		jwt.WithIssuer(manager.issuer),
		jwt.WithValidMethods([]string{manager.signingMethod.Alg()}),
	}

	token, err := jwt.Parse(tokenString, getKeyFunc, parserOptions...)
	if err != nil {
		return nil, ErrTokenValidationFailed(ErrTokenFailedParsing, err)
	}

	if !token.Valid {
		return nil, ErrTokenValidationFailed(ErrTokenInvalid)
	}

	_, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrTokenValidationFailed(ErrTokenEmptyClaims)
	}

	principal := pointer.Zero[Principal]()
	return principal, nil
}
