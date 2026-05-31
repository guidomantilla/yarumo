package claimcheck

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

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

// counter returns a Handler[T] that increments the returned int32
// once per dispatched message.
func counter[T any]() (messaging.Handler[T], func() int32) {
	var n int32

	handler := func(_ context.Context, _ messaging.Message[T]) error {
		atomic.AddInt32(&n, 1)

		return nil
	}

	get := func() int32 {
		return atomic.LoadInt32(&n)
	}

	return handler, get
}

func TestNewClaimCheckIn(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[ClaimCheckReference]()
		ms := store.NewInMemoryMessageStore[int]()

		c := NewClaimCheckIn("test", src, dst, ms)
		if c == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[ClaimCheckReference]()
		ms := store.NewInMemoryMessageStore[int]()

		c := NewClaimCheckIn("orders-claim-in", src, dst, ms)
		if c.Name() != "orders-claim-in" {
			t.Fatalf("expected name orders-claim-in, got %q", c.Name())
		}
	})
}

func TestNewClaimCheckOut(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[ClaimCheckReference]()
		dst := messaging.NewPipelineChannel[int]()
		ms := store.NewInMemoryMessageStore[int]()

		c := NewClaimCheckOut("test", src, dst, ms)
		if c == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[ClaimCheckReference]()
		dst := messaging.NewPipelineChannel[int]()
		ms := store.NewInMemoryMessageStore[int]()

		c := NewClaimCheckOut("orders-claim-out", src, dst, ms)
		if c.Name() != "orders-claim-out" {
			t.Fatalf("expected name orders-claim-out, got %q", c.Name())
		}
	})
}

func TestClaimCheckIn_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("stores original and forwards reference", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[ClaimCheckReference]()
		ms := store.NewInMemoryMessageStore[int]()

		var seen messaging.Message[ClaimCheckReference]

		_, err := dst.Subscribe(func(_ context.Context, ref messaging.Message[ClaimCheckReference]) error {
			seen = ref

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		in := NewClaimCheckIn("in", src, dst, ms, WithKeyGen(func() string { return "k-1" }))

		err = in.Start(context.Background())
		if err != nil {
			t.Fatalf("in start: %v", err)
		}

		original := messaging.Message[int]{
			Payload: 42,
			Headers: messaging.Headers{
				MessageID:     "m-1",
				CorrelationID: "corr-1",
			},
		}

		err = src.Send(context.Background(), original)
		if err != nil {
			t.Fatalf("src send: %v", err)
		}

		if seen.Payload.Key != "k-1" {
			t.Fatalf("expected reference key k-1, got %q", seen.Payload.Key)
		}

		stored, err := ms.Get(context.Background(), "k-1")
		if err != nil {
			t.Fatalf("store get: %v", err)
		}

		if stored.Payload != 42 {
			t.Fatalf("expected stored payload 42, got %d", stored.Payload)
		}

		if stored.Headers.MessageID != "m-1" {
			t.Fatalf("expected stored MessageID m-1, got %q", stored.Headers.MessageID)
		}
	})

	t.Run("preserves CorrelationID on the reference message", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[ClaimCheckReference]()
		ms := store.NewInMemoryMessageStore[int]()

		var seen messaging.Message[ClaimCheckReference]

		_, err := dst.Subscribe(func(_ context.Context, ref messaging.Message[ClaimCheckReference]) error {
			seen = ref

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		in := NewClaimCheckIn("in", src, dst, ms)

		err = in.Start(context.Background())
		if err != nil {
			t.Fatalf("in start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{
			Payload: 1,
			Headers: messaging.Headers{
				MessageID:     "m-1",
				CorrelationID: "trace-xyz",
			},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if seen.Headers.CorrelationID != "trace-xyz" {
			t.Fatalf("expected CorrelationID trace-xyz on reference, got %q", seen.Headers.CorrelationID)
		}
	})
}

func TestClaimCheckIn_StoreErrors(t *testing.T) {
	t.Parallel()

	t.Run("Put failure does NOT forward and reports ErrStorePut", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[ClaimCheckReference]()

		boom := errors.New("store down")
		ms := &fakeMsgStore[int]{putErr: boom}

		h, get := counter[ClaimCheckReference]()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		errHandler, getErrs := captureErrors()

		in := NewClaimCheckIn("in", src, dst, ms, WithErrorHandler(errHandler))

		err = in.Start(context.Background())
		if err != nil {
			t.Fatalf("in start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := get(), int32(0); got != want {
			t.Fatalf("forwarded count: got %d want %d (Put failure must NOT forward)", got, want)
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error, got %d", len(errs))
		}

		if !errors.Is(errs[0], ErrStorePut) {
			t.Fatalf("expected ErrStorePut, got %v", errs[0])
		}

		if !errors.Is(errs[0], boom) {
			t.Fatalf("expected wrapped store error, got %v", errs[0])
		}

		if !errors.Is(errs[0], ErrClaimCheckFailed) {
			t.Fatalf("expected ErrClaimCheckFailed, got %v", errs[0])
		}
	})

	t.Run("forward Send failure is reported via ErrorHandler", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()

		boom := errors.New("dst down")
		dst := &failingChannel[ClaimCheckReference]{err: boom}

		ms := store.NewInMemoryMessageStore[int]()

		errHandler, getErrs := captureErrors()

		in := NewClaimCheckIn("in", src, dst, ms, WithErrorHandler(errHandler))

		err := in.Start(context.Background())
		if err != nil {
			t.Fatalf("in start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
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

func TestClaimCheckIn_KeyGenCustomization(t *testing.T) {
	t.Parallel()

	t.Run("WithKeyGen is used for every claim", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[ClaimCheckReference]()
		ms := store.NewInMemoryMessageStore[int]()

		var (
			mu   sync.Mutex
			n    int
			keys = []string{"a", "b", "c"}
			seen []string
		)

		gen := func() string {
			mu.Lock()
			defer mu.Unlock()

			k := keys[n]
			n++

			return k
		}

		_, err := dst.Subscribe(func(_ context.Context, ref messaging.Message[ClaimCheckReference]) error {
			mu.Lock()
			defer mu.Unlock()

			seen = append(seen, ref.Payload.Key)

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		in := NewClaimCheckIn("in", src, dst, ms, WithKeyGen(gen))

		err = in.Start(context.Background())
		if err != nil {
			t.Fatalf("in start: %v", err)
		}

		for _, p := range []int{1, 2, 3} {
			err = src.Send(context.Background(), messaging.Message[int]{Payload: p})
			if err != nil {
				t.Fatalf("send %d: %v", p, err)
			}
		}

		mu.Lock()
		defer mu.Unlock()

		if len(seen) != 3 {
			t.Fatalf("expected 3 references forwarded, got %d", len(seen))
		}

		for idx, want := range keys {
			if seen[idx] != want {
				t.Fatalf("reference[%d]: got key %q want %q", idx, seen[idx], want)
			}
		}
	})
}

func TestClaimCheckOut_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("retrieves original from store and forwards it", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[ClaimCheckReference]()
		dst := messaging.NewPipelineChannel[int]()
		ms := store.NewInMemoryMessageStore[int]()

		// Seed the store as ClaimCheckIn would have done.
		original := messaging.Message[int]{
			Payload: 99,
			Headers: messaging.Headers{MessageID: "m-1", CorrelationID: "corr-1"},
		}

		err := ms.Put(context.Background(), "k-1", original)
		if err != nil {
			t.Fatalf("seed store: %v", err)
		}

		var seen messaging.Message[int]

		_, err = dst.Subscribe(func(_ context.Context, msg messaging.Message[int]) error {
			seen = msg

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		out := NewClaimCheckOut("out", src, dst, ms)

		err = out.Start(context.Background())
		if err != nil {
			t.Fatalf("out start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[ClaimCheckReference]{
			Payload: ClaimCheckReference{Key: "k-1"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if seen.Payload != 99 {
			t.Fatalf("expected forwarded payload 99, got %d", seen.Payload)
		}

		if seen.Headers.MessageID != "m-1" {
			t.Fatalf("expected forwarded MessageID m-1, got %q", seen.Headers.MessageID)
		}
	})

	t.Run("default deletes entry from store after retrieval", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[ClaimCheckReference]()
		dst := messaging.NewPipelineChannel[int]()
		ms := store.NewInMemoryMessageStore[int]()

		err := ms.Put(context.Background(), "k-1", messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("seed store: %v", err)
		}

		_, err = dst.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		out := NewClaimCheckOut("out", src, dst, ms)

		err = out.Start(context.Background())
		if err != nil {
			t.Fatalf("out start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[ClaimCheckReference]{
			Payload: ClaimCheckReference{Key: "k-1"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		_, err = ms.Get(context.Background(), "k-1")
		if !errors.Is(err, store.ErrStoreNotFound) {
			t.Fatalf("expected key deleted after retrieve, got Get err %v", err)
		}
	})

	t.Run("WithDeleteAfterRetrieve(false) keeps the entry in the store", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[ClaimCheckReference]()
		dst := messaging.NewPipelineChannel[int]()
		ms := store.NewInMemoryMessageStore[int]()

		err := ms.Put(context.Background(), "k-1", messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("seed store: %v", err)
		}

		_, err = dst.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		out := NewClaimCheckOut("out", src, dst, ms, WithDeleteAfterRetrieve(false))

		err = out.Start(context.Background())
		if err != nil {
			t.Fatalf("out start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[ClaimCheckReference]{
			Payload: ClaimCheckReference{Key: "k-1"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		stored, err := ms.Get(context.Background(), "k-1")
		if err != nil {
			t.Fatalf("expected key still present after retrieve, got %v", err)
		}

		if stored.Payload != 1 {
			t.Fatalf("expected stored payload 1, got %d", stored.Payload)
		}
	})
}

func TestClaimCheckOut_StoreErrors(t *testing.T) {
	t.Parallel()

	t.Run("Get error fails closed and reports ErrStoreGet", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[ClaimCheckReference]()
		dst := messaging.NewPipelineChannel[int]()
		ms := store.NewInMemoryMessageStore[int]()

		h, get := counter[int]()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		errHandler, getErrs := captureErrors()

		out := NewClaimCheckOut("out", src, dst, ms, WithErrorHandler(errHandler))

		err = out.Start(context.Background())
		if err != nil {
			t.Fatalf("out start: %v", err)
		}

		// Key that was never Put.
		err = src.Send(context.Background(), messaging.Message[ClaimCheckReference]{
			Payload: ClaimCheckReference{Key: "missing"},
		})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got, want := get(), int32(0); got != want {
			t.Fatalf("forwarded count: got %d want %d (Get failure must NOT forward)", got, want)
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error, got %d", len(errs))
		}

		if !errors.Is(errs[0], ErrStoreGet) {
			t.Fatalf("expected ErrStoreGet, got %v", errs[0])
		}

		if !errors.Is(errs[0], store.ErrStoreNotFound) {
			t.Fatalf("expected wrapped ErrStoreNotFound, got %v", errs[0])
		}
	})
}

func TestClaimCheckRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("In + Out reconstruct the original message end-to-end", func(t *testing.T) {
		t.Parallel()

		// src → [In] → middle (refs) → [Out] → final (originals)
		src := messaging.NewPipelineChannel[int]()
		middle := messaging.NewPipelineChannel[ClaimCheckReference]()
		final := messaging.NewPipelineChannel[int]()
		ms := store.NewInMemoryMessageStore[int]()

		var seen messaging.Message[int]

		_, err := final.Subscribe(func(_ context.Context, msg messaging.Message[int]) error {
			seen = msg

			return nil
		})
		if err != nil {
			t.Fatalf("final subscribe: %v", err)
		}

		in := NewClaimCheckIn("in", src, middle, ms, WithKeyGen(func() string { return "rt-1" }))
		out := NewClaimCheckOut("out", middle, final, ms)

		err = in.Start(context.Background())
		if err != nil {
			t.Fatalf("in start: %v", err)
		}

		err = out.Start(context.Background())
		if err != nil {
			t.Fatalf("out start: %v", err)
		}

		original := messaging.Message[int]{
			Payload: 7,
			Headers: messaging.Headers{
				MessageID:     "rt-mid",
				CorrelationID: "rt-corr",
				Type:          "test.event",
			},
		}

		err = src.Send(context.Background(), original)
		if err != nil {
			t.Fatalf("src send: %v", err)
		}

		if seen.Payload != original.Payload {
			t.Fatalf("payload mismatch end-to-end: got %d want %d", seen.Payload, original.Payload)
		}

		if seen.Headers.MessageID != original.Headers.MessageID {
			t.Fatalf("MessageID mismatch end-to-end: got %q want %q", seen.Headers.MessageID, original.Headers.MessageID)
		}

		if seen.Headers.CorrelationID != original.Headers.CorrelationID {
			t.Fatalf("CorrelationID mismatch end-to-end: got %q want %q", seen.Headers.CorrelationID, original.Headers.CorrelationID)
		}

		if seen.Headers.Type != original.Headers.Type {
			t.Fatalf("Type mismatch end-to-end: got %q want %q", seen.Headers.Type, original.Headers.Type)
		}

		// Default WithDeleteAfterRetrieve(true) means the store
		// should have been cleaned up after Out forwarded.
		_, err = ms.Get(context.Background(), "rt-1")
		if !errors.Is(err, store.ErrStoreNotFound) {
			t.Fatalf("expected store cleaned up after round trip, got %v", err)
		}
	})
}

func TestClaimCheckIn_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Start is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[ClaimCheckReference]()
		ms := store.NewInMemoryMessageStore[int]()

		h, get := counter[ClaimCheckReference]()

		_, err := dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		in := NewClaimCheckIn("test", src, dst, ms)

		err = in.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = in.Start(context.Background())
		if err != nil {
			t.Fatalf("second start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[int]{Payload: 1})
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
		dst := messaging.NewPipelineChannel[ClaimCheckReference]()
		ms := store.NewInMemoryMessageStore[int]()

		in := NewClaimCheckIn("test", src, dst, ms)

		err := in.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = in.Stop(context.Background())
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = in.Stop(context.Background())
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})

	t.Run("Done closes after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[ClaimCheckReference]()
		ms := store.NewInMemoryMessageStore[int]()

		in := NewClaimCheckIn("test", src, dst, ms)

		err := in.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		select {
		case <-in.Done():
			t.Fatal("Done closed before Stop")
		default:
		}

		err = in.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-in.Done():
		default:
			t.Fatal("Done not closed after Stop")
		}
	})

	t.Run("Stop with expired ctx returns ErrShutdownTimeout", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[int]()
		dst := messaging.NewPipelineChannel[ClaimCheckReference]()
		ms := store.NewInMemoryMessageStore[int]()

		in := NewClaimCheckIn("test", src, dst, ms)

		err := in.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = in.Stop(ctx)
		if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Fatalf("expected ErrShutdownTimeout, got %v", err)
		}
	})
}

func TestClaimCheckOut_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Start is idempotent", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[ClaimCheckReference]()
		dst := messaging.NewPipelineChannel[int]()
		ms := store.NewInMemoryMessageStore[int]()

		err := ms.Put(context.Background(), "k-1", messaging.Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("seed store: %v", err)
		}

		h, get := counter[int]()

		_, err = dst.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		out := NewClaimCheckOut("test", src, dst, ms, WithDeleteAfterRetrieve(false))

		err = out.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = out.Start(context.Background())
		if err != nil {
			t.Fatalf("second start: %v", err)
		}

		err = src.Send(context.Background(), messaging.Message[ClaimCheckReference]{
			Payload: ClaimCheckReference{Key: "k-1"},
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

		src := messaging.NewPipelineChannel[ClaimCheckReference]()
		dst := messaging.NewPipelineChannel[int]()
		ms := store.NewInMemoryMessageStore[int]()

		out := NewClaimCheckOut("test", src, dst, ms)

		err := out.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = out.Stop(context.Background())
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = out.Stop(context.Background())
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})

	t.Run("Done closes after Stop", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[ClaimCheckReference]()
		dst := messaging.NewPipelineChannel[int]()
		ms := store.NewInMemoryMessageStore[int]()

		out := NewClaimCheckOut("test", src, dst, ms)

		err := out.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		select {
		case <-out.Done():
			t.Fatal("Done closed before Stop")
		default:
		}

		err = out.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-out.Done():
		default:
			t.Fatal("Done not closed after Stop")
		}
	})

	t.Run("Stop with expired ctx returns ErrShutdownTimeout", func(t *testing.T) {
		t.Parallel()

		src := messaging.NewPipelineChannel[ClaimCheckReference]()
		dst := messaging.NewPipelineChannel[int]()
		ms := store.NewInMemoryMessageStore[int]()

		out := NewClaimCheckOut("test", src, dst, ms)

		err := out.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = out.Stop(ctx)
		if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Fatalf("expected ErrShutdownTimeout, got %v", err)
		}
	})
}

func TestClaimCheck_Options(t *testing.T) {
	t.Parallel()

	t.Run("defaults install messaging.DefaultErrorHandler", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.errorHandler == nil {
			t.Fatal("default error handler should be installed")
		}
	})

	t.Run("defaults install non-nil keyGen", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.keyGen == nil {
			t.Fatal("default key gen should be installed")
		}

		k := opts.keyGen()
		if k == "" {
			t.Fatal("default key gen should produce a non-empty key")
		}
	})

	t.Run("defaults install deleteAfterRetrieve true", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if !opts.deleteAfterRetrieve {
			t.Fatal("default deleteAfterRetrieve should be true")
		}
	})

	t.Run("WithErrorHandler(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithErrorHandler(nil))
		if opts.errorHandler == nil {
			t.Fatal("expected default error handler preserved on nil arg")
		}
	})

	t.Run("WithKeyGen(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithKeyGen(nil))
		if opts.keyGen == nil {
			t.Fatal("expected default key gen preserved on nil arg")
		}
	})

	t.Run("WithDeleteAfterRetrieve(false) flips the flag", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDeleteAfterRetrieve(false))
		if opts.deleteAfterRetrieve {
			t.Fatal("expected deleteAfterRetrieve false")
		}
	})
}

// fakeMsgStore is a MessageStore[T] test double whose Put/Get/Delete
// can each be wired to return a configured error. Empty errors mean
// the call succeeds (Get returns the zero-value Message[T] in that
// case — only useful when the test does not assert on the retrieved
// payload).
type fakeMsgStore[T any] struct {
	putErr error
	getErr error
	delErr error
}

func (f *fakeMsgStore[T]) Put(_ context.Context, _ string, _ messaging.Message[T]) error {
	return f.putErr
}

func (f *fakeMsgStore[T]) Get(_ context.Context, _ string) (messaging.Message[T], error) {
	var zero messaging.Message[T]

	return zero, f.getErr
}

func (f *fakeMsgStore[T]) Delete(_ context.Context, _ string) error {
	return f.delErr
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
