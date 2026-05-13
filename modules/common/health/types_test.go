package health

import (
	"testing"
)

const (
	unknownLabel   = "unknown"
	healthyLabel   = "healthy"
	degradedLabel  = "degraded"
	unhealthyLabel = "unhealthy"
)

func TestStatus_String(t *testing.T) {
	t.Parallel()

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()

		got := StatusUnknown.String()
		if got != unknownLabel {
			t.Fatalf("StatusUnknown.String() = %q, want %q", got, unknownLabel)
		}
	})

	t.Run("healthy", func(t *testing.T) {
		t.Parallel()

		got := StatusHealthy.String()
		if got != healthyLabel {
			t.Fatalf("StatusHealthy.String() = %q, want %q", got, healthyLabel)
		}
	})

	t.Run("degraded", func(t *testing.T) {
		t.Parallel()

		got := StatusDegraded.String()
		if got != degradedLabel {
			t.Fatalf("StatusDegraded.String() = %q, want %q", got, degradedLabel)
		}
	})

	t.Run("unhealthy", func(t *testing.T) {
		t.Parallel()

		got := StatusUnhealthy.String()
		if got != unhealthyLabel {
			t.Fatalf("StatusUnhealthy.String() = %q, want %q", got, unhealthyLabel)
		}
	})

	t.Run("out of range falls back to unknown", func(t *testing.T) {
		t.Parallel()

		bogus := Status(99)

		got := bogus.String()
		if got != unknownLabel {
			t.Fatalf("Status(99).String() = %q, want %q", got, unknownLabel)
		}
	})
}

func TestStatus_Ordering(t *testing.T) {
	t.Parallel()

	// The aggregator relies on the integer ordering of the Status enum:
	// higher integer means worse. This test pins that contract so a future
	// reorder of the const block is caught immediately.
	if StatusUnknown >= StatusHealthy {
		t.Fatalf("expected StatusUnknown < StatusHealthy")
	}

	if StatusHealthy >= StatusDegraded {
		t.Fatalf("expected StatusHealthy < StatusDegraded")
	}

	if StatusDegraded >= StatusUnhealthy {
		t.Fatalf("expected StatusDegraded < StatusUnhealthy")
	}
}
