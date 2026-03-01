package managed

import (
	"context"
	"testing"
	"time"
)

func TestBuildCronWorker(t *testing.T) {
	t.Run("build succeeds and stop completes", func(t *testing.T) {
		errCh := make(chan error, 1)

		sched := newMockSchedulerDoneImmediately()

		component, stopFn, err := BuildCronWorker(t.Context(), "test-cron", sched, errCh)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if component.name != "test-cron" {
			t.Fatalf("expected name test-cron, got %s", component.name)
		}

		if stopFn == nil {
			t.Fatal("expected non-nil stopFn")
		}

		time.Sleep(50 * time.Millisecond)

		stopFn(context.Background(), 5*time.Second)

		select {
		case err := <-errCh:
			t.Fatalf("unexpected error: %v", err)
		default:
		}
	})

	t.Run("stop with short timeout logs error", func(t *testing.T) {
		errCh := make(chan error, 1)

		sched := newMockSchedulerNeverDone()

		_, stopFn, err := BuildCronWorker(context.Background(), "test-cron-timeout", sched, errCh)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		stopFn(ctx, time.Nanosecond)
	})

}
