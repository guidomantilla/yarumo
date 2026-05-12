package hmacs

import (
	"errors"
	"testing"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func TestDigest_ByName(t *testing.T) {
	t.Parallel()

	t.Run("computes digest by name", func(t *testing.T) {
		t.Parallel()

		k := ctypes.Bytes("12345678901234567890123456789012")

		got, err := Digest("HMAC_with_SHA256", k, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 32 {
			t.Fatalf("expected 32 bytes, got %d", len(got))
		}
	})

	t.Run("returns domain error for unknown name", func(t *testing.T) {
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
}

func TestValidate_ByName(t *testing.T) {
	t.Parallel()

	t.Run("validates digest by name", func(t *testing.T) {
		t.Parallel()

		k := ctypes.Bytes("12345678901234567890123456789012")

		mac, err := Digest("HMAC_with_SHA256", k, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error computing digest: %v", err)
		}

		ok, err := Validate("HMAC_with_SHA256", k, mac, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected validation to succeed")
		}
	})

	t.Run("returns domain error for unknown name", func(t *testing.T) {
		t.Parallel()

		ok, err := Validate("UNKNOWN", []byte("k"), []byte("sig"), []byte("data"))
		if err == nil {
			t.Fatal("expected error for unknown name")
		}

		if ok {
			t.Fatal("expected ok to be false on error")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})
}

func TestKey(t *testing.T) {
	t.Parallel()

	t.Run("generates key of correct size", func(t *testing.T) {
		t.Parallel()

		m := HMAC_with_SHA256

		k, err := key(m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(k) != 32 {
			t.Fatalf("expected 32 bytes, got %d", len(k))
		}
	})

	t.Run("returns error for nil method", func(t *testing.T) {
		t.Parallel()

		_, err := key(nil)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrMethodIsNil) {
			t.Fatalf("expected ErrMethodIsNil, got %v", err)
		}
	})
}

func TestDigest(t *testing.T) {
	t.Parallel()

	t.Run("computes digest", func(t *testing.T) {
		t.Parallel()

		m := HMAC_with_SHA256
		k := ctypes.Bytes("12345678901234567890123456789012")

		d, err := digest(m, k, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(d) != 32 {
			t.Fatalf("expected 32 bytes, got %d", len(d))
		}
	})

	t.Run("returns error for nil method", func(t *testing.T) {
		t.Parallel()

		_, err := digest(nil, nil, nil)
		if !errors.Is(err, ErrMethodIsNil) {
			t.Fatalf("expected ErrMethodIsNil, got %v", err)
		}
	})

	t.Run("returns error for nil key", func(t *testing.T) {
		t.Parallel()

		_, err := digest(HMAC_with_SHA256, nil, []byte("data"))
		if !errors.Is(err, ErrKeyIsNil) {
			t.Fatalf("expected ErrKeyIsNil, got %v", err)
		}
	})

	t.Run("returns error for unavailable hash", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("bad-hash", 0, 32)

		_, err := digest(m, []byte("key"), []byte("data"))
		if !errors.Is(err, ErrHashNotAvailable) {
			t.Fatalf("expected ErrHashNotAvailable, got %v", err)
		}
	})
}

func TestValidate(t *testing.T) {
	t.Parallel()

	t.Run("returns true for matching digest", func(t *testing.T) {
		t.Parallel()

		m := HMAC_with_SHA256
		k := ctypes.Bytes("12345678901234567890123456789012")

		d, err := digest(m, k, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := validate(m, k, d, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected valid")
		}
	})

	t.Run("returns false for mismatched digest", func(t *testing.T) {
		t.Parallel()

		m := HMAC_with_SHA256
		k := ctypes.Bytes("12345678901234567890123456789012")

		ok, err := validate(m, k, []byte("wrong"), []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatal("expected invalid")
		}
	})

	t.Run("returns error for nil method", func(t *testing.T) {
		t.Parallel()

		_, err := validate(nil, nil, nil, nil)
		if !errors.Is(err, ErrMethodIsNil) {
			t.Fatalf("expected ErrMethodIsNil, got %v", err)
		}
	})

	t.Run("returns error for nil key", func(t *testing.T) {
		t.Parallel()

		_, err := validate(HMAC_with_SHA256, nil, []byte("sig"), []byte("data"))
		if !errors.Is(err, ErrKeyIsNil) {
			t.Fatalf("expected ErrKeyIsNil, got %v", err)
		}
	})

	t.Run("returns error for unavailable hash", func(t *testing.T) {
		t.Parallel()

		m := NewMethod("bad-hash", 0, 32)

		_, err := validate(m, []byte("key"), []byte("sig"), []byte("data"))
		if !errors.Is(err, ErrHashNotAvailable) {
			t.Fatalf("expected ErrHashNotAvailable, got %v", err)
		}
	})
}
