package expressions

import (
	"errors"
	"testing"
)

func TestNewEvaluator(t *testing.T) {
	t.Parallel()

	t.Run("creates with defaults", func(t *testing.T) {
		t.Parallel()

		ev := NewEvaluator()

		impl, ok := ev.(*evaluator)
		if !ok {
			t.Fatal("expected *evaluator concrete type")
		}

		if impl.options.funcs == nil {
			t.Fatal("expected funcs to be initialized")
		}
	})

	t.Run("creates with custom function", func(t *testing.T) {
		t.Parallel()

		custom := func(args ...any) (any, error) { return 42.0, nil }
		ev := NewEvaluator(WithFunc("custom", custom))

		impl, ok := ev.(*evaluator)
		if !ok {
			t.Fatal("expected *evaluator concrete type")
		}

		_, ok = impl.options.funcs["custom"]
		if !ok {
			t.Fatal("expected custom function in evaluator")
		}
	})
}

func TestEvaluator_Evaluate(t *testing.T) {
	t.Parallel()

	t.Run("simple arithmetic", func(t *testing.T) {
		t.Parallel()

		ev := NewEvaluator()
		result, err := ev.Evaluate("1 + 2", nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != 3.0 {
			t.Fatalf("expected 3.0, got %v", result)
		}
	})

	t.Run("context lookup", func(t *testing.T) {
		t.Parallel()

		ev := NewEvaluator()
		ctx := Context{"x": 10.0}
		result, err := ev.Evaluate("x + 5", ctx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != 15.0 {
			t.Fatalf("expected 15.0, got %v", result)
		}
	})

	t.Run("builtin function", func(t *testing.T) {
		t.Parallel()

		ev := NewEvaluator()
		ctx := Context{"items": []any{1.0, 2.0, 3.0}}
		result, err := ev.Evaluate("sum(items)", ctx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != 6.0 {
			t.Fatalf("expected 6.0, got %v", result)
		}
	})

	t.Run("custom function", func(t *testing.T) {
		t.Parallel()

		double := func(args ...any) (any, error) {
			n, ok := toFloat64(args[0])
			if !ok {
				return nil, ErrEval("expected numeric", ErrTypeMismatch)
			}

			return n * 2, nil
		}

		ev := NewEvaluator(WithFunc("double", double))
		ctx := Context{"x": 5.0}
		result, err := ev.Evaluate("double(x)", ctx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != 10.0 {
			t.Fatalf("expected 10.0, got %v", result)
		}
	})

	t.Run("parse error", func(t *testing.T) {
		t.Parallel()

		ev := NewEvaluator()
		_, err := ev.Evaluate("(((", nil)

		if err == nil {
			t.Fatal("expected parse error")
		}

		var pe *ParseError

		if !errors.As(err, &pe) {
			t.Fatal("expected ParseError")
		}
	})

	t.Run("eval error", func(t *testing.T) {
		t.Parallel()

		ev := NewEvaluator()
		_, err := ev.Evaluate("x + 1", nil)

		if err == nil {
			t.Fatal("expected eval error")
		}

		if !errors.Is(err, ErrUnknownField) {
			t.Fatalf("expected ErrUnknownField, got %v", err)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		t.Parallel()

		ev := NewEvaluator()
		_, err := ev.Evaluate("", nil)

		if err == nil {
			t.Fatal("expected error for empty input")
		}

		if !errors.Is(err, ErrEmptyInput) {
			t.Fatalf("expected ErrEmptyInput, got %v", err)
		}
	})

	t.Run("boolean expression", func(t *testing.T) {
		t.Parallel()

		ev := NewEvaluator()
		ctx := Context{"age": 25.0}
		result, err := ev.Evaluate("age >= 18", ctx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		b, ok := result.(bool)
		if !ok || !b {
			t.Fatalf("expected true, got %v", result)
		}
	})

	t.Run("custom function overrides builtin", func(t *testing.T) {
		t.Parallel()

		customLen := func(args ...any) (any, error) { return 99.0, nil }
		ev := NewEvaluator(WithFunc("len", customLen))
		result, err := ev.Evaluate("len(\"hello\")", nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != 99.0 {
			t.Fatalf("expected 99.0, got %v", result)
		}
	})
}
