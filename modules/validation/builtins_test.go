package validation_test

import (
	"errors"
	"testing"

	cvalidation "github.com/guidomantilla/yarumo/common/validation"

	"github.com/guidomantilla/yarumo/validation"
)

// runByName looks up a leaf in the default registry and runs it. It
// concentrates the dispatch path used by the engine and keeps the tests
// focused on the leaves' contract.
func runByName(t *testing.T, name string, value any, params []any) error {
	t.Helper()

	reg := validation.DefaultRegistry()

	fn, ok := reg.Get(name)
	if !ok {
		t.Fatalf("rule %q not in default registry", name)
	}

	return fn(value, params)
}

func TestBuiltin_RequiredAndMustBeUndefined(t *testing.T) {
	t.Parallel()

	t.Run("required passes", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "required", "x", nil)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("required fails on empty", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "required", "", nil)
		if !errors.Is(err, cvalidation.ErrFieldRequired) {
			t.Fatalf("expected ErrFieldRequired, got %v", err)
		}
	})

	t.Run("must_be_undefined passes empty", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "must_be_undefined", "", nil)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("must_be_undefined fails non-empty", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "must_be_undefined", "x", nil)
		if !errors.Is(err, cvalidation.ErrFieldMustBeUndefined) {
			t.Fatalf("expected ErrFieldMustBeUndefined, got %v", err)
		}
	})
}

func TestBuiltin_StringLeaves(t *testing.T) {
	t.Parallel()

	t.Run("min_len passes", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "min_len", "hello", []any{3})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("min_len fails short", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "min_len", "hi", []any{5})
		if !errors.Is(err, cvalidation.ErrMinLen) {
			t.Fatalf("expected ErrMinLen, got %v", err)
		}
	})

	t.Run("min_len bad value type", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "min_len", 42, []any{1})
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})

	t.Run("min_len missing param", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "min_len", "hi", nil)
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})

	t.Run("max_len passes", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "max_len", "hi", []any{5})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("max_len fails long", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "max_len", "very long", []any{3})
		if !errors.Is(err, cvalidation.ErrMaxLen) {
			t.Fatalf("expected ErrMaxLen, got %v", err)
		}
	})

	t.Run("max_len bad value type", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "max_len", 42, []any{1})
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})

	t.Run("regex passes", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "regex", "abc", []any{`^[a-z]+$`})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("regex fails", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "regex", "ABC", []any{`^[a-z]+$`})
		if !errors.Is(err, cvalidation.ErrRegexMismatch) {
			t.Fatalf("expected ErrRegexMismatch, got %v", err)
		}
	})

	t.Run("regex bad value type", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "regex", 42, []any{`^x$`})
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})

	t.Run("regex bad pattern type", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "regex", "x", []any{42})
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})

	t.Run("email passes", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "email", "a@b.com", nil)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("email bad value", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "email", "x", nil)
		if !errors.Is(err, cvalidation.ErrEmailInvalid) {
			t.Fatalf("expected ErrEmailInvalid, got %v", err)
		}
	})

	t.Run("email bad value type", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "email", 42, nil)
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})

	t.Run("url passes", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "url", "https://example.com", nil)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("url bad value", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "url", "abc", nil)
		if !errors.Is(err, cvalidation.ErrURLInvalid) {
			t.Fatalf("expected ErrURLInvalid, got %v", err)
		}
	})

	t.Run("url bad value type", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "url", 42, nil)
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})
}

func TestBuiltin_NumericLeaves(t *testing.T) {
	t.Parallel()

	t.Run("min passes", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "min", 10, []any{5})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("min fails", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "min", 1, []any{5})
		if !errors.Is(err, cvalidation.ErrMinValue) {
			t.Fatalf("expected ErrMinValue, got %v", err)
		}
	})

	t.Run("min bad value type", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "min", "x", []any{5})
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})

	t.Run("min missing param", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "min", 5, nil)
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})

	t.Run("max passes", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "max", 5, []any{10})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("max fails", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "max", 20, []any{10})
		if !errors.Is(err, cvalidation.ErrMaxValue) {
			t.Fatalf("expected ErrMaxValue, got %v", err)
		}
	})

	t.Run("max bad value type", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "max", "x", []any{10})
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})

	t.Run("max missing param", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "max", 5, nil)
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})

	t.Run("in_range passes", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "in_range", 50, []any{1, 100})
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("in_range fails", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "in_range", 200, []any{1, 100})
		if !errors.Is(err, cvalidation.ErrOutOfRange) {
			t.Fatalf("expected ErrOutOfRange, got %v", err)
		}
	})

	t.Run("in_range missing param", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "in_range", 50, []any{1})
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})

	t.Run("in_range bad value type", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "in_range", "x", []any{1, 100})
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})

	t.Run("in_range bad lo param type", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "in_range", 50, []any{"x", 100})
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})
}

func TestBuiltin_IDLeaves(t *testing.T) {
	t.Parallel()

	t.Run("uuid passes", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "uuid", "550e8400-e29b-41d4-a716-446655440000", nil)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("uuid bad value type", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "uuid", 42, nil)
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})

	t.Run("ulid passes", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "ulid", "01ARZ3NDEKTSV4RRFFQ69G5FAV", nil)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("ulid bad value type", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "ulid", 42, nil)
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})

	t.Run("ulid bad value", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "ulid", "not-ulid", nil)
		if !errors.Is(err, cvalidation.ErrULIDInvalid) {
			t.Fatalf("expected ErrULIDInvalid, got %v", err)
		}
	})
}

func TestBuiltin_NonEmpty(t *testing.T) {
	t.Parallel()

	t.Run("non_empty passes typed slice", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "non_empty", []int{1, 2}, nil)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("non_empty fails", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "non_empty", []int{}, nil)
		if !errors.Is(err, cvalidation.ErrCollectionEmpty) {
			t.Fatalf("expected ErrCollectionEmpty, got %v", err)
		}
	})

	t.Run("non_empty bad value type", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "non_empty", 42, nil)
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})

	t.Run("non_empty any slice", func(t *testing.T) {
		t.Parallel()

		err := runByName(t, "non_empty", []any{"x"}, nil)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})
}

func TestCoerce_AsInt(t *testing.T) {
	t.Parallel()

	reg := validation.DefaultRegistry()

	fn, _ := reg.Get("min_len")

	t.Run("int8 not accepted by asInt", func(t *testing.T) {
		t.Parallel()

		// asInt does not have int8; this exercises the default branch.
		err := fn("hello", []any{int8(3)})
		if !errors.Is(err, validation.ErrBadParams) {
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

	reg := validation.DefaultRegistry()

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
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})
}

func TestCoerce_AsSliceArray(t *testing.T) {
	t.Parallel()

	reg := validation.DefaultRegistry()

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
		if !errors.Is(err, validation.ErrBadParams) {
			t.Fatalf("expected ErrBadParams, got %v", err)
		}
	})
}

func TestRegistry_Names(t *testing.T) {
	t.Parallel()

	reg := validation.NewRegistry()
	reg.Register("a", func(any, []any) error { return nil })
	reg.Register("b", func(any, []any) error { return nil })

	names := reg.Names()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
}

func TestOptions_WithEvaluator(t *testing.T) {
	t.Parallel()

	// WithEvaluator(nil) should silently keep the default; pass a non-nil
	// evaluator and verify the engine still runs.
	rs := validation.Ruleset{Rules: []validation.RuleNode{{
		Field: "Name",
		Rules: []validation.RuleNode{{Name: "required"}},
	}}}

	eng := validation.NewEngine(rs,
		validation.WithEvaluator(nil),
		validation.WithRegistry(nil),
	)

	err := eng.Validate(struct{ Name string }{Name: "x"}, nil)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestErrors_ErrorMethods(t *testing.T) {
	t.Parallel()

	t.Run("engine error", func(t *testing.T) {
		t.Parallel()

		err := validation.ErrEngine(validation.ErrBadRule)

		msg := err.Error()
		if msg == "" {
			t.Fatalf("expected non-empty message")
		}
	})

	t.Run("load error", func(t *testing.T) {
		t.Parallel()

		err := validation.ErrLoad(validation.ErrDataNil)

		msg := err.Error()
		if msg == "" {
			t.Fatalf("expected non-empty message")
		}
	})
}
