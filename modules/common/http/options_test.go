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
	if o.clientTimeout != 30*time.Second {
		t.Fatalf("default timeout = %v, want %v", o.clientTimeout, 30*time.Second)
	}

	if o.clientAttempts != 1 {
		t.Fatalf("default attempts = %d, want 1", o.clientAttempts)
	}

	if o.clientLimiterRate != rate.Inf {
		t.Fatalf("default clientLimiterRate = %v, want %v", o.clientLimiterRate, rate.Inf)
	}

	if o.clientLimiterBurst != 0 {
		t.Fatalf("default clientLimiterBurst = %d, want 0", o.clientLimiterBurst)
	}

	// Transport should be a clone of the default one when timeout is non-zero
	// and must not be the same pointer as stdhttp.DefaultTransport
	defTr, ok := stdhttp.DefaultTransport.(*stdhttp.Transport)
	if !ok {
		t.Fatalf("stdhttp.DefaultTransport is not *http.Transport")
	}

	gotTr, ok := o.clientTransport.(*stdhttp.Transport)
	if !ok {
		t.Fatalf("options.transport is not *http.Transport: %T", o.clientTransport)
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
		WithClientTransport(orig),
		WithClientTimeout(timeout),
	)

	// Ensure we got a clone and different instance
	gotTr, ok := o.clientTransport.(*stdhttp.Transport)
	if !ok {
		t.Fatalf("options.transport is not *http.Transport: %T", o.clientTransport)
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
		WithClientTransport(d),
		WithClientTimeout(2*time.Second), // alignment should not clone since not *Transport
	)
	if o.clientTransport != d {
		t.Fatalf("expected custom RoundTripper to be kept as-is")
	}
}

func TestOptionsSetters_AttemptsTimeoutTransportNil(t *testing.T) {
	// WithAttempts: only if >1
	o1 := NewOptions(WithClientAttempts(1))
	if o1.clientAttempts != 1 {
		t.Fatalf("WithAttempts(1) should not change default; got %d", o1.clientAttempts)
	}

	o2 := NewOptions(WithClientAttempts(5))
	if o2.clientAttempts != 5 {
		t.Fatalf("WithAttempts(5) not applied; got %d", o2.clientAttempts)
	}

	// WithTimeout: only if >0
	o3 := NewOptions(WithClientTimeout(0))
	if o3.clientTimeout != 30*time.Second {
		t.Fatalf("WithTimeout(0) should keep default; got %v", o3.clientTimeout)
	}

	o4 := NewOptions(WithClientTimeout(123 * time.Millisecond))
	if o4.clientTimeout != 123*time.Millisecond {
		t.Fatalf("WithTimeout positive not applied; got %v", o4.clientTimeout)
	}

	// WithTransport(nil) is ignored; should stay at default transport (cloned)
	o5 := NewOptions(WithClientTransport(nil))
	if _, ok := o5.clientTransport.(*stdhttp.Transport); !ok {
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
		WithClientRetryIf(retryIf),
		WithClientRetryHook(retryHook),
		WithClientLimiterRate(10), // finite
		// no burst provided -> should be normalized to 1
	)

	// Exercise the stored funcs to ensure assignment
	if got := o.clientRetryIf(nil); got != true {
		t.Fatalf("retryIf not set or unexpected behavior; got %v", got)
	}

	o.clientRetryHook(0, nil)

	if !calledHook {
		t.Fatalf("retryHook not set or not invoked")
	}

	if o.clientLimiterRate != rate.Limit(10) {
		t.Fatalf("clientLimiterRate = %v, want %v", o.clientLimiterRate, rate.Limit(10))
	}

	if o.clientLimiterBurst != 1 {
		t.Fatalf("clientLimiterBurst hardening failed: got %d, want 1", o.clientLimiterBurst)
	}

	// When burst > 0, it must be preserved
	o2 := NewOptions(
		WithClientLimiterRate(5),
		WithClientLimiterBurst(7),
	)
	if o2.clientLimiterRate != rate.Limit(5) {
		t.Fatalf("clientLimiterRate = %v, want %v", o2.clientLimiterRate, rate.Limit(5))
	}

	if o2.clientLimiterBurst != 7 {
		t.Fatalf("clientLimiterBurst should be preserved when >0; got %d", o2.clientLimiterBurst)
	}

	// Passing Inf must keep default rate and not trigger hardening
	o3 := NewOptions(
		WithClientLimiterRate(float64(rate.Inf)),
	)
	if o3.clientLimiterRate != rate.Inf {
		t.Fatalf("clientLimiterRate with Inf should remain Inf; got %v", o3.clientLimiterRate)
	}

	if o3.clientLimiterBurst != 0 {
		t.Fatalf("clientLimiterBurst should remain 0 when rate is Inf; got %d", o3.clientLimiterBurst)
	}

	// Silence unused import warning for retry package in case signers change
	var _ retry.Option
}

func TestOptions_WithRetryOnResponse_SetAndIgnoreNil(t *testing.T) {
	var called bool

	fn := func(res *stdhttp.Response) bool { called = true; return true }

	// nil should be ignored; non-nil should be applied
	o := NewOptions(
		WithClientRetryOnResponse(nil),
		WithClientRetryOnResponse(RetryOnResponseFn(fn)),
	)
	if !o.clientRetryOnResponse(&stdhttp.Response{StatusCode: 200}) || !called {
		t.Fatalf("retryOnResponse not set or not invoked")
	}
}
