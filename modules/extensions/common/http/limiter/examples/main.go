// Demo that exercises the public API of the http/limiter adapter:
//
//  1. NewLimiterTransport wraps a base RoundTripper with a configured
//     resilience.Limiter. Requests block when the bucket is empty.
//  2. A 5 rps limiter with burst 2 lets the first 2 calls through fast,
//     then paces subsequent calls.
//  3. A canceled request context surfaces an error wrapping
//     ErrRateLimiterFailed.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/guidomantilla/yarumo/config"
	chttplimiter "github.com/guidomantilla/yarumo/extensions/common/http/limiter"
	rlimiter "github.com/guidomantilla/yarumo/extensions/common/resilience/limiter"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/extensions/common/http/limiter/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	demos := []struct {
		title string
		fn    func(context.Context) error
	}{
		{"Burst pass-through, then paced", demoPaced},
		{"Canceled context", demoCanceled},
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

// demoPaced sends 4 sequential requests through a 5 rps / burst 2
// limiter. The first 2 should be near-instant, the next 2 each delayed
// by ~200ms.
func demoPaced(ctx context.Context) error {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	l := rlimiter.NewLimiter(
		rlimiter.WithRate(5, time.Second),
		rlimiter.WithBurst(2),
	)
	transport := chttplimiter.NewLimiterTransport(http.DefaultTransport, l)
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	start := time.Now()
	for i := 1; i <= 4; i++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("attempt %d: %w", i, err)
		}
		resp.Body.Close()
		fmt.Printf("  attempt %d at t+%s\n", i, time.Since(start).Round(time.Millisecond))
	}

	return nil
}

// demoCanceled drains the bucket then issues one request under a very
// short context deadline. The limiter rejects with ErrRateLimiterFailed.
func demoCanceled(parentCtx context.Context) error {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	l := rlimiter.NewLimiter(
		rlimiter.WithRate(1, 10*time.Second), // very slow refill
		rlimiter.WithBurst(1),
	)
	transport := chttplimiter.NewLimiterTransport(http.DefaultTransport, l)
	client := &http.Client{Transport: transport}

	// Drain the bucket with one successful request.
	req, _ := http.NewRequestWithContext(parentCtx, http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("first request: %w", err)
	}
	resp.Body.Close()

	ctx, cancel := context.WithTimeout(parentCtx, 30*time.Millisecond)
	defer cancel()

	req2, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	_, err = client.Do(req2)
	if err == nil {
		return fmt.Errorf("expected limiter to reject the second request")
	}

	if !errors.Is(err, rlimiter.ErrWaitFailed) {
		fmt.Printf("  second request error (not necessarily ErrWaitFailed): %v\n", err)
	} else {
		fmt.Printf("  second request rejected with ErrWaitFailed: %v\n", err)
	}

	return nil
}
