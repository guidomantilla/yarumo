package rsapss

import (
	"crypto"
	"crypto/rsa"
	"testing"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func TestNewMethod(t *testing.T) {
	t.Parallel()

	t.Run("creates method with defaults", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("test-rsapss", crypto.SHA256, rsa.PSSSaltLengthEqualsHash, []int{2048})

		if m == nil {
			t.Fatal("expected non-nil method")
		}

		if m.name != "test-rsapss" {
			t.Fatalf("expected 'test-rsapss', got %q", m.name)
		}
	})
}

func TestMethod_Name(t *testing.T) {
	t.Parallel()

	t.Run("returns the method name", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("my-rsapss", crypto.SHA256, rsa.PSSSaltLengthEqualsHash, []int{2048})

		got := m.Name()
		if got != "my-rsapss" {
			t.Fatalf("expected 'my-rsapss', got %q", got)
		}
	})
}

func TestMethod_GenerateKey(t *testing.T) {
	t.Parallel()

	t.Run("generates RSA key", func(t *testing.T) {
		t.Parallel()

		key, err := RSASSA_PSS_using_SHA256.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if key == nil {
			t.Fatal("expected non-nil key")
		}

		if key.N.BitLen() != 2048 {
			t.Fatalf("expected 2048-bit key, got %d", key.N.BitLen())
		}
	})

	t.Run("rejects disallowed key size", func(t *testing.T) {
		t.Parallel()

		_, err := RSASSA_PSS_using_SHA256.GenerateKey(1024)
		if err == nil {
			t.Fatal("expected error for disallowed key size")
		}
	})

	t.Run("wraps error from keyFn", func(t *testing.T) {
		t.Parallel()

		failKey := func(bits int) (*rsa.PrivateKey, error) {
			return nil, ErrMethodIsNil
		}

		m := NewMethod("fail", crypto.SHA256, rsa.PSSSaltLengthEqualsHash, []int{2048}, WithKeyFn(failKey))

		_, err := m.GenerateKey(2048)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMethod_Sign(t *testing.T) {
	t.Parallel()

	t.Run("signs data", func(t *testing.T) {
		t.Parallel()

		key, err := RSASSA_PSS_using_SHA256.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := RSASSA_PSS_using_SHA256.Sign(key, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(sig) == 0 {
			t.Fatal("expected non-empty signature")
		}
	})

	t.Run("wraps error from signFn", func(t *testing.T) {
		t.Parallel()

		failSign := func(method *Method, key *rsa.PrivateKey, data ctypes.Bytes) (ctypes.Bytes, error) {
			return nil, ErrSignFailed
		}

		m := NewMethod("fail-sign", crypto.SHA256, rsa.PSSSaltLengthEqualsHash, []int{2048}, WithSignFn(failSign))

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

		key, err := RSASSA_PSS_using_SHA256.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := RSASSA_PSS_using_SHA256.Sign(key, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := RSASSA_PSS_using_SHA256.Verify(&key.PublicKey, sig, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected verification to succeed")
		}
	})

	t.Run("rejects invalid signature", func(t *testing.T) {
		t.Parallel()

		key, err := RSASSA_PSS_using_SHA256.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := RSASSA_PSS_using_SHA256.Verify(&key.PublicKey, []byte("bad"), []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatal("expected verification to fail")
		}
	})

	t.Run("wraps error from verifyFn", func(t *testing.T) {
		t.Parallel()

		failVerify := func(method *Method, key *rsa.PublicKey, sig ctypes.Bytes, data ctypes.Bytes) (bool, error) {
			return false, ErrVerifyFailed
		}

		m := NewMethod("fail-verify", crypto.SHA256, rsa.PSSSaltLengthEqualsHash, []int{2048}, WithVerifyFn(failVerify))

		_, err := m.Verify(nil, nil, nil)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
