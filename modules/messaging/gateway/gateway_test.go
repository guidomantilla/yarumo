package gateway

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	cuids "github.com/guidomantilla/yarumo/core/common/uids"
	"github.com/guidomantilla/yarumo/messaging"
)

// newCounterUID returns a UID that emits sequential test ids
// ("id-1", "id-2", ...). Goroutine-safe.
func newCounterUID() cuids.UID {
	var n int64

	return cuids.NewUID("test", func() (string, error) {
		v := atomic.AddInt64(&n, 1)

		return fmt.Sprintf("id-%d", v), nil
	})
}

// newFailingUID returns a UID whose Generate always returns the
// supplied err.
func newFailingUID(err error) cuids.UID {
	return cuids.NewUID("failing", func() (string, error) {
		return "", err
	})
}

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

// wireMockSubscriber subscribes a handler to requestChan that echoes
// each request to replyChan, preserving Headers.CorrelationID so the
// gateway can route the reply. transform maps the inbound Req to the
// outbound Res.
func wireMockSubscriber[Req, Res any](t *testing.T, requestChan messaging.Channel[Req], replyChan messaging.Channel[Res], transform func(Req) Res) {
	t.Helper()

	_, err := requestChan.Subscribe(func(ctx context.Context, msg messaging.Message[Req]) error {
		reply := messaging.Message[Res]{
			Payload: transform(msg.Payload),
			Headers: messaging.Headers{
				CorrelationID: msg.Headers.CorrelationID,
			},
		}

		return replyChan.Send(ctx, reply)
	})
	if err != nil {
		t.Fatalf("subscribe mock: %v", err)
	}
}

func TestNewGateway(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil component", func(t *testing.T) {
		t.Parallel()

		req := messaging.NewPipelineChannel[int]()
		rep := messaging.NewPipelineChannel[int]()

		g := NewGateway[int, int]("test", req, rep, WithUIDGenerator(newCounterUID()))
		if g == nil {
			t.Fatal("expected non-nil component")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		req := messaging.NewPipelineChannel[int]()
		rep := messaging.NewPipelineChannel[int]()

		g := NewGateway[int, int]("api-gw", req, rep, WithUIDGenerator(newCounterUID()))
		if g.Name() != "api-gw" {
			t.Fatalf("expected name api-gw, got %q", g.Name())
		}
	})
}

func TestGateway_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("round-trip returns echoed response", func(t *testing.T) {
		t.Parallel()

		req := messaging.NewPipelineChannel[int]()
		rep := messaging.NewPipelineChannel[int]()

		wireMockSubscriber(t, req, rep, func(in int) int { return in * 2 })

		g := NewGateway[int, int]("api-gw", req, rep, WithUIDGenerator(newCounterUID()))

		err := g.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = g.Stop(context.Background()) })

		res, err := g.Request(context.Background(), 21)
		if err != nil {
			t.Fatalf("request: %v", err)
		}

		if res != 42 {
			t.Fatalf("expected res=42, got %d", res)
		}
	})

	t.Run("stamps ReplyTo and CorrelationID on outbound request", func(t *testing.T) {
		t.Parallel()

		req := messaging.NewPipelineChannel[int]()
		rep := messaging.NewPipelineChannel[int]()

		var seenReplyTo string

		var seenCorrelationID string

		_, err := req.Subscribe(func(ctx context.Context, msg messaging.Message[int]) error {
			seenReplyTo = msg.Headers.ReplyTo
			seenCorrelationID = msg.Headers.CorrelationID

			reply := messaging.Message[int]{
				Payload: msg.Payload,
				Headers: messaging.Headers{CorrelationID: msg.Headers.CorrelationID},
			}

			return rep.Send(ctx, reply)
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		g := NewGateway[int, int]("orders-gw", req, rep, WithUIDGenerator(newCounterUID()))

		err = g.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = g.Stop(context.Background()) })

		_, err = g.Request(context.Background(), 1)
		if err != nil {
			t.Fatalf("request: %v", err)
		}

		if seenReplyTo != "orders-gw" {
			t.Fatalf("expected ReplyTo=orders-gw, got %q", seenReplyTo)
		}

		if seenCorrelationID == "" {
			t.Fatal("expected non-empty CorrelationID on outbound request")
		}
	})
}

func TestGateway_Timeout(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrRequestTimeout when no reply arrives", func(t *testing.T) {
		t.Parallel()

		req := messaging.NewPipelineChannel[int]()
		rep := messaging.NewPipelineChannel[int]()

		// Black-hole subscriber that never replies.
		_, err := req.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		g := NewGateway[int, int]("test", req, rep,
			WithUIDGenerator(newCounterUID()),
			WithRequestTimeout(50*time.Millisecond))

		err = g.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = g.Stop(context.Background()) })

		_, err = g.Request(context.Background(), 1)
		if !errors.Is(err, ErrRequestTimeout) {
			t.Fatalf("expected ErrRequestTimeout, got %v", err)
		}

		if !errors.Is(err, ErrGatewayFailed) {
			t.Fatalf("expected ErrGatewayFailed, got %v", err)
		}
	})
}

func TestGateway_CtxCancellation(t *testing.T) {
	t.Parallel()

	t.Run("Request honours caller ctx cancellation", func(t *testing.T) {
		t.Parallel()

		req := messaging.NewPipelineChannel[int]()
		rep := messaging.NewPipelineChannel[int]()

		_, err := req.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		g := NewGateway[int, int]("test", req, rep,
			WithUIDGenerator(newCounterUID()),
			WithRequestTimeout(10*time.Second))

		err = g.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = g.Stop(context.Background()) })

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(20 * time.Millisecond)
			cancel()
		}()

		_, err = g.Request(ctx, 1)
		if !errors.Is(err, ErrRequestCancelled) {
			t.Fatalf("expected ErrRequestCancelled, got %v", err)
		}

		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected wrapped context.Canceled, got %v", err)
		}
	})
}

func TestGateway_ConcurrentRequests(t *testing.T) {
	t.Parallel()

	t.Run("concurrent requests with distinct correlation ids each get own reply", func(t *testing.T) {
		t.Parallel()

		// Use a TopicChannel (async) for the reply so the mock
		// subscriber can publish without re-entering the gateway's
		// handler synchronously and so dispatches survive concurrent
		// callers.
		req := messaging.NewPipelineChannel[int]()
		rep := messaging.NewBroadcastChannel[int]()

		wireMockSubscriber(t, req, rep, func(in int) int { return in * 10 })

		g := NewGateway[int, int]("test", req, rep,
			WithUIDGenerator(newCounterUID()),
			WithRequestTimeout(2*time.Second))

		err := g.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = g.Stop(context.Background()) })

		const n = 25

		results := make([]int, n)
		errs := make([]error, n)

		var wg sync.WaitGroup

		for i := range n {
			wg.Go(func() {
				res, rerr := g.Request(context.Background(), i)
				results[i] = res
				errs[i] = rerr
			})
		}

		wg.Wait()

		for i := range n {
			if errs[i] != nil {
				t.Fatalf("request %d: %v", i, errs[i])
			}

			if results[i] != i*10 {
				t.Fatalf("request %d: got %d want %d", i, results[i], i*10)
			}
		}
	})
}

func TestGateway_UnknownCorrelationID(t *testing.T) {
	t.Parallel()

	t.Run("reply with unknown correlation id is dropped and reported", func(t *testing.T) {
		t.Parallel()

		req := messaging.NewPipelineChannel[int]()
		rep := messaging.NewPipelineChannel[int]()

		errHandler, getErrs := captureErrors()

		g := NewGateway[int, int]("test", req, rep,
			WithUIDGenerator(newCounterUID()),
			WithErrorHandler(errHandler))

		err := g.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = g.Stop(context.Background()) })

		// Publish a stray reply with a CorrelationID that the gateway
		// has never seen.
		err = rep.Send(context.Background(), messaging.Message[int]{
			Payload: 99,
			Headers: messaging.Headers{CorrelationID: "ghost-id"},
		})
		if err != nil {
			t.Fatalf("send stray reply: %v", err)
		}

		errs := getErrs()
		if len(errs) != 1 {
			t.Fatalf("expected 1 captured error, got %d", len(errs))
		}

		if !errors.Is(errs[0], ErrUnknownCorrelationID) {
			t.Fatalf("expected ErrUnknownCorrelationID, got %v", errs[0])
		}
	})
}

func TestGateway_RequestErrors(t *testing.T) {
	t.Parallel()

	t.Run("Request before Start returns ErrGatewayNotStarted", func(t *testing.T) {
		t.Parallel()

		req := messaging.NewPipelineChannel[int]()
		rep := messaging.NewPipelineChannel[int]()

		g := NewGateway[int, int]("test", req, rep, WithUIDGenerator(newCounterUID()))

		_, err := g.Request(context.Background(), 1)
		if !errors.Is(err, ErrGatewayNotStarted) {
			t.Fatalf("expected ErrGatewayNotStarted, got %v", err)
		}
	})

	t.Run("Request without uid generator returns ErrCorrelationIDFailed", func(t *testing.T) {
		t.Parallel()

		req := messaging.NewPipelineChannel[int]()
		rep := messaging.NewPipelineChannel[int]()

		g := NewGateway[int, int]("test", req, rep)

		err := g.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = g.Stop(context.Background()) })

		_, err = g.Request(context.Background(), 1)
		if !errors.Is(err, ErrCorrelationIDFailed) {
			t.Fatalf("expected ErrCorrelationIDFailed, got %v", err)
		}
	})

	t.Run("Request with failing uid returns ErrCorrelationIDFailed", func(t *testing.T) {
		t.Parallel()

		req := messaging.NewPipelineChannel[int]()
		rep := messaging.NewPipelineChannel[int]()

		boom := errors.New("entropy down")

		g := NewGateway[int, int]("test", req, rep, WithUIDGenerator(newFailingUID(boom)))

		err := g.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = g.Stop(context.Background()) })

		_, err = g.Request(context.Background(), 1)
		if !errors.Is(err, ErrCorrelationIDFailed) {
			t.Fatalf("expected ErrCorrelationIDFailed, got %v", err)
		}

		if !errors.Is(err, boom) {
			t.Fatalf("expected wrapped boom, got %v", err)
		}
	})

	t.Run("Request with nil ctx returns ErrContextNil", func(t *testing.T) {
		t.Parallel()

		req := messaging.NewPipelineChannel[int]()
		rep := messaging.NewPipelineChannel[int]()

		g := NewGateway[int, int]("test", req, rep, WithUIDGenerator(newCounterUID()))

		err := g.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = g.Stop(context.Background()) })

		//nolint:staticcheck // intentionally passing nil ctx
		_, err = g.Request(nil, 1)
		if !errors.Is(err, messaging.ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("Send failure returns ErrRequestSendFailed", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("req down")

		req := &failingChannel[int]{err: boom}
		rep := messaging.NewPipelineChannel[int]()

		g := NewGateway[int, int]("test", req, rep, WithUIDGenerator(newCounterUID()))

		err := g.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		t.Cleanup(func() { _ = g.Stop(context.Background()) })

		_, err = g.Request(context.Background(), 1)
		if !errors.Is(err, ErrRequestSendFailed) {
			t.Fatalf("expected ErrRequestSendFailed, got %v", err)
		}

		if !errors.Is(err, boom) {
			t.Fatalf("expected wrapped boom, got %v", err)
		}
	})
}

func TestGateway_StopSignalsPending(t *testing.T) {
	t.Parallel()

	t.Run("Stop fails every pending Request with ErrGatewayShuttingDown", func(t *testing.T) {
		t.Parallel()

		req := messaging.NewPipelineChannel[int]()
		rep := messaging.NewPipelineChannel[int]()

		// Black-hole subscriber so requests stay pending.
		_, err := req.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
			return nil
		})
		if err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		g := NewGateway[int, int]("test", req, rep,
			WithUIDGenerator(newCounterUID()),
			WithRequestTimeout(10*time.Second))

		err = g.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		const n = 5

		errs := make([]error, n)

		var wg sync.WaitGroup

		for i := range n {
			wg.Go(func() {
				_, rerr := g.Request(context.Background(), i)
				errs[i] = rerr
			})
		}

		// Give the requesters time to register their pending entries.
		time.Sleep(50 * time.Millisecond)

		err = g.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		wg.Wait()

		for i, rerr := range errs {
			if !errors.Is(rerr, ErrGatewayShuttingDown) {
				t.Fatalf("request %d: expected ErrGatewayShuttingDown, got %v", i, rerr)
			}
		}
	})
}

func TestGateway_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Start is idempotent", func(t *testing.T) {
		t.Parallel()

		req := messaging.NewPipelineChannel[int]()
		rep := messaging.NewPipelineChannel[int]()

		wireMockSubscriber(t, req, rep, func(in int) int { return in })

		g := NewGateway[int, int]("test", req, rep, WithUIDGenerator(newCounterUID()))

		err := g.Start(context.Background())
		if err != nil {
			t.Fatalf("first start: %v", err)
		}

		err = g.Start(context.Background())
		if err != nil {
			t.Fatalf("second start: %v", err)
		}

		t.Cleanup(func() { _ = g.Stop(context.Background()) })

		res, err := g.Request(context.Background(), 7)
		if err != nil {
			t.Fatalf("request: %v", err)
		}

		if res != 7 {
			t.Fatalf("expected 7, got %d", res)
		}
	})

	t.Run("Stop is idempotent", func(t *testing.T) {
		t.Parallel()

		req := messaging.NewPipelineChannel[int]()
		rep := messaging.NewPipelineChannel[int]()

		g := NewGateway[int, int]("test", req, rep, WithUIDGenerator(newCounterUID()))

		err := g.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		err = g.Stop(context.Background())
		if err != nil {
			t.Fatalf("first stop: %v", err)
		}

		err = g.Stop(context.Background())
		if err != nil {
			t.Fatalf("second stop: %v", err)
		}
	})

	t.Run("Done closes after Stop", func(t *testing.T) {
		t.Parallel()

		req := messaging.NewPipelineChannel[int]()
		rep := messaging.NewPipelineChannel[int]()

		g := NewGateway[int, int]("test", req, rep, WithUIDGenerator(newCounterUID()))

		err := g.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		select {
		case <-g.Done():
			t.Fatal("Done closed before Stop")
		default:
		}

		err = g.Stop(context.Background())
		if err != nil {
			t.Fatalf("stop: %v", err)
		}

		select {
		case <-g.Done():
		default:
			t.Fatal("Done not closed after Stop")
		}
	})

	t.Run("Stop with expired ctx returns ErrShutdownTimeout", func(t *testing.T) {
		t.Parallel()

		req := messaging.NewPipelineChannel[int]()
		rep := messaging.NewPipelineChannel[int]()

		g := NewGateway[int, int]("test", req, rep, WithUIDGenerator(newCounterUID()))

		err := g.Start(context.Background())
		if err != nil {
			t.Fatalf("start: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = g.Stop(ctx)
		if !errors.Is(err, lifecycle.ErrShutdownTimeout) {
			t.Fatalf("expected ErrShutdownTimeout, got %v", err)
		}
	})
}

func TestGateway_Options(t *testing.T) {
	t.Parallel()

	t.Run("WithErrorHandler(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithErrorHandler(nil))
		if opts.errorHandler == nil {
			t.Fatal("expected default error handler preserved on nil arg")
		}
	})

	t.Run("WithUIDGenerator(nil) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithUIDGenerator(nil))
		if opts.uid != nil {
			t.Fatal("expected uid unchanged on nil arg")
		}
	})

	t.Run("WithRequestTimeout(0) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRequestTimeout(0))
		if opts.requestTimeout != DefaultRequestTimeout {
			t.Fatalf("expected default timeout preserved on 0, got %v", opts.requestTimeout)
		}
	})

	t.Run("WithRequestTimeout(negative) is a no-op", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRequestTimeout(-1))
		if opts.requestTimeout != DefaultRequestTimeout {
			t.Fatalf("expected default timeout preserved on negative, got %v", opts.requestTimeout)
		}
	})

	t.Run("defaults install messaging.DefaultErrorHandler and DefaultRequestTimeout", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.errorHandler == nil {
			t.Fatal("default error handler should be installed")
		}

		if opts.requestTimeout != DefaultRequestTimeout {
			t.Fatalf("expected default timeout %v, got %v", DefaultRequestTimeout, opts.requestTimeout)
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
