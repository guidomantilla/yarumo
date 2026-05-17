package validation

import (
	"errors"
	"testing"
)

func TestAnnotatePath(t *testing.T) {
	t.Parallel()

	t.Run("empty path returns the error unchanged", func(t *testing.T) {
		t.Parallel()

		inner := errors.New("boom")

		got := annotatePath("", inner)
		if got != inner {
			t.Fatalf("expected inner error unchanged, got %v", got)
		}
	})

	t.Run("non-empty path wraps and surfaces via PathOf", func(t *testing.T) {
		t.Parallel()

		inner := errors.New("boom")

		wrapped := annotatePath("Owner.Email", inner)
		if wrapped == nil {
			t.Fatalf("expected wrapped error")
		}

		path := PathOf(wrapped)
		if path != "Owner.Email" {
			t.Fatalf("expected Owner.Email, got %q", path)
		}
	})
}

func TestJoinPath(t *testing.T) {
	t.Parallel()

	t.Run("empty parent yields child", func(t *testing.T) {
		t.Parallel()

		got := joinPath("", "Name")
		if got != "Name" {
			t.Fatalf("expected Name, got %q", got)
		}
	})

	t.Run("non-empty parent joins with dot", func(t *testing.T) {
		t.Parallel()

		got := joinPath("Owner", "Email")
		if got != "Owner.Email" {
			t.Fatalf("expected Owner.Email, got %q", got)
		}
	})

	t.Run("empty child yields parent with trailing dot", func(t *testing.T) {
		t.Parallel()

		// joinPath does not special-case an empty child; document the actual
		// behaviour so callers know what to expect.
		got := joinPath("Owner", "")
		if got != "Owner." {
			t.Fatalf("expected Owner., got %q", got)
		}
	})
}

func TestTryParseYAMLList(t *testing.T) {
	t.Parallel()

	t.Run("valid sequence parses to nodes", func(t *testing.T) {
		t.Parallel()

		nodes, ok := tryParseYAMLList([]byte(`- required` + "\n" + `- email`))
		if !ok {
			t.Fatalf("expected ok=true")
		}

		if len(nodes) != 2 {
			t.Fatalf("expected 2 nodes, got %d", len(nodes))
		}
	})

	t.Run("non-sequence returns false", func(t *testing.T) {
		t.Parallel()

		_, ok := tryParseYAMLList([]byte(`rules: []`))
		if ok {
			t.Fatalf("expected ok=false for non-sequence input")
		}
	})

	t.Run("empty sequence returns false", func(t *testing.T) {
		t.Parallel()

		_, ok := tryParseYAMLList([]byte(`[]`))
		if ok {
			t.Fatalf("expected ok=false for empty sequence")
		}
	})

	t.Run("malformed yaml returns false", func(t *testing.T) {
		t.Parallel()

		_, ok := tryParseYAMLList([]byte(`{[not yaml`))
		if ok {
			t.Fatalf("expected ok=false for malformed YAML")
		}
	})
}

func TestTryParseJSONList(t *testing.T) {
	t.Parallel()

	t.Run("valid array parses to nodes", func(t *testing.T) {
		t.Parallel()

		nodes, ok := tryParseJSONList([]byte(`["required", "email"]`))
		if !ok {
			t.Fatalf("expected ok=true")
		}

		if len(nodes) != 2 {
			t.Fatalf("expected 2 nodes, got %d", len(nodes))
		}
	})

	t.Run("non-array returns false", func(t *testing.T) {
		t.Parallel()

		_, ok := tryParseJSONList([]byte(`{"rules": []}`))
		if ok {
			t.Fatalf("expected ok=false for non-array input")
		}
	})

	t.Run("empty array returns false", func(t *testing.T) {
		t.Parallel()

		_, ok := tryParseJSONList([]byte(`[]`))
		if ok {
			t.Fatalf("expected ok=false for empty array")
		}
	})

	t.Run("malformed json returns false", func(t *testing.T) {
		t.Parallel()

		_, ok := tryParseJSONList([]byte(`{not json`))
		if ok {
			t.Fatalf("expected ok=false for malformed JSON")
		}
	})
}
