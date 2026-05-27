package validation

import (
	"errors"
	"testing"
)

func TestCoerce_AsInt(t *testing.T) {
	t.Parallel()

	reg := DefaultRegistry()

	fn, _ := reg.Get("min_len")

	t.Run("int8 not accepted by asInt", func(t *testing.T) {
		t.Parallel()

		// asInt does not have int8; this exercises the default branch.
		err := fn("hello", []any{int8(3)})
		if !errors.Is(err, ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})

	t.Run("int32 accepted", func(t *testing.T) {
		t.Parallel()

		err := fn("hello", []any{int32(3)})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("int64 accepted", func(t *testing.T) {
		t.Parallel()

		err := fn("hello", []any{int64(3)})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("float32 accepted", func(t *testing.T) {
		t.Parallel()

		err := fn("hello", []any{float32(3)})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("float64 accepted", func(t *testing.T) {
		t.Parallel()

		err := fn("hello", []any{float64(3)})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("uint accepted", func(t *testing.T) {
		t.Parallel()

		err := fn("hello", []any{uint(3)})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("uint32 accepted", func(t *testing.T) {
		t.Parallel()

		err := fn("hello", []any{uint32(3)})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("uint64 accepted", func(t *testing.T) {
		t.Parallel()

		err := fn("hello", []any{uint64(3)})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})
}

func TestCoerce_AsFloat(t *testing.T) {
	t.Parallel()

	reg := DefaultRegistry()

	fn, _ := reg.Get("min")

	t.Run("int32 accepted", func(t *testing.T) {
		t.Parallel()

		err := fn(int32(10), []any{5})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("int64 accepted", func(t *testing.T) {
		t.Parallel()

		err := fn(int64(10), []any{5})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("uint accepted", func(t *testing.T) {
		t.Parallel()

		err := fn(uint(10), []any{5})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("uint32 accepted", func(t *testing.T) {
		t.Parallel()

		err := fn(uint32(10), []any{5})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("uint64 accepted", func(t *testing.T) {
		t.Parallel()

		err := fn(uint64(10), []any{5})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("float32 accepted", func(t *testing.T) {
		t.Parallel()

		err := fn(float32(10), []any{5})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("nil rejected", func(t *testing.T) {
		t.Parallel()

		err := fn(nil, []any{5})
		if !errors.Is(err, ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})
}

func TestCoerce_AsSliceArray(t *testing.T) {
	t.Parallel()

	reg := DefaultRegistry()

	fn, _ := reg.Get("non_empty")

	t.Run("array passes", func(t *testing.T) {
		t.Parallel()

		err := fn([3]int{1, 2, 3}, nil)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("nil rejected", func(t *testing.T) {
		t.Parallel()

		err := fn(nil, nil)
		if !errors.Is(err, ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})
}
