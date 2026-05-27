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
		Register(NewUID("LOOKUP_HIT", func() (string, error) { return "ok", nil }))

		got, err := Lookup("LOOKUP_HIT")
		if err != nil {
			t.Fatalf("Lookup error: %v", err)
		}

		if got.Name() != "LOOKUP_HIT" {
			t.Fatalf("Name() = %q, want %q", got.Name(), "LOOKUP_HIT")
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

	t.Run("empty registry returns empty list", func(t *testing.T) {
		restoreMethods(map[string]UID{})

		list := Supported()
		if len(list) != 0 {
			t.Fatalf("expected empty list, got %d", len(list))
		}
	})

	t.Run("returns all registered UIDs", func(t *testing.T) {
		restoreMethods(map[string]UID{})

		Register(NewUID("A", func() (string, error) { return "a", nil }))
		Register(NewUID("B", func() (string, error) { return "b", nil }))

		list := Supported()
		if len(list) != 2 {
			t.Fatalf("expected 2 UIDs, got %d", len(list))
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
