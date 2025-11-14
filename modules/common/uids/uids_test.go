package uids

import (
	"errors"
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
	var ue *UIDError
	if !errors.As(err, &ue) || ue == nil {
		t.Fatalf("error is not *UIDError: %T", err)
	}
	if ue.Type != UIDNotFound {
		t.Fatalf("UIDError.Type = %q, want %q", ue.Type, UIDNotFound)
	}
	// message should include the provided name per ErrUIDFunctionNotFound
	if msg := ue.Error(); msg == "" || !contains(msg, unknown) {
		t.Fatalf("UIDError message %q does not contain %q", msg, unknown)
	}
}

// small helper to avoid importing strings just for Contains in this file
func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
