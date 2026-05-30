package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

type ctxKey string

func TestMergeContexts(t *testing.T) {
	t.Parallel()

	t.Run("values from publisher ctx reach handler", func(t *testing.T) {
		t.Parallel()

		key := ctxKey("trace-id")
		values := context.WithValue(context.Background(), key, "abc")
		merged := mergeContexts(context.Background(), values)

		got := merged.Value(key)
		if got != "abc" {
			t.Fatalf("expected %q, got %v", "abc", got)
		}
	})

	t.Run("lifecycle ctx values still visible as fallback", func(t *testing.T) {
		t.Parallel()

		key := ctxKey("component-id")
		lifecycleCtx := context.WithValue(context.Background(), key, "topic-1")
		merged := mergeContexts(lifecycleCtx, context.Background())

		got := merged.Value(key)
		if got != "topic-1" {
			t.Fatalf("expected %q, got %v", "topic-1", got)
		}
	})

	t.Run("publisher value overrides lifecycle value", func(t *testing.T) {
		t.Parallel()

		key := ctxKey("k")
		lifecycleCtx := context.WithValue(context.Background(), key, "lifecycle")
		values := context.WithValue(context.Background(), key, "publisher")
		merged := mergeContexts(lifecycleCtx, values)

		got := merged.Value(key)
		if got != "publisher" {
			t.Fatalf("expected publisher value to win, got %v", got)
		}
	})

	t.Run("cancellation follows lifecycle not publisher", func(t *testing.T) {
		t.Parallel()

		lifecycleCtx, cancelLifecycle := context.WithCancel(context.Background())
		defer cancelLifecycle()

		publisherCtx, cancelPublisher := context.WithCancel(context.Background())
		merged := mergeContexts(lifecycleCtx, publisherCtx)

		cancelPublisher()

		select {
		case <-merged.Done():
			t.Fatal("merged ctx must NOT cancel when publisher ctx cancels")
		case <-time.After(50 * time.Millisecond):
		}

		cancelLifecycle()

		select {
		case <-merged.Done():
		case <-time.After(50 * time.Millisecond):
			t.Fatal("merged ctx must cancel when lifecycle ctx cancels")
		}
	})

	t.Run("deadline follows lifecycle not publisher", func(t *testing.T) {
		t.Parallel()

		publisherCtx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		merged := mergeContexts(context.Background(), publisherCtx)

		if _, ok := merged.Deadline(); ok {
			t.Fatal("merged deadline must come from lifecycle, not publisher")
		}
	})

	t.Run("nil values returns lifecycle unchanged", func(t *testing.T) {
		t.Parallel()

		lifecycleCtx := context.Background()
		merged := mergeContexts(lifecycleCtx, nil)

		if merged != lifecycleCtx {
			t.Fatal("expected lifecycle ctx to be returned unchanged when values is nil")
		}
	})

	t.Run("nil lifecycle defaults to background", func(t *testing.T) {
		t.Parallel()

		key := ctxKey("x")
		values := context.WithValue(context.Background(), key, "v")
		merged := mergeContexts(nil, values)

		if merged.Err() != nil {
			t.Fatalf("expected nil Err, got %v", merged.Err())
		}

		if merged.Value(key) != "v" {
			t.Fatal("expected publisher value to remain visible")
		}
	})
}

func TestTopicChannel_HandlerCtxPropagation(t *testing.T) {
	t.Parallel()

	t.Run("publisher ctx values reach handler", func(t *testing.T) {
		t.Parallel()

		ch := NewTopicChannel[int]("topic-ctx-values")
		errChan := make(chan error, 1)

		closeFn, buildErr := lifecycle.Build(context.Background(), ch.(lifecycle.Component), errChan)
		if buildErr != nil {
			t.Fatalf("lifecycle.Build: %v", buildErr)
		}

		defer closeFn(context.Background(), time.Second)

		seen := make(chan any, 1)

		key := ctxKey("trace-id")

		_, err := ch.Subscribe(func(ctx context.Context, _ Message[int]) error {
			seen <- ctx.Value(key)

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		sendCtx := context.WithValue(context.Background(), key, "trace-xyz")

		sendErr := ch.Send(sendCtx, Message[int]{Payload: 1})
		if sendErr != nil {
			t.Fatalf("send: %v", sendErr)
		}

		select {
		case got := <-seen:
			if got != "trace-xyz" {
				t.Fatalf("expected %q, got %v", "trace-xyz", got)
			}
		case <-time.After(time.Second):
			t.Fatal("handler did not run in time")
		}
	})

	t.Run("publisher ctx cancel does not abort handler", func(t *testing.T) {
		t.Parallel()

		ch := NewTopicChannel[int]("topic-ctx-cancel")
		errChan := make(chan error, 1)

		closeFn, buildErr := lifecycle.Build(context.Background(), ch.(lifecycle.Component), errChan)
		if buildErr != nil {
			t.Fatalf("lifecycle.Build: %v", buildErr)
		}

		defer closeFn(context.Background(), time.Second)

		seenDone := make(chan struct{}, 1)

		_, err := ch.Subscribe(func(ctx context.Context, _ Message[int]) error {
			select {
			case <-ctx.Done():
				seenDone <- struct{}{}
			case <-time.After(80 * time.Millisecond):
			}

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		sendCtx, cancel := context.WithCancel(context.Background())

		sendErr := ch.Send(sendCtx, Message[int]{Payload: 1})
		if sendErr != nil {
			t.Fatalf("send: %v", sendErr)
		}

		cancel()

		select {
		case <-seenDone:
			t.Fatal("handler ctx should NOT cancel when publisher cancels")
		case <-time.After(150 * time.Millisecond):
		}
	})
}

func TestQueueChannel_HandlerCtxPropagation(t *testing.T) {
	t.Parallel()

	t.Run("publisher ctx values reach handler", func(t *testing.T) {
		t.Parallel()

		ch := NewQueueChannel[int]("queue-ctx-values")
		errChan := make(chan error, 1)

		closeFn, buildErr := lifecycle.Build(context.Background(), ch.(lifecycle.Component), errChan)
		if buildErr != nil {
			t.Fatalf("lifecycle.Build: %v", buildErr)
		}

		defer closeFn(context.Background(), time.Second)

		seen := make(chan any, 1)

		key := ctxKey("trace-id")

		_, err := ch.Subscribe(func(ctx context.Context, _ Message[int]) error {
			seen <- ctx.Value(key)

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		sendCtx := context.WithValue(context.Background(), key, "queue-xyz")

		sendErr := ch.Send(sendCtx, Message[int]{Payload: 1})
		if sendErr != nil {
			t.Fatalf("send: %v", sendErr)
		}

		select {
		case got := <-seen:
			if got != "queue-xyz" {
				t.Fatalf("expected %q, got %v", "queue-xyz", got)
			}
		case <-time.After(time.Second):
			t.Fatal("handler did not run in time")
		}
	})

	t.Run("publisher ctx cancel does not abort handler", func(t *testing.T) {
		t.Parallel()

		ch := NewQueueChannel[int]("queue-ctx-cancel")
		errChan := make(chan error, 1)

		closeFn, buildErr := lifecycle.Build(context.Background(), ch.(lifecycle.Component), errChan)
		if buildErr != nil {
			t.Fatalf("lifecycle.Build: %v", buildErr)
		}

		defer closeFn(context.Background(), time.Second)

		seenDone := make(chan struct{}, 1)

		_, err := ch.Subscribe(func(ctx context.Context, _ Message[int]) error {
			select {
			case <-ctx.Done():
				seenDone <- struct{}{}
			case <-time.After(80 * time.Millisecond):
			}

			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		sendCtx, cancel := context.WithCancel(context.Background())

		sendErr := ch.Send(sendCtx, Message[int]{Payload: 1})
		if sendErr != nil {
			t.Fatalf("send: %v", sendErr)
		}

		cancel()

		select {
		case <-seenDone:
			t.Fatal("handler ctx should NOT cancel when publisher cancels")
		case <-time.After(150 * time.Millisecond):
		}
	})
}