package jsonwebtoken

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/guidomantilla/yarumo/common/errs"
	"github.com/guidomantilla/yarumo/common/utils"
)

type Claims struct {
	jwt.RegisteredClaims
	Payload any `json:"payload,omitempty"`
}

var (
	ErrClaimsCannotBeNil         = errors.New("claims cannot be nil")
	ErrsSigningKeyCannotBeNil    = errors.New("signing key cannot be nil")
	ErrsSigningMethodCannotBeNil = errors.New("signing method cannot be nil")
	ErrTokenCannotBeEmpty        = errors.New("token cannot be empty")
	ErrTokenFailedParsing        = errors.New("token failed to parse")
	ErrTokenEmptyPrincipal       = errors.New("token principal is empty")
)

func Generate(claims *Claims, signingKey any, signingMethod jwt.SigningMethod) (string, error) {
	if utils.Nil(claims) {
		return "", ErrClaimsCannotBeNil
	}
	if utils.Nil(signingKey) {
		return "", ErrsSigningKeyCannotBeNil
	}
	if utils.Nil(signingMethod) {
		return "", ErrsSigningMethodCannotBeNil
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	return token.SignedString(signingKey)
}

func Validate(token string, verifyingKey any, signingMethod jwt.SigningMethod, options ...jwt.ParserOption) (*Claims, error) {
	if utils.Empty(token) {
		return nil, ErrTokenCannotBeEmpty
	}

	getKeyFunc := func(token *jwt.Token) (any, error) {
		return verifyingKey, nil
	}

	options = append(options, jwt.WithValidMethods([]string{signingMethod.Alg()}))

	jwtToken, err := jwt.ParseWithClaims(token, &Claims{}, getKeyFunc, options...)
	if err != nil {
		return nil, errs.Wrap(ErrTokenFailedParsing, err)
	}

	claims := jwtToken.Claims.(*Claims)
	if utils.Nil(claims.Payload) {
		return nil, ErrTokenEmptyPrincipal
	}

	return claims, nil
}
