package ecdsas

import (
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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

	t.Run("returns ErrKeyTypeMismatch when parsing RSA key as ECDSA", func(t *testing.T) {
		t.Parallel()

		rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		der, err := x509.MarshalPKCS8PrivateKey(rsaKey)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		rsaPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})

		_, err = ParsePrivateKeyPEM(rsaPEM)
		if !errors.Is(err, ErrKeyTypeMismatch) {
			t.Fatalf("expected ErrKeyTypeMismatch, got %v", err)
		}
	})
}

func TestPrivateKeyPEMRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("round-trips an ECDSA private key and signs", func(t *testing.T) {
		t.Parallel()

		orig, err := ECDSA_with_SHA256_over_P256.GenerateKey()
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

		sig, err := ECDSA_with_SHA256_over_P256.Sign(parsed, []byte("data"), ASN1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := ECDSA_with_SHA256_over_P256.Verify(&parsed.PublicKey, sig, []byte("data"), ASN1)
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

	t.Run("returns ErrKeyTypeMismatch when parsing RSA public key as ECDSA", func(t *testing.T) {
		t.Parallel()

		rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		der, err := x509.MarshalPKIXPublicKey(&rsaKey.PublicKey)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		rsaPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})

		_, err = ParsePublicKeyPEM(rsaPEM)
		if !errors.Is(err, ErrKeyTypeMismatch) {
			t.Fatalf("expected ErrKeyTypeMismatch, got %v", err)
		}
	})
}

func TestPublicKeyPEMRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("round-trips an ECDSA public key and verifies", func(t *testing.T) {
		t.Parallel()

		priv, err := ECDSA_with_SHA256_over_P256.GenerateKey()
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

		sig, err := ECDSA_with_SHA256_over_P256.Sign(priv, []byte("data"), ASN1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := ECDSA_with_SHA256_over_P256.Verify(parsedPub, sig, []byte("data"), ASN1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected verification to succeed")
		}

		var _ = parsedPub
	})
}

func TestDigest_ByName(t *testing.T) {
	t.Parallel()

	t.Run("signs and verifies round trip", func(t *testing.T) {
		t.Parallel()

		method, err := Get("ECDSA_with_SHA256_over_P256")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		priv, err := method.GenerateKey()
		if err != nil {
			t.Fatalf("unexpected error generating key: %v", err)
		}

		privPEM, err := MarshalPrivateKeyPEM(priv)
		if err != nil {
			t.Fatalf("unexpected error marshalling private key: %v", err)
		}

		pubPEM, err := MarshalPublicKeyPEM(&priv.PublicKey)
		if err != nil {
			t.Fatalf("unexpected error marshalling public key: %v", err)
		}

		sig, err := Digest("ECDSA_with_SHA256_over_P256", privPEM, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := Validate("ECDSA_with_SHA256_over_P256", pubPEM, sig, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected verification to succeed")
		}
	})

	t.Run("Digest returns domain error for unknown name", func(t *testing.T) {
		t.Parallel()

		got, err := Digest("UNKNOWN", []byte("k"), []byte("data"))
		if err == nil {
			t.Fatal("expected error for unknown name")
		}

		if got != nil {
			t.Fatalf("expected nil bytes, got %v", got)
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("Digest returns PEM codec error for invalid key", func(t *testing.T) {
		t.Parallel()

		_, err := Digest("ECDSA_with_SHA256_over_P256", []byte("not a pem"), []byte("data"))
		if err == nil {
			t.Fatal("expected error for invalid PEM")
		}

		if !errors.Is(err, ErrPEMDecodeFailed) {
			t.Fatalf("expected ErrPEMDecodeFailed, got %v", err)
		}
	})

	t.Run("Validate returns domain error for unknown name", func(t *testing.T) {
		t.Parallel()

		ok, err := Validate("UNKNOWN", []byte("k"), []byte("sig"), []byte("data"))
		if err == nil {
			t.Fatal("expected error for unknown name")
		}

		if ok {
			t.Fatal("expected ok=false on error")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})

	t.Run("Validate returns PEM codec error for invalid key", func(t *testing.T) {
		t.Parallel()

		_, err := Validate("ECDSA_with_SHA256_over_P256", []byte("not a pem"), []byte("sig"), []byte("data"))
		if err == nil {
			t.Fatal("expected error for invalid PEM")
		}

		if !errors.Is(err, ErrPEMDecodeFailed) {
			t.Fatalf("expected ErrPEMDecodeFailed, got %v", err)
		}
	})
}
