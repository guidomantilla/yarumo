// Demo that exercises the public API of the resilience/limiter package:
//
//  1. NewLimiter with a tight rate; Allow returns true while tokens are
//     available and false once the burst is depleted.
//  2. Wait blocks until a token replenishes.
//  3. Wait with a canceled context returns an error wrapping ErrWaitFailed.
//  4. Default limiter (no options) yields ~10 rps with burst 10.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/guidomantilla/yarumo/config"
	"github.com/guidomantilla/yarumo/extension/common/resilience/limiter"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/extension/common/resilience/limiter/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	demos := []struct {
		title string
		fn    func(context.Context) error
	}{
		{"Allow drains the burst, then refuses", demoAllow},
		{"Wait blocks for next token", demoWait},
		{"Wait honours canceled context", demoWaitCanceled},
		{"Default limiter (10 rps, burst 10)", demoDefault},
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

// demoAllow constructs a tight limiter (1 rps, burst 2), then issues
// Allow calls until the burst is exhausted.
func demoAllow(_ context.Context) error {
	l := limiter.NewLimiter(
		limiter.WithRate(1, time.Second),
		limiter.WithBurst(2),
	)

	for i := 1; i <= 4; i++ {
		ok := l.Allow()
		fmt.Printf("  attempt %d: allow=%v\n", i, ok)
	}

	return nil
}

// demoWait constructs a 10 rps limiter, drains the burst, then calls
// Wait — expecting a short blocking pause until the next token is
// available.
func demoWait(ctx context.Context) error {
	l := limiter.NewLimiter(
		limiter.WithRate(10, time.Second),
		limiter.WithBurst(1),
	)

	if !l.Allow() {
		return fmt.Errorf("first Allow should succeed")
	}

	start := time.Now()

	err := l.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Wait: %w", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("  Wait returned after %s\n", elapsed.Round(time.Millisecond))

	return nil
}

// demoWaitCanceled shows that Wait returns an error wrapping
// ErrWaitFailed when the context expires before a token is granted.
func demoWaitCanceled(parentCtx context.Context) error {
	l := limiter.NewLimiter(
		limiter.WithRate(1, 10*time.Second), // 0.1 rps — tokens come slowly
		limiter.WithBurst(1),
	)

	// Drain the burst.
	_ = l.Allow()

	ctx, cancel := context.WithTimeout(parentCtx, 30*time.Millisecond)
	defer cancel()

	err := l.Wait(ctx)
	if err == nil {
		return fmt.Errorf("expected Wait to fail under deadline")
	}

	if !errors.Is(err, limiter.ErrWaitFailed) {
		return fmt.Errorf("expected ErrWaitFailed, got %v", err)
	}

	fmt.Printf("  Wait under canceled ctx: %v\n", err)

	return nil
}

// demoDefault shows the default limiter configuration: 10 rps, burst
// 10. Ten consecutive Allow calls succeed; the eleventh refuses.
func demoDefault(_ context.Context) error {
	l := limiter.NewLimiter()

	allowed := 0
	for range 11 {
		if l.Allow() {
			allowed++
		}
	}

	fmt.Printf("  11 attempts -> %d allowed (expected 10 from burst)\n", allowed)

	return nil
}
