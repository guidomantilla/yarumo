package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	cresilience "github.com/guidomantilla/yarumo/extensions/common/resilience"
)

func main() {
	circuitBreakerExample()
	rateLimiterExample()
}

// circuitBreakerExample demonstrates the lifecycle of a circuit breaker:
// successful Execute, consecutive failures tripping the breaker, fail-fast
// rejection while open, and recovery via half-open probe.
func circuitBreakerExample() {
	fmt.Println("=== Circuit Breaker ===")

	registry := cresilience.NewCircuitBreakerRegistry()

	// Configure a low failure threshold and short open-state timeout so the
	// transitions are visible at runtime.
	err := registry.Use("payments",
		cresilience.WithCircuitBreakerConsecutiveFailures(3),
		cresilience.WithCircuitBreakerTimeout(200*time.Millisecond),
		cresilience.WithCircuitBreakerMaxRequests(1))
	if err != nil {
		fmt.Printf("Use failed: %v\n", err)

		return
	}

	breaker := registry.Get("payments")

	fmt.Printf("initial state: %s\n", breaker.State())

	// One successful call.
	result, err := breaker.Execute(context.Background(), func() (any, error) {
		return "ok", nil
	})
	fmt.Printf("call 1: result=%v err=%v state=%s\n", result, err, breaker.State())

	// Trip the breaker with three consecutive failures.
	for i := range 3 {
		_, err := breaker.Execute(context.Background(), func() (any, error) {
			return nil, errors.New("upstream-failure")
		})

		fmt.Printf("call %d (failure): err=%v state=%s\n", i+2, err, breaker.State())
	}

	// Subsequent calls are rejected fast while the breaker is open.
	_, err = breaker.Execute(context.Background(), func() (any, error) {
		return "never reached", nil
	})

	fmt.Printf("call 5 (open): err=%v state=%s\n", err, breaker.State())

	// Wait for the open-state timeout — the next call probes (half-open).
	time.Sleep(250 * time.Millisecond)

	result, err = breaker.Execute(context.Background(), func() (any, error) {
		return "recovered", nil
	})

	fmt.Printf("call 6 (probe): result=%v err=%v state=%s\n", result, err, breaker.State())
	fmt.Println()
}

// rateLimiterExample demonstrates non-blocking Allow and blocking Wait on a
// token bucket configured with a fast refill rate.
func rateLimiterExample() {
	fmt.Println("=== Rate Limiter ===")

	registry := cresilience.NewRateLimiterRegistry()

	// 10 tokens per second (one every 100 ms) with a burst of 3.
	err := registry.Use("api",
		cresilience.WithRateLimiterInterval(100*time.Millisecond),
		cresilience.WithRateLimiterBurst(3))
	if err != nil {
		fmt.Printf("Use failed: %v\n", err)

		return
	}

	limiter := registry.Get("api")

	// Burst: first three Allow calls succeed immediately.
	for i := range 5 {
		ok := limiter.Allow()

		fmt.Printf("Allow #%d: %v\n", i+1, ok)
	}

	// Wait blocks until the next token is produced.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	start := time.Now()

	err = limiter.Wait(ctx)
	fmt.Printf("Wait: err=%v elapsed=%s\n", err, time.Since(start).Round(10*time.Millisecond))
	fmt.Println()
}
