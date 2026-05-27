package breaker

import (
	"testing"
)

func TestState_String(t *testing.T) {
	t.Parallel()

	t.Run("known states", func(t *testing.T) {
		t.Parallel()

		cases := map[State]string{
			StateClosed:   "closed",
			StateHalfOpen: "half-open",
			StateOpen:     "open",
		}

		for state, want := range cases {
			got := state.String()
			if got != want {
				t.Fatalf("State(%d).String() = %q, want %q", state, got, want)
			}
		}
	})

	t.Run("unknown state falls back to \"unknown\"", func(t *testing.T) {
		t.Parallel()

		got := State(99).String()
		if got != "unknown" {
			t.Fatalf("State(99).String() = %q, want %q", got, "unknown")
		}
	})
}
