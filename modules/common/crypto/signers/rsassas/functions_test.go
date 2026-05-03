package rsassas

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

	t.Run("returns error for unsupported padding", func(t *testing.T) {
		t.Parallel()

		k, err := key(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := &Method{
			padding:         Padding(99),
			kind:            RSASSA_PSS_using_SHA256.kind,
			allowedKeySizes: []int{2048},
		}

		_, err = sign(m, k, []byte("data"))
		if !errors.Is(err, ErrPaddingNotSupported) {
			t.Fatalf("expected ErrPaddingNotSupported, got %v", err)
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

	t.Run("returns false for invalid PSS signature", func(t *testing.T) {
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

	t.Run("returns false for invalid PKCS1v15 signature", func(t *testing.T) {
		t.Parallel()

		k, err := key(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := verify(RSASSA_PKCS1v15_using_SHA256, &k.PublicKey, []byte("bad-sig"), []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatal("expected false for invalid signature")
		}
	})

	t.Run("returns error for unsupported padding", func(t *testing.T) {
		t.Parallel()

		k, err := key(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := &Method{
			padding:         Padding(99),
			kind:            RSASSA_PSS_using_SHA256.kind,
			allowedKeySizes: []int{2048},
		}

		_, err = verify(m, &k.PublicKey, []byte("sig"), []byte("data"))
		if !errors.Is(err, ErrPaddingNotSupported) {
			t.Fatalf("expected ErrPaddingNotSupported, got %v", err)
		}
	})
}
