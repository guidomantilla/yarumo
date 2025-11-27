package macs

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"testing"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"
)

func TestHMAC_SHA256(t *testing.T) {
	key := []byte("key")
	msg := []byte("abc")

	hm := hmac.New(sha256.New, key)
	hm.Write(msg)
	want := hm.Sum(nil)

	got := HMAC_SHA256(key, msg)
	if !hmac.Equal(got, want) {
		t.Fatalf("HMAC_SHA256 mismatch")
	}
}

func TestHMAC_SHA3_256(t *testing.T) {
	key := []byte("key")
	msg := []byte("abc")

	hm := hmac.New(sha3.New256, key)
	hm.Write(msg)
	want := hm.Sum(nil)

	got := HMAC_SHA3_256(key, msg)
	if !hmac.Equal(got, want) {
		t.Fatalf("HMAC_SHA3_256 mismatch")
	}
}

func TestBLAKE2b_256_MAC(t *testing.T) {
	key := []byte("key")
	msg := []byte("abc")

	d, err := blake2b.New256(key)
	if err != nil {
		t.Fatalf("init blake2b-256: %v", err)
	}
	d.Write(msg)
	want := d.Sum(nil)

	got := BLAKE2b_256_MAC(key, msg)
	if !hmac.Equal(got, want) {
		t.Fatalf("BLAKE2b_256_MAC mismatch")
	}
}

func TestHMAC_SHA512(t *testing.T) {
	key := []byte("key")
	msg := []byte("abc")

	hm := hmac.New(sha512.New, key)
	hm.Write(msg)
	want := hm.Sum(nil)

	got := HMAC_SHA512(key, msg)
	if !hmac.Equal(got, want) {
		t.Fatalf("HMAC_SHA512 mismatch")
	}
}

func TestHMAC_SHA3_512(t *testing.T) {
	key := []byte("key")
	msg := []byte("abc")

	hm := hmac.New(sha3.New512, key)
	hm.Write(msg)
	want := hm.Sum(nil)

	got := HMAC_SHA3_512(key, msg)
	if !hmac.Equal(got, want) {
		t.Fatalf("HMAC_SHA3_512 mismatch")
	}
}

func TestBLAKE2b_512(t *testing.T) {
	key := []byte("key")
	msg := []byte("abc")

	d, err := blake2b.New512(key)
	if err != nil {
		t.Fatalf("init blake2b-512: %v", err)
	}
	d.Write(msg)
	want := d.Sum(nil)

	got := BLAKE2b_512_MAC(key, msg)
	if !hmac.Equal(got, want) {
		t.Fatalf("BLAKE2b_512_MAC mismatch")
	}
}

// Edge cases to reach 100% coverage
func TestEdgeCases(t *testing.T) {
	msg := []byte("data")

	t.Run("HMAC_SHA256 empty key", func(t *testing.T) {
		if got := HMAC_SHA256(nil, msg); got != nil {
			t.Fatalf("expected nil for empty key, got %v", got)
		}
	})
	t.Run("HMAC_SHA256 empty data", func(t *testing.T) {
		if got := HMAC_SHA256([]byte("k"), nil); got != nil {
			t.Fatalf("expected nil for empty data, got %v", got)
		}
	})

	t.Run("HMAC_SHA3_256 empty key", func(t *testing.T) {
		if got := HMAC_SHA3_256(nil, msg); got != nil {
			t.Fatalf("expected nil for empty key, got %v", got)
		}
	})
	t.Run("HMAC_SHA3_256 empty data", func(t *testing.T) {
		if got := HMAC_SHA3_256([]byte("k"), nil); got != nil {
			t.Fatalf("expected nil for empty data, got %v", got)
		}
	})

	t.Run("HMAC_SHA512 empty key", func(t *testing.T) {
		if got := HMAC_SHA512(nil, msg); got != nil {
			t.Fatalf("expected nil for empty key, got %v", got)
		}
	})
	t.Run("HMAC_SHA512 empty data", func(t *testing.T) {
		if got := HMAC_SHA512([]byte("k"), nil); got != nil {
			t.Fatalf("expected nil for empty data, got %v", got)
		}
	})

	t.Run("HMAC_SHA3_512 empty key", func(t *testing.T) {
		if got := HMAC_SHA3_512(nil, msg); got != nil {
			t.Fatalf("expected nil for empty key, got %v", got)
		}
	})
	t.Run("HMAC_SHA3_512 empty data", func(t *testing.T) {
		if got := HMAC_SHA3_512([]byte("k"), nil); got != nil {
			t.Fatalf("expected nil for empty data, got %v", got)
		}
	})

	t.Run("BLAKE2b_256 empty key", func(t *testing.T) {
		if got := BLAKE2b_256_MAC(nil, msg); got != nil {
			t.Fatalf("expected nil for empty key, got %v", got)
		}
	})
	t.Run("BLAKE2b_256 empty data", func(t *testing.T) {
		if got := BLAKE2b_256_MAC([]byte("k"), nil); got != nil {
			t.Fatalf("expected nil for empty data, got %v", got)
		}
	})
	t.Run("BLAKE2b_256 long key behavior mirrors library", func(t *testing.T) {
		longKey := make([]byte, 65)
		d, err := blake2b.New256(longKey)
		got := BLAKE2b_256_MAC(longKey, msg)
		if err != nil {
			if got != nil {
				t.Fatalf("expected nil when blake2b.New256 errors, got %v", got)
			}
			return
		}
		d.Write(msg)
		want := d.Sum(nil)
		if !hmac.Equal(got, want) {
			t.Fatalf("BLAKE2b_256_MAC with long key mismatch")
		}
	})

	t.Run("BLAKE2b_512 empty key", func(t *testing.T) {
		if got := BLAKE2b_512_MAC(nil, msg); got != nil {
			t.Fatalf("expected nil for empty key, got %v", got)
		}
	})
	t.Run("BLAKE2b_512 empty data", func(t *testing.T) {
		if got := BLAKE2b_512_MAC([]byte("k"), nil); got != nil {
			t.Fatalf("expected nil for empty data, got %v", got)
		}
	})
	t.Run("BLAKE2b_512 long key behavior mirrors library", func(t *testing.T) {
		longKey := make([]byte, 65)
		d, err := blake2b.New512(longKey)
		got := BLAKE2b_512_MAC(longKey, msg)
		if err != nil {
			if got != nil {
				t.Fatalf("expected nil when blake2b.New512 errors, got %v", got)
			}
			return
		}
		d.Write(msg)
		want := d.Sum(nil)
		if !hmac.Equal(got, want) {
			t.Fatalf("BLAKE2b_512_MAC with long key mismatch")
		}
	})
}
