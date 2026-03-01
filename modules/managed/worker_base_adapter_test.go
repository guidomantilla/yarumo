package managed

import (
	"context"
	"testing"
)

func TestNewBaseWorker(t *testing.T) {
	t.Parallel()

	worker := NewBaseWorker()
	if worker == nil {
		t.Fatal("expected non-nil worker")
	}
}

func Test_baseWorker_Start(t *testing.T) {
	t.Parallel()

	worker := NewBaseWorker()

	err := worker.Start(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func Test_baseWorker_Stop(t *testing.T) {
	t.Parallel()

	t.Run("stop with valid context closes done", func(t *testing.T) {
		t.Parallel()

		worker := NewBaseWorker()

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

		worker := NewBaseWorker()

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

	t.Run("double stop is safe", func(t *testing.T) {
		t.Parallel()

		worker := NewBaseWorker()

		err := worker.Stop(context.Background())
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		err = worker.Stop(context.Background())
		if err != nil {
			t.Fatalf("expected nil error on second stop, got %v", err)
		}
	})
}

func Test_baseWorker_Done(t *testing.T) {
	t.Parallel()

	worker := NewBaseWorker()

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
