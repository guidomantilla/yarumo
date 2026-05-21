package cron

import (
	"testing"
	"time"

	cron "github.com/robfig/cron/v3"
)

func TestNewScheduler(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil scheduler", func(t *testing.T) {
		t.Parallel()

		s := NewScheduler("cron-1")
		if s == nil {
			t.Fatal("expected non-nil scheduler")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		s := NewScheduler("cron-named")
		if s.Name() != "cron-named" {
			t.Fatalf("expected name %q, got %q", "cron-named", s.Name())
		}
	})

	t.Run("with custom location option", func(t *testing.T) {
		t.Parallel()

		s := NewScheduler("cron-utc", cron.WithLocation(time.UTC))
		if s.Location() != time.UTC {
			t.Fatalf("got location %v, want %v", s.Location(), time.UTC)
		}
	})

	t.Run("entries empty initially", func(t *testing.T) {
		t.Parallel()

		s := NewScheduler("cron-empty")
		if len(s.Entries()) != 0 {
			t.Fatalf("got %d entries, want 0", len(s.Entries()))
		}
	})
}
