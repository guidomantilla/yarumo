package retry

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	rretry "github.com/guidomantilla/yarumo/core/common/resilience/retry"
	rretryimpl "github.com/guidomantilla/yarumo/extension/common/resilience/retry"
)

// fakeRoundTripper produces a configurable response sequence. Each call
// to RoundTrip returns the next response from the slice (and the matching
// error). When the slice is exhausted, the last entry is reused.
type fakeRoundTripper struct {
	calls     atomic.Int64
	responses []*http.Response
	errors    []error
}

func (f *fakeRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	i := int(f.calls.Add(1) - 1)
	if i >= len(f.responses) {
		i = len(f.responses) - 1
	}

	return f.responses[i], f.errors[i]
}

func okResponse() *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}
}

func statusResponse(code int) *http.Response {
	return &http.Response{StatusCode: code, Body: io.NopCloser(bytes.NewReader(nil))}
}

// fastRetrier builds a resilience.Retry with a fixed-1ms delay so tests
// stay fast. attempts is the total attempt count (1 original + N-1
// retries).
func fastRetrier(attempts uint) rretry.Retry {
	return rretryimpl.NewRetry(
		rretryimpl.WithAttempts(attempts),
		rretryimpl.WithDelay(time.Millisecond),
		rretryimpl.WithBackoff(rretry.BackoffFixed),
		rretryimpl.WithRetryIf(RetryIfHttpError),
	)
}

func TestNewRetryTransport(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil transport", func(t *testing.T) {
		t.Parallel()

		rt := NewRetryTransport(http.DefaultTransport, fastRetrier(3))
		if rt == nil {
			t.Fatal("expected non-nil transport")
		}
	})

	t.Run("falls back to http.DefaultTransport when base is nil", func(t *testing.T) {
		t.Parallel()

		rt := NewRetryTransport(nil, fastRetrier(3))
		if rt == nil {
			t.Fatal("expected non-nil transport")
		}
	})
}

func TestRetryTransport_RoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("returns first response when no retries are needed", func(t *testing.T) {
		t.Parallel()

		base := &fakeRoundTripper{responses: []*http.Response{okResponse()}, errors: []error{nil}}
		rt := NewRetryTransport(base, fastRetrier(3))

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		res, err := rt.RoundTrip(req)
		if err != nil {
			t.Fatalf("RoundTrip: %v", err)
		}

		if res.StatusCode != http.StatusOK {
			t.Fatalf("StatusCode = %d, want 200", res.StatusCode)
		}

		if base.calls.Load() != 1 {
			t.Fatalf("base.RoundTrip called %d times, want 1", base.calls.Load())
		}
	})

	t.Run("retries on response when RetryOnResponse returns true", func(t *testing.T) {
		t.Parallel()

		base := &fakeRoundTripper{
			responses: []*http.Response{statusResponse(500), statusResponse(500), okResponse()},
			errors:    []error{nil, nil, nil},
		}
		rt := NewRetryTransport(base, fastRetrier(5), WithRetryOnResponse(RetryOn5xxAnd429))

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		res, err := rt.RoundTrip(req)
		if err != nil {
			t.Fatalf("RoundTrip: %v", err)
		}

		if res.StatusCode != http.StatusOK {
			t.Fatalf("StatusCode = %d, want 200", res.StatusCode)
		}

		if base.calls.Load() != 3 {
			t.Fatalf("base.RoundTrip called %d times, want 3", base.calls.Load())
		}
	})

	t.Run("returns rretry.ErrRetryFailed when attempts are exhausted", func(t *testing.T) {
		t.Parallel()

		base := &fakeRoundTripper{
			responses: []*http.Response{statusResponse(500)},
			errors:    []error{nil},
		}
		rt := NewRetryTransport(base, fastRetrier(3), WithRetryOnResponse(RetryOn5xxAnd429))

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		_, err := rt.RoundTrip(req)
		if err == nil {
			t.Fatal("expected error after attempts exhausted")
		}

		if !errors.Is(err, rretry.ErrRetryFailed) {
			t.Fatalf("expected wrap of rretry.ErrRetryFailed, got %v", err)
		}

		if base.calls.Load() != 3 {
			t.Fatalf("base.RoundTrip called %d times, want 3", base.calls.Load())
		}
	})

	t.Run("returns ErrNonReplayableBody when body is set without GetBody", func(t *testing.T) {
		t.Parallel()

		base := &fakeRoundTripper{responses: []*http.Response{okResponse()}, errors: []error{nil}}
		rt := NewRetryTransport(base, fastRetrier(3))

		req, _ := http.NewRequest(http.MethodPost, "http://example.com", strings.NewReader("hello"))
		// Force GetBody to nil to simulate a non-replayable body.
		req.GetBody = nil

		_, err := rt.RoundTrip(req)
		if err == nil {
			t.Fatal("expected error for non-replayable body")
		}

		if !errors.Is(err, ErrNonReplayableBodyFailed) {
			t.Fatalf("expected wrap of ErrNonReplayableBodyFailed, got %v", err)
		}

		if base.calls.Load() != 0 {
			t.Fatalf("base.RoundTrip called %d times, want 0", base.calls.Load())
		}
	})

	t.Run("invokes retrier OnRetry hook before each retry", func(t *testing.T) {
		t.Parallel()

		base := &fakeRoundTripper{
			responses: []*http.Response{statusResponse(500), okResponse()},
			errors:    []error{nil, nil},
		}

		var hookCalls atomic.Int64
		retrier := rretryimpl.NewRetry(
			rretryimpl.WithAttempts(3),
			rretryimpl.WithDelay(time.Millisecond),
			rretryimpl.WithBackoff(rretry.BackoffFixed),
			rretryimpl.WithRetryIf(RetryIfHttpError),
			rretryimpl.WithOnRetry(func(_ uint, _ error) { hookCalls.Add(1) }),
		)
		rt := NewRetryTransport(base, retrier, WithRetryOnResponse(RetryOn5xxAnd429))

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		_, err := rt.RoundTrip(req)
		if err != nil {
			t.Fatalf("RoundTrip: %v", err)
		}

		// avast/retry-go invokes OnRetry on each failed attempt; with 2
		// attempts where the first fails, the hook fires once.
		if hookCalls.Load() < 1 {
			t.Fatalf("retry hook called %d times, want >= 1", hookCalls.Load())
		}
	})

	t.Run("does not retry on 5xx when RetryOnResponse is not configured", func(t *testing.T) {
		t.Parallel()

		base := &fakeRoundTripper{
			responses: []*http.Response{statusResponse(500)},
			errors:    []error{nil},
		}
		// retrier configured with RetryIfHttpError, but no RetryOnResponse
		// means the transport never synthesizes a StatusCodeError → no retry.
		rt := NewRetryTransport(base, fastRetrier(3))

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		res, err := rt.RoundTrip(req)
		if err != nil {
			t.Fatalf("RoundTrip: %v", err)
		}

		if res.StatusCode != 500 {
			t.Fatalf("StatusCode = %d, want 500", res.StatusCode)
		}

		if base.calls.Load() != 1 {
			t.Fatalf("base.RoundTrip called %d times, want 1", base.calls.Load())
		}
	})
}
