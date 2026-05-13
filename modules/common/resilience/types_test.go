package resilience_test

import (
	"testing"

	cresilience "github.com/guidomantilla/yarumo/common/resilience"
)

func TestState_String(t *testing.T) {
	t.Parallel()

	t.Run("closed", func(t *testing.T) {
		t.Parallel()

		got := cresilience.StateClosed.String()
		if got != "closed" {
			t.Fatalf("expected closed, got %s", got)
		}
	})

	t.Run("half-open", func(t *testing.T) {
		t.Parallel()

		got := cresilience.StateHalfOpen.String()
		if got != "half-open" {
			t.Fatalf("expected half-open, got %s", got)
		}
	})

	t.Run("open", func(t *testing.T) {
		t.Parallel()

		got := cresilience.StateOpen.String()
		if got != "open" {
			t.Fatalf("expected open, got %s", got)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()

		got := cresilience.State(99).String()
		if got != "unknown" {
			t.Fatalf("expected unknown, got %s", got)
		}
	})
}
