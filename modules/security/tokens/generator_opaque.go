package tokens

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/utils"

	"github.com/guidomantilla/yarumo/security/cryptos"
)

type opaqueGenerator struct {
	cipherKey []byte
	timeout   time.Duration
}

func NewOpaqueGenerator(opts ...Option) Generator {
	options := NewOptions(opts...)
	return &opaqueGenerator{
		timeout:   options.timeout,
		cipherKey: append([]byte(nil), options.cipherKey...),
	}
}

func (g *opaqueGenerator) Name() Name {
	assert.NotNil(g, "generator is nil")
	return Name(fmt.Sprintf("%s-%s", "OPAQUE", "AES256"))
}

func (g *opaqueGenerator) Generate(subject string, principal Principal) (*string, error) {
	assert.NotNil(g, "generator is nil")

	if utils.Empty(subject) {
		return nil, ErrTokenGeneration(ErrSubjectCannotBeEmpty)
	}
	if utils.Empty(principal) {
		return nil, ErrTokenGeneration(ErrPrincipalCannotBeNil)
	}

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.timeout)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Principal: principal,
	}

	plain, err := json.Marshal(claims)
	if err != nil {
		return nil, ErrTokenGeneration(err)
	}

	ciphered, err := cryptos.AesEncrypt(g.cipherKey, plain)
	if err != nil {
		return nil, ErrTokenGeneration(err)
	}

	token := base64.RawURLEncoding.EncodeToString(ciphered)
	return &token, nil
}

func (g *opaqueGenerator) Validate(tokenString string) (Principal, error) {
	assert.NotNil(g, "generator is nil")

	if utils.Empty(tokenString) {
		return nil, ErrTokenValidation(ErrTokenCannotBeEmpty)
	}

	ciphered, err := base64.RawURLEncoding.DecodeString(tokenString)
	if err != nil {
		return nil, ErrTokenValidation(err)
	}

	plain, err := cryptos.AesDecrypt(g.cipherKey, ciphered)
	if err != nil {
		return nil, ErrTokenValidation(err)
	}

	var claims Claims
	err = json.Unmarshal(plain, &claims)
	if err != nil {
		return nil, ErrTokenValidation(err)
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
