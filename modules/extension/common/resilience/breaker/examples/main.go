// Demo that exercises the public API of the resilience/breaker package:
//
//  1. Execute with a passing fn — observe State() stays Closed.
//  2. Execute with a failing fn enough times to trip the breaker —
//     observe State() flips to Open and subsequent calls fail-fast with
//     ErrBreakerOpen wrapped in ErrBreakerFailed.
//  3. Execute with a state-change hook to log the Closed -> Open
//     transition inline.
//  4. After the configured Timeout, the breaker transitions to
//     Half-Open; a successful probe closes it again.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/guidomantilla/yarumo/config"
	cbreaker "github.com/guidomantilla/yarumo/core/common/resilience/breaker"
	"github.com/guidomantilla/yarumo/extension/common/resilience/breaker"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/extension/common/resilience/breaker/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	demos := []struct {
		title string
		fn    func(context.Context) error
	}{
		{"Execute happy path (Closed)", demoPassing},
		{"Trip the breaker (Closed -> Open)", demoTrip},
		{"OnStateChange hook", demoHook},
		{"Half-Open recovery", demoRecovery},
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

// demoPassing executes a no-op fn through a default breaker. The state stays Closed.
func demoPassing(ctx context.Context) error {
	b := breaker.NewBreaker()

	err := b.Execute(ctx, func() error { return nil })
	if err != nil {
		return fmt.Errorf("Execute happy: %w", err)
	}

	fmt.Printf("  state after success: %s\n", b.State())

	return nil
}

// demoTrip drives the breaker through enough consecutive failures to
// trip it from Closed to Open. The first call after Open returns
// ErrBreakerOpen.
func demoTrip(ctx context.Context) error {
	b := breaker.NewBreaker(
		breaker.WithConsecutiveFailures(3),
		breaker.WithTimeout(200*time.Millisecond),
	)

	boom := errors.New("upstream is on fire")

	for i := 1; i <= 3; i++ {
		err := b.Execute(ctx, func() error { return boom })
		fmt.Printf("  attempt %d: state=%s err=%v\n", i, b.State(), err != nil)
	}

	fmt.Printf("  state after %d failures: %s\n", 3, b.State())

	err := b.Execute(ctx, func() error { return nil })
	if !errors.Is(err, cbreaker.ErrBreakerOpen) {
		return fmt.Errorf("expected ErrBreakerOpen, got %v", err)
	}

	fmt.Printf("  rejected call wraps ErrBreakerOpen: %v\n", errors.Is(err, cbreaker.ErrBreakerOpen))

	return nil
}

// demoHook attaches an OnStateChange hook that prints the transition.
func demoHook(ctx context.Context) error {
	b := breaker.NewBreaker(
		breaker.WithConsecutiveFailures(2),
		breaker.WithTimeout(200*time.Millisecond),
		breaker.WithOnStateChange(func(from, to cbreaker.State) {
			fmt.Printf("  hook: %s -> %s\n", from, to)
		}),
	)

	boom := errors.New("nope")

	for range 3 {
		_ = b.Execute(ctx, func() error { return boom })
	}

	return nil
}

// demoRecovery trips the breaker, waits past Timeout, then issues a
// successful probe and observes the transition back to Closed.
func demoRecovery(ctx context.Context) error {
	b := breaker.NewBreaker(
		breaker.WithConsecutiveFailures(2),
		breaker.WithTimeout(150*time.Millisecond),
	)

	boom := errors.New("bad")

	for range 2 {
		_ = b.Execute(ctx, func() error { return boom })
	}

	fmt.Printf("  state after trip: %s\n", b.State())

	time.Sleep(200 * time.Millisecond)

	err := b.Execute(ctx, func() error { return nil })
	if err != nil {
		return fmt.Errorf("probe should succeed, got %v", err)
	}

	fmt.Printf("  state after successful probe: %s\n", b.State())

	return nil
}
