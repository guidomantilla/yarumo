package health

import (
	"context"
	"testing"
)

// Tests in this file mutate the global DefaultHealth singleton and cannot
// run in parallel. The linter exclusion in .golangci.yml documents this.

func TestDefaultHealth(t *testing.T) {
	t.Run("is non-nil and usable", func(t *testing.T) {
		if DefaultHealth == nil {
			t.Fatalf("DefaultHealth is nil")
		}

		status, _ := DefaultHealth.Status(context.Background())
		_ = status
	})
}

func TestRegister(t *testing.T) {
	t.Run("delegates to DefaultHealth", func(t *testing.T) {
		original := DefaultHealth
		t.Cleanup(func() { DefaultHealth = original })

		DefaultHealth = NewHealth()

		Register(&stubCheck{name: "ok", result: Result{Status: StatusHealthy}})

		status, results := Aggregate(context.Background())
		if status != StatusHealthy {
			t.Fatalf("status = %v, want StatusHealthy", status)
		}

		if len(results) != 1 {
			t.Fatalf("len(results) = %d, want 1", len(results))
		}

		if results[0].Name != "ok" {
			t.Fatalf("results[0].Name = %q, want %q", results[0].Name, "ok")
		}
	})
}

func TestAggregateDefault(t *testing.T) {
	t.Run("delegates to DefaultHealth.Status", func(t *testing.T) {
		original := DefaultHealth
		t.Cleanup(func() { DefaultHealth = original })

		DefaultHealth = NewHealth()

		// Empty aggregator returns StatusUnknown.
		status, results := Aggregate(context.Background())
		if status != StatusUnknown {
			t.Fatalf("status = %v, want StatusUnknown", status)
		}

		if results != nil {
			t.Fatalf("results = %v, want nil", results)
		}
	})
}
