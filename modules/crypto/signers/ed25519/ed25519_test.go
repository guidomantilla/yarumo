package ed25519

import (
	"crypto/ed25519"
	"testing"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func TestNewMethod(t *testing.T) {
	t.Parallel()

	t.Run("creates method with defaults", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("test-ed25519")

		if m == nil {
			t.Fatal("expected non-nil method")
		}

		if m.name != "test-ed25519" {
			t.Fatalf("expected 'test-ed25519', got %q", m.name)
		}
	})

	t.Run("applies custom key function via option", func(t *testing.T) {
		t.Parallel()

		called := false
		custom := func() (ed25519.PublicKey, ed25519.PrivateKey, error) {
			called = true
			return nil, nil, ErrMethodIsNil
		}

		m := NewMethod("custom", WithKeyFn(custom))

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

		m := NewMethod("my-ed25519")

		got := m.Name()
		if got != "my-ed25519" {
			t.Fatalf("expected 'my-ed25519', got %q", got)
		}
	})
}

func TestMethod_GenerateKey(t *testing.T) {
	t.Parallel()

	t.Run("generates Ed25519 key pair", func(t *testing.T) {
		t.Parallel()

		key, err := Ed25519.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(key) != ed25519.PrivateKeySize {
			t.Fatalf("expected %d bytes, got %d", ed25519.PrivateKeySize, len(key))
		}
	})

	t.Run("wraps error from keyFn", func(t *testing.T) {
		t.Parallel()

		failKey := func() (ed25519.PublicKey, ed25519.PrivateKey, error) {
			return nil, nil, ErrMethodIsNil
		}

		m := NewMethod("fail", WithKeyFn(failKey))

		_, err := m.GenerateKey()
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMethod_Sign(t *testing.T) {
	t.Parallel()

	t.Run("signs data", func(t *testing.T) {
		t.Parallel()

		k, err := Ed25519.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := Ed25519.Sign(&k, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(sig) != ed25519.SignatureSize {
			t.Fatalf("expected %d bytes, got %d", ed25519.SignatureSize, len(sig))
		}
	})

	t.Run("wraps error from signFn", func(t *testing.T) {
		t.Parallel()

		failSign := func(method *Method, key *ed25519.PrivateKey, data ctypes.Bytes) (ctypes.Bytes, error) {
			return nil, ErrKeyIsNil
		}

		m := NewMethod("fail-sign", WithSignFn(failSign))

		_, err := m.Sign(nil, []byte("data"))
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMethod_Verify(t *testing.T) {
	t.Parallel()

	t.Run("verifies valid signature", func(t *testing.T) {
		t.Parallel()

		k, err := Ed25519.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := Ed25519.Sign(&k, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pub := k.Public().(ed25519.PublicKey)

		ok, err := Ed25519.Verify(&pub, sig, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected verification to succeed")
		}
	})

	t.Run("wraps error from verifyFn", func(t *testing.T) {
		t.Parallel()

		failVerify := func(method *Method, key *ed25519.PublicKey, sig ctypes.Bytes, data ctypes.Bytes) (bool, error) {
			return false, ErrKeyIsNil
		}

		m := NewMethod("fail-verify", WithVerifyFn(failVerify))

		_, err := m.Verify(nil, nil, nil)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
