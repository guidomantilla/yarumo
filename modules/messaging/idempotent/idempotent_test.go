package idempotent

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
	"github.com/guidomantilla/yarumo/messaging/store"
)

// captureErrors returns a thread-safe ErrorHandler that appends every
// reported error, and a getter that returns a defensive copy.
func captureErrors() (messaging.ErrorHandler, func() []error) {
	var mu sync.Mutex

	captured := []error{}

	handler := func(_ context.Context, _ any, err error) {
		mu.Lock()
		defer mu.Unlock()

		captured = append(captured, err)
	}

	get := func() []error {
		mu.Lock()
		defer mu.Unlock()

		out := make([]error, len(captured))
		copy(out, captured)

		return out
	}

	return handler, get
}

// captureDrops returns a thread-safe DropHandler that records every
// drop reason, and a getter returning the recorded reasons.
func captureDrops() (DropHandler, func() []DropReason) {
	var mu sync.Mutex

	reasons := []DropReason{}

	handler := func(_ context.Context, _ any, reason DropReason) {
		mu.Lock()
		defer mu.Unlock()

		reasons = append(reasons, reason)
	}

	get := func() []DropReason {
		mu.Lock()
		defer mu.Unlock()

		out := make([]DropReason, len(reasons))
		copy(out, reasons)

		return out
	}

	return handler, get
}

// counter returns a Handler[int] that increments the returned int32
// once per dispatched message.
func counter() (messaging.Handler[int], func() int32) {
	var n int32

	handler := func(_ context.Context, _ messaging.Message[int]) error {
		atomic.AddInt32(&n, 1)

		return nil
	}

	get := func() int32 {
		return atomic.LoadInt32(&n)
	}

	return handler, get
}

// startMetaStore starts an in-memory MetadataStore and registers a
// Cleanup hook that stops it. The returned store is ready for use.
func startMetaStore(t *testing.T) store.MetadataStore {
	t.Helper()

	s := store.NewInMemoryMetadataStore("test", store.WithSweepInterval(20*time.Millisecond))

	c, ok := s.(lifecycle.Component)
	if !ok {
		t.Fatal("expected MetadataStore to implement lifecycle.Component")
	}

	err := c.Start(context.Background())
	if err != nil {
		t.Fatalf("metastore start: %v", err)
	}

	t.Cleanup(func() {
		_ = c.Stop(context.Background())
	})

	return s
}

func TestNewIdempotent(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		ms := startMetaStore(t)

		c := NewIdempotent("test", src, dst, ms)
		if c == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		ms := startMetaStore(t)

		c := NewIdempotent("orders-dedup", src, dst, ms)
		if c.Name() != "orders-dedup" {
			t.Fatalf("expected name orders-dedup, got %q", c.Name())
		}
	})
}

func TestIdempotent_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("first occurrence of a key is forwarded", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		ms := startMetaStore(t)

		h, get := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		i := NewIdempotent("test", src, dst, ms)

		err = i.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		msg := messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{MessageID: "abc"},
		}

		err = src.Send(context.Background(), msg)
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := get(), int32(1); got != want {
			t.Fatalf("forwarded count: got %d want %d", got, want)
		}

		seen, err := ms.Has(context.Background(), "abc")
		if err != nil {
			t.Fatalf("metastore has: %v", err)
		}

		if !seen {
			t.Fatal("expected key abc to be recorded after first forward")
		}
	})

	t.Run("second occurrence with same key is dropped as duplicate", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		ms := startMetaStore(t)

		h, get := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		drops, getDrops := captureDrops()

		i := NewIdempotent("test", src, dst, ms, WithDropHandler[int](drops))

		err = i.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		msg := messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{MessageID: "abc"},
		}

		err = src.Send(context.Background(), msg)
		if err != nil {
			t.Fatalf("first send: %v", err)
		}

		err = src.Send(context.Background(), msg)
		if err != nil {
			t.Fatalf("second send: %v", err)
		}

		if got, want := get(), int32(1); got != want {
			t.Fatalf("forwarded count: got %d want %d (duplicate must drop)", got, want)
		}

		drs := getDrops()
		if len(drs) != 1 {
			t.Fatalf("expected exactly 1 drop, got %d", len(drs))
		}

		if drs[0] != DropReasonDuplicate {
			t.Fatalf("expected DropReasonDuplicate, got %q", drs[0])
		}
	})

	t.Run("empty MessageID is dropped as no-key", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		ms := startMetaStore(t)

		h, get := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		drops, getDrops := captureDrops()

		i := NewIdempotent("test", src, dst, ms, WithDropHandler[int](drops))

		err = i.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		// MessageID intentionally empty.
		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := get(), int32(0); got != want {
			t.Fatalf("forwarded count: got %d want %d (no-key must drop)", got, want)
		}

		drs := getDrops()
		if len(drs) != 1 {
			t.Fatalf("expected exactly 1 drop, got %d", len(drs))
		}

		if drs[0] != DropReasonNoKey {
			t.Fatalf("expected DropReasonNoKey, got %q", drs[0])
		}
	})
}

func TestIdempotent_StoreErrors(t *testing.T) {
	t.Parallel()

	t.Run("Has error fails closed and reports ErrStoreCheck", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		h, get := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		boom := errors.New("store down")
		ms := &fakeMetaStore{hasErr: boom}

		errHandler, getErrs := captureErrors()

		i := NewIdempotent("test", src, dst, ms, WithErrorHandler[int](errHandler))

		err = i.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{MessageID: "abc"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := get(), int32(0); got != want {
			t.Fatalf("forwarded count: got %d want %d (Has error must NOT forward)", got, want)
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error, got %d", len(errs))
		}

		if !errors.Is(errs[0], ErrStoreCheck) {
			t.Fatalf("expected ErrStoreCheck, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped store error, got %v", errs[0])
		}

		if !errors.Is(errs[0], ErrIdempotentFailed) {
			t.Fatalf("expected ErrIdempotentFailed, got %v", errs[0])
		}
	})

	t.Run("Add error fails open and reports ErrStoreAdd", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()

		h, get := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		boom := errors.New("store down")
		ms := &fakeMetaStore{addErr: boom}

		errHandler, getErrs := captureErrors()

		i := NewIdempotent("test", src, dst, ms, WithErrorHandler[int](errHandler))

		err = i.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{MessageID: "abc"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		// Fail-open: forwarding still happens despite Add failure.
		if got, want := get(), int32(1); got != want {
			t.Fatalf("forwarded count: got %d want %d (Add error must still forward)", got, want)
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error, got %d", len(errs))
		}

		if !errors.Is(errs[0], ErrStoreAdd) {
			t.Fatalf("expected ErrStoreAdd, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped store error, got %v", errs[0])
		}
	})
}

func TestIdempotent_ForwardError(t *testing.T) {
	t.Parallel()

	t.Run("forward Send failure is reported via ErrorHandler", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dst down")
		dst := &failingChannel[int]{err: boom}

		ms := startMetaStore(t)

		errHandler, getErrs := captureErrors()

		i := NewIdempotent("test", src, dst, ms, WithErrorHandler[int](errHandler))

		err := i.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{MessageID: "abc"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error, got %d", len(errs))
		}

		if !errors.Is(errs[0], ErrForwardFailed) {
			t.Fatalf("expected ErrForwardFailed, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped destination error, got %v", errs[0])
		}
	})
}

func TestIdempotent_CustomKeyFn(t *testing.T) {
	t.Parallel()

	t.Run("custom KeyFn extracts CorrelationID for dedup", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		ms := startMetaStore(t)

		h, get := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		keyFn := func(msg messaging.Message[int]) string {
			return msg.Headers.CorrelationID
		}

		i := NewIdempotent("test", src, dst, ms, WithKeyFn(keyFn))

		err = i.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		// Two messages with the SAME CorrelationID but DIFFERENT
		// MessageID: with the custom KeyFn the second must dedup.
		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{MessageID: "m1", CorrelationID: "saga-1"},
		})
		if err != nil {
			t.Fatalf("first send: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 2,
			Headers: messaging.Headers{MessageID: "m2", CorrelationID: "saga-1"},
		})
		if err != nil {
			t.Fatalf("second send: %v", err)
		}

		if got, want := get(), int32(1); got != want {
			t.Fatalf("forwarded count: got %d want %d (CorrelationID dedup must drop second)", got, want)
		}
	})
}

func TestIdempotent_TTLExpiry(t *testing.T) {
	t.Parallel()

	t.Run("same key after TTL expires is forwarded again", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		ms := startMetaStore(t)

		h, get := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		ttl := 30 * time.Millisecond

		i := NewIdempotent("test", src, dst, ms, WithTTL[int](ttl))

		err = i.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		msg := messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{MessageID: "abc"},
		}

		err = src.Send(context.Background(), msg)
		if err != nil {
			t.Fatalf("first send: %v", err)
		}

		// Wait past the TTL so the entry is no longer fresh.
		time.Sleep(ttl + 50*time.Millisecond)

		err = src.Send(context.Background(), msg)
		if err != nil {
			t.Fatalf("second send: %v", err)
		}

		if got, want := get(), int32(2); got != want {
			t.Fatalf("forwarded count: got %d want %d (post-TTL must forward again)", got, want)
		}
	})
}

func TestIdempotent_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Start is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		ms := startMetaStore(t)

		h, get := counter()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		i := NewIdempotent("test", src, dst, ms)

		err = i.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = i.Start(context.Background())
		if err != nil {
			t.Fatalf("second start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{MessageID: "abc"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := get(), int32(1); got != want {
			t.Fatalf("dest should receive once (no double subscription), got %d", got)
		}
	})

	t.Run("Stop is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		ms := startMetaStore(t)

		i := NewIdempotent("test", src, dst, ms)

		err := i.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = i.Stop(context.Background())
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = i.Stop(context.Background())
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})

	t.Run("Done closes after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		ms := startMetaStore(t)

		i := NewIdempotent("test", src, dst, ms)

		err := i.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		select {
		case <-i.Done():
			t.Fatal("Done closed before Stop")
		default:
		}

		err = i.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-i.Done():
		default:
			t.Fatal("Done not closed after Stop")
		}
	})

	t.Run("Stop with expired ctx returns ErrShutdownTimeout", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[int]()
		ms := startMetaStore(t)

		i := NewIdempotent("test", src, dst, ms)

		err := i.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = i.Stop(ctx)
		if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Fatalf("expected ErrShutdownTimeout, got %v", err)
		}
	})
}

func TestIdempotent_Options(t *testing.T) {
	t.Parallel()

	t.Run("defaults install messaging.DefaultErrorHandler", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int]()
		if opts.errorHandler == nil {
			t.Fatal("default error handler should be installed")
		}
	})

	t.Run("defaults install nil DropHandler (silent)", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int]()
		if opts.dropHandler != nil {
			t.Fatal("default drop handler should be nil")
		}
	})

	t.Run("defaults install DefaultKeyFn", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int]()
		if opts.keyFn == nil {
			t.Fatal("default key fn should be installed")
		}

		msg := messaging.Message[int]{Headers: messaging.Headers{MessageID: "abc"}}
		if got := opts.keyFn(msg); got != "abc" {
			t.Fatalf("default key fn should extract MessageID, got %q", got)
		}
	})

	t.Run("default TTL is 24h", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions[int]()
		if opts.ttl != defaultTTL {
			t.Fatalf("default ttl: got %v, want %v", opts.ttl, defaultTTL)
		}
	})

	t.Run("WithErrorHandler(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithErrorHandler[int](nil))
		if opts.errorHandler == nil {
			t.Fatal("expected default error handler preserved on nil arg")
		}
	})

	t.Run("WithDropHandler(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		drop := DropHandler(func(_ context.Context, _ any, _ DropReason) {})

		opts := NewOptions(WithDropHandler[int](drop), WithDropHandler[int](nil))
		if opts.dropHandler == nil {
			t.Fatal("expected previously installed drop handler preserved on nil arg")
		}
	})

	t.Run("WithKeyFn(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithKeyFn[int](nil))
		if opts.keyFn == nil {
			t.Fatal("expected default key fn preserved on nil arg")
		}
	})

	t.Run("WithTTL ignores non-positive values", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTTL[int](-time.Second))
		if opts.ttl != defaultTTL {
			t.Fatalf("expected default ttl preserved on negative arg, got %v", opts.ttl)
		}

		opts = NewOptions(WithTTL[int](0))
		if opts.ttl != defaultTTL {
			t.Fatalf("expected default ttl preserved on zero arg, got %v", opts.ttl)
		}
	})

	t.Run("WithTTL applies positive values", func(t *testing.T) {
		t.Parallel()

		want := 5 * time.Minute

		opts := NewOptions(WithTTL[int](want))
		if opts.ttl != want {
			t.Fatalf("expected ttl %v, got %v", want, opts.ttl)
		}
	})
}

// fakeMetaStore is a MetadataStore test double whose Has and Add can
// each be wired to return a configured error. Empty errors mean the
// call succeeds.
type fakeMetaStore struct {
	hasErr error
	addErr error
}

func (f *fakeMetaStore) Has(_ context.Context, _ string) (bool, error) {
	return false, f.hasErr
}

func (f *fakeMetaStore) Add(_ context.Context, _ string, _ time.Duration) error {
	return f.addErr
}

// failingChannel is a Channel[T] test double whose Send always returns
// the configured err. Subscribe is a no-op returning a no-op Cancel.
type failingChannel[T any] struct {
	err error
}

func (c *failingChannel[T]) Send(_ context.Context, _ messaging.Message[T]) error {
	return c.err
}

func (c *failingChannel[T]) Subscribe(_ messaging.Handler[T]) (messaging.Cancel, error) {
	return func() {}, nil
}
