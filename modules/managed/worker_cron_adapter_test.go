package managed

import (
	"context"
	"testing"
	"time"

	cron "github.com/robfig/cron/v3"
)

type mockScheduler struct {
	startCalled bool
	stopCtx     context.Context
}

func (m *mockScheduler) Start() { m.startCalled = true }

func (m *mockScheduler) Run() {}

func (m *mockScheduler) Stop() context.Context { return m.stopCtx }

func (m *mockScheduler) AddFunc(_ string, _ func()) (cron.EntryID, error) { return 0, nil }

func (m *mockScheduler) AddJob(_ string, _ cron.Job) (cron.EntryID, error) { return 0, nil }

func (m *mockScheduler) Schedule(_ cron.Schedule, _ cron.Job) cron.EntryID { return 0 }

func (m *mockScheduler) Entries() []cron.Entry { return nil }

func (m *mockScheduler) Location() *time.Location { return time.UTC }

func (m *mockScheduler) Entry(_ cron.EntryID) cron.Entry { return cron.Entry{} }

func (m *mockScheduler) Remove(_ cron.EntryID) {}

func newMockSchedulerDoneImmediately() *mockScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return &mockScheduler{stopCtx: ctx}
}

func newMockSchedulerNeverDone() *mockScheduler {
	return &mockScheduler{stopCtx: context.Background()}
}

func TestNewCronWorker(t *testing.T) {
	t.Parallel()

	sched := newMockSchedulerDoneImmediately()
	worker := NewCronWorker(sched)
	if worker == nil {
		t.Fatal("expected non-nil worker")
	}
}

func Test_cronWorker_Start(t *testing.T) {
	t.Parallel()

	sched := newMockSchedulerDoneImmediately()
	worker := NewCronWorker(sched)

	err := worker.Start(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !sched.startCalled {
		t.Fatal("expected scheduler Start to be called")
	}
}

func Test_cronWorker_Stop(t *testing.T) {
	t.Parallel()

	t.Run("stop completes when scheduler stops", func(t *testing.T) {
		t.Parallel()

		sched := newMockSchedulerDoneImmediately()
		worker := NewCronWorker(sched)

		err := worker.Stop(context.Background())
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		select {
		case <-worker.Done():
		default:
			t.Fatal("expected done channel to be closed")
		}
	})

	t.Run("stop with canceled context returns error and closes done", func(t *testing.T) {
		t.Parallel()

		sched := newMockSchedulerNeverDone()
		worker := NewCronWorker(sched)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := worker.Stop(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		select {
		case <-worker.Done():
		default:
			t.Fatal("expected done channel to be closed after timeout")
		}
	})
}

func Test_cronWorker_Done(t *testing.T) {
	t.Parallel()

	sched := newMockSchedulerDoneImmediately()
	worker := NewCronWorker(sched)

	ch := worker.Done()
	if ch == nil {
		t.Fatal("expected non-nil done channel")
	}

	select {
	case <-ch:
		t.Fatal("expected done channel to be open")
	default:
	}
}
