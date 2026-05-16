package resilience

import (
	"testing"
)

func TestState_String(t *testing.T) {
	t.Parallel()

	t.Run("closed", func(t *testing.T) {
		t.Parallel()

		got := StateClosed.String()
		if got != "closed" {
			t.Fatalf("expected closed, got %s", got)
		}
	})

	t.Run("half-open", func(t *testing.T) {
		t.Parallel()

		got := StateHalfOpen.String()
		if got != "half-open" {
			t.Fatalf("expected half-open, got %s", got)
		}
	})

	t.Run("open", func(t *testing.T) {
		t.Parallel()

		got := StateOpen.String()
		if got != "open" {
			t.Fatalf("expected open, got %s", got)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()

		got := State(99).String()
		if got != "unknown" {
			t.Fatalf("expected unknown, got %s", got)
		}
	})
}
