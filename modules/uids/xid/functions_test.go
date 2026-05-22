package xid

import (
	"testing"

	"github.com/rs/xid"
)

func TestXID(t *testing.T) {
	t.Parallel()

	t.Run("returns non-empty string", func(t *testing.T) {
		t.Parallel()

		got, err := XID()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got == "" {
			t.Fatal("expected non-empty string")
		}
	})

	t.Run("returns unique values", func(t *testing.T) {
		t.Parallel()

		a, errA := XID()
		if errA != nil {
			t.Fatalf("unexpected error: %v", errA)
		}

		b, errB := XID()
		if errB != nil {
			t.Fatalf("unexpected error: %v", errB)
		}

		if a == b {
			t.Fatalf("expected unique values, got same: %s", a)
		}
	})

	t.Run("returns valid XID", func(t *testing.T) {
		t.Parallel()

		got, err := XID()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, parseErr := xid.FromString(got)
		if parseErr != nil {
			t.Fatalf("parse failed: %v", parseErr)
		}
	})
}

func TestIsXID(t *testing.T) {
	t.Parallel()

	t.Run("accepts generated XID", func(t *testing.T) {
		t.Parallel()

		id, err := XID()
		if err != nil {
			t.Fatalf("XID: %v", err)
		}

		if !IsXID(id) {
			t.Fatalf("IsXID(%q) = false, want true", id)
		}
	})

	t.Run("rejects empty string", func(t *testing.T) {
		t.Parallel()

		if IsXID("") {
			t.Fatal("IsXID(\"\") = true, want false")
		}
	})

	t.Run("rejects wrong length", func(t *testing.T) {
		t.Parallel()

		// 19 chars — one short of the 20-char XID.
		if IsXID("9m4e2mr0ui3e8a215n4") {
			t.Fatal("IsXID(too short) = true, want false")
		}
	})

	t.Run("rejects characters outside base32hex alphabet", func(t *testing.T) {
		t.Parallel()

		// 20 chars containing '!' which is outside the XID alphabet.
		id := "9m4e2mr0ui3e8a215n4!"
		if IsXID(id) {
			t.Fatalf("IsXID(%q) = true, want false", id)
		}
	})
}
