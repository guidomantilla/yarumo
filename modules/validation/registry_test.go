package validation

import (
	"testing"
)

func TestRegistry_Names(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	reg.Register("a", func(any, []any) error { return nil })
	reg.Register("b", func(any, []any) error { return nil })

	names := reg.Names()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
}
