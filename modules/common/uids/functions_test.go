package uids

import (
	"testing"

	"github.com/google/uuid"
	ulidpkg "github.com/oklog/ulid/v2"
	"github.com/rs/xid"
)

func generateFor(t *testing.T, name string) string {
	t.Helper()

	algo, err := Get(name)
	if err != nil {
		t.Fatalf("Get(%q) error: %v", name, err)
	}

	got, genErr := algo.Generate()
	if genErr != nil {
		t.Fatalf("Generate(%q) error: %v", name, genErr)
	}

	return got
}

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

func TestIsUUID(t *testing.T) {
	t.Parallel()

	t.Run("accepts generated UUIDv4", func(t *testing.T) {
		t.Parallel()

		id := generateFor(t, "UUIDv4")
		if !IsUUID(id) {
			t.Fatalf("IsUUID(%q) = false, want true", id)
		}
	})

	t.Run("accepts generated UUIDv7", func(t *testing.T) {
		t.Parallel()

		id := generateFor(t, "UUIDv7")
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

	t.Run("rejects ULID format", func(t *testing.T) {
		t.Parallel()

		id := generateFor(t, "ULID")
		if IsUUID(id) {
			t.Fatalf("IsUUID(%q) = true, want false", id)
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

func TestIsULID(t *testing.T) {
	t.Parallel()

	t.Run("accepts generated ULID", func(t *testing.T) {
		t.Parallel()

		id := generateFor(t, "ULID")
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

	t.Run("rejects UUID format", func(t *testing.T) {
		t.Parallel()

		id := generateFor(t, "UUIDv7")
		if IsULID(id) {
			t.Fatalf("IsULID(%q) = true, want false", id)
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

func TestIsNanoID(t *testing.T) {
	t.Parallel()

	t.Run("accepts generated NanoID", func(t *testing.T) {
		t.Parallel()

		id := generateFor(t, "NanoID")
		if !IsNanoID(id) {
			t.Fatalf("IsNanoID(%q) = false, want true", id)
		}
	})

	t.Run("rejects empty string", func(t *testing.T) {
		t.Parallel()

		if IsNanoID("") {
			t.Fatal("IsNanoID(\"\") = true, want false")
		}
	})

	t.Run("rejects wrong length", func(t *testing.T) {
		t.Parallel()

		// 20 chars — one short of the default 21.
		if IsNanoID("abcdefghij0123456789") {
			t.Fatal("IsNanoID(too short) = true, want false")
		}

		// 22 chars — one above the default 21.
		if IsNanoID("abcdefghij012345678901") {
			t.Fatal("IsNanoID(too long) = true, want false")
		}
	})

	t.Run("rejects characters outside URL-safe alphabet", func(t *testing.T) {
		t.Parallel()

		// 21 chars containing '!' which is not in the URL-safe alphabet.
		id := "abcdefghij0123456789!"
		if IsNanoID(id) {
			t.Fatalf("IsNanoID(%q) = true, want false", id)
		}
	})

	t.Run("rejects ULID format", func(t *testing.T) {
		t.Parallel()

		id := generateFor(t, "ULID")
		if IsNanoID(id) {
			t.Fatalf("IsNanoID(%q) = true, want false", id)
		}
	})
}

func TestIsCUID2(t *testing.T) {
	t.Parallel()

	t.Run("accepts generated CUID2", func(t *testing.T) {
		t.Parallel()

		id := generateFor(t, "CUID2")
		if !IsCUID2(id) {
			t.Fatalf("IsCUID2(%q) = false, want true", id)
		}
	})

	t.Run("rejects empty string", func(t *testing.T) {
		t.Parallel()

		if IsCUID2("") {
			t.Fatal("IsCUID2(\"\") = true, want false")
		}
	})

	t.Run("rejects wrong length", func(t *testing.T) {
		t.Parallel()

		// 23 chars — one short of the default 24.
		if IsCUID2("abcdefghijklmnopqrstuvw") {
			t.Fatal("IsCUID2(too short) = true, want false")
		}

		// 25 chars — one above the default 24.
		if IsCUID2("abcdefghijklmnopqrstuvwxy") {
			t.Fatal("IsCUID2(too long) = true, want false")
		}
	})

	t.Run("rejects strings starting with a digit", func(t *testing.T) {
		t.Parallel()

		id := "1bcdefghijklmnopqrstuvwx"
		if IsCUID2(id) {
			t.Fatalf("IsCUID2(%q) = true, want false", id)
		}
	})

	t.Run("rejects uppercase characters", func(t *testing.T) {
		t.Parallel()

		id := "Abcdefghijklmnopqrstuvwx"
		if IsCUID2(id) {
			t.Fatalf("IsCUID2(%q) = true, want false", id)
		}
	})

	t.Run("rejects UUID format", func(t *testing.T) {
		t.Parallel()

		id := generateFor(t, "UUIDv7")
		if IsCUID2(id) {
			t.Fatalf("IsCUID2(%q) = true, want false", id)
		}
	})
}

func TestIsXID(t *testing.T) {
	t.Parallel()

	t.Run("accepts generated XID", func(t *testing.T) {
		t.Parallel()

		id := generateFor(t, "XID")
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

	t.Run("rejects UUID format", func(t *testing.T) {
		t.Parallel()

		id := generateFor(t, "UUIDv7")
		if IsXID(id) {
			t.Fatalf("IsXID(%q) = true, want false", id)
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
