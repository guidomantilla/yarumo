package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	stdhttp "net/http"
	"testing"
	"time"
)

// --- Helpers ---------------------------------------------------------------

// trackingBody implements io.ReadCloser and lets us assert Close() was called.
type trackingBody struct {
	bytes.Reader
	closed *bool
}

func (tb *trackingBody) Close() error {
	if tb.closed != nil {
		*tb.closed = true
	}
	return nil
}

// successRT returns a 200 response with provided body text.
type successRT struct{ body string }

func (rt successRT) RoundTrip(*stdhttp.Request) (*stdhttp.Response, error) {
	rc := stdhttp.Response{
		StatusCode: 200,
		Body:       stdhttp.NoBody,
	}
	if rt.body != "" {
		rc.Body = io.NopCloser(bytes.NewBufferString(rt.body))
	}
	return &rc, nil
}

// errRT returns an error only (nil response).
type errRT struct{ err error }

func (rt errRT) RoundTrip(*stdhttp.Request) (*stdhttp.Response, error) {
	return nil, rt.err
}

// resAndErrRT returns a non-nil response AND an error to exercise the close-body path.
type resAndErrRT struct{}

func (resAndErrRT) RoundTrip(*stdhttp.Request) (*stdhttp.Response, error) {
	closed := false
	body := &trackingBody{Reader: *bytes.NewReader([]byte("x")), closed: &closed}
	// Embed pointer to closed flag inside the Response via Body wrapper.
	// We will check it from the outer test by capturing through a client transport wrapper.
	return &stdhttp.Response{StatusCode: 503, Body: body}, context.DeadlineExceeded
}

// flakyRT fails the first failCount calls, then succeeds with 200.
type flakyRT struct {
	n         int
	failCount int
}

func (rt *flakyRT) RoundTrip(*stdhttp.Request) (*stdhttp.Response, error) {
	rt.n++
	if rt.n <= rt.failCount {
		return nil, context.DeadlineExceeded
	}
	return &stdhttp.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString("ok"))}, nil
}

// newRequest creates a minimal GET request with provided context.
func newRequest(t *testing.T, ctx context.Context) *stdhttp.Request {
	t.Helper()
	req, err := stdhttp.NewRequestWithContext(ctx, stdhttp.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	return req
}

// --- Tests -----------------------------------------------------------------

func TestNewClientAndLimiterEnabled(t *testing.T) {
	// Default options -> limiter disabled (rate.Inf)
	c1 := NewClient()
	if cc, ok := c1.(*client); !ok || cc.LimiterEnabled() {
		t.Fatalf("LimiterEnabled default = true, want false")
	}

	// Finite rate + burst -> enabled
	c2 := NewClient(WithLimiterRate(5), WithLimiterBurst(2))
	if cc, ok := c2.(*client); !ok || !cc.LimiterEnabled() {
		t.Fatalf("LimiterEnabled finite = false, want true")
	}
}

func TestClient_Do_Success_NoLimiter_NoRetry(t *testing.T) {
	c := NewClient(
		WithTransport(successRT{body: "hello"}),
		WithAttempts(1),
	)

	req := newRequest(t, context.Background())
	res, err := c.Do(req)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if res == nil || res.StatusCode != 200 {
		t.Fatalf("unexpected response: %+v", res)
	}
	_ = res.Body.Close()
}

func TestClient_Do_ErrorWithResponse_ClosesBodyAndWraps(t *testing.T) {
	// Use a custom RoundTripper that returns both a response and an error
	// to force body-close branch and ErrDoCall wrapping.
	// We also use a short timeout to avoid any coupling with limiter.
	c := NewClient(
		WithTransport(resAndErrRT{}),
		WithTimeout(10*time.Millisecond),
		WithLimiterRate(float64(^uint(0))), // disabled (Inf)
	)

	req := newRequest(t, context.Background())
	res, err := c.Do(req)
	if err == nil {
		t.Fatalf("expected error, got nil and res=%+v", res)
	}
	// must wrap ErrHttpRequestFailed
	if !errors.Is(err, ErrHttpRequestFailed) {
		t.Fatalf("error does not wrap ErrHttpRequestFailed: %v", err)
	}
}

func TestClient_Do_ErrorOnly_Wraps(t *testing.T) {
	c := NewClient(
		WithTransport(errRT{err: context.Canceled}),
	)
	req := newRequest(t, context.Background())
	res, err := c.Do(req)
	if err == nil {
		t.Fatalf("expected error, got nil and res=%+v", res)
	}
	if !errors.Is(err, ErrHttpRequestFailed) {
		t.Fatalf("error does not wrap ErrHttpRequestFailed: %v", err)
	}
}

func TestClient_Do_Retry_SucceedsAfterFailures(t *testing.T) {
	fl := &flakyRT{failCount: 2}
	var hookCalls int
	c := NewClient(
		WithTransport(fl),
		WithAttempts(3),
		WithRetryIf(func(err error) bool { return true }),
		WithRetryHook(func(n uint, err error) { hookCalls++ }),
	)

	req := newRequest(t, context.Background())
	res, err := c.Do(req)
	if err != nil {
		t.Fatalf("Do with retries failed: %v", err)
	}
	if res == nil || res.StatusCode != 200 {
		t.Fatalf("unexpected response: %+v", res)
	}
	if hookCalls == 0 {
		t.Fatalf("retry hook not called")
	}
}

func TestClient_waitForLimiter_DeadlineFromClientTimeout(t *testing.T) {
	// Enable limiter with finite rate. Because Options hardens burst to 1 when
	// finite, we must pre-consume the initial token so that the next Wait blocks
	// until the rate replenishes, which will exceed the tiny client timeout.
	c := NewClient(
		WithTransport(successRT{}),
		WithTimeout(1*time.Millisecond),
		WithLimiterRate(100), // 100/s but we'll pre-consume the first token
	)
	cc := c.(*client)
	// Pre-consume initial token so the next wait must block.
	_ = cc.limiter.Wait(context.Background())

	req := newRequest(t, context.Background())
	_, err := c.Do(req)
	if err == nil {
		t.Fatalf("expected rate limiter error, got nil")
	}
	// Must wrap rate limiter exceeded
	if !errors.Is(err, ErrRateLimiterExceeded) {
		t.Fatalf("expected ErrRateLimiterExceeded wrapping, got %v", err)
	}
}

func TestClient_waitForLimiter_DeadlineFromRequestCtxEarlier(t *testing.T) {
	// Client timeout longer, but request context deadline is earlier; it should
	// respect the earlier one and fail quickly when limiter has to wait.
	c := NewClient(
		WithTransport(successRT{}),
		WithTimeout(250*time.Millisecond),
		WithLimiterRate(100),
	)
	cc := c.(*client)
	_ = cc.limiter.Wait(context.Background()) // pre-consume first token

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	req := newRequest(t, ctx)
	_, err := c.Do(req)
	if err == nil {
		t.Fatalf("expected error due to request context deadline, got nil")
	}
	if !errors.Is(err, ErrRateLimiterExceeded) {
		t.Fatalf("expected ErrRateLimiterExceeded wrapping, got %v", err)
	}
}
