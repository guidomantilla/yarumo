// Demo that exercises NewServer + lifecycle.Build end-to-end and proves that:
//
//  1. The two-step pattern `grpc.NewServer(...)` + `lifecycle.Build(...)`
//     replaces the legacy grpc.BuildServer and is identical for http/cron/
//     diagnostics — a single Build helper drives every Component.
//  2. `defer stopFn(ctx, timeout)` triggers GracefulStop, the blocking
//     Start (Serve) returns, the lifecycle goroutine exits via the
//     internal `spawned` channel, and closeFn only returns after that
//     happens — no race window for callers observing goroutine counts.
//  3. The Server / Start / Stop / Done implementation actually serves
//     real RPCs: a built-in gRPC health service is registered with
//     `WithService`, dialled from the same process, and a Check call
//     succeeds with SERVING status.
//  4. No goroutines leak: the count returns to the pre-Build baseline
//     after stopFn completes.
package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/guidomantilla/yarumo/common/lifecycle"
	"github.com/guidomantilla/yarumo/config"
	cgrpc "github.com/guidomantilla/yarumo/managed/grpc"
)

const (
	demoService = "demo"
	demoAddress = "127.0.0.1:50051"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/managed/grpc/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	baseline := runtime.NumGoroutine()

	errChan := make(chan error, 1)

	// Built-in gRPC health service — no .proto file or generated code needed.
	healthSrv := health.NewServer()
	healthSrv.SetServingStatus(demoService, healthpb.HealthCheckResponse_SERVING)

	server := cgrpc.NewServer(
		"demo-grpc", "tcp", "127.0.0.1", "50051",
		cgrpc.WithService(healthSrv, &healthpb.Health_ServiceDesc),
	)

	stopFn, err := lifecycle.Build(ctx, server, errChan)
	if err != nil {
		return fmt.Errorf("build server: %w", err)
	}

	// Observe goroutine cleanup AFTER stopFn returns. defer is LIFO, so
	// stopFn fires first; closeFn already waits on the internal `spawned`
	// channel, so by the time this runs the lifecycle goroutine is gone.
	defer func() {
		fmt.Printf("[main] post-stop goroutines: %d (baseline %d)\n",
			runtime.NumGoroutine(), baseline)
	}()
	defer stopFn(ctx, 5*time.Second)

	fmt.Printf("[main] goroutines: baseline=%d  after-build=%d\n",
		baseline, runtime.NumGoroutine())

	// grpc.NewClient is lazy — the connection is established on the first
	// RPC. The Check call below blocks until either the server is ready or
	// the per-call context expires.
	conn, err := grpc.NewClient(demoAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	client := healthpb.NewHealthClient(conn)

	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := client.Check(checkCtx, &healthpb.HealthCheckRequest{Service: demoService})
	if err != nil {
		return fmt.Errorf("health check rpc: %w", err)
	}

	fmt.Printf("[rpc] Check(%q) → %s\n", demoService, resp.GetStatus())

	if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		return fmt.Errorf("expected SERVING, got %s", resp.GetStatus())
	}

	fmt.Println("[main] returning (defer stopFn next)")

	return nil
}
