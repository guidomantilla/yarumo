package ecdsas

import (
	"crypto/elliptic"
	"errors"
	"testing"
)

func TestSign(t *testing.T) {
	t.Parallel()

	t.Run("returns error for nil method", func(t *testing.T) {
		t.Parallel()

		_, err := sign(nil, nil, nil, RS)
		if !errors.Is(err, ErrMethodIsNil) {
			t.Fatalf("expected ErrMethodIsNil, got %v", err)
		}
	})

	t.Run("returns error for nil key", func(t *testing.T) {
		t.Parallel()

		_, err := sign(ECDSA_with_SHA256_over_P256, nil, nil, RS)
		if !errors.Is(err, ErrKeyIsNil) {
			t.Fatalf("expected ErrKeyIsNil, got %v", err)
		}
	})

	t.Run("returns error for wrong curve", func(t *testing.T) {
		t.Parallel()

		k, err := key(ECDSA_with_SHA256_over_P256)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m521 := ECDSA_with_SHA512_over_P521

		_, err = sign(m521, k, []byte("data"), RS)
		if !errors.Is(err, ErrKeyCurveIsInvalid) {
			t.Fatalf("expected ErrKeyCurveIsInvalid, got %v", err)
		}
	})

	t.Run("returns error for unsupported format", func(t *testing.T) {
		t.Parallel()

		k, err := key(ECDSA_with_SHA256_over_P256)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = sign(ECDSA_with_SHA256_over_P256, k, []byte("data"), Format(99))
		if !errors.Is(err, ErrFormatUnsupported) {
			t.Fatalf("expected ErrFormatUnsupported, got %v", err)
		}
	})
}

func TestVerify(t *testing.T) {
	t.Parallel()

	t.Run("returns error for nil method", func(t *testing.T) {
		t.Parallel()

		_, err := verify(nil, nil, nil, nil, RS)
		if !errors.Is(err, ErrMethodIsNil) {
			t.Fatalf("expected ErrMethodIsNil, got %v", err)
		}
	})

	t.Run("returns error for nil key", func(t *testing.T) {
		t.Parallel()

		_, err := verify(ECDSA_with_SHA256_over_P256, nil, nil, nil, RS)
		if !errors.Is(err, ErrKeyIsNil) {
			t.Fatalf("expected ErrKeyIsNil, got %v", err)
		}
	})

	t.Run("returns error for wrong curve", func(t *testing.T) {
		t.Parallel()

		k, err := key(ECDSA_with_SHA256_over_P256)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m521 := ECDSA_with_SHA512_over_P521

		_, err = verify(m521, &k.PublicKey, nil, nil, RS)
		if !errors.Is(err, ErrKeyCurveIsInvalid) {
			t.Fatalf("expected ErrKeyCurveIsInvalid, got %v", err)
		}
	})

	t.Run("returns error for invalid RS signature length", func(t *testing.T) {
		t.Parallel()

		k, err := key(ECDSA_with_SHA256_over_P256)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = verify(ECDSA_with_SHA256_over_P256, &k.PublicKey, []byte("short"), []byte("data"), RS)
		if !errors.Is(err, ErrSignatureInvalid) {
			t.Fatalf("expected ErrSignatureInvalid, got %v", err)
		}
	})

	t.Run("returns error for unsupported format", func(t *testing.T) {
		t.Parallel()

		k, err := key(ECDSA_with_SHA256_over_P256)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = verify(ECDSA_with_SHA256_over_P256, &k.PublicKey, nil, []byte("data"), Format(99))
		if !errors.Is(err, ErrFormatUnsupported) {
			t.Fatalf("expected ErrFormatUnsupported, got %v", err)
		}
	})

	t.Run("returns false for invalid ASN1 signature", func(t *testing.T) {
		t.Parallel()

		k, err := key(ECDSA_with_SHA256_over_P256)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := verify(ECDSA_with_SHA256_over_P256, &k.PublicKey, []byte("bad-asn1"), []byte("data"), ASN1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatal("expected false for invalid ASN1 signature")
		}
	})
}

func TestKey(t *testing.T) {
	t.Parallel()

	t.Run("generates key for P256", func(t *testing.T) {
		t.Parallel()

		k, err := key(ECDSA_with_SHA256_over_P256)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if k.Curve != elliptic.P256() {
			t.Fatal("expected P256 curve")
		}
	})
}
