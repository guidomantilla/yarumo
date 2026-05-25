// Package main demonstrates the in-process messaging primitives end-to-
// end: DirectChannel synchronous dispatch, QueueChannel async dispatch
// with lifecycle.Build + graceful drain, the Publisher/Subscriber facade
// routed by Go type, and subscription cancel.
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

// UserSignedUp is a sample domain event published by the auth service.
type UserSignedUp struct {
	UserID string
	Email  string
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

	err := demoDirectChannel(ctx)
	if err != nil {
		return fmt.Errorf("direct channel: %w", err)
	}

	err = demoQueueChannel(ctx)
	if err != nil {
		return fmt.Errorf("queue channel: %w", err)
	}

	err = demoPubSub(ctx)
	if err != nil {
		return fmt.Errorf("pubsub: %w", err)
	}

	err = demoCancel(ctx)
	if err != nil {
		return fmt.Errorf("cancel: %w", err)
	}

	return nil
}

// demoDirectChannel shows synchronous in-goroutine dispatch: Send invokes
// every subscribed handler in the caller's goroutine, in registration
// order.
func demoDirectChannel(ctx context.Context) error {
	fmt.Println("=== DirectChannel (synchronous) ===")

	channel := messaging.NewDirectChannel[OrderCreated]()

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

// demoQueueChannel shows asynchronous dispatch via a worker goroutine,
// wired into the lifecycle with lifecycle.Build, and the graceful drain
// performed by Stop.
func demoQueueChannel(ctx context.Context) error {
	fmt.Println("=== QueueChannel (async + lifecycle.Build + drain) ===")

	queue := messaging.NewQueueChannel[OrderCreated]("orders-queue",
		messaging.WithBufferSize(8),
		messaging.WithDrainTimeout(2*time.Second),
	)

	errChan := make(chan error, 1)

	closeFn, err := messaging.BuildQueueChannel(ctx, queue, errChan)
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

// demoPubSub shows the Publisher/Subscriber facade routing payloads by
// Go type: each subscriber only receives events of its declared type.
func demoPubSub(ctx context.Context) error {
	fmt.Println("=== Publisher/Subscriber (routed by Go type) ===")

	bus := messaging.NewPubSub()

	_, err := messaging.Subscribe[OrderCreated](bus, func(_ context.Context, msg messaging.Message[OrderCreated]) error {
		fmt.Printf("  order handler: ID=%s amount=$%.2f\n", msg.Payload.ID, msg.Payload.Amount)
		return nil
	})
	if err != nil {
		return err
	}

	_, err = messaging.Subscribe[UserSignedUp](bus, func(_ context.Context, msg messaging.Message[UserSignedUp]) error {
		fmt.Printf("  user handler:  ID=%s email=%s\n", msg.Payload.UserID, msg.Payload.Email)
		return nil
	})
	if err != nil {
		return err
	}

	err = messaging.Publish[OrderCreated](ctx, bus, OrderCreated{ID: "ord-200", Amount: 49.99})
	if err != nil {
		return err
	}

	err = messaging.Publish[UserSignedUp](ctx, bus, UserSignedUp{UserID: "usr-77", Email: "ana@example.com"})
	if err != nil {
		return err
	}

	fmt.Println("  (each handler only fired for its own type)")
	fmt.Println()

	return nil
}

// demoCancel shows that the Cancel returned by Subscribe detaches the
// handler: subsequent Sends do not reach it.
func demoCancel(ctx context.Context) error {
	fmt.Println("=== Cancel subscription ===")

	channel := messaging.NewDirectChannel[OrderCreated]()

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
