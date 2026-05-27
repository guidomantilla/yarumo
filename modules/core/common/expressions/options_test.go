package expressions

import (
	"testing"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("creates default options with built-in functions", func(t *testing.T) {
		t.Parallel()
		opts := NewOptions()
		if opts.funcs == nil {
			t.Fatal("expected funcs to be initialized")
		}
		_, ok := opts.funcs["len"]
		if !ok {
			t.Fatal("expected len function in defaults")
		}
		_, ok = opts.funcs["sum"]
		if !ok {
			t.Fatal("expected sum function in defaults")
		}
	})

	t.Run("WithFunc adds a custom function", func(t *testing.T) {
		t.Parallel()
		custom := func(args ...any) (any, error) { return 42.0, nil }
		opts := NewOptions(WithFunc("myFunc", custom))
		_, ok := opts.funcs["myFunc"]
		if !ok {
			t.Fatal("expected myFunc in options")
		}
	})

	t.Run("WithFunc with nil does not add", func(t *testing.T) {
		t.Parallel()
		opts := NewOptions(WithFunc("bad", nil))
		_, ok := opts.funcs["bad"]
		if ok {
			t.Fatal("nil function should not be registered")
		}
	})

	t.Run("WithFunc overrides built-in", func(t *testing.T) {
		t.Parallel()
		custom := func(args ...any) (any, error) { return 0.0, nil }
		opts := NewOptions(WithFunc("len", custom))
		result, err := opts.funcs["len"]()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != 0.0 {
			t.Fatalf("expected 0.0 from custom len, got %v", result)
		}
	})
}
