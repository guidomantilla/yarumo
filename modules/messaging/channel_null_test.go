package messaging

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

func TestNewNullChannel(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil channel", func(t *testing.T) {
		t.Parallel()

		ch := NewNullChannel[int]()
		if ch == nil {
			t.Fatal("NewNullChannel returned nil")
		}
	})

	t.Run("send returns nil error", func(t *testing.T) {
		t.Parallel()

		ch := NewNullChannel[int](WithErrorHandler(SilentErrorHandler))

		err := ch.Send(context.Background(), Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("Send returned %v, want nil", err)
		}
	})

	t.Run("send fires error handler with ErrDropped", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int32
		var captured atomic.Pointer[error]

		hook := func(_ context.Context, _ any, err error) {
			calls.Add(1)
			captured.Store(&err)
		}

		ch := NewNullChannel[int](WithErrorHandler(hook))

		err := ch.Send(context.Background(), Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("Send returned %v, want nil", err)
		}

		if calls.Load() != 1 {
			t.Fatalf("error handler called %d times, want 1", calls.Load())
		}

		got := captured.Load()
		if got == nil {
			t.Fatal("error handler did not capture an error")
		}

		if !errors.Is(*got, ErrDropped) {
			t.Fatalf("captured error %v does not match ErrDropped", *got)
		}
	})

	t.Run("send with nil ctx returns ErrSend wrapping ErrContextNil", func(t *testing.T) {
		t.Parallel()

		ch := NewNullChannel[int]()

		//nolint:staticcheck // intentional nil ctx to validate guard
		err := ch.Send(nil, Message[int]{Payload: 1})
		if err == nil {
			t.Fatal("Send(nil ctx) returned nil, want error")
		}

		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("error %v does not match ErrContextNil", err)
		}

		if !errors.Is(err, ErrSendFailed) {
			t.Fatalf("error %v does not match ErrSendFailed", err)
		}
	})

	t.Run("multiple sends fire handler per call", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int32

		ch := NewNullChannel[int](WithErrorHandler(func(_ context.Context, _ any, _ error) {
			calls.Add(1)
		}))

		ctx := context.Background()

		for range 5 {
			_ = ch.Send(ctx, Message[int]{Payload: 1})
		}

		if calls.Load() != 5 {
			t.Fatalf("error handler called %d times, want 5", calls.Load())
		}
	})

	t.Run("subscribe nil handler returns ErrHandlerNil", func(t *testing.T) {
		t.Parallel()

		ch := NewNullChannel[int]()

		_, err := ch.Subscribe(nil)
		if err == nil {
			t.Fatal("Subscribe(nil) returned nil error, want error")
		}

		if !errors.Is(err, ErrHandlerNil) {
			t.Fatalf("error %v does not match ErrHandlerNil", err)
		}
	})

	t.Run("subscribed handler never invoked", func(t *testing.T) {
		t.Parallel()

		var invocations atomic.Int32

		ch := NewNullChannel[int](WithErrorHandler(SilentErrorHandler))

		cancel, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error {
			invocations.Add(1)

			return nil
		})
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		defer cancel()

		ctx := context.Background()

		for range 3 {
			_ = ch.Send(ctx, Message[int]{Payload: 1})
		}

		if invocations.Load() != 0 {
			t.Fatalf("handler invoked %d times, want 0 (NullChannel drops every message)", invocations.Load())
		}
	})

	t.Run("cancel is idempotent", func(t *testing.T) {
		t.Parallel()

		ch := NewNullChannel[int]()

		cancel, err := ch.Subscribe(func(_ context.Context, _ Message[int]) error { return nil })
		if err != nil {
			t.Fatalf("Subscribe returned %v", err)
		}

		// Calling Cancel more than once must not panic and must remain a
		// no-op after the first call.
		cancel()
		cancel()
		cancel()
	})

	t.Run("default error handler installed when no option passed", func(t *testing.T) {
		t.Parallel()

		// Without WithErrorHandler, NewNullChannel defaults to
		// DefaultErrorHandler (logs via common/log). Send must complete
		// without panicking — output goes through the global slot.
		ch := NewNullChannel[int]()

		err := ch.Send(context.Background(), Message[int]{Payload: 1})
		if err != nil {
			t.Fatalf("Send returned %v, want nil", err)
		}
	})
}
