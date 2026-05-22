package kdfs

import (
	"crypto"
	"encoding/hex"
	"errors"
	"testing"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()

	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("invalid hex: %v", err)
	}

	return b
}

func TestNewMethod(t *testing.T) {
	t.Parallel()

	t.Run("creates HKDF method with defaults", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("custom-hkdf", crypto.SHA256)

		if m == nil {
			t.Fatal("expected non-nil method")
		}

		if m.name != "custom-hkdf" {
			t.Fatalf("expected name 'custom-hkdf', got %q", m.name)
		}

		if m.kind != crypto.SHA256 {
			t.Fatalf("expected SHA256, got %v", m.kind)
		}
	})

	t.Run("creates PBKDF2 method via option", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("custom-pbkdf2", crypto.SHA256, WithPbkdf2Iterations(1000))

		if m.pbkdf2Params == nil {
			t.Fatal("expected pbkdf2Params to be set")
		}

		if m.pbkdf2Params.iterations != 1000 {
			t.Fatalf("expected iterations=1000, got %d", m.pbkdf2Params.iterations)
		}
	})

	t.Run("creates Scrypt method via option", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("custom-scrypt", 0, WithScryptParams(1024, 8, 1))

		if m.scryptParams == nil {
			t.Fatal("expected scryptParams to be set")
		}

		if m.scryptParams.n != 1024 {
			t.Fatalf("expected n=1024, got %d", m.scryptParams.n)
		}
	})
}

func TestMethod_Name(t *testing.T) {
	t.Parallel()

	t.Run("returns the method name", func(t *testing.T) {
		t.Parallel()

		got := HKDF_with_SHA256.Name()
		if got != "HKDF_with_SHA256" {
			t.Fatalf("expected 'HKDF_with_SHA256', got %q", got)
		}
	})
}

func TestMethod_Derive_HKDF_RFC5869_TestCase1(t *testing.T) {
	t.Parallel()

	// RFC 5869 §A.1 Test Case 1 — basic test vector for HKDF-SHA-256.
	// https://datatracker.ietf.org/doc/html/rfc5869#appendix-A.1
	t.Run("RFC 5869 Test Case 1 (HKDF-SHA-256)", func(t *testing.T) {
		t.Parallel()

		ikm := mustHex(t, "0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b")
		salt := mustHex(t, "000102030405060708090a0b0c")
		info := mustHex(t, "f0f1f2f3f4f5f6f7f8f9")
		expected := mustHex(t, "3cb25f25faacd57a90434f64d0362f2a2d2d0a90cf1a5a4c5db02d56ecc4c5bf34007208d5b887185865")

		got, err := HKDF_with_SHA256.Derive(ikm, salt, info, 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if hex.EncodeToString(got) != hex.EncodeToString(expected) {
			t.Fatalf("expected %s, got %s", hex.EncodeToString(expected), hex.EncodeToString(got))
		}
	})
}

func TestMethod_Derive_HKDF_RFC5869_TestCase2(t *testing.T) {
	t.Parallel()

	// RFC 5869 §A.2 Test Case 2 — HKDF-SHA-256 with longer inputs.
	t.Run("RFC 5869 Test Case 2 (HKDF-SHA-256 long)", func(t *testing.T) {
		t.Parallel()

		ikm := mustHex(t, "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f404142434445464748494a4b4c4d4e4f")
		salt := mustHex(t, "606162636465666768696a6b6c6d6e6f707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f909192939495969798999a9b9c9d9e9fa0a1a2a3a4a5a6a7a8a9aaabacadaeaf")
		info := mustHex(t, "b0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c8c9cacbcccdcecfd0d1d2d3d4d5d6d7d8d9dadbdcdddedfe0e1e2e3e4e5e6e7e8e9eaebecedeeeff0f1f2f3f4f5f6f7f8f9fafbfcfdfeff")
		expected := mustHex(t, "b11e398dc80327a1c8e7f78c596a49344f012eda2d4efad8a050cc4c19afa97c59045a99cac7827271cb41c65e590e09da3275600c2f09b8367793a9aca3db71cc30c58179ec3e87c14c01d5c1f3434f1d87")

		got, err := HKDF_with_SHA256.Derive(ikm, salt, info, 82)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if hex.EncodeToString(got) != hex.EncodeToString(expected) {
			t.Fatalf("expected %s, got %s", hex.EncodeToString(expected), hex.EncodeToString(got))
		}
	})
}

func TestMethod_Derive_HKDF_RFC5869_TestCase3(t *testing.T) {
	t.Parallel()

	// RFC 5869 §A.3 Test Case 3 — HKDF-SHA-256 with zero-length salt/info.
	t.Run("RFC 5869 Test Case 3 (HKDF-SHA-256 zero salt/info)", func(t *testing.T) {
		t.Parallel()

		ikm := mustHex(t, "0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b")
		expected := mustHex(t, "8da4e775a563c18f715f802a063c5a31b8a11f5c5ee1879ec3454e5f3c738d2d9d201395faa4b61a96c8")

		got, err := HKDF_with_SHA256.Derive(ikm, []byte{}, []byte{}, 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if hex.EncodeToString(got) != hex.EncodeToString(expected) {
			t.Fatalf("expected %s, got %s", hex.EncodeToString(expected), hex.EncodeToString(got))
		}
	})
}

func TestMethod_Derive_HKDF_SHA384_RoundTrip(t *testing.T) {
	t.Parallel()

	// No standard RFC 5869 vector for SHA-384; verify deterministic output.
	t.Run("HKDF-SHA-384 produces deterministic 48-byte output", func(t *testing.T) {
		t.Parallel()

		ikm := []byte("input keying material for SHA-384 test")
		salt := []byte("salt-384")
		info := []byte("yarumo.test.sha384")

		a, err := HKDF_with_SHA384.Derive(ikm, salt, info, 48)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		b, err := HKDF_with_SHA384.Derive(ikm, salt, info, 48)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(a) != 48 {
			t.Fatalf("expected 48 bytes, got %d", len(a))
		}

		if hex.EncodeToString(a) != hex.EncodeToString(b) {
			t.Fatal("expected deterministic output across calls")
		}
	})
}

func TestMethod_Derive_HKDF_SHA512_RoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("HKDF-SHA-512 produces deterministic 64-byte output", func(t *testing.T) {
		t.Parallel()

		ikm := []byte("input keying material for SHA-512 test")
		salt := []byte("salt-512")
		info := []byte("yarumo.test.sha512")

		a, err := HKDF_with_SHA512.Derive(ikm, salt, info, 64)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		b, err := HKDF_with_SHA512.Derive(ikm, salt, info, 64)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(a) != 64 {
			t.Fatalf("expected 64 bytes, got %d", len(a))
		}

		if hex.EncodeToString(a) != hex.EncodeToString(b) {
			t.Fatal("expected deterministic output across calls")
		}
	})
}

func TestMethod_Derive_HKDF_InfoBinding(t *testing.T) {
	t.Parallel()

	// HKDF info argument must bind the derived key to a label: distinct
	// info values MUST yield distinct outputs for the same ikm/salt.
	t.Run("different info produces different output", func(t *testing.T) {
		t.Parallel()

		ikm := []byte("master")
		salt := []byte("salt")

		k1, err := HKDF_with_SHA256.Derive(ikm, salt, []byte("encryption"), 32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		k2, err := HKDF_with_SHA256.Derive(ikm, salt, []byte("authentication"), 32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if hex.EncodeToString(k1) == hex.EncodeToString(k2) {
			t.Fatal("expected distinct outputs for distinct info labels")
		}
	})
}

func TestMethod_Derive_PBKDF2_RFC8018_VerifiedVector(t *testing.T) {
	t.Parallel()

	// Known-good PBKDF2-HMAC-SHA-256 test vector (RFC 7914 §11 / IETF
	// errata-derived from RFC 8018) — password "passwd", salt "salt",
	// c=1, dkLen=64. Verified against multiple independent implementations.
	t.Run("PBKDF2-HMAC-SHA-256 known vector c=1", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("PBKDF2_TEST_SHA256_1", crypto.SHA256, WithPbkdf2Iterations(1))

		expected := mustHex(t, "55ac046e56e3089fec1691c22544b605f94185216dde0465e68b9d57c20dacbc49ca9cccf179b645991664b39d77ef317c71b845b1e30bd509112041d3a19783")

		got, err := m.Derive([]byte("passwd"), []byte("salt"), nil, 64)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if hex.EncodeToString(got) != hex.EncodeToString(expected) {
			t.Fatalf("expected %s, got %s", hex.EncodeToString(expected), hex.EncodeToString(got))
		}
	})

	t.Run("PBKDF2_with_SHA256 deterministic round-trip", func(t *testing.T) {
		t.Parallel()

		password := []byte("correct horse battery staple")
		salt := []byte("public-salt-1234")

		a, err := PBKDF2_with_SHA256.Derive(password, salt, nil, 32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		b, err := PBKDF2_with_SHA256.Derive(password, salt, nil, 32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if hex.EncodeToString(a) != hex.EncodeToString(b) {
			t.Fatal("expected deterministic output across calls")
		}

		if len(a) != 32 {
			t.Fatalf("expected 32 bytes, got %d", len(a))
		}
	})

	t.Run("PBKDF2_with_SHA512 deterministic round-trip", func(t *testing.T) {
		t.Parallel()

		password := []byte("correct horse battery staple")
		salt := []byte("public-salt-1234")

		a, err := PBKDF2_with_SHA512.Derive(password, salt, nil, 64)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		b, err := PBKDF2_with_SHA512.Derive(password, salt, nil, 64)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if hex.EncodeToString(a) != hex.EncodeToString(b) {
			t.Fatal("expected deterministic output across calls")
		}

		if len(a) != 64 {
			t.Fatalf("expected 64 bytes, got %d", len(a))
		}
	})
}

func TestMethod_Derive_Scrypt_RFC7914_TestVector(t *testing.T) {
	t.Parallel()

	// RFC 7914 §12 Test Vector 1: password "", salt "", N=16, r=1, p=1,
	// dkLen=64. Verified against the reference implementation.
	// https://datatracker.ietf.org/doc/html/rfc7914#section-12
	t.Run("RFC 7914 Test Vector 1 (N=16, r=1, p=1)", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("Scrypt_TEST_V1", 0, WithScryptParams(16, 1, 1))

		expected := mustHex(t, "77d6576238657b203b19ca42c18a0497f16b4844e3074ae8dfdffa3fede21442fcd0069ded0948f8326a753a0fc81f17e8d3e0fb2e0d3628cf35e20c38d18906")

		got, err := m.Derive([]byte(""), []byte(""), nil, 64)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if hex.EncodeToString(got) != hex.EncodeToString(expected) {
			t.Fatalf("expected %s, got %s", hex.EncodeToString(expected), hex.EncodeToString(got))
		}
	})

	t.Run("RFC 7914 Test Vector 2 (password=password, N=1024)", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("Scrypt_TEST_V2", 0, WithScryptParams(1024, 8, 16))

		// RFC 7914 §12 Test Vector 2: password "password", salt "NaCl",
		// N=1024, r=8, p=16, dkLen=64.
		expected := mustHex(t, "fdbabe1c9d3472007856e7190d01e9fe7c6ad7cbc8237830e77376634b3731622eaf30d92e22a3886ff109279d9830dac727afb94a83ee6d8360cbdfa2cc0640")

		got, err := m.Derive([]byte("password"), []byte("NaCl"), nil, 64)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if hex.EncodeToString(got) != hex.EncodeToString(expected) {
			t.Fatalf("expected %s, got %s", hex.EncodeToString(expected), hex.EncodeToString(got))
		}
	})
}

func TestMethod_Derive_RejectsInvalidInput(t *testing.T) {
	t.Parallel()

	t.Run("HKDF rejects nil secret", func(t *testing.T) {
		t.Parallel()

		_, err := HKDF_with_SHA256.Derive(nil, []byte("salt"), []byte("info"), 32)
		if err == nil {
			t.Fatal("expected error for nil secret")
		}

		if !errors.Is(err, ErrSecretIsNil) {
			t.Fatalf("expected ErrSecretIsNil chain, got %v", err)
		}
	})

	t.Run("HKDF rejects zero length", func(t *testing.T) {
		t.Parallel()

		_, err := HKDF_with_SHA256.Derive([]byte("ikm"), []byte("salt"), []byte("info"), 0)
		if err == nil {
			t.Fatal("expected error for zero length")
		}

		if !errors.Is(err, ErrLengthInvalid) {
			t.Fatalf("expected ErrLengthInvalid chain, got %v", err)
		}
	})

	t.Run("HKDF rejects negative length", func(t *testing.T) {
		t.Parallel()

		_, err := HKDF_with_SHA256.Derive([]byte("ikm"), []byte("salt"), []byte("info"), -1)
		if err == nil {
			t.Fatal("expected error for negative length")
		}

		if !errors.Is(err, ErrLengthInvalid) {
			t.Fatalf("expected ErrLengthInvalid chain, got %v", err)
		}
	})

	t.Run("PBKDF2 rejects nil salt", func(t *testing.T) {
		t.Parallel()

		_, err := PBKDF2_with_SHA256.Derive([]byte("password"), nil, nil, 32)
		if err == nil {
			t.Fatal("expected error for nil salt")
		}

		if !errors.Is(err, ErrSaltIsNil) {
			t.Fatalf("expected ErrSaltIsNil chain, got %v", err)
		}
	})

	t.Run("Scrypt rejects nil salt", func(t *testing.T) {
		t.Parallel()

		_, err := Scrypt_KDF.Derive([]byte("password"), nil, nil, 32)
		if err == nil {
			t.Fatal("expected error for nil salt")
		}

		if !errors.Is(err, ErrSaltIsNil) {
			t.Fatalf("expected ErrSaltIsNil chain, got %v", err)
		}
	})

	t.Run("wraps error chain via ErrDerive", func(t *testing.T) {
		t.Parallel()

		_, err := HKDF_with_SHA256.Derive(nil, []byte("salt"), []byte("info"), 32)
		if err == nil {
			t.Fatal("expected error")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !errors.Is(err, ErrDeriveFailed) {
			t.Fatalf("expected ErrDeriveFailed in chain, got %v", err)
		}
	})
}

func TestMethod_Derive_AEADKeySize(t *testing.T) {
	t.Parallel()

	// HKDF is commonly used to derive a 32-byte key for AES-256-GCM /
	// ChaCha20-Poly1305 from a high-entropy master secret.
	t.Run("derives 32-byte AEAD key via HKDF-SHA-256", func(t *testing.T) {
		t.Parallel()

		master := ctypes.Bytes("master-secret-from-ecdh-or-tls-handshake")
		salt := ctypes.Bytes("yarumo-app-v1")
		info := ctypes.Bytes("yarumo.aead.key.v1")

		key, err := HKDF_with_SHA256.Derive(master, salt, info, 32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(key) != 32 {
			t.Fatalf("expected 32-byte AEAD key, got %d", len(key))
		}
	})
}
