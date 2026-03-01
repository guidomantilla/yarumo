package tokens

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	cutils "github.com/guidomantilla/yarumo/common/utils"
)

func generate(method *Method, subject string, payload Payload) (string, error) {

	if cutils.Empty(subject) {
		return "", ErrSubjectEmpty
	}

	if cutils.Nil(payload) {
		return "", ErrPayloadNil
	}

	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    method.issuer,
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(now.Add(method.timeout)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Payload: payload,
	}

	if cutils.Nil(method.signingKey) {
		return "", ErrSigningKeyNil
	}

	if cutils.Nil(method.signingMethod) {
		return "", ErrSigningMethodNil
	}

	token := jwt.NewWithClaims(method.signingMethod, claims)
	signed, err := token.SignedString(method.signingKey)
	if err != nil {
		return "", err
	}

	return signed, nil
}

func validate(method *Method, tokenString string) (Payload, error) {

	if cutils.Empty(tokenString) {
		return nil, ErrTokenEmpty
	}

	getKeyFunc := func(_ *jwt.Token) (any, error) {
		return method.verifyingKey, nil
	}

	parserOptions := []jwt.ParserOption{
		jwt.WithValidMethods([]string{method.signingMethod.Alg()}),
	}

	if cutils.NotEmpty(method.issuer) {
		parserOptions = append(parserOptions, jwt.WithIssuer(method.issuer))
	}

	jwtToken, err := jwt.ParseWithClaims(tokenString, &Claims{}, getKeyFunc, parserOptions...)
	if err != nil {
		return nil, cerrs.Wrap(ErrTokenParseFailed, err)
	}

	claims, ok := jwtToken.Claims.(*Claims)
	if !ok {
		return nil, ErrTokenParseFailed
	}

	if cutils.Nil(claims.Payload) {
		return nil, ErrTokenPayloadEmpty
	}

	return claims.Payload, nil
}
