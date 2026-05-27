// Package main demonstrates the four in-process messaging primitives
// end-to-end:
//
//   - PipelineChannel: synchronous, sequential fan-out (fail-fast,
//     ChainError trace).
//   - BroadcastChannel: synchronous, parallel fan-out with barrier
//     semantics (joined errors).
//   - TopicChannel: asynchronous fan-out via a worker goroutine
//     (lifecycle.Build + graceful drain).
//   - QueueChannel: asynchronous point-to-point with worker pool and
//     round-robin distribution.
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

	err = demoQueueChannel(ctx)
	if err != nil {
		return fmt.Errorf("queue channel: %w", err)
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

	err = channel.Send(ctx, messaging.Message[OrderCreated]{Payload: OrderCreated{ID: "ord-001", Amount: 19.95}})
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
	err = channel.Send(ctx, messaging.Message[OrderCreated]{Payload: OrderCreated{ID: "ord-100", Amount: 200.00}})
	elapsed := time.Since(start)

	fmt.Printf("  Send returned after %v (= max handler time, not sum)\n", elapsed.Round(time.Millisecond))

	if err != nil {
		fmt.Printf("  errors from failing handlers: %v\n", err)
	}

	fmt.Println()

	return nil
}

// demoTopicChannel shows asynchronous fan-out: Send returns
// immediately, a single worker goroutine dispatches each message to
// every subscriber, lifecycle.Build wires Start/Stop into the app
// lifecycle.
func demoTopicChannel(ctx context.Context) error {
	fmt.Println("=== TopicChannel (async fan-out + lifecycle drain) ===")

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
		err = channel.Send(ctx, messaging.Message[OrderCreated]{Payload: order})
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

	var workerA, workerB int32

	_, err = channel.Subscribe(func(_ context.Context, msg messaging.Message[OrderCreated]) error {
		n := atomic.AddInt32(&workerA, 1)
		fmt.Printf("  worker-A picks %s  (worker-A total: %d)\n", msg.Payload.ID, n)
		return nil
	})
	if err != nil {
		return err
	}

	_, err = channel.Subscribe(func(_ context.Context, msg messaging.Message[OrderCreated]) error {
		n := atomic.AddInt32(&workerB, 1)
		fmt.Printf("  worker-B picks %s  (worker-B total: %d)\n", msg.Payload.ID, n)
		return nil
	})
	if err != nil {
		return err
	}

	for i := range 6 {
		err = channel.Send(ctx, messaging.Message[OrderCreated]{Payload: OrderCreated{ID: fmt.Sprintf("ord-3%02d", i)}})
		if err != nil {
			return err
		}
	}

	closeFn(ctx, 5*time.Second)
	<-component.Done()

	a := atomic.LoadInt32(&workerA)
	b := atomic.LoadInt32(&workerB)
	fmt.Printf("  final totals: A=%d B=%d  (each msg went to exactly ONE subscriber, total=%d)\n", a, b, a+b)
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

	err = channel.Send(ctx, messaging.Message[OrderCreated]{Payload: OrderCreated{ID: "ord-400"}})
	if err != nil {
		return err
	}

	cancel()

	err = channel.Send(ctx, messaging.Message[OrderCreated]{Payload: OrderCreated{ID: "ord-401"}})
	if err != nil {
		return err
	}

	fmt.Println("  (ord-401 was sent but no handler is attached anymore)")

	cancel()
	fmt.Println("  Cancel is idempotent — second call did nothing.")

	return nil
}
