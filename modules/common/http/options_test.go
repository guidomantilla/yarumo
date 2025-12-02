package http

import (
	stdhttp "net/http"
	"testing"
	"time"

	retry "github.com/avast/retry-go/v4"
	"golang.org/x/time/rate"
)

// dummyRT is a simple RoundTripper that is NOT a *http.Transport, to exercise the non-clone path
type dummyRT struct{}

func (dummyRT) RoundTrip(*stdhttp.Request) (*stdhttp.Response, error) { return nil, nil }

func TestNewOptions_DefaultsAndClone(t *testing.T) {
	o := NewOptions()
	if o == nil {
		t.Fatalf("NewOptions returned nil")
	}

	// Defaults
	if o.timeout != 30*time.Second {
		t.Fatalf("default timeout = %v, want %v", o.timeout, 30*time.Second)
	}
	if o.attempts != 1 {
		t.Fatalf("default attempts = %d, want 1", o.attempts)
	}
	if o.limiterRate != rate.Inf {
		t.Fatalf("default limiterRate = %v, want %v", o.limiterRate, rate.Inf)
	}
	if o.limiterBurst != 0 {
		t.Fatalf("default limiterBurst = %d, want 0", o.limiterBurst)
	}

	// Transport should be a clone of the default one when timeout is non-zero
	// and must not be the same pointer as stdhttp.DefaultTransport
	defTr, ok := stdhttp.DefaultTransport.(*stdhttp.Transport)
	if !ok {
		t.Fatalf("stdhttp.DefaultTransport is not *http.Transport")
	}
	gotTr, ok := o.transport.(*stdhttp.Transport)
	if !ok {
		t.Fatalf("options.transport is not *http.Transport: %T", o.transport)
	}
	if gotTr == defTr {
		t.Fatalf("transport must be a cloned instance, got same pointer")
	}
	// Since default timeouts are not greater than 30s, the values should remain equal
	if gotTr.TLSHandshakeTimeout != defTr.TLSHandshakeTimeout {
		t.Fatalf("TLSHandshakeTimeout changed unexpectedly: got %v want %v", gotTr.TLSHandshakeTimeout, defTr.TLSHandshakeTimeout)
	}
	if gotTr.ResponseHeaderTimeout != defTr.ResponseHeaderTimeout {
		t.Fatalf("ResponseHeaderTimeout changed unexpectedly: got %v want %v", gotTr.ResponseHeaderTimeout, defTr.ResponseHeaderTimeout)
	}
	if gotTr.ExpectContinueTimeout != defTr.ExpectContinueTimeout {
		t.Fatalf("ExpectContinueTimeout changed unexpectedly: got %v want %v", gotTr.ExpectContinueTimeout, defTr.ExpectContinueTimeout)
	}
}

func TestNewOptions_TimeoutAlignmentCapsTransport(t *testing.T) {
	// Prepare custom transport with timeouts larger than the client timeout
	orig := &stdhttp.Transport{
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 3 * time.Second,
	}
	timeout := 500 * time.Millisecond

	o := NewOptions(
		WithTransport(orig),
		WithTimeout(timeout),
	)

	// Ensure we got a clone and different instance
	gotTr, ok := o.transport.(*stdhttp.Transport)
	if !ok {
		t.Fatalf("options.transport is not *http.Transport: %T", o.transport)
	}
	if gotTr == orig {
		t.Fatalf("transport was not cloned; same pointer returned")
	}

	// The non-zero transport timeouts above client timeout must be capped to client timeout
	if gotTr.TLSHandshakeTimeout != timeout {
		t.Fatalf("TLSHandshakeTimeout not capped: got %v want %v", gotTr.TLSHandshakeTimeout, timeout)
	}
	if gotTr.ResponseHeaderTimeout != timeout {
		t.Fatalf("ResponseHeaderTimeout not capped: got %v want %v", gotTr.ResponseHeaderTimeout, timeout)
	}
	if gotTr.ExpectContinueTimeout != timeout {
		t.Fatalf("ExpectContinueTimeout not capped: got %v want %v", gotTr.ExpectContinueTimeout, timeout)
	}

	// Original transport must remain unmodified
	if orig.TLSHandshakeTimeout != 10*time.Second || orig.ResponseHeaderTimeout != 5*time.Second || orig.ExpectContinueTimeout != 3*time.Second {
		t.Fatalf("original transport mutated: %+v", orig)
	}
}

func TestNewOptions_CustomNonTransportKeepsRoundTripper(t *testing.T) {
	d := dummyRT{}
	o := NewOptions(
		WithTransport(d),
		WithTimeout(2*time.Second), // alignment should not clone since not *Transport
	)
	if o.transport != d {
		t.Fatalf("expected custom RoundTripper to be kept as-is")
	}
}

func TestOptionsSetters_AttemptsTimeoutTransportNil(t *testing.T) {
	// WithAttempts: only if >1
	o1 := NewOptions(WithAttempts(1))
	if o1.attempts != 1 {
		t.Fatalf("WithAttempts(1) should not change default; got %d", o1.attempts)
	}
	o2 := NewOptions(WithAttempts(5))
	if o2.attempts != 5 {
		t.Fatalf("WithAttempts(5) not applied; got %d", o2.attempts)
	}

	// WithTimeout: only if >0
	o3 := NewOptions(WithTimeout(0))
	if o3.timeout != 30*time.Second {
		t.Fatalf("WithTimeout(0) should keep default; got %v", o3.timeout)
	}
	o4 := NewOptions(WithTimeout(123 * time.Millisecond))
	if o4.timeout != 123*time.Millisecond {
		t.Fatalf("WithTimeout positive not applied; got %v", o4.timeout)
	}

	// WithTransport(nil) is ignored; should stay at default transport (cloned)
	o5 := NewOptions(WithTransport(nil))
	if _, ok := o5.transport.(*stdhttp.Transport); !ok {
		t.Fatalf("WithTransport(nil) should keep default *http.Transport")
	}
}

func TestOptions_RetryAndLimiter(t *testing.T) {
	// Provide non-nil retryIf and retryHook and ensure they are set
	var calledHook bool
	retryIf := func(err error) bool { return err == nil }
	retryHook := func(n uint, err error) { calledHook = true }

	// Finite limiter rate without burst triggers hardening (burst normalized to 1)
	o := NewOptions(
		WithRetryIf(retryIf),
		WithRetryHook(retryHook),
		WithLimiterRate(10), // finite
		// no burst provided -> should be normalized to 1
	)

	// Exercise the stored funcs to ensure assignment
	if got := o.retryIf(nil); got != true {
		t.Fatalf("retryIf not set or unexpected behavior; got %v", got)
	}
	o.retryHook(0, nil)
	if !calledHook {
		t.Fatalf("retryHook not set or not invoked")
	}

	if o.limiterRate != rate.Limit(10) {
		t.Fatalf("limiterRate = %v, want %v", o.limiterRate, rate.Limit(10))
	}
	if o.limiterBurst != 1 {
		t.Fatalf("limiterBurst hardening failed: got %d, want 1", o.limiterBurst)
	}

	// When burst > 0, it must be preserved
	o2 := NewOptions(
		WithLimiterRate(5),
		WithLimiterBurst(7),
	)
	if o2.limiterRate != rate.Limit(5) {
		t.Fatalf("limiterRate = %v, want %v", o2.limiterRate, rate.Limit(5))
	}
	if o2.limiterBurst != 7 {
		t.Fatalf("limiterBurst should be preserved when >0; got %d", o2.limiterBurst)
	}

	// Passing Inf must keep default rate and not trigger hardening
	o3 := NewOptions(
		WithLimiterRate(float64(rate.Inf)),
	)
	if o3.limiterRate != rate.Inf {
		t.Fatalf("limiterRate with Inf should remain Inf; got %v", o3.limiterRate)
	}
	if o3.limiterBurst != 0 {
		t.Fatalf("limiterBurst should remain 0 when rate is Inf; got %d", o3.limiterBurst)
	}

	// Silence unused import warning for retry package in case signers change
	var _ retry.Option
}

func TestOptions_WithRetryOnResponse_SetAndIgnoreNil(t *testing.T) {
	var called bool
	fn := func(res *stdhttp.Response) bool { called = true; return true }

	// nil should be ignored; non-nil should be applied
	o := NewOptions(
		WithRetryOnResponse(nil),
		WithRetryOnResponse(RetryOnResponseFn(fn)),
	)
	if !o.retryOnResponse(&stdhttp.Response{StatusCode: 200}) || !called {
		t.Fatalf("retryOnResponse not set or not invoked")
	}
}
