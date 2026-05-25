// Package main demonstrates the in-process messaging primitives end-to-
// end: PipelineChannel synchronous dispatch, TopicChannel async dispatch
// with lifecycle.Build + graceful drain, and subscription cancel.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/guidomantilla/yarumo/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// OrderCreated is a sample domain event published by the order service.
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
		return fmt.Errorf("direct channel: %w", err)
	}

	err = demoTopicChannel(ctx)
	if err != nil {
		return fmt.Errorf("queue channel: %w", err)
	}

	err = demoCancel(ctx)
	if err != nil {
		return fmt.Errorf("cancel: %w", err)
	}

	return nil
}

// demoPipelineChannel shows synchronous in-goroutine dispatch: Send invokes
// every subscribed handler in the caller's goroutine, in registration
// order.
func demoPipelineChannel(ctx context.Context) error {
	fmt.Println("=== PipelineChannel (synchronous) ===")

	channel := messaging.NewPipelineChannel[OrderCreated]()

	_, err := channel.Subscribe(func(_ context.Context, msg messaging.Message[OrderCreated]) error {
		fmt.Printf("  audit: order %s recorded\n", msg.Payload.ID)
		return nil
	})
	if err != nil {
		return err
	}

	_, err = channel.Subscribe(func(_ context.Context, msg messaging.Message[OrderCreated]) error {
		fmt.Printf("  billing: charge $%.2f for order %s\n", msg.Payload.Amount, msg.Payload.ID)
		return nil
	})
	if err != nil {
		return err
	}

	orders := []OrderCreated{
		{ID: "ord-001", Amount: 19.95},
		{ID: "ord-002", Amount: 249.00},
	}

	for _, order := range orders {
		err = channel.Send(ctx, messaging.Message[OrderCreated]{Payload: order})
		if err != nil {
			return err
		}
	}

	fmt.Println()

	return nil
}

// demoTopicChannel shows asynchronous dispatch via a worker goroutine,
// wired into the lifecycle with lifecycle.Build, and the graceful drain
// performed by Stop.
func demoTopicChannel(ctx context.Context) error {
	fmt.Println("=== TopicChannel (async + lifecycle.Build + drain) ===")

	queue := messaging.NewTopicChannel[OrderCreated]("orders-queue",
		messaging.WithBufferSize(8),
		messaging.WithDrainTimeout(2*time.Second),
	)

	errChan := make(chan error, 1)

	closeFn, err := lifecycle.Build(ctx, queue, errChan)
	if err != nil {
		return err
	}

	_, err = queue.Subscribe(func(_ context.Context, msg messaging.Message[OrderCreated]) error {
		// Simulate non-trivial work to make async visible.
		time.Sleep(50 * time.Millisecond)
		fmt.Printf("  fulfillment: shipped order %s ($%.2f)\n", msg.Payload.ID, msg.Payload.Amount)
		return nil
	})
	if err != nil {
		return err
	}

	orders := []OrderCreated{
		{ID: "ord-100", Amount: 12.50},
		{ID: "ord-101", Amount: 75.00},
		{ID: "ord-102", Amount: 320.00},
	}

	for _, order := range orders {
		err = queue.Send(ctx, messaging.Message[OrderCreated]{Payload: order})
		if err != nil {
			return err
		}

		fmt.Printf("  enqueued order %s (Send returned immediately)\n", order.ID)
	}

	fmt.Println("  closing channel — drain pending messages...")
	closeFn(ctx, 5*time.Second)

	// Wait for the worker to fully exit so its prints appear in order.
	<-queue.Done()

	fmt.Println("  drain complete; worker exited.")
	fmt.Println()

	return nil
}

// demoCancel shows that the Cancel returned by Subscribe detaches the
// handler: subsequent Sends do not reach it.
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

	err = channel.Send(ctx, messaging.Message[OrderCreated]{Payload: OrderCreated{ID: "ord-300"}})
	if err != nil {
		return err
	}

	fmt.Println("  calling Cancel()...")
	cancel()

	err = channel.Send(ctx, messaging.Message[OrderCreated]{Payload: OrderCreated{ID: "ord-301"}})
	if err != nil {
		return err
	}

	fmt.Println("  (ord-301 was sent but no handler is attached anymore)")

	// Idempotent cancel: calling it again is a no-op.
	cancel()
	fmt.Println("  Cancel is idempotent — second call did nothing.")

	// Quiet linter: assert the CloseFn type at compile time without using
	// it (the queue demo already exercised it at runtime).
	var _ lifecycle.CloseFn = func(context.Context, time.Duration) {}

	return nil
}
