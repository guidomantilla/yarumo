// Demo that exercises the public API of the resilience/retry package:
//
//  1. Do with an fn that succeeds on the third attempt: WithAttempts(5)
//     budget + OnRetry hook shows the retry trace.
//  2. Do with a permanent error short-circuits via WithRetryIf.
//  3. Do with a fixed backoff produces evenly-spaced delays.
//  4. Do with an fn that always fails returns ErrRetryFailed.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/guidomantilla/yarumo/config"
	"github.com/guidomantilla/yarumo/extensions/common/resilience/retry"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/extensions/common/resilience/retry/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	demos := []struct {
		title string
		fn    func(context.Context) error
	}{
		{"Succeed on 3rd attempt", demoEventualSuccess},
		{"Short-circuit on permanent error", demoShortCircuit},
		{"BackoffFixed delays", demoFixedBackoff},
		{"Budget exhausted", demoExhausted},
	}

	for _, d := range demos {
		fmt.Printf("=== Demo: %s ===\n", d.title)
		err := d.fn(ctx)
		if err != nil {
			return fmt.Errorf("%s: %w", d.title, err)
		}
		fmt.Println()
	}

	return nil
}

// demoEventualSuccess returns nil only on the third call. The retrier
// is configured for up to 5 attempts; the OnRetry hook prints each
// retry trigger.
func demoEventualSuccess(ctx context.Context) error {
	r := retry.NewRetry(
		retry.WithAttempts(5),
		retry.WithDelay(20*time.Millisecond),
		retry.WithBackoff(retry.BackoffFixed),
		retry.WithOnRetry(func(attempt uint, err error) {
			fmt.Printf("  retry attempt #%d: err=%v\n", attempt, err)
		}),
	)

	calls := 0
	err := r.Do(ctx, func() error {
		calls++
		if calls < 3 {
			return fmt.Errorf("transient %d", calls)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Do: %w", err)
	}

	fmt.Printf("  succeeded after %d total calls\n", calls)

	return nil
}

// demoShortCircuit returns a permanent error and configures
// WithRetryIf to refuse that error. The retrier gives up immediately.
func demoShortCircuit(ctx context.Context) error {
	permanent := errors.New("permanent: bad input")

	r := retry.NewRetry(
		retry.WithAttempts(5),
		retry.WithDelay(10*time.Millisecond),
		retry.WithRetryIf(func(err error) bool {
			return !errors.Is(err, permanent)
		}),
	)

	calls := 0
	err := r.Do(ctx, func() error {
		calls++
		return permanent
	})
	if err == nil {
		return fmt.Errorf("expected Do to fail")
	}

	fmt.Printf("  permanent error short-circuited after %d call(s); err=%v\n", calls, err)

	return nil
}

// demoFixedBackoff configures a fixed 50ms delay and prints the gap
// between attempts.
func demoFixedBackoff(ctx context.Context) error {
	r := retry.NewRetry(
		retry.WithAttempts(3),
		retry.WithDelay(50*time.Millisecond),
		retry.WithBackoff(retry.BackoffFixed),
	)

	var last time.Time
	err := r.Do(ctx, func() error {
		now := time.Now()
		if !last.IsZero() {
			fmt.Printf("  gap since previous attempt: %s\n", now.Sub(last).Round(time.Millisecond))
		}
		last = now
		return errors.New("always fail")
	})
	if err == nil {
		return fmt.Errorf("expected Do to fail")
	}

	return nil
}

// demoExhausted always returns an error so the retrier exhausts its
// attempt budget and returns ErrRetryFailed.
func demoExhausted(ctx context.Context) error {
	r := retry.NewRetry(
		retry.WithAttempts(3),
		retry.WithDelay(10*time.Millisecond),
	)

	err := r.Do(ctx, func() error {
		return errors.New("nope")
	})
	if err == nil {
		return fmt.Errorf("expected retry to give up")
	}

	if !errors.Is(err, retry.ErrRetryFailed) {
		return fmt.Errorf("expected ErrRetryFailed, got %v", err)
	}

	fmt.Printf("  budget exhausted: %v\n", err)

	return nil
}
