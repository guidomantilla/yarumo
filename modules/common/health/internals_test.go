package health

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// stubCheck is a minimal Check used by tests. It returns a fixed Result and
// optionally counts invocations.
type stubCheck struct {
	name   string
	result Result
	delay  time.Duration
	calls  atomic.Int64
}

func (s *stubCheck) Name() string { return s.name }

func (s *stubCheck) Probe(ctx context.Context) Result {
	s.calls.Add(1)

	if s.delay > 0 {
		select {
		case <-ctx.Done():
			return Result{Status: StatusUnhealthy, Message: "ctx done"}
		case <-time.After(s.delay):
		}
	}

	return s.result
}

// concurrencyProbe records the peak number of concurrent in-flight invocations
// across all instances sharing the same inFlight / peak counters.
type concurrencyProbe struct {
	name     string
	inFlight *atomic.Int64
	peak     *atomic.Int64
	delay    time.Duration
	status   Status
}

func (c *concurrencyProbe) Name() string { return c.name }

func (c *concurrencyProbe) Probe(ctx context.Context) Result {
	cur := c.inFlight.Add(1)
	defer c.inFlight.Add(-1)

	for {
		peak := c.peak.Load()
		if cur <= peak {
			break
		}

		swapped := c.peak.CompareAndSwap(peak, cur)
		if swapped {
			break
		}
	}

	if c.delay > 0 {
		select {
		case <-ctx.Done():
			return Result{Status: StatusUnhealthy, Message: "ctx done"}
		case <-time.After(c.delay):
		}
	}

	return Result{Status: c.status}
}

func TestAggregate(t *testing.T) {
	t.Parallel()

	t.Run("empty results aggregate to unknown", func(t *testing.T) {
		t.Parallel()

		got := aggregate(nil)
		if got != StatusUnknown {
			t.Fatalf("aggregate(nil) = %v, want StatusUnknown", got)
		}

		got = aggregate([]Result{})
		if got != StatusUnknown {
			t.Fatalf("aggregate([]) = %v, want StatusUnknown", got)
		}
	})

	t.Run("all healthy aggregate to healthy", func(t *testing.T) {
		t.Parallel()

		got := aggregate([]Result{
			{Status: StatusHealthy},
			{Status: StatusHealthy},
			{Status: StatusHealthy},
		})

		if got != StatusHealthy {
			t.Fatalf("aggregate(all healthy) = %v, want StatusHealthy", got)
		}
	})

	t.Run("worst status wins — mixed", func(t *testing.T) {
		t.Parallel()

		got := aggregate([]Result{
			{Status: StatusHealthy},
			{Status: StatusDegraded},
			{Status: StatusHealthy},
		})

		if got != StatusDegraded {
			t.Fatalf("aggregate = %v, want StatusDegraded", got)
		}

		got = aggregate([]Result{
			{Status: StatusHealthy},
			{Status: StatusDegraded},
			{Status: StatusUnhealthy},
		})

		if got != StatusUnhealthy {
			t.Fatalf("aggregate = %v, want StatusUnhealthy", got)
		}
	})

	t.Run("unknown is overridden by any classified status", func(t *testing.T) {
		t.Parallel()

		got := aggregate([]Result{
			{Status: StatusUnknown},
			{Status: StatusHealthy},
		})

		if got != StatusHealthy {
			t.Fatalf("aggregate = %v, want StatusHealthy", got)
		}
	})
}

func TestProbeOne(t *testing.T) {
	t.Parallel()

	t.Run("nil check returns unknown with check-nil message", func(t *testing.T) {
		t.Parallel()

		got := probeOne(context.Background(), nil)
		if got.Status != StatusUnknown {
			t.Fatalf("probeOne(nil check).Status = %v, want StatusUnknown", got.Status)
		}

		if got.Message != ErrCheckNil.Error() {
			t.Fatalf("probeOne(nil check).Message = %q, want %q", got.Message, ErrCheckNil.Error())
		}
	})

	t.Run("nil context returns unknown with context-nil message", func(t *testing.T) {
		t.Parallel()

		stub := &stubCheck{name: "x", result: Result{Status: StatusHealthy}}

		// Pass a typed nil context — go vet would complain about a literal nil.
		var ctx context.Context

		got := probeOne(ctx, stub)
		if got.Status != StatusUnknown {
			t.Fatalf("probeOne(nil ctx).Status = %v, want StatusUnknown", got.Status)
		}

		if got.Message != ErrContextNil.Error() {
			t.Fatalf("probeOne(nil ctx).Message = %q, want %q", got.Message, ErrContextNil.Error())
		}

		if got.Name != "x" {
			t.Fatalf("probeOne(nil ctx).Name = %q, want %q", got.Name, "x")
		}

		if stub.calls.Load() != 0 {
			t.Fatalf("probe was invoked despite nil ctx")
		}
	})

	t.Run("forces canonical Name and Duration", func(t *testing.T) {
		t.Parallel()

		stub := &stubCheck{
			name:  "canonical",
			delay: 5 * time.Millisecond,
			result: Result{
				Name:     "ignored-name", // must be overwritten by probeOne
				Status:   StatusHealthy,
				Duration: 999 * time.Hour, // must be overwritten by probeOne
			},
		}

		got := probeOne(context.Background(), stub)

		if got.Name != "canonical" {
			t.Fatalf("Name = %q, want %q (probeOne must force the canonical name)", got.Name, "canonical")
		}

		if got.Duration <= 0 {
			t.Fatalf("Duration = %v, want > 0", got.Duration)
		}

		if got.Duration >= time.Hour {
			t.Fatalf("Duration = %v, want measured time, not the probe-provided value", got.Duration)
		}

		if got.Status != StatusHealthy {
			t.Fatalf("Status = %v, want StatusHealthy", got.Status)
		}
	})
}
