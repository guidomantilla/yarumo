package controlbus

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// captureErrors returns a thread-safe ErrorHandler that appends every
// reported error and a getter that returns a defensive copy.
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

// captureResults returns a Handler[Result] subscriber that appends every
// dispatched Result to a slice, and a getter for a defensive copy.
func captureResults() (messaging.Handler[Result], func() []Result) {
	var mu sync.Mutex

	captured := []Result{}

	handler := func(_ context.Context, msg messaging.Message[Result]) error {
		mu.Lock()
		defer mu.Unlock()

		captured = append(captured, msg.Payload)

		return nil
	}

	get := func() []Result {
		mu.Lock()
		defer mu.Unlock()

		out := make([]Result, len(captured))
		copy(out, captured)

		return out
	}

	return handler, get
}

func TestNewControlBus(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		cmdChan := messaging.NewPipelineChannel[Command]()
		resChan := messaging.NewPipelineChannel[Result]()

		c := NewControlBus("test", cmdChan, resChan, map[string]Handler{})
		if c == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		cmdChan := messaging.NewPipelineChannel[Command]()
		resChan := messaging.NewPipelineChannel[Result]()

		c := NewControlBus("ops-bus", cmdChan, resChan, map[string]Handler{})
		if c.Name() != "ops-bus" {
			t.Fatalf("expected name ops-bus, got %q", c.Name())
		}
	})

	t.Run("clones handlers map so post-construction mutation is ignored", func(t *testing.T) {
		t.Parallel()

		cmdChan := messaging.NewPipelineChannel[Command]()
		resChan := messaging.NewPipelineChannel[Result]()

		handlers := map[string]Handler{
			"stats": func(_ context.Context, cmd Command) Result {
				return Result{Command: cmd, Success: true, Message: "ok"}
			},
		}

		h, get := captureResults()

		_, err := resChan.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		b := NewControlBus("test", cmdChan, resChan, handlers)

		// Mutate after construction: this must not affect dispatch.
		delete(handlers, "stats")

		err = b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = cmdChan.Send(context.Background(), messaging.Message[Command]{Payload: Command{Verb: "stats"}})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		results := get()
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}

		if !results[0].Success {
			t.Fatalf("expected Success=true, got Result=%+v", results[0])
		}
	})
}

func TestControlBus_KnownVerb(t *testing.T) {
	t.Parallel()

	t.Run("dispatches known verb and publishes Result", func(t *testing.T) {
		t.Parallel()

		cmdChan := messaging.NewPipelineChannel[Command]()
		resChan := messaging.NewPipelineChannel[Result]()

		var called int32

		handlers := map[string]Handler{
			"stats": func(_ context.Context, cmd Command) Result {
				atomic.AddInt32(&called, 1)

				return Result{
					Command: cmd,
					Success: true,
					Message: "uptime 3h",
					Data:    map[string]any{"reqs": 42},
				}
			},
		}

		h, get := captureResults()

		_, err := resChan.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		b := NewControlBus("test", cmdChan, resChan, handlers)

		err = b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		cmd := Command{Verb: "stats", Target: "web", Args: map[string]string{"format": "short"}}

		err = cmdChan.Send(context.Background(), messaging.Message[Command]{Payload: cmd})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got := atomic.LoadInt32(&called); got != 1 {
			t.Fatalf("handler call count: got %d want 1", got)
		}

		results := get()
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}

		got := results[0]
		if !got.Success {
			t.Fatal("expected Success=true")
		}

		if got.Message != "uptime 3h" {
			t.Fatalf("expected Message %q, got %q", "uptime 3h", got.Message)
		}

		if got.Command.Verb != "stats" {
			t.Fatalf("expected echoed Verb stats, got %q", got.Command.Verb)
		}

		if got.Command.Target != "web" {
			t.Fatalf("expected echoed Target web, got %q", got.Command.Target)
		}

		if got.Data["reqs"] != 42 {
			t.Fatalf("expected Data[reqs]=42, got %v", got.Data["reqs"])
		}
	})
}

func TestControlBus_UnknownVerb(t *testing.T) {
	t.Parallel()

	t.Run("default fallback returns Success=false unknown verb", func(t *testing.T) {
		t.Parallel()

		cmdChan := messaging.NewPipelineChannel[Command]()
		resChan := messaging.NewPipelineChannel[Result]()

		h, get := captureResults()

		_, err := resChan.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		b := NewControlBus("test", cmdChan, resChan, map[string]Handler{})

		err = b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = cmdChan.Send(context.Background(), messaging.Message[Command]{Payload: Command{Verb: "mystery"}})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		results := get()
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}

		if results[0].Success {
			t.Fatal("expected Success=false for unknown verb")
		}

		if results[0].Message != "unknown verb" {
			t.Fatalf("expected default message %q, got %q", "unknown verb", results[0].Message)
		}

		if results[0].Command.Verb != "mystery" {
			t.Fatalf("expected echoed verb mystery, got %q", results[0].Command.Verb)
		}
	})

	t.Run("custom UnknownVerbHandler is invoked", func(t *testing.T) {
		t.Parallel()

		cmdChan := messaging.NewPipelineChannel[Command]()
		resChan := messaging.NewPipelineChannel[Result]()

		fallback := func(_ context.Context, cmd Command) Result {
			return Result{Command: cmd, Success: false, Message: "try /help"}
		}

		h, get := captureResults()

		_, err := resChan.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		b := NewControlBus("test", cmdChan, resChan,
			map[string]Handler{},
			WithUnknownVerbHandler(fallback))

		err = b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = cmdChan.Send(context.Background(), messaging.Message[Command]{Payload: Command{Verb: "what"}})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		results := get()
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}

		if results[0].Message != "try /help" {
			t.Fatalf("expected custom message, got %q", results[0].Message)
		}
	})
}

func TestControlBus_HandlerPanic(t *testing.T) {
	t.Parallel()

	t.Run("panic yields Success=false Result and fires ErrorHandler", func(t *testing.T) {
		t.Parallel()

		cmdChan := messaging.NewPipelineChannel[Command]()
		resChan := messaging.NewPipelineChannel[Result]()

		handlers := map[string]Handler{
			"boom": func(_ context.Context, _ Command) Result {
				panic("kaboom")
			},
		}

		errHandler, getErrs := captureErrors()

		h, getResults := captureResults()

		_, err := resChan.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		b := NewControlBus("test", cmdChan, resChan, handlers, WithErrorHandler(errHandler))

		err = b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = cmdChan.Send(context.Background(), messaging.Message[Command]{Payload: Command{Verb: "boom"}})
		if err != nil {
			t.Fatalf("send (should not propagate panic): %v", err)
		}

		results := getResults()
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}

		if results[0].Success {
			t.Fatal("expected Success=false on panic")
		}

		if results[0].Message == "" {
			t.Fatal("expected Message to record the panic value")
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error, got %d", len(errs))
		}

		if !errors.Is(errs[0], ErrHandlerPanic) {
			t.Fatalf("expected ErrHandlerPanic, got %v", errs[0])
		}

		if !errors.Is(errs[0], ErrControlBusFailed) {
			t.Fatalf("expected ErrControlBusFailed, got %v", errs[0])
		}
	})
}

func TestControlBus_ForwardFailure(t *testing.T) {
	t.Parallel()

	t.Run("reports ErrForwardFailed when reply channel errs", func(t *testing.T) {
		t.Parallel()

		cmdChan := messaging.NewPipelineChannel[Command]()

		boom := errors.New("reply down")
		resChan := &failingChannel[Result]{err: boom}

		handlers := map[string]Handler{
			"stats": func(_ context.Context, cmd Command) Result {
				return Result{Command: cmd, Success: true}
			},
		}

		errHandler, getErrs := captureErrors()

		b := NewControlBus("test", cmdChan, resChan, handlers, WithErrorHandler(errHandler))

		err := b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = cmdChan.Send(context.Background(), messaging.Message[Command]{Payload: Command{Verb: "stats"}})
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
			t.Fatalf("expected wrapped boom, got %v", errs[0])
		}
	})

	t.Run("forward failure does not propagate to cmd caller", func(t *testing.T) {
		t.Parallel()

		cmdChan := messaging.NewPipelineChannel[Command]()
		resChan := &failingChannel[Result]{err: errors.New("reply down")}

		handlers := map[string]Handler{
			"x": func(_ context.Context, cmd Command) Result {
				return Result{Command: cmd, Success: true}
			},
		}

		b := NewControlBus("test", cmdChan, resChan, handlers, WithErrorHandler(messaging.SilentErrorHandler))

		err := b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = cmdChan.Send(context.Background(), messaging.Message[Command]{Payload: Command{Verb: "x"}})
		if err != nil {
			t.Fatalf("expected nil from cmdChan.Send, got %v", err)
		}
	})
}

func TestControlBus_MultipleCommands(t *testing.T) {
	t.Parallel()

	t.Run("processes multiple commands in order", func(t *testing.T) {
		t.Parallel()

		cmdChan := messaging.NewPipelineChannel[Command]()
		resChan := messaging.NewPipelineChannel[Result]()

		handlers := map[string]Handler{
			"a": func(_ context.Context, cmd Command) Result {
				return Result{Command: cmd, Success: true, Message: "A"}
			},
			"b": func(_ context.Context, cmd Command) Result {
				return Result{Command: cmd, Success: true, Message: "B"}
			},
		}

		h, get := captureResults()

		_, err := resChan.Subscribe(h)
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		b := NewControlBus("test", cmdChan, resChan, handlers)

		err = b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		verbs := []string{"a", "b", "a", "b", "a"}
		for _, v := range verbs {
			err = cmdChan.Send(context.Background(), messaging.Message[Command]{Payload: Command{Verb: v}})
			if err != nil {
				t.Fatalf("send %q: %v", v, err)
			}
		}

		results := get()
		if len(results) != len(verbs) {
			t.Fatalf("expected %d results, got %d", len(verbs), len(results))
		}

		want := []string{"A", "B", "A", "B", "A"}
		for i := range results {
			if results[i].Message != want[i] {
				t.Fatalf("result[%d] message: got %q want %q", i, results[i].Message, want[i])
			}
		}
	})
}

func TestControlBus_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Start is idempotent", func(t *testing.T) {
		t.Parallel()

		cmdChan := messaging.NewPipelineChannel[Command]()
		resChan := messaging.NewPipelineChannel[Result]()

		var calls int32

		handlers := map[string]Handler{
			"x": func(_ context.Context, cmd Command) Result {
				atomic.AddInt32(&calls, 1)

				return Result{Command: cmd, Success: true}
			},
		}

		b := NewControlBus("test", cmdChan, resChan, handlers)

		err := b.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = b.Start(context.Background())
		if err != nil {
			t.Fatalf("second start: %v", err)
		}

		err = cmdChan.Send(context.Background(), messaging.Message[Command]{Payload: Command{Verb: "x"}})
		if err != nil {
			t.Fatalf("send: %v", err)
		}

		if got := atomic.LoadInt32(&calls); got != 1 {
			t.Fatalf("handler should fire once (no double subscription), got %d", got)
		}
	})

	t.Run("Stop is idempotent", func(t *testing.T) {
		t.Parallel()

		cmdChan := messaging.NewPipelineChannel[Command]()
		resChan := messaging.NewPipelineChannel[Result]()

		b := NewControlBus("test", cmdChan, resChan, map[string]Handler{})

		err := b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = b.Stop(context.Background())
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = b.Stop(context.Background())
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})

	t.Run("Done closes after Stop", func(t *testing.T) {
		t.Parallel()

		cmdChan := messaging.NewPipelineChannel[Command]()
		resChan := messaging.NewPipelineChannel[Result]()

		b := NewControlBus("test", cmdChan, resChan, map[string]Handler{})

		err := b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		select {
		case <-b.Done():
			t.Fatal("Done closed before Stop")
		default:
		}

		err = b.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-b.Done():
		default:
			t.Fatal("Done not closed after Stop")
		}
	})

	t.Run("Stop with expired ctx returns ErrShutdownTimeout", func(t *testing.T) {
		t.Parallel()

		cmdChan := messaging.NewPipelineChannel[Command]()
		resChan := messaging.NewPipelineChannel[Result]()

		b := NewControlBus("test", cmdChan, resChan, map[string]Handler{})

		err := b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = b.Stop(ctx)
		if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Fatalf("expected ErrShutdownTimeout, got %v", err)
		}
	})

	t.Run("Subscription stops receiving after Stop", func(t *testing.T) {
		t.Parallel()

		cmdChan := messaging.NewPipelineChannel[Command]()
		resChan := messaging.NewPipelineChannel[Result]()

		var calls int32

		handlers := map[string]Handler{
			"x": func(_ context.Context, cmd Command) Result {
				atomic.AddInt32(&calls, 1)

				return Result{Command: cmd, Success: true}
			},
		}

		b := NewControlBus("test", cmdChan, resChan, handlers)

		err := b.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = cmdChan.Send(context.Background(), messaging.Message[Command]{Payload: Command{Verb: "x"}})
		if err != nil {
			t.Fatalf("send pre-stop: %v", err)
		}

		err = b.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		err = cmdChan.Send(context.Background(), messaging.Message[Command]{Payload: Command{Verb: "x"}})
		if err != nil {
			t.Fatalf("send post-stop: %v", err)
		}

		if got := atomic.LoadInt32(&calls); got != 1 {
			t.Fatalf("expected 1 call (post-stop ignored), got %d", got)
		}
	})
}

func TestControlBus_Options(t *testing.T) {
	t.Parallel()

	t.Run("WithErrorHandler(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithErrorHandler(nil))
		if opts.errorHandler == nil {
			t.Fatal("expected default error handler preserved on nil arg")
		}
	})

	t.Run("WithUnknownVerbHandler(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithUnknownVerbHandler(nil))
		if opts.unknownVerbHandler == nil {
			t.Fatal("expected default unknown-verb handler preserved on nil arg")
		}
	})

	t.Run("defaults install messaging.DefaultErrorHandler", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.errorHandler == nil {
			t.Fatal("default error handler should be installed")
		}

		if opts.unknownVerbHandler == nil {
			t.Fatal("default unknown-verb handler should be installed")
		}
	})
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
