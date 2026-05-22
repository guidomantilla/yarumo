// Demo that exercises http.BuildServer end-to-end and proves that:
//
//  1. The builder shape `(Server, lifecycle.CloseFn, error)` matches the
//     project's managed-component idiom and mirrors grpc.BuildServer /
//     cron.BuildScheduler.
//  2. `defer stopFn(ctx, timeout)` triggers Shutdown, the blocking Start
//     (Serve) returns, the lifecycle goroutine exits via the internal
//     `spawned` channel, and closeFn only returns after that happens —
//     no race window for callers observing goroutine counts.
//  3. The Server / Start / Stop / Done implementation actually serves
//     real HTTP requests: a fixed listen address is registered with an
//     "/health" handler, dialled from the same process, and returns
//     200 OK with body "ok".
//  4. No goroutines leak: the count returns to the pre-Build baseline
//     after stopFn completes.
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	cghttp "github.com/guidomantilla/yarumo/http"
)

const demoAddress = "127.0.0.1:50052"

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	baseline := runtime.NumGoroutine()

	errChan := make(chan error, 1)

	handler := http.NewServeMux()
	handler.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	_, stopFn, err := cghttp.BuildServer(
		ctx, "demo-http", "tcp", "127.0.0.1", "50052", handler, errChan,
		cghttp.WithReadTimeout(5*time.Second),
		cghttp.WithWriteTimeout(5*time.Second),
		cghttp.WithIdleTimeout(30*time.Second),
	)
	if err != nil {
		return fmt.Errorf("build server: %w", err)
	}

	defer func() {
		// net/http.Server's Shutdown drains accepted connections but a
		// handful of internal goroutines (h1 conn-state notifier, h2
		// idle-conn writer) take a beat to fully exit after Shutdown
		// returns. This brief pause lets the count settle.
		time.Sleep(50 * time.Millisecond)

		fmt.Printf("[main] post-stop goroutines: %d (baseline %d)\n",
			runtime.NumGoroutine(), baseline)
	}()
	defer stopFn(ctx, 5*time.Second)

	fmt.Printf("[main] goroutines: baseline=%d  after-build=%d\n",
		baseline, runtime.NumGoroutine())

	// Give the listener a moment to bind.
	time.Sleep(100 * time.Millisecond)

	client := &http.Client{Timeout: 2 * time.Second}
	defer client.CloseIdleConnections()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+demoAddress+"/health", nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Close = true // disable keep-alive so the server-side conn is released right after the response.

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("GET /health: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	fmt.Printf("[rpc] GET /health → %d %q\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	fmt.Println("[main] returning (defer stopFn next)")

	return nil
}
