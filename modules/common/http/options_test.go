package http

import (
	"crypto/tls"
	"errors"
	stdhttp "net/http"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// dummyRT is a simple RoundTripper that is NOT a *http.Transport, to exercise the non-clone path.
type dummyRT struct{}

var errDummyRT = errors.New("dummy round tripper")

func (dummyRT) RoundTrip(*stdhttp.Request) (*stdhttp.Response, error) { return nil, errDummyRT }

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("applies client defaults", func(t *testing.T) {
		t.Parallel()

		o := NewOptions()
		if o == nil {
			t.Fatalf("NewOptions returned nil")
		}

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
	})

	t.Run("applies server defaults", func(t *testing.T) {
		t.Parallel()

		o := NewOptions()

		if o.serverReadHeaderTimeout != 5*time.Second {
			t.Fatalf("serverReadHeaderTimeout = %v, want 5s", o.serverReadHeaderTimeout)
		}

		if o.serverReadTimeout != 15*time.Second {
			t.Fatalf("serverReadTimeout = %v, want 15s", o.serverReadTimeout)
		}

		if o.serverWriteTimeout != 15*time.Second {
			t.Fatalf("serverWriteTimeout = %v, want 15s", o.serverWriteTimeout)
		}

		if o.serverIdleTimeout != 60*time.Second {
			t.Fatalf("serverIdleTimeout = %v, want 60s", o.serverIdleTimeout)
		}

		if o.serverMaxHeaderBytes != 1<<20 {
			t.Fatalf("serverMaxHeaderBytes = %d, want %d", o.serverMaxHeaderBytes, 1<<20)
		}

		if o.serverTLSConfig != nil {
			t.Fatalf("serverTLSConfig = %v, want nil", o.serverTLSConfig)
		}
	})

	t.Run("clones default transport", func(t *testing.T) {
		t.Parallel()

		o := NewOptions()

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

		if gotTr.TLSHandshakeTimeout != defTr.TLSHandshakeTimeout {
			t.Fatalf("TLSHandshakeTimeout changed unexpectedly: got %v want %v", gotTr.TLSHandshakeTimeout, defTr.TLSHandshakeTimeout)
		}

		if gotTr.ResponseHeaderTimeout != defTr.ResponseHeaderTimeout {
			t.Fatalf("ResponseHeaderTimeout changed unexpectedly: got %v want %v", gotTr.ResponseHeaderTimeout, defTr.ResponseHeaderTimeout)
		}

		if gotTr.ExpectContinueTimeout != defTr.ExpectContinueTimeout {
			t.Fatalf("ExpectContinueTimeout changed unexpectedly: got %v want %v", gotTr.ExpectContinueTimeout, defTr.ExpectContinueTimeout)
		}
	})

	t.Run("caps transport timeouts to client timeout", func(t *testing.T) {
		t.Parallel()

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

		gotTr, ok := o.clientTransport.(*stdhttp.Transport)
		if !ok {
			t.Fatalf("options.transport is not *http.Transport: %T", o.clientTransport)
		}

		if gotTr == orig {
			t.Fatalf("transport was not cloned; same pointer returned")
		}

		if gotTr.TLSHandshakeTimeout != timeout {
			t.Fatalf("TLSHandshakeTimeout not capped: got %v want %v", gotTr.TLSHandshakeTimeout, timeout)
		}

		if gotTr.ResponseHeaderTimeout != timeout {
			t.Fatalf("ResponseHeaderTimeout not capped: got %v want %v", gotTr.ResponseHeaderTimeout, timeout)
		}

		if gotTr.ExpectContinueTimeout != timeout {
			t.Fatalf("ExpectContinueTimeout not capped: got %v want %v", gotTr.ExpectContinueTimeout, timeout)
		}

		if orig.TLSHandshakeTimeout != 10*time.Second || orig.ResponseHeaderTimeout != 5*time.Second || orig.ExpectContinueTimeout != 3*time.Second {
			t.Fatalf("original transport mutated: %+v", orig)
		}
	})

	t.Run("keeps custom non-transport round tripper", func(t *testing.T) {
		t.Parallel()

		d := dummyRT{}

		o := NewOptions(
			WithClientTransport(d),
			WithClientTimeout(2*time.Second),
		)
		if o.clientTransport != d {
			t.Fatalf("expected custom RoundTripper to be kept as-is")
		}
	})

	t.Run("auto-wires RetryIfHttpError when retryOnResponse set", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithClientRetryOnResponse(RetryOn5xxAnd429Response))

		got := o.clientRetryIf(&StatusCodeError{StatusCode: 503})
		if !got {
			t.Fatalf("expected auto-wired RetryIfHttpError to return true for StatusCodeError; got %v", got)
		}
	})

	t.Run("keeps user retryIf when both set", func(t *testing.T) {
		t.Parallel()

		custom := func(err error) bool { return false }

		o := NewOptions(
			WithClientRetryOnResponse(RetryOn5xxAnd429Response),
			WithClientRetryIf(custom),
		)

		got := o.clientRetryIf(&StatusCodeError{StatusCode: 503})
		if got {
			t.Fatalf("expected user-provided retryIf to be kept; got %v", got)
		}
	})

	t.Run("keeps NoopRetryIf when neither set", func(t *testing.T) {
		t.Parallel()

		o := NewOptions()

		got := o.clientRetryIf(&StatusCodeError{StatusCode: 503})
		if got {
			t.Fatalf("expected NoopRetryIf default; got %v", got)
		}
	})

	t.Run("hardens burst to one when rate is finite", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithClientLimiterRate(10))

		if o.clientLimiterRate != rate.Limit(10) {
			t.Fatalf("clientLimiterRate = %v, want %v", o.clientLimiterRate, rate.Limit(10))
		}

		if o.clientLimiterBurst != 1 {
			t.Fatalf("clientLimiterBurst = %d, want 1", o.clientLimiterBurst)
		}
	})
}

func TestWithClientTimeout(t *testing.T) {
	t.Parallel()

	t.Run("applies positive value", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithClientTimeout(123 * time.Millisecond))
		if o.clientTimeout != 123*time.Millisecond {
			t.Fatalf("WithTimeout positive not applied; got %v", o.clientTimeout)
		}
	})

	t.Run("ignores zero", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithClientTimeout(0))
		if o.clientTimeout != 30*time.Second {
			t.Fatalf("WithTimeout(0) should keep default; got %v", o.clientTimeout)
		}
	})
}

func TestWithClientTransport(t *testing.T) {
	t.Parallel()

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithClientTransport(nil))

		_, ok := o.clientTransport.(*stdhttp.Transport)
		if !ok {
			t.Fatalf("WithTransport(nil) should keep default *http.Transport")
		}
	})
}

func TestWithClientAttempts(t *testing.T) {
	t.Parallel()

	t.Run("ignores one", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithClientAttempts(1))
		if o.clientAttempts != 1 {
			t.Fatalf("WithAttempts(1) should not change default; got %d", o.clientAttempts)
		}
	})

	t.Run("applies value greater than one", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithClientAttempts(5))
		if o.clientAttempts != 5 {
			t.Fatalf("WithAttempts(5) not applied; got %d", o.clientAttempts)
		}
	})
}

func TestWithClientRetryIf(t *testing.T) {
	t.Parallel()

	t.Run("applies function", func(t *testing.T) {
		t.Parallel()

		retryIf := func(err error) bool { return err == nil }

		o := NewOptions(WithClientRetryIf(retryIf))

		got := o.clientRetryIf(nil)
		if !got {
			t.Fatalf("retryIf not set or unexpected behavior; got %v", got)
		}
	})

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithClientRetryIf(nil))

		got := o.clientRetryIf(nil)
		if got {
			t.Fatalf("expected NoopRetryIf default; got %v", got)
		}
	})
}

func TestWithClientRetryHook(t *testing.T) {
	t.Parallel()

	t.Run("applies function", func(t *testing.T) {
		t.Parallel()

		var called bool

		retryHook := func(n uint, err error) { called = true }

		o := NewOptions(WithClientRetryHook(retryHook))
		o.clientRetryHook(0, nil)

		if !called {
			t.Fatalf("retryHook not set or not invoked")
		}
	})

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithClientRetryHook(nil))
		o.clientRetryHook(0, nil)
	})
}

func TestWithClientRetryOnResponse(t *testing.T) {
	t.Parallel()

	t.Run("applies function", func(t *testing.T) {
		t.Parallel()

		var called bool

		fn := func(res *stdhttp.Response) bool { called = true; return true }

		o := NewOptions(WithClientRetryOnResponse(RetryOnResponseFn(fn)))

		got := o.clientRetryOnResponse(&stdhttp.Response{StatusCode: 200})
		if !got || !called {
			t.Fatalf("retryOnResponse not set or not invoked")
		}
	})

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithClientRetryOnResponse(nil))

		got := o.clientRetryOnResponse(&stdhttp.Response{StatusCode: 200})
		if got {
			t.Fatalf("expected NoopRetryOnResponse default; got %v", got)
		}
	})
}

func TestWithClientLimiterRate(t *testing.T) {
	t.Parallel()

	t.Run("applies positive value", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithClientLimiterRate(5))
		if o.clientLimiterRate != rate.Limit(5) {
			t.Fatalf("clientLimiterRate = %v, want %v", o.clientLimiterRate, rate.Limit(5))
		}
	})

	t.Run("ignores rate.Inf", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithClientLimiterRate(float64(rate.Inf)))
		if o.clientLimiterRate != rate.Inf {
			t.Fatalf("clientLimiterRate with Inf should remain Inf; got %v", o.clientLimiterRate)
		}

		if o.clientLimiterBurst != 0 {
			t.Fatalf("clientLimiterBurst should remain 0 when rate is Inf; got %d", o.clientLimiterBurst)
		}
	})

	t.Run("ignores negative", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithClientLimiterRate(-1))
		if o.clientLimiterRate != rate.Inf {
			t.Fatalf("WithClientLimiterRate(-1) should keep default Inf; got %v", o.clientLimiterRate)
		}
	})

	t.Run("ignores zero", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithClientLimiterRate(0))
		if o.clientLimiterRate != rate.Inf {
			t.Fatalf("WithClientLimiterRate(0) should keep default Inf; got %v", o.clientLimiterRate)
		}
	})
}

func TestWithClientLimiterBurst(t *testing.T) {
	t.Parallel()

	t.Run("preserves explicit burst", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(
			WithClientLimiterRate(5),
			WithClientLimiterBurst(7),
		)
		if o.clientLimiterBurst != 7 {
			t.Fatalf("clientLimiterBurst should be preserved when >0; got %d", o.clientLimiterBurst)
		}
	})
}

func TestWithServerReadHeaderTimeout(t *testing.T) {
	t.Parallel()

	t.Run("applies positive value", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithServerReadHeaderTimeout(10 * time.Second))
		if o.serverReadHeaderTimeout != 10*time.Second {
			t.Fatalf("serverReadHeaderTimeout = %v, want 10s", o.serverReadHeaderTimeout)
		}
	})

	t.Run("ignores zero", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithServerReadHeaderTimeout(0))
		if o.serverReadHeaderTimeout != 5*time.Second {
			t.Fatalf("WithServerReadHeaderTimeout(0) should keep default; got %v", o.serverReadHeaderTimeout)
		}
	})
}

func TestWithServerReadTimeout(t *testing.T) {
	t.Parallel()

	t.Run("applies positive value", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithServerReadTimeout(30 * time.Second))
		if o.serverReadTimeout != 30*time.Second {
			t.Fatalf("serverReadTimeout = %v, want 30s", o.serverReadTimeout)
		}
	})

	t.Run("ignores zero", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithServerReadTimeout(0))
		if o.serverReadTimeout != 15*time.Second {
			t.Fatalf("WithServerReadTimeout(0) should keep default; got %v", o.serverReadTimeout)
		}
	})
}

func TestWithServerWriteTimeout(t *testing.T) {
	t.Parallel()

	t.Run("applies positive value", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithServerWriteTimeout(45 * time.Second))
		if o.serverWriteTimeout != 45*time.Second {
			t.Fatalf("serverWriteTimeout = %v, want 45s", o.serverWriteTimeout)
		}
	})

	t.Run("ignores zero", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithServerWriteTimeout(0))
		if o.serverWriteTimeout != 15*time.Second {
			t.Fatalf("WithServerWriteTimeout(0) should keep default; got %v", o.serverWriteTimeout)
		}
	})
}

func TestWithServerIdleTimeout(t *testing.T) {
	t.Parallel()

	t.Run("applies positive value", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithServerIdleTimeout(120 * time.Second))
		if o.serverIdleTimeout != 120*time.Second {
			t.Fatalf("serverIdleTimeout = %v, want 120s", o.serverIdleTimeout)
		}
	})

	t.Run("ignores zero", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithServerIdleTimeout(0))
		if o.serverIdleTimeout != 60*time.Second {
			t.Fatalf("WithServerIdleTimeout(0) should keep default; got %v", o.serverIdleTimeout)
		}
	})
}

func TestWithServerMaxHeaderBytes(t *testing.T) {
	t.Parallel()

	t.Run("applies positive value", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithServerMaxHeaderBytes(2 << 20))
		if o.serverMaxHeaderBytes != 2<<20 {
			t.Fatalf("serverMaxHeaderBytes = %d, want %d", o.serverMaxHeaderBytes, 2<<20)
		}
	})

	t.Run("ignores zero", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithServerMaxHeaderBytes(0))
		if o.serverMaxHeaderBytes != 1<<20 {
			t.Fatalf("WithServerMaxHeaderBytes(0) should keep default; got %d", o.serverMaxHeaderBytes)
		}
	})
}

func TestWithServerTLSConfig(t *testing.T) {
	t.Parallel()

	t.Run("applies config", func(t *testing.T) {
		t.Parallel()

		cfg := &tls.Config{MinVersion: tls.VersionTLS13}

		o := NewOptions(WithServerTLSConfig(cfg))
		if o.serverTLSConfig != cfg {
			t.Fatalf("serverTLSConfig not set")
		}
	})

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithServerTLSConfig(nil))
		if o.serverTLSConfig != nil {
			t.Fatalf("WithServerTLSConfig(nil) should keep default nil; got %v", o.serverTLSConfig)
		}
	})
}
