package ed25519

import (
	"crypto/ed25519"
	"errors"
	"testing"
)

func TestKey(t *testing.T) {
	t.Parallel()

	t.Run("generates key pair", func(t *testing.T) {
		t.Parallel()

		pub, priv, err := key()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(pub) != ed25519.PublicKeySize {
			t.Fatalf("expected %d bytes public key, got %d", ed25519.PublicKeySize, len(pub))
		}

		if len(priv) != ed25519.PrivateKeySize {
			t.Fatalf("expected %d bytes private key, got %d", ed25519.PrivateKeySize, len(priv))
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

		_, err := sign(Ed25519, nil, nil)
		if !errors.Is(err, ErrKeyIsNil) {
			t.Fatalf("expected ErrKeyIsNil, got %v", err)
		}
	})

	t.Run("returns error for invalid key length", func(t *testing.T) {
		t.Parallel()

		badKey := ed25519.PrivateKey(make([]byte, 10))

		_, err := sign(Ed25519, &badKey, []byte("data"))
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

		_, err := verify(Ed25519, nil, nil, nil)
		if !errors.Is(err, ErrKeyIsNil) {
			t.Fatalf("expected ErrKeyIsNil, got %v", err)
		}
	})

	t.Run("returns error for invalid key length", func(t *testing.T) {
		t.Parallel()

		badKey := ed25519.PublicKey(make([]byte, 10))

		_, err := verify(Ed25519, &badKey, nil, nil)
		if !errors.Is(err, ErrKeyLengthIsInvalid) {
			t.Fatalf("expected ErrKeyLengthIsInvalid, got %v", err)
		}
	})

	t.Run("returns error for invalid signature length", func(t *testing.T) {
		t.Parallel()

		_, priv, err := key()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pub := priv.Public().(ed25519.PublicKey)

		_, err = verify(Ed25519, &pub, []byte("short"), []byte("data"))
		if !errors.Is(err, ErrSignatureLengthInvalid) {
			t.Fatalf("expected ErrSignatureLengthInvalid, got %v", err)
		}
	})

	t.Run("returns false for invalid signature", func(t *testing.T) {
		t.Parallel()

		_, priv, err := key()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pub := priv.Public().(ed25519.PublicKey)
		fakeSig := make([]byte, ed25519.SignatureSize)

		ok, err := verify(Ed25519, &pub, fakeSig, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatal("expected false for invalid signature")
		}
	})
}
