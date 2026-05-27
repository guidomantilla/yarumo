// Demo that exercises every public entry point of modules/managed/keep-alive:
//
//  1. NewKeepAlive               — constructs a no-op lifecycle.Component
//                                  whose Start blocks until Stop. Useful as a
//                                  "main holder" so the program does not exit
//                                  while real components run in the background.
//  2. lifecycle.Build (keep-alive) — wires the component into the unified
//                                    lifecycle pipeline. Stop is exercised via
//                                    the returned CloseFn after a short delay.
//
// The example does not block waiting for SIGINT — it drives the lifecycle
// itself so it can run as a one-shot demo.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/config"
	keepalive "github.com/guidomantilla/yarumo/managed/keep-alive"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/managed/keep-alive/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	demos := []struct {
		title string
		fn    func(context.Context) error
	}{
		{"NewKeepAlive (raw component)", demoNewKeepAlive},
		{"lifecycle.Build (managed lifecycle)", demoBuild},
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

// demoNewKeepAlive constructs the component directly and drives Start/Stop
// without lifecycle.Build. Start runs in its own goroutine because it blocks
// until Stop releases it. Done is closed when Start returns.
func demoNewKeepAlive(ctx context.Context) error {
	component := keepalive.NewKeepAlive("demo-keepalive")

	fmt.Printf("[keepalive] component name: %q\n", component.Name())

	startErr := make(chan error, 1)
	go func() {
		startErr <- component.Start(ctx)
	}()

	// Let Start enter its wait loop before signalling Stop.
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("[keepalive] Start is parked, signalling Stop\n")

	stopCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := component.Stop(stopCtx)
	if err != nil {
		return fmt.Errorf("Stop: %w", err)
	}

	err = <-startErr
	if err != nil {
		return fmt.Errorf("Start returned: %w", err)
	}

	<-component.Done()
	fmt.Printf("[keepalive] Done channel closed; Start returned cleanly\n")

	return nil
}

// demoBuild wires the component through lifecycle.Build — the canonical way
// to manage every Component in the workspace. The returned CloseFn handles
// Stop + drain timeout in one call.
func demoBuild(ctx context.Context) error {
	errChan := make(chan error, 1)

	component := keepalive.NewKeepAlive("demo-keepalive-managed")

	closeFn, err := lifecycle.Build(ctx, component, errChan)
	if err != nil {
		return fmt.Errorf("lifecycle.Build: %w", err)
	}

	fmt.Printf("[build] component built and Start dispatched\n")

	// Real apps block here until a signal or another component errors.
	time.Sleep(50 * time.Millisecond)

	closeFn(ctx, 2*time.Second)
	fmt.Printf("[build] CloseFn returned; component fully stopped\n")

	return nil
}
