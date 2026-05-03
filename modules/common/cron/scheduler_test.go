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

		s := NewScheduler()
		if s == nil {
			t.Fatal("expected non-nil scheduler")
		}
	})

	t.Run("with custom location option", func(t *testing.T) {
		t.Parallel()

		s := NewScheduler(cron.WithLocation(time.UTC))
		if s.Location() != time.UTC {
			t.Fatalf("got location %v, want %v", s.Location(), time.UTC)
		}
	})

	t.Run("entries empty initially", func(t *testing.T) {
		t.Parallel()

		s := NewScheduler()
		if len(s.Entries()) != 0 {
			t.Fatalf("got %d entries, want 0", len(s.Entries()))
		}
	})
}
