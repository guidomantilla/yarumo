package uuid

import (
	"testing"

	"github.com/google/uuid"
)

func TestUUIDv4(t *testing.T) {
	t.Parallel()

	t.Run("returns non-empty string", func(t *testing.T) {
		t.Parallel()

		got, err := UUIDv4()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got == "" {
			t.Fatal("expected non-empty string")
		}
	})

	t.Run("returns unique values", func(t *testing.T) {
		t.Parallel()

		a, errA := UUIDv4()
		if errA != nil {
			t.Fatalf("unexpected error: %v", errA)
		}

		b, errB := UUIDv4()
		if errB != nil {
			t.Fatalf("unexpected error: %v", errB)
		}

		if a == b {
			t.Fatalf("expected unique values, got same: %s", a)
		}
	})

	t.Run("returns valid UUID v4", func(t *testing.T) {
		t.Parallel()

		got, err := UUIDv4()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		u, parseErr := uuid.Parse(got)
		if parseErr != nil {
			t.Fatalf("parse failed: %v", parseErr)
		}

		if u.Version() != 4 {
			t.Fatalf("version = %v, want 4", u.Version())
		}
	})
}

func TestUUIDv7(t *testing.T) {
	t.Parallel()

	t.Run("returns non-empty string", func(t *testing.T) {
		t.Parallel()

		got, err := UUIDv7()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got == "" {
			t.Fatal("expected non-empty string")
		}
	})

	t.Run("returns unique values", func(t *testing.T) {
		t.Parallel()

		a, errA := UUIDv7()
		if errA != nil {
			t.Fatalf("unexpected error: %v", errA)
		}

		b, errB := UUIDv7()
		if errB != nil {
			t.Fatalf("unexpected error: %v", errB)
		}

		if a == b {
			t.Fatalf("expected unique values, got same: %s", a)
		}
	})

	t.Run("returns valid UUID v7", func(t *testing.T) {
		t.Parallel()

		got, err := UUIDv7()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		u, parseErr := uuid.Parse(got)
		if parseErr != nil {
			t.Fatalf("parse failed: %v", parseErr)
		}

		if u.Version() != 7 {
			t.Fatalf("version = %v, want 7", u.Version())
		}
	})
}

func TestIsUUID(t *testing.T) {
	t.Parallel()

	t.Run("accepts generated UUIDv4", func(t *testing.T) {
		t.Parallel()

		id, err := UUIDv4()
		if err != nil {
			t.Fatalf("UUIDv4: %v", err)
		}

		if !IsUUID(id) {
			t.Fatalf("IsUUID(%q) = false, want true", id)
		}
	})

	t.Run("accepts generated UUIDv7", func(t *testing.T) {
		t.Parallel()

		id, err := UUIDv7()
		if err != nil {
			t.Fatalf("UUIDv7: %v", err)
		}

		if !IsUUID(id) {
			t.Fatalf("IsUUID(%q) = false, want true", id)
		}
	})

	t.Run("accepts well-known UUIDv1", func(t *testing.T) {
		t.Parallel()

		// Sample v1 UUID from RFC 4122 examples.
		id := "f81d4fae-7dec-11d0-a765-00a0c91e6bf6"
		if !IsUUID(id) {
			t.Fatalf("IsUUID(%q) = false, want true", id)
		}
	})

	t.Run("rejects empty string", func(t *testing.T) {
		t.Parallel()

		if IsUUID("") {
			t.Fatal("IsUUID(\"\") = true, want false")
		}
	})

	t.Run("rejects wrong length", func(t *testing.T) {
		t.Parallel()

		if IsUUID("abc") {
			t.Fatal("IsUUID(\"abc\") = true, want false")
		}
	})

	t.Run("rejects invalid characters", func(t *testing.T) {
		t.Parallel()

		// Same length and shape as a UUID but with characters outside hex.
		id := "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz"
		if IsUUID(id) {
			t.Fatalf("IsUUID(%q) = true, want false", id)
		}
	})
}
