package rsapss

import (
	"errors"
	"testing"
)

func TestKey(t *testing.T) {
	t.Parallel()

	t.Run("generates RSA key", func(t *testing.T) {
		t.Parallel()

		k, err := key(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if k.N.BitLen() != 2048 {
			t.Fatalf("expected 2048-bit key, got %d", k.N.BitLen())
		}
	})
}

func TestSign(t *testing.T) {
	t.Parallel()

	t.Run("returns error for nil method", func(t *testing.T) {
		t.Parallel()

		_, err := sign(nil, nil, nil)
		if !errors.Is(err, ErrMethodIsNil) {
			t.Fatalf("expected ErrMethodIsNil, got %v", err)
		}
	})

	t.Run("returns error for nil key", func(t *testing.T) {
		t.Parallel()

		_, err := sign(RSASSA_PSS_using_SHA256, nil, nil)
		if !errors.Is(err, ErrKeyIsNil) {
			t.Fatalf("expected ErrKeyIsNil, got %v", err)
		}
	})

	t.Run("returns error for invalid key size", func(t *testing.T) {
		t.Parallel()

		k, err := key(1024)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = sign(RSASSA_PSS_using_SHA256, k, []byte("data"))
		if !errors.Is(err, ErrKeyLengthIsInvalid) {
			t.Fatalf("expected ErrKeyLengthIsInvalid, got %v", err)
		}
	})
}

func TestVerify(t *testing.T) {
	t.Parallel()

	t.Run("returns error for nil method", func(t *testing.T) {
		t.Parallel()

		_, err := verify(nil, nil, nil, nil)
		if !errors.Is(err, ErrMethodIsNil) {
			t.Fatalf("expected ErrMethodIsNil, got %v", err)
		}
	})

	t.Run("returns error for nil key", func(t *testing.T) {
		t.Parallel()

		_, err := verify(RSASSA_PSS_using_SHA256, nil, nil, nil)
		if !errors.Is(err, ErrKeyIsNil) {
			t.Fatalf("expected ErrKeyIsNil, got %v", err)
		}
	})

	t.Run("returns error for invalid key size", func(t *testing.T) {
		t.Parallel()

		k, err := key(1024)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = verify(RSASSA_PSS_using_SHA256, &k.PublicKey, nil, nil)
		if !errors.Is(err, ErrKeyLengthIsInvalid) {
			t.Fatalf("expected ErrKeyLengthIsInvalid, got %v", err)
		}
	})

	t.Run("returns false for invalid signature (rsa.ErrVerification)", func(t *testing.T) {
		t.Parallel()

		k, err := key(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := verify(RSASSA_PSS_using_SHA256, &k.PublicKey, []byte("bad-sig"), []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatal("expected false for invalid signature")
		}
	})
}
