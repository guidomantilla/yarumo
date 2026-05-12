package expressions

import (
	"errors"
	"testing"
)

func TestToFloat64(t *testing.T) {
	t.Parallel()

	t.Run("float64 passthrough", func(t *testing.T) {
		t.Parallel()
		v, ok := toFloat64(3.14)
		if !ok || v != 3.14 {
			t.Fatalf("expected 3.14, got %v", v)
		}
	})

	t.Run("int converts", func(t *testing.T) {
		t.Parallel()
		v, ok := toFloat64(42)
		if !ok || v != 42.0 {
			t.Fatalf("expected 42.0, got %v", v)
		}
	})

	t.Run("int64 converts", func(t *testing.T) {
		t.Parallel()
		v, ok := toFloat64(int64(100))
		if !ok || v != 100.0 {
			t.Fatalf("expected 100.0, got %v", v)
		}
	})

	t.Run("float32 converts", func(t *testing.T) {
		t.Parallel()
		v, ok := toFloat64(float32(1.5))
		if !ok {
			t.Fatal("expected ok")
		}
		if v < 1.4 || v > 1.6 {
			t.Fatalf("expected ~1.5, got %v", v)
		}
	})

	t.Run("string fails", func(t *testing.T) {
		t.Parallel()
		_, ok := toFloat64("hello")
		if ok {
			t.Fatal("expected not ok for string")
		}
	})

	t.Run("nil fails", func(t *testing.T) {
		t.Parallel()
		_, ok := toFloat64(nil)
		if ok {
			t.Fatal("expected not ok for nil")
		}
	})
}

func TestToBool(t *testing.T) {
	t.Parallel()

	t.Run("true passthrough", func(t *testing.T) {
		t.Parallel()
		v, ok := toBool(true)
		if !ok || !v {
			t.Fatal("expected true")
		}
	})

	t.Run("false passthrough", func(t *testing.T) {
		t.Parallel()
		v, ok := toBool(false)
		if !ok || v {
			t.Fatal("expected false")
		}
	})

	t.Run("string fails", func(t *testing.T) {
		t.Parallel()
		_, ok := toBool("true")
		if ok {
			t.Fatal("expected not ok for string")
		}
	})
}

func TestToString(t *testing.T) {
	t.Parallel()

	t.Run("string passthrough", func(t *testing.T) {
		t.Parallel()
		v, ok := toString("hello")
		if !ok || v != "hello" {
			t.Fatalf("expected hello, got %s", v)
		}
	})

	t.Run("int fails", func(t *testing.T) {
		t.Parallel()
		_, ok := toString(42)
		if ok {
			t.Fatal("expected not ok for int")
		}
	})
}

func TestToSlice(t *testing.T) {
	t.Parallel()

	t.Run("slice passthrough", func(t *testing.T) {
		t.Parallel()
		input := []any{1.0, 2.0}
		v, ok := toSlice(input)
		if !ok || len(v) != 2 {
			t.Fatalf("expected slice of 2, got %v", v)
		}
	})

	t.Run("string fails", func(t *testing.T) {
		t.Parallel()
		_, ok := toSlice("abc")
		if ok {
			t.Fatal("expected not ok for string")
		}
	})
}

func TestFormatValue(t *testing.T) {
	t.Parallel()

	t.Run("nil returns nil string", func(t *testing.T) {
		t.Parallel()
		got := formatValue(nil)
		if got != "nil" {
			t.Fatalf("expected nil, got %s", got)
		}
	})

	t.Run("number returns formatted", func(t *testing.T) {
		t.Parallel()
		got := formatValue(42.5)
		if got != "42.5" {
			t.Fatalf("expected 42.5, got %s", got)
		}
	})
}

func TestResolveProperty(t *testing.T) {
	t.Parallel()

	t.Run("simple field access", func(t *testing.T) {
		t.Parallel()
		obj := map[string]any{"name": "Alice"}
		v, err := resolveProperty(obj, "name")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != "Alice" {
			t.Fatalf("expected Alice, got %v", v)
		}
	})

	t.Run("nested field access", func(t *testing.T) {
		t.Parallel()
		obj := map[string]any{
			"customer": map[string]any{
				"age": 30.0,
			},
		}
		v, err := resolveProperty(obj, "customer.age")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 30.0 {
			t.Fatalf("expected 30, got %v", v)
		}
	})

	t.Run("nil object returns error", func(t *testing.T) {
		t.Parallel()
		_, err := resolveProperty(nil, "field")
		if err == nil {
			t.Fatal("expected error for nil object")
		}
		if !errors.Is(err, ErrNilAccess) {
			t.Fatalf("expected ErrNilAccess, got %v", err)
		}
	})

	t.Run("non-map object returns error", func(t *testing.T) {
		t.Parallel()
		_, err := resolveProperty(42, "field")
		if err == nil {
			t.Fatal("expected error for non-map object")
		}
		if !errors.Is(err, ErrTypeMismatch) {
			t.Fatalf("expected ErrTypeMismatch, got %v", err)
		}
	})

	t.Run("unknown field returns error", func(t *testing.T) {
		t.Parallel()
		obj := map[string]any{"name": "Alice"}
		_, err := resolveProperty(obj, "age")
		if err == nil {
			t.Fatal("expected error for unknown field")
		}
		if !errors.Is(err, ErrUnknownField) {
			t.Fatalf("expected ErrUnknownField, got %v", err)
		}
	})

	t.Run("Context type works", func(t *testing.T) {
		t.Parallel()
		ctx := Context{"x": 10.0}
		v, err := resolveProperty(ctx, "x")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 10.0 {
			t.Fatalf("expected 10.0, got %v", v)
		}
	})
}
