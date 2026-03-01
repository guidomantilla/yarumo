package ecdsas

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"testing"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func TestNewMethod(t *testing.T) {
	t.Parallel()

	t.Run("creates method with defaults", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("test-ecdsa", crypto.SHA256, 32, elliptic.P256())

		if m == nil {
			t.Fatal("expected non-nil method")
		}

		if m.name != "test-ecdsa" {
			t.Fatalf("expected name 'test-ecdsa', got %q", m.name)
		}
	})

	t.Run("applies custom key function via option", func(t *testing.T) {
		t.Parallel()

		called := false
		custom := func(method *Method) (*ecdsa.PrivateKey, error) {
			called = true
			return nil, ErrMethodIsNil
		}

		m := NewMethod("custom", crypto.SHA256, 32, elliptic.P256(), WithKeyFn(custom))

		_, _ = m.GenerateKey()

		if !called {
			t.Fatal("expected custom keyFn to be called")
		}
	})
}

func TestMethod_Name(t *testing.T) {
	t.Parallel()

	t.Run("returns the method name", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("my-ecdsa", crypto.SHA256, 32, elliptic.P256())

		got := m.Name()
		if got != "my-ecdsa" {
			t.Fatalf("expected 'my-ecdsa', got %q", got)
		}
	})
}

func TestMethod_GenerateKey(t *testing.T) {
	t.Parallel()

	t.Run("generates ECDSA key pair", func(t *testing.T) {
		t.Parallel()

		key, err := ECDSA_with_SHA256_over_P256.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if key == nil {
			t.Fatal("expected non-nil key")
		}
	})

	t.Run("wraps error from keyFn", func(t *testing.T) {
		t.Parallel()

		failKey := func(method *Method) (*ecdsa.PrivateKey, error) {
			return nil, ErrMethodIsNil
		}

		m := NewMethod("fail", crypto.SHA256, 32, elliptic.P256(), WithKeyFn(failKey))

		_, err := m.GenerateKey()
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMethod_Sign(t *testing.T) {
	t.Parallel()

	t.Run("signs with RS format", func(t *testing.T) {
		t.Parallel()

		key, err := ECDSA_with_SHA256_over_P256.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := ECDSA_with_SHA256_over_P256.Sign(key, []byte("data"), RS)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(sig) != 64 {
			t.Fatalf("expected 64 bytes for RS format, got %d", len(sig))
		}
	})

	t.Run("signs with ASN1 format", func(t *testing.T) {
		t.Parallel()

		key, err := ECDSA_with_SHA256_over_P256.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := ECDSA_with_SHA256_over_P256.Sign(key, []byte("data"), ASN1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(sig) == 0 {
			t.Fatal("expected non-empty signature")
		}
	})

	t.Run("wraps error from signFn", func(t *testing.T) {
		t.Parallel()

		failSign := func(method *Method, key *ecdsa.PrivateKey, data ctypes.Bytes, format Format) (ctypes.Bytes, error) {
			return nil, ErrSignFailed
		}

		m := NewMethod("fail-sign", crypto.SHA256, 32, elliptic.P256(), WithSignFn(failSign))

		_, err := m.Sign(nil, []byte("data"), RS)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMethod_Verify(t *testing.T) {
	t.Parallel()

	t.Run("verifies RS signature", func(t *testing.T) {
		t.Parallel()

		key, err := ECDSA_with_SHA256_over_P256.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := ECDSA_with_SHA256_over_P256.Sign(key, []byte("data"), RS)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := ECDSA_with_SHA256_over_P256.Verify(&key.PublicKey, sig, []byte("data"), RS)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected verification to succeed")
		}
	})

	t.Run("verifies ASN1 signature", func(t *testing.T) {
		t.Parallel()

		key, err := ECDSA_with_SHA256_over_P256.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := ECDSA_with_SHA256_over_P256.Sign(key, []byte("data"), ASN1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := ECDSA_with_SHA256_over_P256.Verify(&key.PublicKey, sig, []byte("data"), ASN1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected verification to succeed")
		}
	})

	t.Run("rejects invalid signature", func(t *testing.T) {
		t.Parallel()

		key, err := ECDSA_with_SHA256_over_P256.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wrongSig := make([]byte, 64)

		ok, err := ECDSA_with_SHA256_over_P256.Verify(&key.PublicKey, wrongSig, []byte("data"), RS)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatal("expected verification to fail")
		}
	})

	t.Run("wraps error from verifyFn", func(t *testing.T) {
		t.Parallel()

		failVerify := func(method *Method, key *ecdsa.PublicKey, sig ctypes.Bytes, data ctypes.Bytes, format Format) (bool, error) {
			return false, ErrFormatUnsupported
		}

		m := NewMethod("fail-verify", crypto.SHA256, 32, elliptic.P256(), WithVerifyFn(failVerify))

		_, err := m.Verify(nil, nil, nil, RS)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
