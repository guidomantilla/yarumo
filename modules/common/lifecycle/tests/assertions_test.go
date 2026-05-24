package tests

import (
	"context"
	"sync"
	"testing"

	"github.com/guidomantilla/yarumo/common/lifecycle"
)

// fakeComponent is a minimal lifecycle.Component used to exercise the
// assertion helpers without importing a concrete implementation from
// outside common/ (which would invert the dependency direction).
type fakeComponent struct {
	name string
	done chan struct{}
	once sync.Once
}

func newFakeComponent(name string) lifecycle.Component {
	return &fakeComponent{name: name, done: make(chan struct{})}
}

func (f *fakeComponent) Name() string                  { return f.name }
func (f *fakeComponent) Start(_ context.Context) error { return nil }
func (f *fakeComponent) Done() <-chan struct{}         { return f.done }

func (f *fakeComponent) Stop(_ context.Context) error {
	f.once.Do(func() { close(f.done) })
	return nil
}

func TestAssertIdempotentStop(t *testing.T) {
	t.Parallel()

	t.Run("passes for a fresh component", func(t *testing.T) {
		t.Parallel()

		c := newFakeComponent("idempotent-1")
		AssertIdempotentStop(t, c)
	})

	t.Run("passes for a worker-style component already Started", func(t *testing.T) {
		t.Parallel()

		c := newFakeComponent("idempotent-2")

		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("Start returned %v", err)
		}

		AssertIdempotentStop(t, c)
	})

	t.Run("passes for a component already Stopped once", func(t *testing.T) {
		t.Parallel()

		c := newFakeComponent("idempotent-3")

		err := c.Stop(context.Background())
		if err != nil {
			t.Fatalf("first Stop returned %v", err)
		}

		AssertIdempotentStop(t, c)
	})
}
