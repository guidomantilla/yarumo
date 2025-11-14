package uids

import (
	"testing"

	"github.com/google/uuid"
	ulidpkg "github.com/oklog/ulid/v2"
	"github.com/rs/xid"
)

func TestUUIDv4(t *testing.T) {
	a := UUIDv4()
	b := UUIDv4()

	if a == "" || b == "" {
		t.Fatalf("UUIDv4 returned empty string")
	}
	if a == b {
		t.Fatalf("UUIDv4 should return unique values, got same: %s", a)
	}

	ua, err := uuid.Parse(a)
	if err != nil {
		t.Fatalf("UUIDv4 parse failed: %v", err)
	}
	if ua.Version() != 4 {
		t.Fatalf("UUIDv4 wrong version: %v", ua.Version())
	}
}

func TestUUIDv7(t *testing.T) {
	a := UUIDv7()
	b := UUIDv7()

	if a == "" || b == "" {
		t.Fatalf("UUIDv7 returned empty string")
	}
	if a == b {
		t.Fatalf("UUIDv7 should return unique values, got same: %s", a)
	}

	ua, err := uuid.Parse(a)
	if err != nil {
		t.Fatalf("UUIDv7 parse failed: %v", err)
	}
	if ua.Version() != 7 {
		t.Fatalf("UUIDv7 wrong version: %v", ua.Version())
	}
}

func TestULID(t *testing.T) {
	id := ULID()
	if id == "" {
		t.Fatalf("ULID returned empty string")
	}
	if _, err := ulidpkg.Parse(id); err != nil {
		t.Fatalf("ULID parse failed: %v", err)
	}
}

func TestXID(t *testing.T) {
	id := XID()
	if id == "" {
		t.Fatalf("XID returned empty string")
	}
	if _, err := xid.FromString(id); err != nil {
		t.Fatalf("XID parse failed: %v", err)
	}
}

func TestNANOID(t *testing.T) {
	a := NANOID()
	b := NANOID()
	if a == "" || b == "" {
		t.Fatalf("NANOID returned empty string")
	}
	if a == b {
		t.Fatalf("NANOID should return unique values, got same: %s", a)
	}
}

func TestCUID2(t *testing.T) {
	a := CUID2()
	b := CUID2()
	if a == "" || b == "" {
		t.Fatalf("CUID2 returned empty string")
	}
	if a == b {
		t.Fatalf("CUID2 should return unique values, got same: %s", a)
	}
}
