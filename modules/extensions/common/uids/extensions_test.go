package uids

import (
	"errors"
	"maps"
	"testing"
)

func snapshotMethods() map[string]UID {
	cp := make(map[string]UID, len(methods))
	maps.Copy(cp, methods)

	return cp
}

func restoreMethods(m map[string]UID) {
	methods = make(map[string]UID, len(m))
	maps.Copy(methods, m)
}

func TestRegister(t *testing.T) {
	snap := snapshotMethods()
	defer restoreMethods(snap)

	t.Run("registers new UID generator", func(t *testing.T) {
		u := NewUID("CUSTOM", func() (string, error) { return "custom", nil })
		Register(u)

		got, err := Lookup("CUSTOM")
		if err != nil {
			t.Fatalf("Lookup after Register: %v", err)
		}

		generated, genErr := got.Generate()
		if genErr != nil {
			t.Fatalf("Generate after Register: %v", genErr)
		}

		if generated != "custom" {
			t.Fatalf("Generate() = %q, want %q", generated, "custom")
		}
	})

	t.Run("overwrites existing registration", func(t *testing.T) {
		Register(NewUID("OVER", func() (string, error) { return "v1", nil }))
		Register(NewUID("OVER", func() (string, error) { return "v2", nil }))

		got, err := Lookup("OVER")
		if err != nil {
			t.Fatalf("Lookup after overwrite: %v", err)
		}

		generated, genErr := got.Generate()
		if genErr != nil {
			t.Fatalf("Generate after overwrite: %v", genErr)
		}

		if generated != "v2" {
			t.Fatalf("Generate() = %q, want %q", generated, "v2")
		}
	})
}

func TestLookup(t *testing.T) {
	snap := snapshotMethods()
	defer restoreMethods(snap)

	t.Run("returns registered UID", func(t *testing.T) {
		uid, err := Lookup("UUIDv4")
		if err != nil {
			t.Fatalf("Lookup(UUIDv4) error: %v", err)
		}

		if uid.Name() != "UUIDv4" {
			t.Fatalf("Name() = %q, want %q", uid.Name(), "UUIDv4")
		}
	})

	t.Run("returns error for unknown name", func(t *testing.T) {
		_, err := Lookup("DOES_NOT_EXIST")
		if err == nil {
			t.Fatal("expected error for unknown name")
		}

		var e *Error
		if !errors.As(err, &e) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})
}

func TestSupported(t *testing.T) {
	snap := snapshotMethods()
	defer restoreMethods(snap)

	t.Run("returns all default UIDs", func(t *testing.T) {
		list := Supported()
		if len(list) != 6 {
			t.Fatalf("expected 6 UIDs, got %d", len(list))
		}
	})

	t.Run("includes newly registered UID", func(t *testing.T) {
		Register(NewUID("NEW", func() (string, error) { return "new", nil }))

		list := Supported()
		found := false

		for _, u := range list {
			if u.Name() == "NEW" {
				found = true
				break
			}
		}

		if !found {
			t.Fatal("Supported() does not include newly registered UID")
		}
	})
}
