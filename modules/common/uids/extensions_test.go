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
		u := NewUID("CUSTOM", func() string { return "custom" })
		Register(u)

		got, err := Get("CUSTOM")
		if err != nil {
			t.Fatalf("Get after Register: %v", err)
		}

		if got.Generate() != "custom" {
			t.Fatalf("Generate() = %q, want %q", got.Generate(), "custom")
		}
	})

	t.Run("overwrites existing registration", func(t *testing.T) {
		Register(NewUID("OVER", func() string { return "v1" }))
		Register(NewUID("OVER", func() string { return "v2" }))

		got, err := Get("OVER")
		if err != nil {
			t.Fatalf("Get after overwrite: %v", err)
		}

		if got.Generate() != "v2" {
			t.Fatalf("Generate() = %q, want %q", got.Generate(), "v2")
		}
	})
}

func TestGet(t *testing.T) {
	snap := snapshotMethods()
	defer restoreMethods(snap)

	t.Run("returns registered UID", func(t *testing.T) {
		uid, err := Get("UUIDv4")
		if err != nil {
			t.Fatalf("Get(UUIDv4) error: %v", err)
		}

		if uid.Name() != "UUIDv4" {
			t.Fatalf("Name() = %q, want %q", uid.Name(), "UUIDv4")
		}
	})

	t.Run("returns error for unknown name", func(t *testing.T) {
		_, err := Get("DOES_NOT_EXIST")
		if err == nil {
			t.Fatal("expected error for unknown name")
		}

		var e *Error
		if !errors.As(err, &e) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})
}

func TestUse(t *testing.T) {
	snap := snapshotMethods()
	origCurrent := current

	defer func() {
		restoreMethods(snap)

		current = origCurrent
	}()

	t.Run("selects registered generator as default", func(t *testing.T) {
		err := Use("UUIDv4")
		if err != nil {
			t.Fatalf("Use(UUIDv4) error: %v", err)
		}

		if current.Name() != "UUIDv4" {
			t.Fatalf("current.Name() = %q, want %q", current.Name(), "UUIDv4")
		}
	})

	t.Run("returns error for unknown name", func(t *testing.T) {
		err := Use("DOES_NOT_EXIST")
		if err == nil {
			t.Fatal("expected error for unknown name")
		}

		var e *Error
		if !errors.As(err, &e) {
			t.Fatalf("expected *Error, got %T", err)
		}
	})
}

func TestGenerate(t *testing.T) {
	snap := snapshotMethods()
	origCurrent := current

	defer func() {
		restoreMethods(snap)

		current = origCurrent
	}()

	t.Run("delegates to current default generator", func(t *testing.T) {
		id := Generate()
		if id == "" {
			t.Fatal("Generate() returned empty string")
		}
	})

	t.Run("uses selected generator after Use", func(t *testing.T) {
		Register(NewUID("FIXED", func() string { return "fixed-id" }))

		err := Use("FIXED")
		if err != nil {
			t.Fatalf("Use(FIXED) error: %v", err)
		}

		got := Generate()
		if got != "fixed-id" {
			t.Fatalf("Generate() = %q, want %q", got, "fixed-id")
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
		Register(NewUID("NEW", func() string { return "new" }))

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
