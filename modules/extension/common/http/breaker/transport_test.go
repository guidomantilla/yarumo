package breaker

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	rbreaker "github.com/guidomantilla/yarumo/extension/common/resilience/breaker"
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
	return &http.Response{StatusCode: code, Body: io.NopCloser(nil)}
}

func TestNewBreakerTransport(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil transport", func(t *testing.T) {
		t.Parallel()

		rt := NewBreakerTransport(http.DefaultTransport, rbreaker.NewBreaker())
		if rt == nil {
			t.Fatal("expected non-nil transport")
		}
	})

	t.Run("falls back to http.DefaultTransport when base is nil", func(t *testing.T) {
		t.Parallel()

		rt := NewBreakerTransport(nil, rbreaker.NewBreaker())
		if rt == nil {
			t.Fatal("expected non-nil transport")
		}
	})
}

func TestBreakerTransport_RoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("delegates to base and returns response when breaker is closed", func(t *testing.T) {
		t.Parallel()

		base := &fakeRoundTripper{responses: []*http.Response{okResponse()}, errors: []error{nil}}
		rt := NewBreakerTransport(base, rbreaker.NewBreaker())

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

	t.Run("transport errors count toward the breaker threshold", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("conn refused")
		base := &fakeRoundTripper{
			responses: []*http.Response{nil, nil, nil},
			errors:    []error{boom, boom, boom},
		}
		b := rbreaker.NewBreaker(rbreaker.WithConsecutiveFailures(2))
		rt := NewBreakerTransport(base, b)

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

		// Two failures trip the breaker.
		_, _ = rt.RoundTrip(req)
		_, _ = rt.RoundTrip(req)

		if b.State() != rbreaker.StateOpen {
			t.Fatalf("breaker state = %s, want open", b.State())
		}
	})

	t.Run("reports synthetic StatusCodeError to the breaker when FailOnResponse fires", func(t *testing.T) {
		t.Parallel()

		base := &fakeRoundTripper{
			responses: []*http.Response{statusResponse(500), statusResponse(500), statusResponse(500)},
			errors:    []error{nil, nil, nil},
		}
		b := rbreaker.NewBreaker(rbreaker.WithConsecutiveFailures(2))
		rt := NewBreakerTransport(base, b, WithFailOnResponse(FailOn5xxAnd429))

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

		// First 500 → reported as failure.
		_, err := rt.RoundTrip(req)
		if err == nil {
			t.Fatal("expected StatusCodeError on first failure")
		}
		var sce *StatusCodeError
		if !errors.As(err, &sce) {
			t.Fatalf("expected *StatusCodeError in chain, got %v", err)
		}
		if sce.StatusCode != 500 {
			t.Fatalf("StatusCode = %d, want 500", sce.StatusCode)
		}

		// Second 500 trips the breaker.
		_, _ = rt.RoundTrip(req)

		if b.State() != rbreaker.StateOpen {
			t.Fatalf("breaker state = %s, want open", b.State())
		}
	})

	t.Run("returns ErrBreakerRejectedFailed wrapping ErrBreakerOpen when breaker is open", func(t *testing.T) {
		t.Parallel()

		base := &fakeRoundTripper{
			responses: []*http.Response{nil, nil, nil},
			errors:    []error{errors.New("fail"), errors.New("fail"), errors.New("fail")},
		}
		b := rbreaker.NewBreaker(rbreaker.WithConsecutiveFailures(1))
		rt := NewBreakerTransport(base, b)

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

		// First call trips the breaker.
		_, _ = rt.RoundTrip(req)
		if b.State() != rbreaker.StateOpen {
			t.Fatalf("expected breaker open after first failure, got %s", b.State())
		}

		// Second call rejected fast.
		baseCallsBefore := base.calls.Load()
		_, err := rt.RoundTrip(req)
		if err == nil {
			t.Fatal("expected ErrBreakerRejected when breaker is open")
		}
		if !errors.Is(err, ErrBreakerRejectedFailed) {
			t.Fatalf("expected wrap of ErrBreakerRejectedFailed, got %v", err)
		}
		if !errors.Is(err, rbreaker.ErrBreakerOpen) {
			t.Fatalf("expected wrap of rbreaker.ErrBreakerOpen, got %v", err)
		}
		if base.calls.Load() != baseCallsBefore {
			t.Fatal("expected base NOT to be called while breaker is open")
		}
	})

	t.Run("does not report success-status responses as failures by default", func(t *testing.T) {
		t.Parallel()

		base := &fakeRoundTripper{
			responses: []*http.Response{statusResponse(500), statusResponse(500), statusResponse(500)},
			errors:    []error{nil, nil, nil},
		}
		b := rbreaker.NewBreaker(rbreaker.WithConsecutiveFailures(2))
		rt := NewBreakerTransport(base, b) // no FailOnResponse

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

		// 5xx returned to caller; breaker does NOT count it (default).
		for range 5 {
			_, _ = rt.RoundTrip(req)
		}

		if b.State() != rbreaker.StateClosed {
			t.Fatalf("breaker state = %s, want closed (no fail predicate configured)", b.State())
		}
	})

	t.Run("recovers via half-open probe when breaker timeout elapses", func(t *testing.T) {
		t.Parallel()

		// First 2 calls fail (trip), then succeeds.
		base := &fakeRoundTripper{
			responses: []*http.Response{nil, nil, okResponse()},
			errors:    []error{errors.New("fail"), errors.New("fail"), nil},
		}
		b := rbreaker.NewBreaker(
			rbreaker.WithConsecutiveFailures(2),
			rbreaker.WithTimeout(30*time.Millisecond),
			rbreaker.WithMaxRequests(1),
		)
		rt := NewBreakerTransport(base, b)

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

		// Trip.
		_, _ = rt.RoundTrip(req)
		_, _ = rt.RoundTrip(req)
		if b.State() != rbreaker.StateOpen {
			t.Fatalf("expected open after trip, got %s", b.State())
		}

		// Wait for timeout → half-open. Probe succeeds → closed.
		time.Sleep(50 * time.Millisecond)

		_, err := rt.RoundTrip(req)
		if err != nil {
			t.Fatalf("probe call: %v", err)
		}
		if b.State() != rbreaker.StateClosed {
			t.Fatalf("breaker state = %s, want closed after successful probe", b.State())
		}
	})

	t.Run("passes context cancellation through to the breaker", func(t *testing.T) {
		t.Parallel()

		base := &fakeRoundTripper{responses: []*http.Response{okResponse()}, errors: []error{nil}}
		b := rbreaker.NewBreaker()
		rt := NewBreakerTransport(base, b)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)
		_, err := rt.RoundTrip(req)
		if err == nil {
			t.Fatal("expected error when ctx is canceled")
		}
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected wrap of context.Canceled, got %v", err)
		}
	})
}
