package cache

import (
	"testing"
)

func TestStringKey(t *testing.T) {
	t.Parallel()

	t.Run("string keys are returned verbatim", func(t *testing.T) {
		t.Parallel()

		got, err := stringKey("hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "hello" {
			t.Fatalf("got %q, want hello", got)
		}
	})

	t.Run("int keys are rendered", func(t *testing.T) {
		t.Parallel()

		got, err := stringKey(42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "42" {
			t.Fatalf("got %q, want 42", got)
		}
	})

	t.Run("struct keys are rendered", func(t *testing.T) {
		t.Parallel()

		type k struct {
			Name string
			ID   int
		}
		got, err := stringKey(k{Name: "x", ID: 7})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == "" {
			t.Fatal("expected non-empty rendering of struct key")
		}
	})
}

func TestValidateOptions(t *testing.T) {
	t.Parallel()

	t.Run("nil options returns error", func(t *testing.T) {
		t.Parallel()

		err := validateOptions(nil)
		if err == nil {
			t.Fatal("expected error for nil options")
		}
	})

	t.Run("non-nil options is accepted", func(t *testing.T) {
		t.Parallel()

		err := validateOptions(NewOptions())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
