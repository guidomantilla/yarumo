package tokens

import (
	"encoding/base64"
	"encoding/json"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	cutils "github.com/guidomantilla/yarumo/common/utils"
)

// generate is the package-default GenerateFn. It dispatches on the method's
// Algorithm family:
//
//   - opaque family (AlgorithmOpaque*) → opaque AEAD-encrypted token.
//   - JWT family (everything else)    → JWT signed token via golang-jwt/v5.
//
// Method.Generate (the struct method) already enforces method != nil via
// cassert.NotNil, so this private helper trusts the receiver. Callers that
// need a different impl can inject via WithGenerateFn; the Method.Generate
// wrapper still funnels everything through ErrGeneration.
func generate(method *Method, subject string, payload Payload) (string, error) {
	if method.algorithm.isOpaque() {
		return generateOpaque(method, subject, payload)
	}
	return generateJWT(method, subject, payload)
}

// validate is the package-default ValidateFn. It dispatches on the method's
// Algorithm family; see generate for the rules.
func validate(method *Method, tokenString string) (Payload, error) {
	if method.algorithm.isOpaque() {
		return validateOpaque(method, tokenString)
	}
	return validateJWT(method, tokenString)
}

func generateJWT(method *Method, subject string, payload Payload) (string, error) {

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
		return "", cerrs.Wrap(ErrTokenSignFailed, err)
	}

	return signed, nil
}

func validateJWT(method *Method, tokenString string) (Payload, error) {

	if cutils.Empty(tokenString) {
		return nil, ErrTokenEmpty
	}

	if cutils.Nil(method.verifyingKey) {
		return nil, ErrVerifyingKeyNil
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

// generateOpaque builds an opaque (AEAD-encrypted) token for the given
// subject and payload. The full claims envelope (iss, sub, iat, nbf, exp,
// payload) is serialized to JSON, encrypted with the method's cipher under
// signingKey, and base64url-encoded. The AEAD nonce is prepended internally
// by caead.Method.Encrypt.
func generateOpaque(method *Method, subject string, payload Payload) (string, error) {

	if cutils.Empty(subject) {
		return "", ErrSubjectEmpty
	}

	if cutils.Nil(payload) {
		return "", ErrPayloadNil
	}

	if cutils.Nil(method.cipher) {
		return "", ErrCipherNil
	}

	if cutils.Nil(method.signingKey) {
		return "", ErrSigningKeyNil
	}

	// Opaque AEAD requires a symmetric byte-slice key. The Method.signingKey
	// field is widened to any so the JWT-asymmetric path can hold *rsa /
	// *ecdsa / ed25519 keys; for opaque it must be []byte.
	signingKey, ok := method.signingKey.([]byte)
	if !ok {
		return "", ErrSigningKeyNil
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

	jsonBytes, err := json.Marshal(claims)
	if err != nil {
		return "", cerrs.Wrap(ErrTokenMarshalFailed, err)
	}

	ciphertext, err := method.cipher.Encrypt(signingKey, jsonBytes, nil)
	if err != nil {
		return "", cerrs.Wrap(ErrTokenSignFailed, err)
	}

	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

// validateOpaque decodes, decrypts, and validates an opaque token. It
// returns the embedded payload on success, or one of the opaque-specific
// sentinels (ErrTokenDecodeFailed, ErrTokenDecryptFailed,
// ErrTokenUnmarshalFailed, ErrTokenExpired, ErrTokenNotYetValid,
// ErrTokenIssuerMismatch) wrapped via cerrs.Wrap on failure.
func validateOpaque(method *Method, tokenString string) (Payload, error) {

	if cutils.Empty(tokenString) {
		return nil, ErrTokenEmpty
	}

	if cutils.Nil(method.cipher) {
		return nil, ErrCipherNil
	}

	if cutils.Nil(method.verifyingKey) {
		return nil, ErrVerifyingKeyNil
	}

	// Opaque AEAD requires a symmetric byte-slice key. See the matching
	// assertion in generateOpaque for the rationale.
	verifyingKey, ok := method.verifyingKey.([]byte)
	if !ok {
		return nil, ErrVerifyingKeyNil
	}

	ciphertext, err := base64.RawURLEncoding.DecodeString(tokenString)
	if err != nil {
		return nil, cerrs.Wrap(ErrTokenDecodeFailed, err)
	}

	jsonBytes, err := method.cipher.Decrypt(verifyingKey, ciphertext, nil)
	if err != nil {
		return nil, cerrs.Wrap(ErrTokenDecryptFailed, err)
	}

	claims := &Claims{}

	err = json.Unmarshal(jsonBytes, claims)
	if err != nil {
		return nil, cerrs.Wrap(ErrTokenUnmarshalFailed, err)
	}

	err = checkOpaqueClaims(method, claims)
	if err != nil {
		return nil, err
	}

	if cutils.Nil(claims.Payload) {
		return nil, ErrTokenPayloadEmpty
	}

	return claims.Payload, nil
}

// checkOpaqueClaims mirrors the temporal and issuer checks performed by
// jwt.ParseWithClaims, but against the freshly-decrypted Claims envelope so
// the opaque path enforces the same invariants as the JWT path.
func checkOpaqueClaims(method *Method, claims *Claims) error {
	now := time.Now()

	if claims.ExpiresAt != nil && now.After(claims.ExpiresAt.Time) {
		return ErrTokenExpired
	}

	if claims.NotBefore != nil && now.Before(claims.NotBefore.Time) {
		return ErrTokenNotYetValid
	}

	if cutils.NotEmpty(method.issuer) && claims.Issuer != method.issuer {
		return ErrTokenIssuerMismatch
	}

	return nil
}
