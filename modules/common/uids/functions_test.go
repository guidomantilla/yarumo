package uids

import (
	"testing"

	"github.com/google/uuid"
	ulidpkg "github.com/oklog/ulid/v2"
	"github.com/rs/xid"
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

func TestNANOID(t *testing.T) {
	t.Parallel()

	t.Run("returns non-empty string", func(t *testing.T) {
		t.Parallel()

		got, err := NANOID()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got == "" {
			t.Fatal("expected non-empty string")
		}
	})

	t.Run("returns unique values", func(t *testing.T) {
		t.Parallel()

		a, errA := NANOID()
		if errA != nil {
			t.Fatalf("unexpected error: %v", errA)
		}

		b, errB := NANOID()
		if errB != nil {
			t.Fatalf("unexpected error: %v", errB)
		}

		if a == b {
			t.Fatalf("expected unique values, got same: %s", a)
		}
	})
}

func TestCUID2(t *testing.T) {
	t.Parallel()

	t.Run("returns non-empty string", func(t *testing.T) {
		t.Parallel()

		got, err := CUID2()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got == "" {
			t.Fatal("expected non-empty string")
		}
	})

	t.Run("returns unique values", func(t *testing.T) {
		t.Parallel()

		a, errA := CUID2()
		if errA != nil {
			t.Fatalf("unexpected error: %v", errA)
		}

		b, errB := CUID2()
		if errB != nil {
			t.Fatalf("unexpected error: %v", errB)
		}

		if a == b {
			t.Fatalf("expected unique values, got same: %s", a)
		}
	})
}

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
