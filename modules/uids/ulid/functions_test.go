package ulid

import (
	"testing"

	ulidpkg "github.com/oklog/ulid/v2"
)

func TestULID(t *testing.T) {
	t.Parallel()

	t.Run("returns non-empty string", func(t *testing.T) {
		t.Parallel()

		got, err := ULID()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got == "" {
			t.Fatal("expected non-empty string")
		}
	})

	t.Run("returns unique values", func(t *testing.T) {
		t.Parallel()

		a, errA := ULID()
		if errA != nil {
			t.Fatalf("unexpected error: %v", errA)
		}

		b, errB := ULID()
		if errB != nil {
			t.Fatalf("unexpected error: %v", errB)
		}

		if a == b {
			t.Fatalf("expected unique values, got same: %s", a)
		}
	})

	t.Run("returns valid ULID", func(t *testing.T) {
		t.Parallel()

		got, err := ULID()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, parseErr := ulidpkg.Parse(got)
		if parseErr != nil {
			t.Fatalf("parse failed: %v", parseErr)
		}
	})
}

func TestIsULID(t *testing.T) {
	t.Parallel()

	t.Run("accepts generated ULID", func(t *testing.T) {
		t.Parallel()

		id, err := ULID()
		if err != nil {
			t.Fatalf("ULID: %v", err)
		}

		if !IsULID(id) {
			t.Fatalf("IsULID(%q) = false, want true", id)
		}
	})

	t.Run("rejects empty string", func(t *testing.T) {
		t.Parallel()

		if IsULID("") {
			t.Fatal("IsULID(\"\") = true, want false")
		}
	})

	t.Run("rejects wrong length", func(t *testing.T) {
		t.Parallel()

		// 25 chars — one short of the 26-char ULID.
		if IsULID("01ARZ3NDEKTSV4RRFFQ69G5FA") {
			t.Fatal("IsULID(too short) = true, want false")
		}
	})

	t.Run("rejects characters outside Crockford Base32", func(t *testing.T) {
		t.Parallel()

		// 26 chars containing '!' which is outside Crockford Base32.
		id := "01ARZ3NDEKTSV4RRFFQ69G5F!!"
		if IsULID(id) {
			t.Fatalf("IsULID(%q) = true, want false", id)
		}
	})
}
