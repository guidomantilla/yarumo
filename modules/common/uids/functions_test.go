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

		if UUIDv4() == "" {
			t.Fatal("expected non-empty string")
		}
	})

	t.Run("returns unique values", func(t *testing.T) {
		t.Parallel()

		a, b := UUIDv4(), UUIDv4()
		if a == b {
			t.Fatalf("expected unique values, got same: %s", a)
		}
	})

	t.Run("returns valid UUID v4", func(t *testing.T) {
		t.Parallel()

		u, err := uuid.Parse(UUIDv4())
		if err != nil {
			t.Fatalf("parse failed: %v", err)
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

		if UUIDv7() == "" {
			t.Fatal("expected non-empty string")
		}
	})

	t.Run("returns unique values", func(t *testing.T) {
		t.Parallel()

		a, b := UUIDv7(), UUIDv7()
		if a == b {
			t.Fatalf("expected unique values, got same: %s", a)
		}
	})

	t.Run("returns valid UUID v7", func(t *testing.T) {
		t.Parallel()

		u, err := uuid.Parse(UUIDv7())
		if err != nil {
			t.Fatalf("parse failed: %v", err)
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

		if NANOID() == "" {
			t.Fatal("expected non-empty string")
		}
	})

	t.Run("returns unique values", func(t *testing.T) {
		t.Parallel()

		a, b := NANOID(), NANOID()
		if a == b {
			t.Fatalf("expected unique values, got same: %s", a)
		}
	})
}

func TestCUID2(t *testing.T) {
	t.Parallel()

	t.Run("returns non-empty string", func(t *testing.T) {
		t.Parallel()

		if CUID2() == "" {
			t.Fatal("expected non-empty string")
		}
	})

	t.Run("returns unique values", func(t *testing.T) {
		t.Parallel()

		a, b := CUID2(), CUID2()
		if a == b {
			t.Fatalf("expected unique values, got same: %s", a)
		}
	})
}

func TestULID(t *testing.T) {
	t.Parallel()

	t.Run("returns non-empty string", func(t *testing.T) {
		t.Parallel()

		if ULID() == "" {
			t.Fatal("expected non-empty string")
		}
	})

	t.Run("returns unique values", func(t *testing.T) {
		t.Parallel()

		a, b := ULID(), ULID()
		if a == b {
			t.Fatalf("expected unique values, got same: %s", a)
		}
	})

	t.Run("returns valid ULID", func(t *testing.T) {
		t.Parallel()

		_, err := ulidpkg.Parse(ULID())
		if err != nil {
			t.Fatalf("parse failed: %v", err)
		}
	})
}

func TestXID(t *testing.T) {
	t.Parallel()

	t.Run("returns non-empty string", func(t *testing.T) {
		t.Parallel()

		if XID() == "" {
			t.Fatal("expected non-empty string")
		}
	})

	t.Run("returns unique values", func(t *testing.T) {
		t.Parallel()

		a, b := XID(), XID()
		if a == b {
			t.Fatalf("expected unique values, got same: %s", a)
		}
	})

	t.Run("returns valid XID", func(t *testing.T) {
		t.Parallel()

		_, err := xid.FromString(XID())
		if err != nil {
			t.Fatalf("parse failed: %v", err)
		}
	})
}
