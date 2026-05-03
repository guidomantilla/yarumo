package expressions

import (
	"errors"
	"testing"
)

func TestDefaultFuncs(t *testing.T) {
	t.Parallel()

	t.Run("returns all 9 built-in functions", func(t *testing.T) {
		t.Parallel()
		funcs := DefaultFuncs()
		expected := []string{"len", "sum", "min", "max", "avg", "abs", "contains", "lower", "upper"}
		for _, name := range expected {
			if _, ok := funcs[name]; !ok {
				t.Fatalf("expected function %s", name)
			}
		}
	})
}

func TestBuiltinLen(t *testing.T) {
	t.Parallel()

	t.Run("string length", func(t *testing.T) {
		t.Parallel()
		v, err := builtinLen("hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 5.0 {
			t.Fatalf("expected 5, got %v", v)
		}
	})

	t.Run("slice length", func(t *testing.T) {
		t.Parallel()
		v, err := builtinLen([]any{1.0, 2.0, 3.0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 3.0 {
			t.Fatalf("expected 3, got %v", v)
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinLen()
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrArgCount) {
			t.Fatalf("expected ErrArgCount, got %v", err)
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		t.Parallel()
		_, err := builtinLen(42)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, ErrTypeMismatch) {
			t.Fatalf("expected ErrTypeMismatch, got %v", err)
		}
	})
}

func TestBuiltinSum(t *testing.T) {
	t.Parallel()

	t.Run("sums numeric slice", func(t *testing.T) {
		t.Parallel()
		v, err := builtinSum([]any{1.0, 2.0, 3.0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 6.0 {
			t.Fatalf("expected 6, got %v", v)
		}
	})

	t.Run("empty slice returns 0", func(t *testing.T) {
		t.Parallel()
		v, err := builtinSum([]any{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 0.0 {
			t.Fatalf("expected 0, got %v", v)
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinSum()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-slice arg", func(t *testing.T) {
		t.Parallel()
		_, err := builtinSum(42)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-numeric element", func(t *testing.T) {
		t.Parallel()
		_, err := builtinSum([]any{1.0, "bad"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestBuiltinMin(t *testing.T) {
	t.Parallel()

	t.Run("finds minimum", func(t *testing.T) {
		t.Parallel()
		v, err := builtinMin([]any{3.0, 1.0, 2.0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 1.0 {
			t.Fatalf("expected 1, got %v", v)
		}
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()
		v, err := builtinMin([]any{5.0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 5.0 {
			t.Fatalf("expected 5, got %v", v)
		}
	})

	t.Run("empty slice error", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMin([]any{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMin()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-slice arg", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMin(42)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-numeric first element", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMin([]any{"bad"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-numeric later element", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMin([]any{1.0, "bad"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestBuiltinMax(t *testing.T) {
	t.Parallel()

	t.Run("finds maximum", func(t *testing.T) {
		t.Parallel()
		v, err := builtinMax([]any{1.0, 3.0, 2.0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 3.0 {
			t.Fatalf("expected 3, got %v", v)
		}
	})

	t.Run("empty slice error", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMax([]any{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMax()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-slice arg", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMax(42)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-numeric first element", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMax([]any{"bad"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-numeric later element", func(t *testing.T) {
		t.Parallel()
		_, err := builtinMax([]any{1.0, "bad"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestBuiltinAvg(t *testing.T) {
	t.Parallel()

	t.Run("computes average", func(t *testing.T) {
		t.Parallel()
		v, err := builtinAvg([]any{1.0, 2.0, 3.0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 2.0 {
			t.Fatalf("expected 2, got %v", v)
		}
	})

	t.Run("empty slice error", func(t *testing.T) {
		t.Parallel()
		_, err := builtinAvg([]any{})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinAvg()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-slice arg", func(t *testing.T) {
		t.Parallel()
		_, err := builtinAvg(42)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-numeric element", func(t *testing.T) {
		t.Parallel()
		_, err := builtinAvg([]any{1.0, "bad"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestBuiltinAbs(t *testing.T) {
	t.Parallel()

	t.Run("positive number", func(t *testing.T) {
		t.Parallel()
		v, err := builtinAbs(5.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 5.0 {
			t.Fatalf("expected 5, got %v", v)
		}
	})

	t.Run("negative number", func(t *testing.T) {
		t.Parallel()
		v, err := builtinAbs(-5.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != 5.0 {
			t.Fatalf("expected 5, got %v", v)
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinAbs()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-numeric", func(t *testing.T) {
		t.Parallel()
		_, err := builtinAbs("bad")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestBuiltinContains(t *testing.T) {
	t.Parallel()

	t.Run("string contains substring", func(t *testing.T) {
		t.Parallel()
		v, err := builtinContains("hello world", "world")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("string does not contain", func(t *testing.T) {
		t.Parallel()
		v, err := builtinContains("hello", "world")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != false {
			t.Fatalf("expected false, got %v", v)
		}
	})

	t.Run("slice contains element", func(t *testing.T) {
		t.Parallel()
		v, err := builtinContains([]any{"a", "b"}, "b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Fatalf("expected true, got %v", v)
		}
	})

	t.Run("slice does not contain", func(t *testing.T) {
		t.Parallel()
		v, err := builtinContains([]any{"a", "b"}, "c")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != false {
			t.Fatalf("expected false, got %v", v)
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinContains("a")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("unsupported collection type", func(t *testing.T) {
		t.Parallel()
		_, err := builtinContains(42, "x")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("string contains with non-string search", func(t *testing.T) {
		t.Parallel()
		_, err := builtinContains("hello", 42)
		if err == nil {
			t.Fatal("expected error for non-string search in string")
		}
	})
}

func TestBuiltinLower(t *testing.T) {
	t.Parallel()

	t.Run("converts to lowercase", func(t *testing.T) {
		t.Parallel()
		v, err := builtinLower("HELLO")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != "hello" {
			t.Fatalf("expected hello, got %v", v)
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinLower()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-string", func(t *testing.T) {
		t.Parallel()
		_, err := builtinLower(42)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestBuiltinUpper(t *testing.T) {
	t.Parallel()

	t.Run("converts to uppercase", func(t *testing.T) {
		t.Parallel()
		v, err := builtinUpper("hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != "HELLO" {
			t.Fatalf("expected HELLO, got %v", v)
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		t.Parallel()
		_, err := builtinUpper()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-string", func(t *testing.T) {
		t.Parallel()
		_, err := builtinUpper(42)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
