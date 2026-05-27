// Demo that exercises the public API of the http/retry adapter:
//
//  1. NewRetryTransport wraps a base RoundTripper with a configured
//     resilience.Retry. Transport errors are retried per policy.
//  2. WithRetryOnResponse(RetryOn5xxAnd429) + WithRetryIf(RetryIfHttpError)
//     turn 5xx responses into retry triggers.
//  3. A server that returns 503 twice then 200 demonstrates the chain
//     succeeding after retries.
//  4. Exhausting the attempt budget surfaces the rretry domain error.
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"time"

	"github.com/guidomantilla/yarumo/config"
	chttpretry "github.com/guidomantilla/yarumo/extension/common/http/retry"
	rretry "github.com/guidomantilla/yarumo/extension/common/resilience/retry"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/extension/common/http/retry/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	demos := []struct {
		title string
		fn    func(context.Context) error
	}{
		{"Eventual success after 2x 503", demoEventualSuccess},
		{"Budget exhausted on persistent 503", demoExhausted},
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

// demoEventualSuccess wires a server that flaps 503 -> 503 -> 200. The
// retry transport observes the 503s via the RetryOnResponse predicate,
// re-issues the request, and eventually returns the 200 body.
func demoEventualSuccess(ctx context.Context) error {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := attempts.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	r := rretry.NewRetry(
		rretry.WithAttempts(5),
		rretry.WithDelay(20*time.Millisecond),
		rretry.WithBackoff(rretry.BackoffFixed),
		rretry.WithRetryIf(chttpretry.RetryIfHttpError),
		rretry.WithOnRetry(func(attempt uint, err error) {
			fmt.Printf("  retry #%d triggered by: %v\n", attempt, err)
		}),
	)

	transport := chttpretry.NewRetryTransport(http.DefaultTransport, r,
		chttpretry.WithRetryOnResponse(chttpretry.RetryOn5xxAnd429),
	)
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("client.Do: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  final response: %d %q (after %d server attempts)\n", resp.StatusCode, string(body), attempts.Load())

	return nil
}

// demoExhausted wires a server that always 503s and a retrier with only
// 2 attempts. The chain fails after both attempts; the server saw
// exactly 2 hits.
func demoExhausted(ctx context.Context) error {
	var hits atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	r := rretry.NewRetry(
		rretry.WithAttempts(2),
		rretry.WithDelay(10*time.Millisecond),
		rretry.WithBackoff(rretry.BackoffFixed),
		rretry.WithRetryIf(chttpretry.RetryIfHttpError),
	)

	transport := chttpretry.NewRetryTransport(http.DefaultTransport, r,
		chttpretry.WithRetryOnResponse(chttpretry.RetryOn5xxAnd429),
	)
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	_, err := client.Do(req)
	if err == nil {
		return fmt.Errorf("expected exhaustion to fail")
	}

	fmt.Printf("  exhausted after %d server hits; err=%v\n", hits.Load(), err)

	return nil
}
