// Demo that exercises the public API of the http/breaker adapter:
//
//  1. NewBreakerTransport wraps a base RoundTripper with a configured
//     resilience.Breaker. Happy-path requests pass through unchanged.
//  2. WithFailOnResponse(FailOn5xxAnd429) makes the breaker observe 5xx
//     responses as failures.
//  3. Enough 5xx responses in a row trip the breaker; subsequent
//     requests fail-fast with ErrBreakerRejectedFailed wrapping
//     ErrBreakerOpen.
//  4. The httptest.Server toggles between 500 (trip mode) and 200
//     (recovery) under flag control.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"time"

	"github.com/guidomantilla/yarumo/config"
	chttpbreaker "github.com/guidomantilla/yarumo/extensions/common/http/breaker"
	rbreaker "github.com/guidomantilla/yarumo/extensions/common/resilience/breaker"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/extensions/common/http/breaker/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	demos := []struct {
		title string
		fn    func(context.Context) error
	}{
		{"Happy path through breaker", demoHappyPath},
		{"5xx trips the breaker", demoTrip},
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

// demoHappyPath builds a breaker + transport, fires a request against
// an httptest server that responds 200, and prints the result.
func demoHappyPath(ctx context.Context) error {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	b := rbreaker.NewBreaker()
	transport := chttpbreaker.NewBreakerTransport(
		http.DefaultTransport, b,
		chttpbreaker.WithFailOnResponse(chttpbreaker.FailOn5xxAnd429),
	)
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("client.Do: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  GET %s -> %d %q\n", server.URL, resp.StatusCode, string(body))
	fmt.Printf("  breaker state: %s\n", b.State())

	return nil
}

// demoTrip uses a server that always returns 500 and a breaker tuned to
// trip after 2 consecutive failures. The third request is rejected
// before reaching the server (counter stays at 2).
func demoTrip(ctx context.Context) error {
	var serverHits atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		serverHits.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	b := rbreaker.NewBreaker(
		rbreaker.WithConsecutiveFailures(2),
		rbreaker.WithTimeout(500*time.Millisecond),
	)
	transport := chttpbreaker.NewBreakerTransport(
		http.DefaultTransport, b,
		chttpbreaker.WithFailOnResponse(chttpbreaker.FailOn5xxAnd429),
	)
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	for i := 1; i <= 3; i++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
		resp, err := client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}
		fmt.Printf("  attempt %d: state=%s err=%v\n", i, b.State(), err != nil)
		if i == 3 {
			var statusErr *chttpbreaker.StatusCodeError
			if errors.As(err, &statusErr) {
				return fmt.Errorf("attempt 3 should be fast-rejected, got StatusCodeError")
			}
		}
	}

	fmt.Printf("  server received %d requests total (expected 2)\n", serverHits.Load())
	fmt.Printf("  final breaker state: %s\n", b.State())

	return nil
}
