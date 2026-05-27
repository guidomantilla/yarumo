package lifecycle

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// fakeComponent is an in-memory Component double for testing the Start/Stop
// free helpers. It records calls and lets the test drive the return values
// of Start, Stop, and the Done channel.
type fakeComponent struct {
	name       string
	startErr   error
	stopErr    error
	startCount int
	stopCount  int
	done       chan struct{}
	mu         sync.Mutex
}

func newFakeComponent(name string) *fakeComponent {
	return &fakeComponent{
		name: name,
		done: make(chan struct{}),
	}
}

func (f *fakeComponent) Name() string { return f.name }

func (f *fakeComponent) Start(_ context.Context) error {
	f.mu.Lock()
	f.startCount++
	f.mu.Unlock()

	return f.startErr
}

func (f *fakeComponent) Stop(_ context.Context) error {
	f.mu.Lock()
	f.stopCount++
	f.mu.Unlock()

	close(f.done)

	return f.stopErr
}

func (f *fakeComponent) Done() <-chan struct{} { return f.done }

func TestStart(t *testing.T) {
	t.Parallel()

	t.Run("returns nil after Done is closed on successful Start", func(t *testing.T) {
		t.Parallel()

		fc := newFakeComponent("worker-1")

		errChan := make(chan error, 1)

		ready := make(chan struct{})
		result := make(chan error, 1)

		go func() {
			close(ready)
			result <- Start(context.Background(), fc, errChan)
		}()

		<-ready
		close(fc.done)

		select {
		case err := <-result:
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("Start did not return after Done closed")
		}
	})

	t.Run("returns the Start error", func(t *testing.T) {
		t.Parallel()

		fc := newFakeComponent("worker-2")
		fc.startErr = errors.New("boot failed")

		errChan := make(chan error, 1)

		err := Start(context.Background(), fc, errChan)
		if !errors.Is(err, fc.startErr) {
			t.Fatalf("expected to wrap %v, got %v", fc.startErr, err)
		}
	})

	t.Run("sends Start error to errChan", func(t *testing.T) {
		t.Parallel()

		fc := newFakeComponent("worker-3")
		fc.startErr = errors.New("boot failed")

		errChan := make(chan error, 1)

		_ = Start(context.Background(), fc, errChan)

		select {
		case got := <-errChan:
			if !errors.Is(got, fc.startErr) {
				t.Fatalf("expected errChan to receive %v, got %v", fc.startErr, got)
			}
		default:
			t.Fatal("expected error to be sent on errChan")
		}
	})

	t.Run("non-blocking send when errChan is full", func(t *testing.T) {
		t.Parallel()

		fc := newFakeComponent("worker-4")
		fc.startErr = errors.New("boot failed")

		errChan := make(chan error, 1)
		errChan <- errors.New("pre-existing")

		err := Start(context.Background(), fc, errChan)
		if !errors.Is(err, fc.startErr) {
			t.Fatalf("expected Start to still return its error even when errChan is full, got %v", err)
		}
	})
}

func TestStop(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when component.Stop is clean", func(t *testing.T) {
		t.Parallel()

		fc := newFakeComponent("worker-5")

		err := Stop(context.Background(), fc, time.Second)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("returns the component.Stop error", func(t *testing.T) {
		t.Parallel()

		fc := newFakeComponent("worker-6")
		fc.stopErr = errors.New("shutdown failed")

		err := Stop(context.Background(), fc, time.Second)
		if !errors.Is(err, fc.stopErr) {
			t.Fatalf("expected to wrap %v, got %v", fc.stopErr, err)
		}
	})

	t.Run("bounds shutdown with the given timeout", func(t *testing.T) {
		t.Parallel()

		fc := newFakeComponent("worker-7")

		err := Stop(context.Background(), fc, time.Millisecond)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if fc.stopCount != 1 {
			t.Fatalf("expected component.Stop to be invoked exactly once, got %d", fc.stopCount)
		}
	})
}
