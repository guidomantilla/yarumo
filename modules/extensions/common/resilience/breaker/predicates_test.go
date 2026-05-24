package breaker

import (
	"testing"
)

func TestNoopOnStateChange(t *testing.T) {
	t.Parallel()

	t.Run("does not panic on any input", func(t *testing.T) {
		t.Parallel()

		NoopOnStateChange("", StateClosed, StateOpen)
		NoopOnStateChange("payments", StateOpen, StateHalfOpen)
		NoopOnStateChange("any-name", StateHalfOpen, StateClosed)
	})
}
