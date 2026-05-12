package rsassas

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
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

	t.Run("returns ErrKeyTypeMismatch when parsing ECDSA key as RSA", func(t *testing.T) {
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

	t.Run("round-trips an RSA private key and signs", func(t *testing.T) {
		t.Parallel()

		orig, err := RSASSA_PSS_using_SHA256.GenerateKey(2048)
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

		sig, err := RSASSA_PSS_using_SHA256.Sign(parsed, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := RSASSA_PSS_using_SHA256.Verify(&parsed.PublicKey, sig, []byte("data"))
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

	t.Run("returns ErrKeyTypeMismatch when parsing ECDSA public key as RSA", func(t *testing.T) {
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

	t.Run("round-trips an RSA public key and verifies", func(t *testing.T) {
		t.Parallel()

		priv, err := RSASSA_PSS_using_SHA256.GenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pubPEM, err := MarshalPublicKeyPEM(&priv.PublicKey)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		parsedPub, err := ParsePublicKeyPEM(pubPEM)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sig, err := RSASSA_PSS_using_SHA256.Sign(priv, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := RSASSA_PSS_using_SHA256.Verify(parsedPub, sig, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected verification to succeed")
		}
	})
}
