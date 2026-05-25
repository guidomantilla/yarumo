package messaging

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

type evtA struct {
	v int
}

type evtB struct {
	s string //nolint:unused // distinct type used only for routing isolation tests
}

func TestNewPubSub(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil facade", func(t *testing.T) {
		t.Parallel()

		ps := NewPubSub()
		if ps == nil {
			t.Fatal("expected non-nil PubSub")
		}
	})

	t.Run("default factory used when none given", func(t *testing.T) {
		t.Parallel()

		ps := NewPubSub()
		if ps.factory == nil {
			t.Fatal("expected default factory")
		}
	})

	t.Run("custom factory applied", func(t *testing.T) {
		t.Parallel()

		var called int32
		factory := func(_ reflect.Type) Channel[any] {
			atomic.AddInt32(&called, 1)
			return NewPipelineChannel[any]()
		}

		ps := NewPubSub(WithChannelFactory(factory))

		_, err := Subscribe[evtA](ps, func(_ context.Context, _ Message[evtA]) error { return nil })
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		if atomic.LoadInt32(&called) != 1 {
			t.Fatalf("expected factory called once, got %d", called)
		}
	})

	t.Run("nil factory ignored", func(t *testing.T) {
		t.Parallel()

		ps := NewPubSub(WithChannelFactory(nil))
		if ps.factory == nil {
			t.Fatal("expected default factory preserved on nil override")
		}
	})
}

func TestPublish(t *testing.T) {
	t.Parallel()

	t.Run("delivers payload to matching subscriber", func(t *testing.T) {
		t.Parallel()

		ps := NewPubSub()

		got := make(chan evtA, 1)
		_, err := Subscribe[evtA](ps, func(_ context.Context, msg Message[evtA]) error {
			got <- msg.Payload
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = Publish[evtA](context.Background(), ps, evtA{v: 7})
		if err != nil {
			t.Fatalf("Publish returned %v", err)
		}

		select {
		case g := <-got:
			if g.v != 7 {
				t.Fatalf("expected 7, got %d", g.v)
			}
		default:
			t.Fatal("expected payload delivered (direct channel is synchronous)")
		}
	})

	t.Run("routes by Go type — wrong-typed subscriber not invoked", func(t *testing.T) {
		t.Parallel()

		ps := NewPubSub()

		var firedA, firedB int32
		_, err := Subscribe[evtA](ps, func(_ context.Context, _ Message[evtA]) error {
			atomic.AddInt32(&firedA, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe[evtA] returned %v", err)
		}

		_, err = Subscribe[evtB](ps, func(_ context.Context, _ Message[evtB]) error {
			atomic.AddInt32(&firedB, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe[evtB] returned %v", err)
		}

		err = Publish[evtA](context.Background(), ps, evtA{v: 1})
		if err != nil {
			t.Fatalf("Publish returned %v", err)
		}

		if atomic.LoadInt32(&firedA) != 1 {
			t.Fatalf("expected evtA handler fired, got %d", firedA)
		}
		if atomic.LoadInt32(&firedB) != 0 {
			t.Fatalf("expected evtB handler not fired, got %d", firedB)
		}
	})

	t.Run("returns ErrClosed on nil publisher", func(t *testing.T) {
		t.Parallel()

		err := Publish[evtA](context.Background(), nil, evtA{})
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("expected ErrClosed, got %v", err)
		}
	})

	t.Run("publish with no subscribers returns nil", func(t *testing.T) {
		t.Parallel()

		ps := NewPubSub()
		err := Publish[evtA](context.Background(), ps, evtA{v: 1})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("propagates handler error wrapped in ErrSend", func(t *testing.T) {
		t.Parallel()

		ps := NewPubSub()

		boom := errors.New("boom")
		_, err := Subscribe[evtA](ps, func(_ context.Context, _ Message[evtA]) error {
			return boom
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = Publish[evtA](context.Background(), ps, evtA{})
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, boom) {
			t.Fatalf("expected error to wrap boom, got %v", err)
		}
	})
}

func TestSubscribe(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrHandlerNil on nil handler", func(t *testing.T) {
		t.Parallel()

		ps := NewPubSub()
		_, err := Subscribe[evtA](ps, nil)
		if !errors.Is(err, ErrHandlerNil) {
			t.Fatalf("expected ErrHandlerNil, got %v", err)
		}
	})

	t.Run("returns ErrClosed on nil subscriber", func(t *testing.T) {
		t.Parallel()

		_, err := Subscribe[evtA](nil, func(_ context.Context, _ Message[evtA]) error { return nil })
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("expected ErrClosed, got %v", err)
		}
	})

	t.Run("cancel detaches subscription", func(t *testing.T) {
		t.Parallel()

		ps := NewPubSub()

		var fired int32
		cancel, err := Subscribe[evtA](ps, func(_ context.Context, _ Message[evtA]) error {
			atomic.AddInt32(&fired, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		err = Publish[evtA](context.Background(), ps, evtA{})
		if err != nil {
			t.Fatalf("Publish returned %v", err)
		}

		cancel()

		err = Publish[evtA](context.Background(), ps, evtA{})
		if err != nil {
			t.Fatalf("Publish returned %v", err)
		}

		got := atomic.LoadInt32(&fired)
		if got != 1 {
			t.Fatalf("expected fired=1 after cancel, got %d", got)
		}
	})
}

func TestPubSub_LazyChannelAllocation(t *testing.T) {
	t.Parallel()

	t.Run("channel allocated once per type", func(t *testing.T) {
		t.Parallel()

		var built int32
		factory := func(_ reflect.Type) Channel[any] {
			atomic.AddInt32(&built, 1)
			return NewPipelineChannel[any]()
		}

		ps := NewPubSub(WithChannelFactory(factory))

		for range 3 {
			_, err := Subscribe[evtA](ps, func(_ context.Context, _ Message[evtA]) error { return nil })
			if err != nil {
				t.Fatalf("Subscribe returned %v", err)
			}
		}

		err := Publish[evtA](context.Background(), ps, evtA{})
		if err != nil {
			t.Fatalf("Publish returned %v", err)
		}

		got := atomic.LoadInt32(&built)
		if got != 1 {
			t.Fatalf("expected factory called once, got %d", got)
		}
	})
}

func TestPubSub_Concurrent(t *testing.T) {
	t.Parallel()

	ps := NewPubSub()

	var fired int64
	const pubs = 100
	const subs = 4

	for range subs {
		_, err := Subscribe[evtA](ps, func(_ context.Context, _ Message[evtA]) error {
			atomic.AddInt64(&fired, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}
	}

	var wg sync.WaitGroup
	wg.Add(pubs)
	for i := range pubs {
		go func(n int) {
			defer wg.Done()
			err := Publish[evtA](context.Background(), ps, evtA{v: n})
			if err != nil {
				t.Errorf("Publish returned %v", err)
			}
		}(i)
	}

	wg.Wait()

	got := atomic.LoadInt64(&fired)
	want := int64(pubs * subs)
	if got != want {
		t.Fatalf("expected %d deliveries, got %d", want, got)
	}
}
