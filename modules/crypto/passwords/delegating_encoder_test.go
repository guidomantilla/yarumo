package passwords

import (
	"errors"
	"strings"
	"testing"
)

func TestNewDelegatingEncoder(t *testing.T) {
	t.Parallel()

	t.Run("constructs encoder with primary method", func(t *testing.T) {
		t.Parallel()

		d := NewDelegatingEncoder(Argon2id)

		if d == nil {
			t.Fatal("expected non-nil delegating encoder")
		}
		if d.primary != Argon2id {
			t.Fatalf("expected primary to be Argon2id, got %v", d.primary)
		}
	})

	t.Run("invariant logs when primary is nil but does not crash", func(t *testing.T) {
		t.Parallel()

		// cassert.NotNil emits a log entry without panicking when assertions
		// are disabled; the returned encoder still holds the nil primary so
		// that callers observe the misuse via subsequent operations.
		d := NewDelegatingEncoder(nil)
		if d == nil {
			t.Fatal("expected non-nil encoder even when primary is nil")
		}
		if d.primary != nil {
			t.Fatalf("expected primary to be nil, got %v", d.primary)
		}
	})
}

func TestDelegatingEncoder_Encode(t *testing.T) {
	t.Parallel()

	t.Run("delegates encoding to primary method", func(t *testing.T) {
		t.Parallel()

		d := NewDelegatingEncoder(Bcrypt)

		encoded, err := d.Encode("delegate-encode")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.HasPrefix(encoded, BcryptPrefixKey) {
			t.Fatalf("expected prefix %q, got %q", BcryptPrefixKey, encoded)
		}
	})

	t.Run("wraps encode error in domain Error", func(t *testing.T) {
		t.Parallel()

		d := NewDelegatingEncoder(Bcrypt)

		_, err := d.Encode("")
		if err == nil {
			t.Fatal("expected error")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !errors.Is(err, ErrDelegateFailed) {
			t.Fatalf("expected ErrDelegateFailed in chain, got %v", err)
		}
	})
}

func TestDelegatingEncoder_Verify(t *testing.T) {
	t.Parallel()

	t.Run("verifies hash produced by primary", func(t *testing.T) {
		t.Parallel()

		d := NewDelegatingEncoder(Bcrypt)

		encoded, err := d.Encode("same-primary")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ok, err := d.Verify(encoded, "same-primary")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatal("expected verify to succeed")
		}
	})

	t.Run("verifies legacy hash via prefix routing", func(t *testing.T) {
		t.Parallel()

		legacy, err := Bcrypt.Encode("legacy-secret")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		d := NewDelegatingEncoder(Argon2id)

		ok, err := d.Verify(legacy, "legacy-secret")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatal("expected legacy bcrypt hash to verify via delegating encoder with argon2 primary")
		}
	})

	t.Run("returns unknown prefix error for unrecognised hash", func(t *testing.T) {
		t.Parallel()

		d := NewDelegatingEncoder(Bcrypt)

		ok, err := d.Verify("{unknown}garbage", "raw")
		if err == nil {
			t.Fatal("expected error")
		}
		if ok {
			t.Fatal("expected ok=false on unknown prefix")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !errors.Is(err, ErrUnknownEncodingPrefix) {
			t.Fatalf("expected ErrUnknownEncodingPrefix in chain, got %v", err)
		}
		if !errors.Is(err, ErrDelegateFailed) {
			t.Fatalf("expected ErrDelegateFailed in chain, got %v", err)
		}
	})

	t.Run("returns false for wrong password against legacy hash", func(t *testing.T) {
		t.Parallel()

		legacy, err := Bcrypt.Encode("correct")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		d := NewDelegatingEncoder(Argon2id)

		ok, err := d.Verify(legacy, "wrong")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatal("expected verify to fail for wrong password")
		}
	})
}

func TestDelegatingEncoder_UpgradeNeeded(t *testing.T) {
	t.Parallel()

	t.Run("returns true when encoded uses different algorithm than primary", func(t *testing.T) {
		t.Parallel()

		legacy, err := Bcrypt.Encode("upgrade-me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		d := NewDelegatingEncoder(Argon2id)

		needed, err := d.UpgradeNeeded(legacy)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !needed {
			t.Fatal("expected upgrade needed when prefix differs from primary algorithm")
		}
	})

	t.Run("delegates to primary when algorithm matches", func(t *testing.T) {
		t.Parallel()

		d := NewDelegatingEncoder(Bcrypt)

		encoded, err := d.Encode("no-upgrade")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		needed, err := d.UpgradeNeeded(encoded)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if needed {
			t.Fatal("expected no upgrade needed for same algorithm and same parameters")
		}
	})

	t.Run("returns unknown prefix error for unrecognised hash", func(t *testing.T) {
		t.Parallel()

		d := NewDelegatingEncoder(Bcrypt)

		_, err := d.UpgradeNeeded("{nope}garbage")
		if err == nil {
			t.Fatal("expected error")
		}

		var domErr *Error
		if !errors.As(err, &domErr) {
			t.Fatalf("expected *Error, got %T", err)
		}

		if !errors.Is(err, ErrUnknownEncodingPrefix) {
			t.Fatalf("expected ErrUnknownEncodingPrefix in chain, got %v", err)
		}
	})
}

// TestDelegatingEncoder_MigrationRoundTrip exercises the full login-time
// upgrade pattern: an existing Bcrypt hash verifies under a delegating
// encoder whose primary is Argon2, UpgradeNeeded reports true, the caller
// re-encodes under the primary, and the new hash continues to verify.
func TestDelegatingEncoder_MigrationRoundTrip(t *testing.T) {
	t.Parallel()

	const raw = "migration-round-trip"

	legacy, err := Bcrypt.Encode(raw)
	if err != nil {
		t.Fatalf("unexpected error encoding legacy hash: %v", err)
	}
	if !strings.HasPrefix(legacy, BcryptPrefixKey) {
		t.Fatalf("expected bcrypt prefix on legacy hash, got %q", legacy)
	}

	d := NewDelegatingEncoder(Argon2id)

	ok, err := d.Verify(legacy, raw)
	if err != nil {
		t.Fatalf("unexpected error verifying legacy hash: %v", err)
	}
	if !ok {
		t.Fatal("expected legacy bcrypt hash to verify under delegating encoder")
	}

	needed, err := d.UpgradeNeeded(legacy)
	if err != nil {
		t.Fatalf("unexpected error checking upgrade: %v", err)
	}
	if !needed {
		t.Fatal("expected UpgradeNeeded to return true for legacy bcrypt hash with argon2 primary")
	}

	upgraded, err := d.Encode(raw)
	if err != nil {
		t.Fatalf("unexpected error re-encoding under primary: %v", err)
	}
	if !strings.HasPrefix(upgraded, Argon2idPrefixKey) {
		t.Fatalf("expected argon2id prefix on upgraded hash, got %q", upgraded)
	}

	ok, err = d.Verify(upgraded, raw)
	if err != nil {
		t.Fatalf("unexpected error verifying upgraded hash: %v", err)
	}
	if !ok {
		t.Fatal("expected upgraded hash to verify under delegating encoder")
	}

	needed, err = d.UpgradeNeeded(upgraded)
	if err != nil {
		t.Fatalf("unexpected error checking upgrade for upgraded hash: %v", err)
	}
	if needed {
		t.Fatal("expected UpgradeNeeded to return false after migration to primary")
	}
}
