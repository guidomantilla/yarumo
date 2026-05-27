package rsassas

import (
	"crypto"
	"crypto/rsa"
	"testing"

	ctypes "github.com/guidomantilla/yarumo/core/common/types"
)

func TestNewMethod(t *testing.T) {
	t.Parallel()

	t.Run("creates PSS method with defaults", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("test-pss", crypto.SHA256, PSS, []int{2048})

		if m.name != "test-pss" {
			t.Fatalf("expected 'test-pss', got %q", m.name)
		}

		if m.padding != PSS {
			t.Fatalf("expected PSS padding, got %d", m.padding)
		}
	})

	t.Run("creates PKCS1v15 method with defaults", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("test-pkcs1v15", crypto.SHA256, PKCS1v15, []int{2048})

		if m.padding != PKCS1v15 {
			t.Fatalf("expected PKCS1v15 padding, got %d", m.padding)
		}
	})
}

func TestMethod_Name(t *testing.T) {
	t.Parallel()

	t.Run("returns the method name", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("my-rsa", crypto.SHA256, PSS, []int{2048})

		got := m.Name()
		if got != "my-rsa" {
			t.Fatalf("expected 'my-rsa', got %q", got)
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

		m := NewMethod("fail", crypto.SHA256, PSS, []int{2048}, WithKeyFn(failKey))

		_, err := m.GenerateKey(2048)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMethod_Sign(t *testing.T) {
	t.Parallel()

	t.Run("signs data with PSS", func(t *testing.T) {
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

	t.Run("signs data with PKCS1v15 SHA256", func(t *testing.T) {
		t.Parallel()

		key, err := RSASSA_PKCS1v15_using_SHA256.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := RSASSA_PKCS1v15_using_SHA256.Sign(key, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(sig) == 0 {
			t.Fatal("expected non-empty signature")
		}
	})

	t.Run("signs data with PKCS1v15 SHA384", func(t *testing.T) {
		t.Parallel()

		key, err := RSASSA_PKCS1v15_using_SHA384.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := RSASSA_PKCS1v15_using_SHA384.Sign(key, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(sig) == 0 {
			t.Fatal("expected non-empty signature")
		}
	})

	t.Run("signs data with PSS SHA384", func(t *testing.T) {
		t.Parallel()

		key, err := RSASSA_PSS_using_SHA384.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := RSASSA_PSS_using_SHA384.Sign(key, []byte("data"))
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

		m := NewMethod("fail-sign", crypto.SHA256, PSS, []int{2048}, WithSignFn(failSign))

		_, err := m.Sign(nil, []byte("data"))
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMethod_Verify(t *testing.T) {
	t.Parallel()

	t.Run("verifies valid PSS signature", func(t *testing.T) {
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

	t.Run("verifies valid PKCS1v15 SHA256 signature", func(t *testing.T) {
		t.Parallel()

		key, err := RSASSA_PKCS1v15_using_SHA256.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := RSASSA_PKCS1v15_using_SHA256.Sign(key, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := RSASSA_PKCS1v15_using_SHA256.Verify(&key.PublicKey, sig, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected verification to succeed")
		}
	})

	t.Run("verifies valid PKCS1v15 SHA384 signature", func(t *testing.T) {
		t.Parallel()

		key, err := RSASSA_PKCS1v15_using_SHA384.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := RSASSA_PKCS1v15_using_SHA384.Sign(key, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := RSASSA_PKCS1v15_using_SHA384.Verify(&key.PublicKey, sig, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected verification to succeed")
		}
	})

	t.Run("round-trips PSS SHA384 with 2048-bit key", func(t *testing.T) {
		t.Parallel()

		key, err := RSASSA_PSS_using_SHA384.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := RSASSA_PSS_using_SHA384.Sign(key, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := RSASSA_PSS_using_SHA384.Verify(&key.PublicKey, sig, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected verification to succeed")
		}
	})

	t.Run("rejects invalid PSS signature", func(t *testing.T) {
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

	t.Run("rejects invalid PKCS1v15 signature", func(t *testing.T) {
		t.Parallel()

		key, err := RSASSA_PKCS1v15_using_SHA256.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := RSASSA_PKCS1v15_using_SHA256.Verify(&key.PublicKey, []byte("bad"), []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatal("expected verification to fail")
		}
	})

	t.Run("PSS and PKCS1v15 signatures are not interchangeable", func(t *testing.T) {
		t.Parallel()

		key, err := RSASSA_PSS_using_SHA256.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := RSASSA_PSS_using_SHA256.Sign(key, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := RSASSA_PKCS1v15_using_SHA256.Verify(&key.PublicKey, sig, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatal("expected PSS signature to fail PKCS1v15 verification")
		}
	})

	t.Run("wraps error from verifyFn", func(t *testing.T) {
		t.Parallel()

		failVerify := func(method *Method, key *rsa.PublicKey, sig ctypes.Bytes, data ctypes.Bytes) (bool, error) {
			return false, ErrVerifyFailed
		}

		m := NewMethod("fail-verify", crypto.SHA256, PSS, []int{2048}, WithVerifyFn(failVerify))

		_, err := m.Verify(nil, nil, nil)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
