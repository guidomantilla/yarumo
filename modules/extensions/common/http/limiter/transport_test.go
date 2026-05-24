package limiter

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// fakeRoundTripper records the number of RoundTrip calls and returns a
// configured response/error.
type fakeRoundTripper struct {
	calls    atomic.Int64
	response *http.Response
	err      error
}

func (f *fakeRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	f.calls.Add(1)
	return f.response, f.err
}

func newOKResponse() *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}
}

// permissiveLimiter is a limiter whose Wait never blocks under typical
// test usage (infinite tokens). Used by tests that do not exercise
// throttling behavior.
func permissiveLimiter() *rate.Limiter {
	return rate.NewLimiter(rate.Inf, 0)
}

func TestNewLimiterTransport(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil RoundTripper", func(t *testing.T) {
		t.Parallel()

		rt := NewLimiterTransport(http.DefaultTransport, permissiveLimiter())
		if rt == nil {
			t.Fatal("expected non-nil RoundTripper")
		}
	})

	t.Run("falls back to http.DefaultTransport when base is nil", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		rt := NewLimiterTransport(nil, permissiveLimiter())

		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}

		res, rtErr := rt.RoundTrip(req)
		if rtErr != nil {
			t.Fatalf("RoundTrip: %v", rtErr)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("StatusCode = %d, want %d", res.StatusCode, http.StatusNoContent)
		}
	})
}

func TestLimiterTransport_RoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("delegates to base after acquiring token", func(t *testing.T) {
		t.Parallel()

		base := &fakeRoundTripper{response: newOKResponse()}
		rt := NewLimiterTransport(base, permissiveLimiter())

		req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}

		res, rtErr := rt.RoundTrip(req)
		if rtErr != nil {
			t.Fatalf("RoundTrip: %v", rtErr)
		}

		if res.StatusCode != http.StatusOK {
			t.Fatalf("StatusCode = %d, want %d", res.StatusCode, http.StatusOK)
		}

		if base.calls.Load() != 1 {
			t.Fatalf("base.RoundTrip called %d times, want 1", base.calls.Load())
		}
	})

	t.Run("waits on limiter before delegating", func(t *testing.T) {
		t.Parallel()

		base := &fakeRoundTripper{response: newOKResponse()}
		// 1 token, burst 1. Second immediate request waits.
		limiter := rate.NewLimiter(rate.Every(50*time.Millisecond), 1)
		rt := NewLimiterTransport(base, limiter)

		req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}

		// Consume the initial burst token.
		_, _ = rt.RoundTrip(req)

		start := time.Now()
		_, rtErr := rt.RoundTrip(req)
		elapsed := time.Since(start)

		if rtErr != nil {
			t.Fatalf("RoundTrip: %v", rtErr)
		}

		if elapsed < 40*time.Millisecond {
			t.Fatalf("expected wait >= 40ms, got %v", elapsed)
		}

		if base.calls.Load() != 2 {
			t.Fatalf("base.RoundTrip called %d times, want 2", base.calls.Load())
		}
	})

	t.Run("returns ErrRateLimiterExceeded when context expires while waiting", func(t *testing.T) {
		t.Parallel()

		base := &fakeRoundTripper{response: newOKResponse()}
		// Effectively unreachable second token within a short context.
		limiter := rate.NewLimiter(rate.Every(time.Hour), 1)
		rt := NewLimiterTransport(base, limiter)

		// Burn the initial token.
		ctx := context.Background()
		warmReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)
		_, _ = rt.RoundTrip(warmReq)

		shortCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		req, _ := http.NewRequestWithContext(shortCtx, http.MethodGet, "http://example.com", nil)
		_, err := rt.RoundTrip(req)
		if err == nil {
			t.Fatal("expected error when context expires while waiting on limiter")
		}

		if !errors.Is(err, ErrRateLimiterFailed) {
			t.Fatalf("expected wrap of ErrRateLimiterFailed, got %v", err)
		}
	})

	t.Run("propagates base RoundTrip error", func(t *testing.T) {
		t.Parallel()

		baseErr := errors.New("base failed")
		base := &fakeRoundTripper{err: baseErr}
		rt := NewLimiterTransport(base, permissiveLimiter())

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		_, rtErr := rt.RoundTrip(req)
		if rtErr == nil {
			t.Fatal("expected error from base")
		}

		if !errors.Is(rtErr, baseErr) {
			t.Fatalf("expected wrap of base error, got %v", rtErr)
		}
	})
}
