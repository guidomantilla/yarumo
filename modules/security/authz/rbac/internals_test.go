package rbac

import (
	"errors"
	"sort"
	"testing"
)

func TestBuildClosure(t *testing.T) {
	t.Parallel()

	t.Run("single edge", func(t *testing.T) {
		t.Parallel()

		closure, err := buildClosure(map[string][]string{
			"admin": {"editor"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ancestors := closure["admin"]
		if len(ancestors) != 1 || ancestors[0] != "editor" {
			t.Fatalf("expected [editor], got %v", ancestors)
		}
	})

	t.Run("transitive closure", func(t *testing.T) {
		t.Parallel()

		closure, err := buildClosure(map[string][]string{
			"admin":  {"editor"},
			"editor": {"viewer"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ancestors := closure["admin"]
		sort.Strings(ancestors)

		if len(ancestors) != 2 || ancestors[0] != "editor" || ancestors[1] != "viewer" {
			t.Fatalf("expected [editor viewer], got %v", ancestors)
		}
	})

	t.Run("cycle returns ErrInheritanceCycle", func(t *testing.T) {
		t.Parallel()

		_, err := buildClosure(map[string][]string{
			"a": {"b"},
			"b": {"a"},
		})

		if err == nil {
			t.Fatal("expected cycle error")
		}

		if !errors.Is(err, ErrInheritanceCycle) {
			t.Fatalf("expected ErrInheritanceCycle, got %v", err)
		}
	})

	t.Run("self-loop is a cycle", func(t *testing.T) {
		t.Parallel()

		_, err := buildClosure(map[string][]string{
			"a": {"a"},
		})

		if !errors.Is(err, ErrInheritanceCycle) {
			t.Fatalf("expected ErrInheritanceCycle, got %v", err)
		}
	})

	t.Run("empty hierarchy", func(t *testing.T) {
		t.Parallel()

		closure, err := buildClosure(map[string][]string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(closure) != 0 {
			t.Fatalf("expected empty closure, got %v", closure)
		}
	})

	t.Run("diamond inheritance", func(t *testing.T) {
		t.Parallel()

		closure, err := buildClosure(map[string][]string{
			"admin":   {"editor", "auditor"},
			"editor":  {"viewer"},
			"auditor": {"viewer"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ancestors := closure["admin"]
		sort.Strings(ancestors)

		if len(ancestors) != 3 {
			t.Fatalf("expected 3 ancestors, got %v", ancestors)
		}
	})
}
