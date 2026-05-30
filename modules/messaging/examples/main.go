// Package main demonstrates the five in-process messaging primitives
// end-to-end:
//
//   - PipelineChannel: synchronous, sequential fan-out (fail-fast,
//     ChainError trace).
//   - BroadcastChannel: synchronous, parallel fan-out with barrier
//     semantics (joined errors).
//   - TopicChannel: asynchronous fan-out with per-subscriber inbox +
//     dedicated worker goroutine (lifecycle.Build + graceful drain).
//   - QueueChannel: asynchronous point-to-point with worker pool and
//     round-robin distribution.
//   - NullChannel: /dev/null sink (drops every message, fires hook).
//
// It also showcases the cross-cutting features: per-subscriber
// isolation in TopicChannel (slow vs fast handler), OverflowPolicy
// (Reject default vs Block), WithDLQChannel (Dead Letter Channel),
// and Cancel semantics.
//
// Each demo prints a labeled section so the runtime ordering is
// obvious.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// OrderCreated is a sample domain event published throughout the demos.
type OrderCreated struct {
	ID     string
	Amount float64
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	err := demoPipelineChannel(ctx)
	if err != nil {
		return fmt.Errorf("pipeline channel: %w", err)
	}

	err = demoBroadcastChannel(ctx)
	if err != nil {
		return fmt.Errorf("broadcast channel: %w", err)
	}

	err = demoTopicChannel(ctx)
	if err != nil {
		return fmt.Errorf("topic channel: %w", err)
	}

	err = demoTopicIsolation(ctx)
	if err != nil {
		return fmt.Errorf("topic isolation: %w", err)
	}

	err = demoQueueChannel(ctx)
	if err != nil {
		return fmt.Errorf("queue channel: %w", err)
	}

	err = demoNullChannel(ctx)
	if err != nil {
		return fmt.Errorf("null channel: %w", err)
	}

	err = demoOverflowPolicy(ctx)
	if err != nil {
		return fmt.Errorf("overflow policy: %w", err)
	}

	err = demoDLQ(ctx)
	if err != nil {
		return fmt.Errorf("dlq: %w", err)
	}

	err = demoCancel(ctx)
	if err != nil {
		return fmt.Errorf("cancel: %w", err)
	}

	return nil
}

// demoPipelineChannel shows synchronous sequential dispatch: Send
// invokes every subscribed handler in the caller's goroutine, in
// registration order, fail-fast.
func demoPipelineChannel(ctx context.Context) error {
	fmt.Println("=== PipelineChannel (sync, sequential, fail-fast) ===")

	channel := messaging.NewPipelineChannel[OrderCreated]()

	_, err := channel.Subscribe(func(_ context.Context, msg messaging.Message[OrderCreated]) error {
		fmt.Printf("  [0] audit:   order %s recorded\n", msg.Payload.ID)
		return nil
	})
	if err != nil {
		return err
	}

	_, err = channel.Subscribe(func(_ context.Context, msg messaging.Message[OrderCreated]) error {
		fmt.Printf("  [1] billing: charge $%.2f for order %s\n", msg.Payload.Amount, msg.Payload.ID)
		return nil
	})
	if err != nil {
		return err
	}

	err = channel.Send(ctx, messaging.NewMessage(OrderCreated{ID: "ord-001", Amount: 19.95}, nil))
	if err != nil {
		return err
	}

	fmt.Println()

	return nil
}

// demoBroadcastChannel shows synchronous parallel dispatch: Send
// spawns one goroutine per subscriber, waits at a barrier for all
// handlers to finish, and returns the joined errors of failures (no
// fail-fast — every handler runs).
func demoBroadcastChannel(ctx context.Context) error {
	fmt.Println("=== BroadcastChannel (sync, parallel barrier) ===")

	channel := messaging.NewBroadcastChannel[OrderCreated]()

	_, err := channel.Subscribe(func(_ context.Context, msg messaging.Message[OrderCreated]) error {
		time.Sleep(40 * time.Millisecond) // slow validator
		fmt.Printf("  [0] inventory check: ok for %s\n", msg.Payload.ID)
		return nil
	})
	if err != nil {
		return err
	}

	_, err = channel.Subscribe(func(_ context.Context, msg messaging.Message[OrderCreated]) error {
		time.Sleep(10 * time.Millisecond) // fast validator
		fmt.Printf("  [1] fraud check: ok for %s\n", msg.Payload.ID)
		return nil
	})
	if err != nil {
		return err
	}

	_, err = channel.Subscribe(func(_ context.Context, _ messaging.Message[OrderCreated]) error {
		time.Sleep(20 * time.Millisecond)
		return errors.New("limit exceeded")
	})
	if err != nil {
		return err
	}

	start := time.Now()
	err = channel.Send(ctx, messaging.NewMessage(OrderCreated{ID: "ord-100", Amount: 200.00}, nil))
	elapsed := time.Since(start)

	fmt.Printf("  Send returned after %v (= max handler time, not sum)\n", elapsed.Round(time.Millisecond))

	if err != nil {
		fmt.Printf("  errors from failing handlers: %v\n", err)
	}

	fmt.Println()

	return nil
}

// demoTopicChannel shows asynchronous fan-out with per-subscriber
// inboxes: Send returns immediately, each subscriber owns its own
// inbox + worker goroutine. lifecycle.Build wires Start/Stop into the
// app lifecycle.
func demoTopicChannel(ctx context.Context) error {
	fmt.Println("=== TopicChannel (async fan-out + per-sub workers + lifecycle drain) ===")

	channel := messaging.NewTopicChannel[OrderCreated]("orders-topic",
		messaging.WithBufferSize(8),
		messaging.WithDrainTimeout(2*time.Second),
	)
	component, _ := channel.(lifecycle.Component)

	errChan := make(chan error, 1)

	closeFn, err := lifecycle.Build(ctx, component, errChan)
	if err != nil {
		return err
	}

	_, err = channel.Subscribe(func(_ context.Context, msg messaging.Message[OrderCreated]) error {
		fmt.Printf("  [analytics]    order %s captured\n", msg.Payload.ID)
		return nil
	})
	if err != nil {
		return err
	}

	_, err = channel.Subscribe(func(_ context.Context, msg messaging.Message[OrderCreated]) error {
		fmt.Printf("  [notification] email queued for order %s\n", msg.Payload.ID)
		return nil
	})
	if err != nil {
		return err
	}

	orders := []OrderCreated{
		{ID: "ord-200", Amount: 12.50},
		{ID: "ord-201", Amount: 75.00},
	}

	for _, order := range orders {
		err = channel.Send(ctx, messaging.NewMessage(order, nil))
		if err != nil {
			return err
		}

		fmt.Printf("  Send(%s) returned immediately\n", order.ID)
	}

	closeFn(ctx, 5*time.Second)
	<-component.Done()

	fmt.Println()

	return nil
}

// demoTopicIsolation showcases the killer feature of TopicChannel's
// per-subscriber worker model: a slow handler stays in its own inbox
// and does NOT block fast handlers from receiving messages at line
// rate. Without per-sub isolation, a single dispatcher would serialize
// the slow handler before reaching the fast one.
func demoTopicIsolation(ctx context.Context) error {
	fmt.Println("=== TopicChannel per-subscriber isolation (slow handler does NOT block fast) ===")

	channel := messaging.NewTopicChannel[int]("isolation",
		messaging.WithBufferSize(32),
		messaging.WithDrainTimeout(2*time.Second),
	)
	component, _ := channel.(lifecycle.Component)

	slowReleased := make(chan struct{})

	_, err := channel.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
		<-slowReleased // hold until we release
		return nil
	})
	if err != nil {
		return err
	}

	var fastDone atomic.Int32

	fastComplete := make(chan struct{})

	_, err = channel.Subscribe(func(_ context.Context, _ messaging.Message[int]) error {
		n := fastDone.Add(1)
		if n == 10 {
			close(fastComplete)
		}
		return nil
	})
	if err != nil {
		return err
	}

	closeFn, err := lifecycle.Build(ctx, component, make(chan error, 1))
	if err != nil {
		return err
	}

	for i := 1; i <= 10; i++ {
		err = channel.Send(ctx, messaging.NewMessage(i, nil))
		if err != nil {
			return err
		}
	}

	select {
	case <-fastComplete:
		fmt.Printf("  fast handler processed 10 msgs in parallel with slow handler still blocked\n")
	case <-time.After(time.Second):
		return fmt.Errorf("fast handler stuck — isolation failed")
	}

	close(slowReleased)

	closeFn(ctx, 2*time.Second)
	<-component.Done()

	fmt.Println()

	return nil
}

// demoQueueChannel shows asynchronous point-to-point distribution:
// each message goes to EXACTLY ONE subscriber via round-robin. With 3
// workers across 2 subscribers, six messages alternate between them.
func demoQueueChannel(ctx context.Context) error {
	fmt.Println("=== QueueChannel (async point-to-point round-robin) ===")

	channel := messaging.NewQueueChannel[OrderCreated]("orders-queue",
		messaging.WithBufferSize(16),
		messaging.WithWorkerCount(3),
		messaging.WithDrainTimeout(2*time.Second),
	)
	component, _ := channel.(lifecycle.Component)

	errChan := make(chan error, 1)

	closeFn, err := lifecycle.Build(ctx, component, errChan)
	if err != nil {
		return err
	}

	var workerA, workerB atomic.Int32

	_, err = channel.Subscribe(func(_ context.Context, msg messaging.Message[OrderCreated]) error {
		n := workerA.Add(1)
		fmt.Printf("  worker-A picks %s  (worker-A total: %d)\n", msg.Payload.ID, n)
		return nil
	})
	if err != nil {
		return err
	}

	_, err = channel.Subscribe(func(_ context.Context, msg messaging.Message[OrderCreated]) error {
		n := workerB.Add(1)
		fmt.Printf("  worker-B picks %s  (worker-B total: %d)\n", msg.Payload.ID, n)
		return nil
	})
	if err != nil {
		return err
	}

	for i := range 6 {
		err = channel.Send(ctx, messaging.NewMessage(OrderCreated{ID: fmt.Sprintf("ord-3%02d", i)}, nil))
		if err != nil {
			return err
		}
	}

	closeFn(ctx, 5*time.Second)
	<-component.Done()

	a := workerA.Load()
	b := workerB.Load()
	fmt.Printf("  final totals: A=%d B=%d  (each msg went to exactly ONE subscriber, total=%d)\n", a, b, a+b)
	fmt.Println()

	return nil
}

// demoNullChannel shows the /dev/null sink: Send drops every message
// and fires the ErrorHandler hook with ErrDropped. Subscribed handlers
// are accepted for interface compatibility but are never invoked.
func demoNullChannel(ctx context.Context) error {
	fmt.Println("=== NullChannel (sink — drop every message, fire hook) ===")

	var drops atomic.Int32

	channel := messaging.NewNullChannel[OrderCreated](
		messaging.WithErrorHandler(func(_ context.Context, _ any, _ error) {
			drops.Add(1)
		}),
	)

	_, err := channel.Subscribe(func(_ context.Context, _ messaging.Message[OrderCreated]) error {
		fmt.Println("  THIS SHOULD NEVER PRINT — NullChannel does not invoke handlers")
		return nil
	})
	if err != nil {
		return err
	}

	for i := range 3 {
		err = channel.Send(ctx, messaging.NewMessage(OrderCreated{ID: fmt.Sprintf("ord-5%02d", i)}, nil))
		if err != nil {
			return err
		}
	}

	fmt.Printf("  sent 3 messages, hook fired %d times, subscribed handler never invoked\n", drops.Load())
	fmt.Println()

	return nil
}

// demoOverflowPolicy contrasts two of the four OverflowPolicy
// strategies: Reject (the default, Send returns ErrBufferFull
// immediately when the inbox is full) and Block (Send waits until a
// slot opens or ctx expires). DropNewest / DropOldest behave like
// Reject but drop silently and fire the hook with ErrOverflow.
func demoOverflowPolicy(ctx context.Context) error {
	fmt.Println("=== OverflowPolicy (Reject default vs Block) ===")

	// Reject path: subscriber with bufferSize=1 and no Start.
	// The subscriber's inbox fills with the first Send; second
	// Send returns ErrBufferFull immediately.
	chReject := messaging.NewTopicChannel[int]("policy-reject",
		messaging.WithBufferSize(1),
	)

	_, err := chReject.Subscribe(func(_ context.Context, _ messaging.Message[int]) error { return nil })
	if err != nil {
		return err
	}

	_ = chReject.Send(ctx, messaging.NewMessage(1, nil))

	start := time.Now()

	err = chReject.Send(ctx, messaging.NewMessage(2, nil))
	elapsed := time.Since(start)

	fmt.Printf("  Reject (default): Send #2 returned %q after %v\n", err, elapsed.Round(time.Microsecond))

	// Block path: same setup but OverflowBlock. Send #2 blocks until
	// ctx expires.
	chBlock := messaging.NewTopicChannel[int]("policy-block",
		messaging.WithBufferSize(1),
		messaging.WithOverflowPolicy(messaging.OverflowBlock),
	)

	_, err = chBlock.Subscribe(func(_ context.Context, _ messaging.Message[int]) error { return nil })
	if err != nil {
		return err
	}

	_ = chBlock.Send(ctx, messaging.NewMessage(1, nil))

	shortCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	start = time.Now()

	err = chBlock.Send(shortCtx, messaging.NewMessage(2, nil))
	elapsed = time.Since(start)

	fmt.Printf("  Block:            Send #2 blocked %v then returned %q\n", elapsed.Round(time.Millisecond), err)
	fmt.Println()

	return nil
}

// demoDLQ shows WithDLQChannel: when a handler returns a non-nil
// error, the Topic publishes a DeadLetter[T] envelope to a configured
// DLQ channel for downstream reprocessing or audit. The ErrorHandler
// hook (observability) fires INDEPENDENTLY of the DLQ publish — they
// are complementary, not alternatives.
func demoDLQ(ctx context.Context) error {
	fmt.Println("=== WithDLQChannel (failed handler → DeadLetter[T] → DLQ Channel) ===")

	// The DLQ is a Pipeline so we can subscribe synchronously and see
	// the envelope land before the main channel exits.
	dlq := messaging.NewPipelineChannel[messaging.DeadLetter[OrderCreated]]()

	var dlqHits atomic.Int32

	_, err := dlq.Subscribe(func(_ context.Context, m messaging.Message[messaging.DeadLetter[OrderCreated]]) error {
		dlqHits.Add(1)
		fmt.Printf("  DLQ received: id=%s err=%v at=%s\n",
			m.Payload.Original.Payload.ID,
			m.Payload.LastError,
			m.Payload.FailedAt.Format("15:04:05"),
		)
		return nil
	})
	if err != nil {
		return err
	}

	channel := messaging.NewTopicChannel[OrderCreated]("orders-with-dlq",
		messaging.WithBufferSize(8),
		messaging.WithDrainTimeout(time.Second),
		messaging.WithDLQChannel(dlq),
		messaging.WithErrorHandler(func(_ context.Context, _ any, err error) {
			fmt.Printf("  observability hook fired (independent of DLQ): %v\n", err)
		}),
	)
	component, _ := channel.(lifecycle.Component)

	_, err = channel.Subscribe(func(_ context.Context, msg messaging.Message[OrderCreated]) error {
		if msg.Payload.Amount > 100 {
			return fmt.Errorf("amount %.2f exceeds limit", msg.Payload.Amount)
		}
		fmt.Printf("  processed: %s ($%.2f)\n", msg.Payload.ID, msg.Payload.Amount)
		return nil
	})
	if err != nil {
		return err
	}

	closeFn, err := lifecycle.Build(ctx, component, make(chan error, 1))
	if err != nil {
		return err
	}

	// First Send succeeds, second fails → DLQ
	_ = channel.Send(ctx, messaging.NewMessage(OrderCreated{ID: "ord-600", Amount: 50}, nil))
	_ = channel.Send(ctx, messaging.NewMessage(OrderCreated{ID: "ord-601", Amount: 500}, nil))

	closeFn(ctx, 2*time.Second)
	<-component.Done()

	fmt.Printf("  total DLQ publications: %d\n", dlqHits.Load())
	fmt.Println()

	return nil
}

// demoCancel shows that the Cancel returned by Subscribe detaches
// the handler.
func demoCancel(ctx context.Context) error {
	fmt.Println("=== Cancel subscription ===")

	channel := messaging.NewPipelineChannel[OrderCreated]()

	cancel, err := channel.Subscribe(func(_ context.Context, msg messaging.Message[OrderCreated]) error {
		fmt.Printf("  handler received order %s\n", msg.Payload.ID)
		return nil
	})
	if err != nil {
		return err
	}

	err = channel.Send(ctx, messaging.NewMessage(OrderCreated{ID: "ord-400"}, nil))
	if err != nil {
		return err
	}

	cancel()

	err = channel.Send(ctx, messaging.NewMessage(OrderCreated{ID: "ord-401"}, nil))
	if err != nil {
		return err
	}

	fmt.Println("  (ord-401 was sent but no handler is attached anymore)")

	cancel()
	fmt.Println("  Cancel is idempotent — second call did nothing.")

	return nil
}
