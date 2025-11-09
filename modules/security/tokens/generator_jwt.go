package tokens

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/guidomantilla/yarumo/common/utils"
)

type Claims struct {
	jwt.RegisteredClaims
	Principal Principal `json:"principal,omitempty"`
}

type jwtGenerator struct {
	issuer        string
	timeout       time.Duration
	signingKey    []byte
	verifyingKey  []byte
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
	if utils.Empty(subject) {
		return nil, ErrTokenGeneration(ErrSubjectCannotBeEmpty)
	}
	if utils.Empty(principal) {
		return nil, ErrTokenGeneration(ErrPrincipalCannotBeNil)
	}

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
		return nil, ErrTokenGeneration(err)
	}

	return &tokenString, nil
}

func (manager *jwtGenerator) Validate(tokenString string) (Principal, error) {
	if utils.Empty(tokenString) {
		return nil, ErrTokenValidation(ErrTokenCannotBeEmpty)
	}

	getKeyFunc := func(token *jwt.Token) (any, error) {
		return manager.verifyingKey, nil
	}

	parserOptions := []jwt.ParserOption{
		jwt.WithIssuer(manager.issuer),
		jwt.WithValidMethods([]string{manager.signingMethod.Alg()}),
	}

	token, err := jwt.Parse(tokenString, getKeyFunc, parserOptions...)
	if err != nil {
		return nil, ErrTokenValidation(ErrTokenFailedParsing, err)
	}

	if !token.Valid {
		return nil, ErrTokenValidation(ErrTokenInvalid)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrTokenValidation(ErrTokenEmptyClaims)
	}

	value, ok := claims["principal"]
	if !ok {
		return nil, ErrTokenValidation(ErrTokenEmptyPrincipal)
	}

	principal := Principal(value.(map[string]any))
	return principal, nil
}
