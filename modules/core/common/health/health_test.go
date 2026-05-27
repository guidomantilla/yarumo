package health

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewHealth(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil Health", func(t *testing.T) {
		t.Parallel()

		h := NewHealth()
		if h == nil {
			t.Fatalf("NewHealth() returned nil")
		}
	})

	t.Run("respects WithConcurrency option", func(t *testing.T) {
		t.Parallel()

		h := NewHealth(WithConcurrency(3))

		hh, ok := h.(*health)
		if !ok {
			t.Fatalf("NewHealth returned %T, want *health", h)
		}

		if hh.concurrency != 3 {
			t.Fatalf("concurrency = %d, want 3", hh.concurrency)
		}
	})
}

func TestHealth_Register(t *testing.T) {
	t.Parallel()

	t.Run("appends checks in registration order", func(t *testing.T) {
		t.Parallel()

		h := NewHealth().(*health)

		c1 := &stubCheck{name: "c1", result: Result{Status: StatusHealthy}}
		c2 := &stubCheck{name: "c2", result: Result{Status: StatusHealthy}}

		h.Register(c1)
		h.Register(c2)

		if len(h.checks) != 2 {
			t.Fatalf("len(checks) = %d, want 2", len(h.checks))
		}

		if h.checks[0].Name() != "c1" || h.checks[1].Name() != "c2" {
			t.Fatalf("registration order broken: %s, %s", h.checks[0].Name(), h.checks[1].Name())
		}
	})

	t.Run("nil check is silently ignored", func(t *testing.T) {
		t.Parallel()

		h := NewHealth().(*health)

		h.Register(nil)

		if len(h.checks) != 0 {
			t.Fatalf("len(checks) = %d, want 0 (nil must be ignored)", len(h.checks))
		}
	})

	t.Run("concurrent registration is race-free", func(t *testing.T) {
		t.Parallel()

		h := NewHealth()

		var wg sync.WaitGroup
		const n = 50

		for range n {
			wg.Go(func() {
				h.Register(&stubCheck{name: "c", result: Result{Status: StatusHealthy}})
			})
		}

		wg.Wait()

		hh := h.(*health)
		if len(hh.checks) != n {
			t.Fatalf("len(checks) = %d, want %d", len(hh.checks), n)
		}
	})
}

func TestHealth_Status(t *testing.T) {
	t.Parallel()

	t.Run("empty registry returns StatusUnknown and nil slice", func(t *testing.T) {
		t.Parallel()

		h := NewHealth()

		status, results := h.Status(context.Background())
		if status != StatusUnknown {
			t.Fatalf("status = %v, want StatusUnknown", status)
		}

		if results != nil {
			t.Fatalf("results = %v, want nil", results)
		}
	})

	t.Run("nil context returns StatusUnknown and nil slice", func(t *testing.T) {
		t.Parallel()

		h := NewHealth()
		h.Register(&stubCheck{name: "x", result: Result{Status: StatusHealthy}})

		var ctx context.Context

		status, results := h.Status(ctx)
		if status != StatusUnknown {
			t.Fatalf("status = %v, want StatusUnknown", status)
		}

		if results != nil {
			t.Fatalf("results = %v, want nil", results)
		}
	})

	t.Run("worst-status wins — three mixed-status checks", func(t *testing.T) {
		t.Parallel()

		h := NewHealth()
		h.Register(&stubCheck{name: "healthy", result: Result{Status: StatusHealthy}})
		h.Register(&stubCheck{name: "degraded", result: Result{Status: StatusDegraded}})
		h.Register(&stubCheck{name: "unhealthy", result: Result{Status: StatusUnhealthy}})

		status, results := h.Status(context.Background())
		if status != StatusUnhealthy {
			t.Fatalf("status = %v, want StatusUnhealthy (worst wins)", status)
		}

		if len(results) != 3 {
			t.Fatalf("len(results) = %d, want 3", len(results))
		}

		// Results must be in registration order.
		wantNames := []string{"healthy", "degraded", "unhealthy"}
		for i, r := range results {
			if r.Name != wantNames[i] {
				t.Fatalf("results[%d].Name = %q, want %q (registration order broken)", i, r.Name, wantNames[i])
			}
		}
	})

	t.Run("two healthy + one degraded — aggregate is degraded", func(t *testing.T) {
		t.Parallel()

		h := NewHealth()
		h.Register(&stubCheck{name: "a", result: Result{Status: StatusHealthy}})
		h.Register(&stubCheck{name: "b", result: Result{Status: StatusDegraded}})
		h.Register(&stubCheck{name: "c", result: Result{Status: StatusHealthy}})

		status, _ := h.Status(context.Background())
		if status != StatusDegraded {
			t.Fatalf("status = %v, want StatusDegraded", status)
		}
	})

	t.Run("all healthy — aggregate is healthy", func(t *testing.T) {
		t.Parallel()

		h := NewHealth()
		h.Register(&stubCheck{name: "a", result: Result{Status: StatusHealthy}})
		h.Register(&stubCheck{name: "b", result: Result{Status: StatusHealthy}})

		status, _ := h.Status(context.Background())
		if status != StatusHealthy {
			t.Fatalf("status = %v, want StatusHealthy", status)
		}
	})

	t.Run("bounded concurrency — 100 checks, max parallelism <= limit", func(t *testing.T) {
		t.Parallel()

		const (
			total = 100
			limit = 4
		)

		h := NewHealth(WithConcurrency(limit))

		var inFlight atomic.Int64
		var peak atomic.Int64

		checks := make([]*concurrencyProbe, total)
		for i := range total {
			checks[i] = &concurrencyProbe{
				name:     "c",
				inFlight: &inFlight,
				peak:     &peak,
				delay:    2 * time.Millisecond,
				status:   StatusHealthy,
			}
			h.Register(checks[i])
		}

		status, results := h.Status(context.Background())
		if status != StatusHealthy {
			t.Fatalf("status = %v, want StatusHealthy", status)
		}

		if len(results) != total {
			t.Fatalf("len(results) = %d, want %d", len(results), total)
		}

		observedPeak := peak.Load()
		if observedPeak > int64(limit) {
			t.Fatalf("peak in-flight = %d, exceeded limit %d", observedPeak, limit)
		}

		if observedPeak < 2 {
			t.Fatalf("peak in-flight = %d, expected concurrent execution (>= 2)", observedPeak)
		}
	})

	t.Run("limit larger than #checks is clamped — does not deadlock", func(t *testing.T) {
		t.Parallel()

		h := NewHealth(WithConcurrency(1024))

		h.Register(&stubCheck{name: "a", result: Result{Status: StatusHealthy}})
		h.Register(&stubCheck{name: "b", result: Result{Status: StatusHealthy}})

		done := make(chan struct{})

		go func() {
			h.Status(context.Background())
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatalf("Status with oversized limit deadlocked")
		}
	})

	t.Run("pre-cancelled context — no probe invoked", func(t *testing.T) {
		t.Parallel()

		h := NewHealth()

		c1 := &stubCheck{name: "a", result: Result{Status: StatusHealthy}, delay: 50 * time.Millisecond}
		c2 := &stubCheck{name: "b", result: Result{Status: StatusHealthy}, delay: 50 * time.Millisecond}

		h.Register(c1)
		h.Register(c2)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		status, _ := h.Status(ctx)
		if status != StatusUnknown {
			t.Fatalf("status = %v, want StatusUnknown (pre-cancelled ctx)", status)
		}

		if c1.calls.Load() != 0 || c2.calls.Load() != 0 {
			t.Fatalf("probe was invoked despite pre-cancelled ctx: c1=%d c2=%d", c1.calls.Load(), c2.calls.Load())
		}
	})

	t.Run("context cancelled mid-aggregation — clean exit", func(t *testing.T) {
		t.Parallel()

		h := NewHealth(WithConcurrency(2))

		// Fill with slow probes so cancellation actually catches some of them.
		for range 10 {
			h.Register(&stubCheck{
				name:   "slow",
				delay:  100 * time.Millisecond,
				result: Result{Status: StatusHealthy},
			})
		}

		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan struct{})

		var (
			gotStatus  Status
			gotResults []Result
		)

		go func() {
			gotStatus, gotResults = h.Status(ctx)
			close(done)
		}()

		// Cancel almost immediately so the aggregator stops scheduling new probes.
		time.Sleep(5 * time.Millisecond)
		cancel()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatalf("Status did not return promptly after cancellation")
		}

		// Behavioural contract: Status returns; we do not strictly assert which
		// status because already-running probes may have completed before
		// cancellation. The length of gotResults must equal the number of
		// registered checks (results slice is pre-allocated with len = N).
		if len(gotResults) != 10 {
			t.Fatalf("len(gotResults) = %d, want 10", len(gotResults))
		}

		_ = gotStatus
	})
}
