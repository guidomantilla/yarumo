package ed25519

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
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

func TestMarshalPrivateKeyPEM(t *testing.T) {
	t.Parallel()

	t.Run("returns error for nil key", func(t *testing.T) {
		t.Parallel()

		_, err := MarshalPrivateKeyPEM(nil)

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !errors.Is(err, ErrKeyIsNil) {
			t.Fatalf("expected ErrKeyIsNil, got %v", err)
		}
	})
}

func TestParsePrivateKeyPEM(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrPEMDecodeFailed on malformed PEM", func(t *testing.T) {
		t.Parallel()

		_, err := ParsePrivateKeyPEM([]byte("not a pem"))

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !errors.Is(err, ErrPEMDecodeFailed) {
			t.Fatalf("expected ErrPEMDecodeFailed, got %v", err)
		}
	})

	t.Run("returns ErrPEMBlockTypeMismatch on wrong block type", func(t *testing.T) {
		t.Parallel()

		block := &pem.Block{Type: "CERTIFICATE", Bytes: []byte("dummy")}

		_, err := ParsePrivateKeyPEM(pem.EncodeToMemory(block))
		if !errors.Is(err, ErrPEMBlockTypeMismatch) {
			t.Fatalf("expected ErrPEMBlockTypeMismatch, got %v", err)
		}
	})

	t.Run("returns ErrKeyTypeMismatch when parsing ECDSA key as Ed25519", func(t *testing.T) {
		t.Parallel()

		ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		der, err := x509.MarshalPKCS8PrivateKey(ecKey)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ecPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})

		_, err = ParsePrivateKeyPEM(ecPEM)
		if !errors.Is(err, ErrKeyTypeMismatch) {
			t.Fatalf("expected ErrKeyTypeMismatch, got %v", err)
		}
	})
}

func TestPrivateKeyPEMRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("round-trips an Ed25519 private key and signs", func(t *testing.T) {
		t.Parallel()

		orig, err := Ed25519.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pemBytes, err := MarshalPrivateKeyPEM(orig)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		parsed, err := ParsePrivateKeyPEM(pemBytes)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := Ed25519.Sign(&parsed, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pub := parsed.Public().(ed25519.PublicKey)

		ok, err := Ed25519.Verify(&pub, sig, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected verification to succeed")
		}
	})
}

func TestMarshalPublicKeyPEM(t *testing.T) {
	t.Parallel()

	t.Run("returns error for nil key", func(t *testing.T) {
		t.Parallel()

		_, err := MarshalPublicKeyPEM(nil)
		if !errors.Is(err, ErrKeyIsNil) {
			t.Fatalf("expected ErrKeyIsNil, got %v", err)
		}
	})
}

func TestParsePublicKeyPEM(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrPEMDecodeFailed on malformed PEM", func(t *testing.T) {
		t.Parallel()

		_, err := ParsePublicKeyPEM([]byte("garbage"))
		if !errors.Is(err, ErrPEMDecodeFailed) {
			t.Fatalf("expected ErrPEMDecodeFailed, got %v", err)
		}
	})

	t.Run("returns ErrPEMBlockTypeMismatch on wrong block type", func(t *testing.T) {
		t.Parallel()

		block := &pem.Block{Type: "CERTIFICATE", Bytes: []byte("dummy")}

		_, err := ParsePublicKeyPEM(pem.EncodeToMemory(block))
		if !errors.Is(err, ErrPEMBlockTypeMismatch) {
			t.Fatalf("expected ErrPEMBlockTypeMismatch, got %v", err)
		}
	})

	t.Run("returns ErrKeyTypeMismatch when parsing ECDSA public key as Ed25519", func(t *testing.T) {
		t.Parallel()

		ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		der, err := x509.MarshalPKIXPublicKey(&ecKey.PublicKey)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ecPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})

		_, err = ParsePublicKeyPEM(ecPEM)
		if !errors.Is(err, ErrKeyTypeMismatch) {
			t.Fatalf("expected ErrKeyTypeMismatch, got %v", err)
		}
	})
}

func TestPublicKeyPEMRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("round-trips an Ed25519 public key and verifies", func(t *testing.T) {
		t.Parallel()

		priv, err := Ed25519.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pub := priv.Public().(ed25519.PublicKey)

		pubPEM, err := MarshalPublicKeyPEM(pub)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		parsedPub, err := ParsePublicKeyPEM(pubPEM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := Ed25519.Sign(&priv, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := Ed25519.Verify(&parsedPub, sig, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected verification to succeed")
		}
	})
}
