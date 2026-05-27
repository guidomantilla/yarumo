// Package tests publishes test helpers for verifying that
// lifecycle.Component implementations satisfy the contract documented
// on the Component interface.
//
// It is intended for consumption from _test.go files only. Importing
// it from production code links the testing package into the final
// binary and is not supported.
package tests

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

// goroutineSettleTimeout bounds how long the helper waits for spawned
// goroutines to wind down after the second Stop before declaring a leak.
const goroutineSettleTimeout = time.Second

// stopReturnTimeout bounds how long the second Stop is allowed to take.
// It is intentionally generous: implementations that delegate to network
// shutdowns may take a few milliseconds even on the "already closed"
// fast path.
const stopReturnTimeout = time.Second

// AssertIdempotentStop verifies that c satisfies the idempotent-Stop
// contract documented on lifecycle.Component:
//
//   - Calling Stop twice does not panic.
//   - After the first Stop, Done() is closed.
//   - The second Stop returns within stopReturnTimeout.
//   - No goroutine survives across the two calls beyond a small
//     tolerance for runtime scheduling jitter.
//
// The helper does not assert on the returned error value. Some
// implementations surface a wrapped ErrShutdownTimeout or a driver-
// level "already closed" error on the second call; both are
// permissible per invariant 1. Tests that need to pin the error
// behavior of a specific implementation should do so in a dedicated
// test, not via this helper.
//
// The caller is responsible for any Start orchestration before
// invoking the helper; the helper exercises only Stop and Done. For
// server-style components, a typical pattern is to spawn Start in a
// goroutine, wait for readiness, then call this helper.
func AssertIdempotentStop(t *testing.T, c lifecycle.Component) {
	t.Helper()

	if c == nil {
		t.Fatal("AssertIdempotentStop: component is nil")
	}

	runtime.GC()
	baseline := runtime.NumGoroutine()

	firstCtx, firstCancel := context.WithTimeout(context.Background(), stopReturnTimeout)
	defer firstCancel()

	_ = c.Stop(firstCtx)

	select {
	case <-c.Done():
	case <-time.After(stopReturnTimeout):
		t.Fatal("AssertIdempotentStop: Done not closed after first Stop")
	}

	secondDone := make(chan struct{})

	go func() {
		defer close(secondDone)
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("AssertIdempotentStop: second Stop panicked: %v", r)
			}
		}()

		secondCtx, secondCancel := context.WithTimeout(context.Background(), stopReturnTimeout)
		defer secondCancel()

		_ = c.Stop(secondCtx)
	}()

	select {
	case <-secondDone:
	case <-time.After(stopReturnTimeout * 2):
		t.Fatalf("AssertIdempotentStop: second Stop did not return within %s", stopReturnTimeout*2)
	}

	deadline := time.Now().Add(goroutineSettleTimeout)
	for time.Now().Before(deadline) {
		runtime.GC()
		if runtime.NumGoroutine() <= baseline+1 {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("AssertIdempotentStop: goroutine leak: baseline=%d after=%d", baseline, runtime.NumGoroutine())
}
