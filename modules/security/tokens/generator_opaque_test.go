package tokens

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/guidomantilla/yarumo/security/cryptos"
)

type BadClaims struct {
	A int
}

func newOpaqueWithKeyTimeout(key []byte, timeout time.Duration) Generator {
	return NewOpaqueGenerator(WithOpaqueKey(key), WithTimeout(timeout))
}

func TestOpaqueGenerateValidate_Success(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef") // 32 bytes
	g := newOpaqueWithKeyTimeout(key, time.Hour)

	tok, err := g.Generate("subject", Principal{"role": "admin"})
	if err != nil || tok == nil {
		t.Fatalf("unexpected error generating: %v", err)
	}
	p, err := g.Validate(*tok)
	if err != nil {
		t.Fatalf("unexpected error validating: %v", err)
	}
	if p["role"].(string) != "admin" {
		t.Fatalf("unexpected principal: %v", p)
	}
}

func TestOpaqueGenerate_Errors(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	g := newOpaqueWithKeyTimeout(key, time.Hour)

	if _, err := g.Generate("", Principal{"x": 1}); err == nil || !errors.Is(err, ErrTokenGenerationFailed) {
		t.Fatalf("expected generation error for empty subject, got %v", err)
	}
	if _, err := g.Generate("sub", nil); err == nil || !errors.Is(err, ErrTokenGenerationFailed) {
		t.Fatalf("expected generation error for nil principal, got %v", err)
	}
}

func TestOpaqueGenerate_ClaimMarshalError(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	g := newOpaqueWithKeyTimeout(key, time.Hour)

	fn := func() {}

	if _, err := g.Generate("sub", Principal{"x": fn}); err == nil || !errors.Is(err, ErrTokenGenerationFailed) {
		t.Fatalf("expected generation error for empty subject, got %v", err)
	}
}

func TestOpaqueGenerate_AesEncryptError(t *testing.T) {
	key := []byte("1")
	g := newOpaqueWithKeyTimeout(key, time.Hour)

	if _, err := g.Generate("sub", Principal{"x": 1}); err == nil || !errors.Is(err, ErrTokenGenerationFailed) {
		t.Fatalf("expected generation error for empty subject, got %v", err)
	}
}

func TestOpaqueValidate_InputErrors(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	g := newOpaqueWithKeyTimeout(key, time.Hour)

	if _, err := g.Validate(""); err == nil || !errors.Is(err, ErrTokenValidationFailed) {
		t.Fatalf("expected validation error for empty token, got %v", err)
	}

	if _, err := g.Validate("***not-base64***"); err == nil || !errors.Is(err, ErrTokenValidationFailed) {
		t.Fatalf("expected validation error for bad base64, got %v", err)
	}
}

func TestOpaqueValidate_DecryptErrorWithWrongKey(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	g1 := newOpaqueWithKeyTimeout(key, time.Hour)
	g2 := newOpaqueWithKeyTimeout([]byte("abcdef0123456789abcdef0123456789"), time.Hour)

	tok, err := g1.Generate("s", Principal{"a": 1})
	if err != nil {
		t.Fatalf("gen err: %v", err)
	}
	if _, err := g2.Validate(*tok); err == nil || !errors.Is(err, ErrTokenValidationFailed) {
		t.Fatalf("expected decrypt failure, got %v", err)
	}
}

func TestOpaqueValidate_JSONError(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	badJSON := []byte("{not-json}")
	cipher, err := cryptos.AesEncrypt(key, badJSON)
	if err != nil {
		t.Fatalf("encrypt err: %v", err)
	}
	token := base64.RawURLEncoding.EncodeToString(cipher)

	g := newOpaqueWithKeyTimeout(key, time.Hour)
	if _, err := g.Validate(token); err == nil || !errors.Is(err, ErrTokenValidationFailed) {
		t.Fatalf("expected json unmarshal error, got %v", err)
	}
}

func TestOpaqueValidate_ExpiredAndEmptyPrincipal(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")

	gExpired := newOpaqueWithKeyTimeout(key, -time.Minute)
	tok, err := gExpired.Generate("s", Principal{"a": 1})
	if err != nil {
		t.Fatalf("gen err: %v", err)
	}
	if _, err := gExpired.Validate(*tok); err == nil || !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expected expired error, got %v", err)
	}

	claims := Claims{}
	claims.RegisteredClaims.Subject = "s"
	claims.RegisteredClaims.ExpiresAt = nil
	plain, _ := json.Marshal(claims)
	cipher, err := cryptos.AesEncrypt(key, plain)
	if err != nil {
		t.Fatalf("encrypt err: %v", err)
	}
	token := base64.RawURLEncoding.EncodeToString(cipher)

	g := newOpaqueWithKeyTimeout(key, time.Hour)
	if _, err := g.Validate(token); err == nil || !errors.Is(err, ErrTokenValidationFailed) || !errors.Is(err, ErrTokenEmptyPrincipal) {
		t.Fatalf("expected empty principal error, got %v", err)
	}
}
