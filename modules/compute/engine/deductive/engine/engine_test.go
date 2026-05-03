package engine

import "testing"

func TestNewEngine(t *testing.T) {
	t.Parallel()

	t.Run("creates engine with defaults", func(t *testing.T) {
		t.Parallel()

		e := NewEngine()
		if e == nil {
			t.Fatal("expected non-nil engine")
		}
	})

	t.Run("creates engine with options", func(t *testing.T) {
		t.Parallel()

		e := NewEngine(WithMaxIterations(10), WithStrategy(FirstMatch))
		if e == nil {
			t.Fatal("expected non-nil engine")
		}
	})
}
