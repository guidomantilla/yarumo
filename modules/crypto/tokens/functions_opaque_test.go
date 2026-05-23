package tokens

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	caead "github.com/guidomantilla/yarumo/crypto/ciphers/aead"
)

// scopeRead is a fixture scope value used by the XChaCha20-Poly1305 round-trip
// test. Extracted into a const so the same literal does not repeat across the
// payload construction, the assertion, and the failure message (goconst).
const scopeRead = "read"

// aes256Key returns a 32-byte key suitable for AES-256-GCM.
func aes256Key() []byte {
	return []byte("0123456789abcdef0123456789abcdef")
}

// xchachaKey returns a 32-byte key suitable for XChaCha20-Poly1305.
func xchachaKey() []byte {
	return []byte("abcdef0123456789abcdef0123456789")
}

func TestOpaque_NewMethod(t *testing.T) {
	t.Parallel()

	t.Run("creates method with cipher and default options", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("custom-opaque", AlgorithmOpaqueAESGCM)

		if m == nil {
			t.Fatal("expected non-nil method")
		}
		if m.cipher == nil {
			t.Fatal("expected non-nil cipher on method")
		}
		if m.signingMethod != nil {
			t.Fatalf("expected nil signingMethod, got %v", m.signingMethod)
		}
		if m.name != "custom-opaque" {
			t.Fatalf("expected name 'custom-opaque', got %q", m.name)
		}
	})

	t.Run("predefined OPAQUE_AES_256_GCM has cipher", func(t *testing.T) {
		t.Parallel()

		if OPAQUE_AES_256_GCM.cipher == nil {
			t.Fatal("expected non-nil cipher on OPAQUE_AES_256_GCM")
		}
		if OPAQUE_AES_256_GCM.signingMethod != nil {
			t.Fatal("expected nil signingMethod on opaque method")
		}
	})

	t.Run("predefined OPAQUE_XCHACHA20_POLY1305 has cipher", func(t *testing.T) {
		t.Parallel()

		if OPAQUE_XCHACHA20_POLY1305.cipher == nil {
			t.Fatal("expected non-nil cipher on OPAQUE_XCHACHA20_POLY1305")
		}
	})
}

func TestOpaque_RoundTrip_AES256GCM(t *testing.T) {
	t.Parallel()

	key := aes256Key()
	m := NewMethod("aes-gcm-roundtrip", AlgorithmOpaqueAESGCM, WithKey(key))

	token, err := m.Generate("user@test.com", Payload{"role": roleAdmin, "tenant": "acme"})
	if err != nil {
		t.Fatalf("unexpected generate error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Sanity check: the token must be base64url-decodable and decidedly
	// NOT a JWT (no dots).
	_, err = base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("opaque token must be base64url-decodable: %v", err)
	}

	payload, err := m.Validate(token)
	if err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}

	if payload["role"] != roleAdmin {
		t.Fatalf("expected role 'admin', got %v", payload["role"])
	}
	if payload["tenant"] != "acme" {
		t.Fatalf("expected tenant 'acme', got %v", payload["tenant"])
	}
}

func TestOpaque_RoundTrip_XChaCha20Poly1305(t *testing.T) {
	t.Parallel()

	key := xchachaKey()
	m := NewMethod("xchacha-roundtrip", AlgorithmOpaqueXChaCha20Poly1305, WithKey(key))

	token, err := m.Generate("user@test.com", Payload{"role": "auditor", "scope": scopeRead})
	if err != nil {
		t.Fatalf("unexpected generate error: %v", err)
	}

	payload, err := m.Validate(token)
	if err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}

	if payload["role"] != "auditor" {
		t.Fatalf("expected role 'auditor', got %v", payload["role"])
	}
	if payload["scope"] != scopeRead {
		t.Fatalf("expected scope %q, got %v", scopeRead, payload["scope"])
	}
}

func TestOpaque_RoundTrip_WithIssuer(t *testing.T) {
	t.Parallel()

	key := aes256Key()
	m := NewMethod("aes-gcm-issuer", AlgorithmOpaqueAESGCM,
		WithKey(key),
		WithIssuer("yarumo-test"),
	)

	token, err := m.Generate("user@test.com", Payload{"role": roleAdmin})
	if err != nil {
		t.Fatalf("unexpected generate error: %v", err)
	}

	payload, err := m.Validate(token)
	if err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}

	if payload["role"] != roleAdmin {
		t.Fatalf("expected role 'admin', got %v", payload["role"])
	}
}

func TestOpaque_Generate_EmptySubject(t *testing.T) {
	t.Parallel()

	m := NewMethod("e", AlgorithmOpaqueAESGCM, WithKey(aes256Key()))

	_, err := m.Generate("", Payload{"k": "v"})
	if err == nil {
		t.Fatal("expected error for empty subject")
	}
	if !errors.Is(err, ErrSubjectEmpty) {
		t.Fatalf("expected ErrSubjectEmpty, got %v", err)
	}
}

func TestOpaque_Generate_NilPayload(t *testing.T) {
	t.Parallel()

	m := NewMethod("e", AlgorithmOpaqueAESGCM, WithKey(aes256Key()))

	_, err := m.Generate("user@test.com", nil)
	if err == nil {
		t.Fatal("expected error for nil payload")
	}
	if !errors.Is(err, ErrPayloadNil) {
		t.Fatalf("expected ErrPayloadNil, got %v", err)
	}
}

func TestOpaque_Generate_NilSigningKey(t *testing.T) {
	t.Parallel()

	// Construct without a key — Generate must surface ErrSigningKeyNil.
	m := NewMethod("no-key", AlgorithmOpaqueAESGCM)

	_, err := m.Generate("user@test.com", Payload{"k": "v"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrSigningKeyNil) {
		t.Fatalf("expected ErrSigningKeyNil, got %v", err)
	}
}

func TestOpaque_Generate_NilCipher(t *testing.T) {
	t.Parallel()

	// Force the opaque generate path directly with a nil cipher by
	// hand-constructing a Method whose generateFn is generateOpaque.
	m := &Method{
		name:       "nil-cipher",
		cipher:     nil,
		signingKey: aes256Key(),
		generateFn: generateOpaque,
	}

	_, err := m.Generate("user@test.com", Payload{"k": "v"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrCipherNil) {
		t.Fatalf("expected ErrCipherNil, got %v", err)
	}
}

func TestOpaque_Validate_EmptyToken(t *testing.T) {
	t.Parallel()

	m := NewMethod("e", AlgorithmOpaqueAESGCM, WithKey(aes256Key()))

	_, err := m.Validate("")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrTokenEmpty) {
		t.Fatalf("expected ErrTokenEmpty, got %v", err)
	}
}

func TestOpaque_Validate_NilVerifyingKey(t *testing.T) {
	t.Parallel()

	m := NewMethod("no-vkey", AlgorithmOpaqueAESGCM)

	_, err := m.Validate("dGhpcy1pcy1qdW5r")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrVerifyingKeyNil) {
		t.Fatalf("expected ErrVerifyingKeyNil, got %v", err)
	}
}

func TestOpaque_Validate_DecodeFailure(t *testing.T) {
	t.Parallel()

	m := NewMethod("decode-fail", AlgorithmOpaqueAESGCM, WithKey(aes256Key()))

	// "!!!" is not a valid base64url string.
	_, err := m.Validate("!!!not-base64url!!!")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrTokenDecodeFailed) {
		t.Fatalf("expected ErrTokenDecodeFailed in chain, got %v", err)
	}
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed in chain, got %v", err)
	}
}

func TestOpaque_Validate_ExpiredToken(t *testing.T) {
	t.Parallel()

	key := aes256Key()
	m := NewMethod("expired", AlgorithmOpaqueAESGCM, WithKey(key))

	// Hand-craft Claims with exp in the past, then encrypt with the same
	// cipher/key so decryption succeeds and the temporal check fires.
	past := time.Now().Add(-2 * time.Hour)
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user@test.com",
			IssuedAt:  jwt.NewNumericDate(past),
			NotBefore: jwt.NewNumericDate(past),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
		Payload: Payload{"role": roleAdmin},
	}

	jsonBytes, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	ciphertext, err := caead.AES_256_GCM.Encrypt(key, jsonBytes, nil)
	if err != nil {
		t.Fatalf("unexpected encrypt error: %v", err)
	}

	token := base64.RawURLEncoding.EncodeToString(ciphertext)

	_, err = m.Validate(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	if !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired in chain, got %v", err)
	}
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed in chain, got %v", err)
	}
}

func TestOpaque_Validate_NotYetValid(t *testing.T) {
	t.Parallel()

	key := aes256Key()
	m := NewMethod("nbf", AlgorithmOpaqueAESGCM, WithKey(key))

	future := time.Now().Add(1 * time.Hour)
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user@test.com",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(future),
			ExpiresAt: jwt.NewNumericDate(future.Add(1 * time.Hour)),
		},
		Payload: Payload{"role": roleAdmin},
	}

	jsonBytes, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	ciphertext, err := caead.AES_256_GCM.Encrypt(key, jsonBytes, nil)
	if err != nil {
		t.Fatalf("unexpected encrypt error: %v", err)
	}

	token := base64.RawURLEncoding.EncodeToString(ciphertext)

	_, err = m.Validate(token)
	if err == nil {
		t.Fatal("expected error for not-yet-valid token")
	}
	if !errors.Is(err, ErrTokenNotYetValid) {
		t.Fatalf("expected ErrTokenNotYetValid in chain, got %v", err)
	}
}

func TestOpaque_Validate_WrongKey(t *testing.T) {
	t.Parallel()

	keyA := aes256Key()
	keyB := []byte("fedcba9876543210fedcba9876543210")

	mGen := NewMethod("gen", AlgorithmOpaqueAESGCM, WithKey(keyA))
	mVal := NewMethod("val", AlgorithmOpaqueAESGCM, WithKey(keyB))

	token, err := mGen.Generate("user@test.com", Payload{"role": roleAdmin})
	if err != nil {
		t.Fatalf("unexpected generate error: %v", err)
	}

	_, err = mVal.Validate(token)
	if err == nil {
		t.Fatal("expected error for wrong key")
	}
	if !errors.Is(err, ErrTokenDecryptFailed) {
		t.Fatalf("expected ErrTokenDecryptFailed in chain, got %v", err)
	}
}

func TestOpaque_Validate_TamperedCiphertext(t *testing.T) {
	t.Parallel()

	key := aes256Key()
	m := NewMethod("tamper", AlgorithmOpaqueAESGCM, WithKey(key))

	token, err := m.Generate("user@test.com", Payload{"role": roleAdmin})
	if err != nil {
		t.Fatalf("unexpected generate error: %v", err)
	}

	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}

	// Flip one byte in the ciphertext region (skip the 12-byte nonce so
	// AEAD authenticates the modified data path, not a different nonce).
	if len(raw) < 16 {
		t.Fatalf("ciphertext too short to tamper: %d bytes", len(raw))
	}
	raw[len(raw)-1] ^= 0x01

	tampered := base64.RawURLEncoding.EncodeToString(raw)

	_, err = m.Validate(tampered)
	if err == nil {
		t.Fatal("expected error for tampered ciphertext")
	}
	if !errors.Is(err, ErrTokenDecryptFailed) {
		t.Fatalf("expected ErrTokenDecryptFailed in chain, got %v", err)
	}
}

func TestOpaque_Validate_IssuerMismatch(t *testing.T) {
	t.Parallel()

	key := aes256Key()
	mGen := NewMethod("gen", AlgorithmOpaqueAESGCM, WithKey(key), WithIssuer("issuer-A"))
	mVal := NewMethod("val", AlgorithmOpaqueAESGCM, WithKey(key), WithIssuer("issuer-B"))

	token, err := mGen.Generate("user@test.com", Payload{"role": roleAdmin})
	if err != nil {
		t.Fatalf("unexpected generate error: %v", err)
	}

	_, err = mVal.Validate(token)
	if err == nil {
		t.Fatal("expected error for issuer mismatch")
	}
	if !errors.Is(err, ErrTokenIssuerMismatch) {
		t.Fatalf("expected ErrTokenIssuerMismatch in chain, got %v", err)
	}
}

func TestOpaque_Validate_EmptyPayload(t *testing.T) {
	t.Parallel()

	key := aes256Key()
	m := NewMethod("empty-payload", AlgorithmOpaqueAESGCM, WithKey(key))

	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user@test.com",
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
		},
		Payload: nil,
	}

	jsonBytes, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	ciphertext, err := caead.AES_256_GCM.Encrypt(key, jsonBytes, nil)
	if err != nil {
		t.Fatalf("unexpected encrypt error: %v", err)
	}

	token := base64.RawURLEncoding.EncodeToString(ciphertext)

	_, err = m.Validate(token)
	if err == nil {
		t.Fatal("expected error for empty payload")
	}
	if !errors.Is(err, ErrTokenPayloadEmpty) {
		t.Fatalf("expected ErrTokenPayloadEmpty, got %v", err)
	}
}

func TestOpaque_Validate_GarbageAfterDecodeIsDecryptFailed(t *testing.T) {
	t.Parallel()

	// A short but valid base64url string that decodes to garbage shorter
	// than the AEAD nonce — the caead.Method.Decrypt path surfaces
	// ErrCiphertextTooShort wrapped by ErrDecryption, and we wrap with
	// ErrTokenDecryptFailed.
	m := NewMethod("garbage", AlgorithmOpaqueAESGCM, WithKey(aes256Key()))

	_, err := m.Validate("YQ") // base64url for "a"
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrTokenDecryptFailed) {
		t.Fatalf("expected ErrTokenDecryptFailed in chain, got %v", err)
	}
}

func TestOpaque_Dispatch_DefaultGenerateUsesOpaque(t *testing.T) {
	t.Parallel()

	// generate() — the package-default GenerateFn — must dispatch to
	// generateOpaque when method.cipher != nil. Build a Method through
	// NewMethod with an opaque Algorithm and confirm the resulting token is
	// opaque (no dots).
	m := NewMethod("dispatch", AlgorithmOpaqueAESGCM, WithKey(aes256Key()))

	token, err := m.Generate("user@test.com", Payload{"x": "y"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// JWT tokens have two dots; opaque tokens have none.
	for _, b := range []byte(token) {
		if b == '.' {
			t.Fatalf("expected opaque token without dots, got JWT-shaped %q", token)
		}
	}
}

func TestOpaque_Registry(t *testing.T) {
	t.Parallel()

	t.Run("OPAQUE_AES_256_GCM is retrievable via Get", func(t *testing.T) {
		t.Parallel()

		got, err := Get("OPAQUE_AES_256_GCM")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name() != "OPAQUE_AES_256_GCM" {
			t.Fatalf("expected name OPAQUE_AES_256_GCM, got %q", got.Name())
		}
		if got.cipher == nil {
			t.Fatal("expected non-nil cipher on registered opaque method")
		}
	})

	t.Run("OPAQUE_XCHACHA20_POLY1305 is retrievable via Get", func(t *testing.T) {
		t.Parallel()

		got, err := Get("OPAQUE_XCHACHA20_POLY1305")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name() != "OPAQUE_XCHACHA20_POLY1305" {
			t.Fatalf("expected name OPAQUE_XCHACHA20_POLY1305, got %q", got.Name())
		}
	})
}
