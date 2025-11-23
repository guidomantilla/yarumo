package tokens

import (
	"errors"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func newJwtWith(key []byte, issuer string, method jwt.SigningMethod, timeout time.Duration) *jwtGenerator {
	// Construimos el generador real usando opciones públicas
	g := NewJwtGenerator(
		WithJwtKey(key),
		WithJwtIssuer(issuer),
		WithJwtSigningMethod(method),
		WithTimeout(timeout),
	)
	return g.(*jwtGenerator)
}

func TestJwtGenerateValidate_Success(t *testing.T) {
	key := []byte("secret-key")
	g := newJwtWith(key, "issuer-x", jwt.SigningMethodHS512, time.Hour)

	tok, err := g.Generate("sub", Principal{"r": "u"})
	if err != nil || tok == nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := g.Validate(*tok)
	if err != nil || p["r"].(string) != "u" {
		t.Fatalf("unexpected validate result: %v %v", p, err)
	}
}

func TestJwtGenerate_InputErrors(t *testing.T) {
	g := newJwtWith([]byte("k"), "", jwt.SigningMethodHS512, time.Hour)
	if _, err := g.Generate("", Principal{"a": 1}); err == nil || !errors.Is(err, ErrTokenGenerationFailed) {
		t.Fatalf("expected error for empty subject, got %v", err)
	}
	if _, err := g.Generate("sub", nil); err == nil || !errors.Is(err, ErrTokenGenerationFailed) {
		t.Fatalf("expected error for nil principal, got %v", err)
	}
}

func TestJwtGenerate_SigningMethodErrors(t *testing.T) {
	g := newJwtWith([]byte("k"), "", jwt.SigningMethodNone, time.Hour)
	if _, err := g.Generate("sub", Principal{"a": 1}); err == nil || !errors.Is(err, ErrTokenGenerationFailed) {
		t.Fatalf("expected error for signing method none, got %v", err)
	}
}

func TestJwtValidate_InputEmptyToken(t *testing.T) {
	g := newJwtWith([]byte("k"), "", jwt.SigningMethodHS512, time.Hour)
	if _, err := g.Validate(""); err == nil || !errors.Is(err, ErrTokenValidationFailed) {
		t.Fatalf("expected error for empty token, got %v", err)
	}
}

func TestJwtValidate_SignatureWrongKey(t *testing.T) {
	g1 := newJwtWith([]byte("key-1"), "", jwt.SigningMethodHS512, time.Hour)
	g2 := newJwtWith([]byte("key-2"), "", jwt.SigningMethodHS512, time.Hour)

	tok, err := g1.Generate("s", Principal{"a": 1})
	if err != nil {
		t.Fatalf("gen err: %v", err)
	}
	if _, err := g2.Validate(*tok); err == nil || !errors.Is(err, ErrTokenValidationFailed) {
		t.Fatalf("expected parsing/verification error, got %v", err)
	}
}

func TestJwtValidate_Expired(t *testing.T) {
	g := newJwtWith([]byte("key"), "", jwt.SigningMethodHS512, -time.Minute)
	tok, err := g.Generate("s", Principal{"a": 1})
	if err != nil {
		t.Fatalf("gen err: %v", err)
	}
	if _, err := g.Validate(*tok); err == nil || !errors.Is(err, ErrTokenFailedParsing) {
		t.Fatalf("expected expired error, got %v", err)
	}
}

func TestJwtValidate_PrincipalEmpty(t *testing.T) {
	// Usamos el generador para firmar manualmente un token con principal = nil
	g := newJwtWith([]byte("key-x"), "issuer-y", jwt.SigningMethodHS512, time.Hour)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    g.issuer,
			Subject:   "s",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Principal: nil,
	}
	token := jwt.NewWithClaims(g.signingMethod, claims)
	tokenString, err := token.SignedString(g.signingKey)
	if err != nil {
		t.Fatalf("sign err: %v", err)
	}

	if _, err := g.Validate(tokenString); err == nil || !errors.Is(err, ErrTokenEmptyPrincipal) {
		t.Fatalf("expected empty principal error, got %v", err)
	}
}

func TestJwtValidate_InvalidMethod(t *testing.T) {
	// El validador solo acepta HS512, construimos un token HS256
	g := newJwtWith([]byte("k"), "issuer-z", jwt.SigningMethodHS512, time.Hour)

	claims := Claims{RegisteredClaims: jwt.RegisteredClaims{Issuer: g.issuer, Subject: "s"}, Principal: Principal{"x": 1}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(g.signingKey)
	if err != nil {
		t.Fatalf("sign err: %v", err)
	}

	if _, err := g.Validate(s); err == nil || !errors.Is(err, ErrTokenValidationFailed) {
		t.Fatalf("expected method invalid parsing error, got %v", err)
	}
}
