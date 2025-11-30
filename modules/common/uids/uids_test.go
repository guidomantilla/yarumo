package uids

import (
	"testing"

	"github.com/google/uuid"
	ulidpkg "github.com/oklog/ulid/v2"
	"github.com/rs/xid"
)

func TestGetByName_SuccessCases(t *testing.T) {
	tests := []struct {
		name  string
		check func(t *testing.T, fn UIDFn)
	}{
		{
			name: UuidV4,
			check: func(t *testing.T, fn UIDFn) {
				if fn == nil {
					t.Fatalf("UUIDv4 function is nil")
				}
				id := fn()
				if id == "" {
					t.Fatalf("UUIDv4 returned empty")
				}
				u, err := uuid.Parse(id)
				if err != nil {
					t.Fatalf("UUIDv4 parse failed: %v", err)
				}
				if u.Version() != 4 {
					t.Fatalf("UUIDv4 wrong version: %v", u.Version())
				}
			},
		},
		{
			name: UuidV7,
			check: func(t *testing.T, fn UIDFn) {
				if fn == nil {
					t.Fatalf("UUIDv7 function is nil")
				}
				id := fn()
				if id == "" {
					t.Fatalf("UUIDv7 returned empty")
				}
				u, err := uuid.Parse(id)
				if err != nil {
					t.Fatalf("UUIDv7 parse failed: %v", err)
				}
				if u.Version() != 7 {
					t.Fatalf("UUIDv7 wrong version: %v", u.Version())
				}
			},
		},
		{
			name: Ulid,
			check: func(t *testing.T, fn UIDFn) {
				if fn == nil {
					t.Fatalf("ULID function is nil")
				}
				id := fn()
				if id == "" {
					t.Fatalf("ULID returned empty")
				}
				if _, err := ulidpkg.Parse(id); err != nil {
					t.Fatalf("ULID parse failed: %v", err)
				}
			},
		},
		{
			name: XId,
			check: func(t *testing.T, fn UIDFn) {
				if fn == nil {
					t.Fatalf("XID function is nil")
				}
				id := fn()
				if id == "" {
					t.Fatalf("XID returned empty")
				}
				if _, err := xid.FromString(id); err != nil {
					t.Fatalf("XID parse failed: %v", err)
				}
			},
		},
		{
			name: NanoID,
			check: func(t *testing.T, fn UIDFn) {
				if fn == nil {
					t.Fatalf("NANOID function is nil")
				}
				a := fn()
				b := fn()
				if a == "" || b == "" {
					t.Fatalf("NANOID returned empty")
				}
				if a == b {
					t.Fatalf("NANOID should be unique between calls, got same: %s", a)
				}
			},
		},
		{
			name: Cuid2,
			check: func(t *testing.T, fn UIDFn) {
				if fn == nil {
					t.Fatalf("CUID2 function is nil")
				}
				a := fn()
				b := fn()
				if a == "" || b == "" {
					t.Fatalf("CUID2 returned empty")
				}
				if a == b {
					t.Fatalf("CUID2 should be unique between calls, got same: %s", a)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, err := GetByName(tt.name)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tt.name, err)
			}
			tt.check(t, fn)
		})
	}
}

func TestGetByName_Unknown(t *testing.T) {
	const unknown = "NOPE"
	fn, err := GetByName(unknown)
	if fn != nil {
		t.Fatalf("expected nil function for unknown, got non-nil")
	}
	if err == nil {
		t.Fatalf("expected error for unknown name")
	}
	// Ajustado: ahora ErrUIDFunctionNotFound devuelve un error simple (fmt.Errorf)
	expected := "uid function " + unknown + " not found"
	if got := err.Error(); got != expected {
		t.Fatalf("error message = %q, want %q", got, expected)
	}
}
