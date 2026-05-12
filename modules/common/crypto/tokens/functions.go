package tokens

import (
	"encoding/base64"
	"encoding/json"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	cutils "github.com/guidomantilla/yarumo/common/utils"
)

// Generate is the recommended entry point for callers that receive the
// algorithm name as a string (e.g. loaded from config, a request header, or
// a database column). It performs a single registry Get and forwards to
// Method.Generate.
//
// The named Method must have been registered with keys configured (via
// WithKey / WithSigningKey / WithGeneratedKey). The predefined templates
// (JWT_HS256, JWT_RS256, OPAQUE_AES_256_GCM, ...) carry no key material;
// register a custom Method via Register before calling Generate by name.
func Generate(name, subject string, payload Payload) (string, error) {
	method, err := Get(name)
	if err != nil {
		return "", err
	}

	return method.Generate(subject, payload)
}

// Validate is the recommended entry point for callers that receive the
// algorithm name as a string. It performs a single registry Get and forwards
// to Method.Validate.
//
// The named Method must have been registered with verifying-key material
// configured. See the Generate doc-string for the registry contract.
func Validate(name, tokenString string) (Payload, error) {
	method, err := Get(name)
	if err != nil {
		return nil, err
	}

	return method.Validate(tokenString)
}

// generate is the package-default GenerateFn. It dispatches on the method's
// Algorithm family:
//
//   - opaque family (AlgorithmOpaque*) → opaque AEAD-encrypted token.
//   - JWT family (everything else)    → JWT signed token via golang-jwt/v5.
//
// Per the workspace defensive-validation standard, every private helper that
// takes *Method validates it independently — even though Method.Generate
// already cassert.NotNil-checks the receiver upstream.
func generate(method *Method, subject string, payload Payload) (string, error) {
	if method == nil {
		return "", ErrMethodIsNil
	}

	if method.algorithm.isOpaque() {
		return generateOpaque(method, subject, payload)
	}
	return generateJWT(method, subject, payload)
}

// validate is the package-default ValidateFn. It dispatches on the method's
// Algorithm family; see generate for the rules.
func validate(method *Method, tokenString string) (Payload, error) {
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if method.algorithm.isOpaque() {
		return validateOpaque(method, tokenString)
	}
	return validateJWT(method, tokenString)
}

func generateJWT(method *Method, subject string, payload Payload) (string, error) {
	if method == nil {
		return "", ErrMethodIsNil
	}

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
	if method == nil {
		return nil, ErrMethodIsNil
	}

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
	if method == nil {
		return "", ErrMethodIsNil
	}

	if cutils.Empty(subject) {
		return "", ErrSubjectEmpty
	}

	if cutils.Nil(payload) {
		return "", ErrPayloadNil
	}

	if cutils.Nil(method.cipher) {
		return "", ErrCipherNil
	}

	signingKey, err := opaqueAEADKey(method.signingKey, ErrSigningKeyNil)
	if err != nil {
		return "", err
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
	if method == nil {
		return nil, ErrMethodIsNil
	}

	if cutils.Empty(tokenString) {
		return nil, ErrTokenEmpty
	}

	if cutils.Nil(method.cipher) {
		return nil, ErrCipherNil
	}

	verifyingKey, err := opaqueAEADKey(method.verifyingKey, ErrVerifyingKeyNil)
	if err != nil {
		return nil, err
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

// opaqueAEADKey converts a Method's signing/verifying key (typed as any to
// accommodate the JWT-asymmetric variants that use *rsa.PrivateKey /
// *ecdsa.PrivateKey / ed25519.PrivateKey) into the []byte required by the
// opaque AEAD path. Returns the provided nilSentinel (ErrSigningKeyNil or
// ErrVerifyingKeyNil) if the value is nil or not a []byte.
func opaqueAEADKey(key any, nilSentinel error) ([]byte, error) {
	if key == nil {
		return nil, nilSentinel
	}

	bytes, ok := key.([]byte)
	if !ok {
		return nil, nilSentinel
	}

	return bytes, nil
}
