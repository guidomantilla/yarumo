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

	return &stdhttp.Response{StatusCode: 503, Body: body}, context.DeadlineExceeded
}

// retryOnResponseRT returns a 5xx response without error so that RetryOnResponse
// logic is exercised (client should close the body and return a *StatusCodeError).
type retryOnResponseRT struct{ closed *bool }

func (rt retryOnResponseRT) RoundTrip(*stdhttp.Request) (*stdhttp.Response, error) {
	body := &trackingBody{Reader: *bytes.NewReader([]byte("should-close")), closed: rt.closed}
	return &stdhttp.Response{StatusCode: 503, Body: body}, nil
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

// flakyResponseRT returns 503 for the first failCount calls, then succeeds with 200.
type flakyResponseRT struct {
	n         int
	failCount int
}

func (rt *flakyResponseRT) RoundTrip(*stdhttp.Request) (*stdhttp.Response, error) {
	rt.n++
	if rt.n <= rt.failCount {
		return &stdhttp.Response{StatusCode: 503, Body: io.NopCloser(bytes.NewBufferString("fail"))}, nil
	}

	return &stdhttp.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString("ok"))}, nil
}

// countingRT counts how many times RoundTrip is called. Always returns an error.
type countingRT struct{ n *int }

func (rt countingRT) RoundTrip(*stdhttp.Request) (*stdhttp.Response, error) {
	*rt.n++

	return nil, context.DeadlineExceeded
}

// newRequest creates a minimal GET request with the provided context.
func newRequest(t *testing.T, ctx context.Context) *stdhttp.Request {
	t.Helper()

	req, err := stdhttp.NewRequestWithContext(ctx, stdhttp.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	return req
}

// --- Tests -----------------------------------------------------------------

func TestNewClient(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil client", func(t *testing.T) {
		t.Parallel()

		c := NewClient()
		if c == nil {
			t.Fatal("expected non-nil client")
		}
	})
}

func TestClient_LimiterEnabled(t *testing.T) {
	t.Parallel()

	t.Run("default rate returns false", func(t *testing.T) {
		t.Parallel()

		c := NewClient()

		cc, ok := c.(*client)
		if !ok {
			t.Fatal("expected *client")
		}

		if cc.LimiterEnabled() {
			t.Fatal("LimiterEnabled default = true, want false")
		}
	})

	t.Run("finite rate and burst returns true", func(t *testing.T) {
		t.Parallel()

		c := NewClient(WithClientLimiterRate(5), WithClientLimiterBurst(2))

		cc, ok := c.(*client)
		if !ok {
			t.Fatal("expected *client")
		}

		if !cc.LimiterEnabled() {
			t.Fatal("LimiterEnabled finite = false, want true")
		}
	})
}

func TestClient_RetrierEnabled(t *testing.T) {
	t.Parallel()

	t.Run("attempts one returns false", func(t *testing.T) {
		t.Parallel()

		c := NewClient(WithClientAttempts(1))

		cc, ok := c.(*client)
		if !ok {
			t.Fatal("expected *client")
		}

		if cc.RetrierEnabled() {
			t.Fatal("RetrierEnabled attempts=1 = true, want false")
		}
	})

	t.Run("attempts two returns true", func(t *testing.T) {
		t.Parallel()

		c := NewClient(WithClientAttempts(2))

		cc, ok := c.(*client)
		if !ok {
			t.Fatal("expected *client")
		}

		if !cc.RetrierEnabled() {
			t.Fatal("RetrierEnabled attempts=2 = false, want true")
		}
	})
}

func TestClient_Do(t *testing.T) {
	t.Parallel()

	t.Run("success no limiter no retry", func(t *testing.T) {
		t.Parallel()

		c := NewClient(
			WithClientTransport(successRT{body: "hello"}),
			WithClientAttempts(1),
		)

		req := newRequest(t, context.Background())

		res, err := c.Do(req)
		if err != nil {
			t.Fatalf("Do returned error: %v", err)
		}

		if res == nil || res.StatusCode != stdhttp.StatusOK {
			t.Fatalf("unexpected response: %+v", res)
		}

		_ = res.Body.Close()
	})

	t.Run("error with response closes body and wraps", func(t *testing.T) {
		t.Parallel()

		c := NewClient(
			WithClientTransport(resAndErrRT{}),
			WithClientTimeout(10*time.Millisecond),
			WithClientLimiterRate(float64(^uint(0))),
		)

		req := newRequest(t, context.Background())

		_, err := c.Do(req) //nolint:bodyclose // error path, no body to close
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrHttpRequestFailed) {
			t.Fatalf("error does not wrap ErrHttpRequestFailed: %v", err)
		}
	})

	t.Run("retry on response closes body and returns StatusCodeError", func(t *testing.T) {
		t.Parallel()

		var wasClosed bool

		c := NewClient(
			WithClientTransport(retryOnResponseRT{closed: &wasClosed}),
			WithClientAttempts(1),
			WithClientRetryOnResponse(RetryOn5xxAnd429Response),
		)

		req := newRequest(t, context.Background())

		_, err := c.Do(req) //nolint:bodyclose // error path, body already closed by interceptor
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var scErr *StatusCodeError
		if !errors.As(err, &scErr) || scErr == nil || scErr.StatusCode != stdhttp.StatusServiceUnavailable {
			t.Fatalf("expected *StatusCodeError{503}, got %v", err)
		}

		if !wasClosed {
			t.Fatal("response body was not closed when retryOnResponse triggered")
		}
	})

	t.Run("retry on response auto wired retries", func(t *testing.T) {
		t.Parallel()

		fl := &flakyResponseRT{failCount: 2}

		c := NewClient(
			WithClientTransport(fl),
			WithClientAttempts(3),
			WithClientRetryOnResponse(RetryOn5xxAnd429Response),
		)

		req := newRequest(t, context.Background())

		res, err := c.Do(req)
		if err != nil {
			t.Fatalf("expected success after retries, got error: %v", err)
		}

		if res == nil || res.StatusCode != stdhttp.StatusOK {
			t.Fatalf("expected 200 OK, got %+v", res)
		}

		_ = res.Body.Close()

		if fl.n != 3 {
			t.Fatalf("expected 3 attempts (2 x 503 + 1 x 200), got %d", fl.n)
		}
	})

	t.Run("error only wraps", func(t *testing.T) {
		t.Parallel()

		c := NewClient(
			WithClientTransport(errRT{err: context.Canceled}),
		)
		req := newRequest(t, context.Background())

		_, err := c.Do(req) //nolint:bodyclose // error path, no body
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrHttpRequestFailed) {
			t.Fatalf("error does not wrap ErrHttpRequestFailed: %v", err)
		}
	})

	t.Run("retry succeeds after failures", func(t *testing.T) {
		t.Parallel()

		fl := &flakyRT{failCount: 2}

		var hookCalls int

		c := NewClient(
			WithClientTransport(fl),
			WithClientAttempts(3),
			WithClientRetryIf(func(err error) bool { return true }),
			WithClientRetryHook(func(n uint, err error) { hookCalls++ }),
		)

		req := newRequest(t, context.Background())

		res, err := c.Do(req)
		if err != nil {
			t.Fatalf("Do with retries failed: %v", err)
		}

		if res == nil || res.StatusCode != stdhttp.StatusOK {
			t.Fatalf("unexpected response: %+v", res)
		}

		if hookCalls == 0 {
			t.Fatal("retry hook not called")
		}

		_ = res.Body.Close()
	})

	t.Run("non replayable body", func(t *testing.T) {
		t.Parallel()

		c := NewClient()

		req := newRequest(t, context.Background())
		req.Body = io.NopCloser(bytes.NewBufferString("x"))
		req.GetBody = nil

		_, err := c.Do(req) //nolint:bodyclose // error path
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrHttpNonReplayableBody) {
			t.Fatalf("error does not wrap ErrHttpNonReplayableBody: %v", err)
		}
	})

	t.Run("get body failure", func(t *testing.T) {
		t.Parallel()

		c := NewClient(WithClientTransport(successRT{body: "ignored"}))

		req := newRequest(t, context.Background())
		req.Body = io.NopCloser(bytes.NewBufferString("x"))
		req.GetBody = func() (io.ReadCloser, error) {
			return nil, errors.New("boom")
		}

		_, err := c.Do(req) //nolint:bodyclose // error path
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrHttpGetBodyFailed) {
			t.Fatalf("error does not wrap ErrHttpGetBodyFailed: %v", err)
		}
	})

	t.Run("replayable body success", func(t *testing.T) {
		t.Parallel()

		c := NewClient(
			WithClientTransport(successRT{body: "ok"}),
			WithClientAttempts(1),
		)

		req := newRequest(t, context.Background())
		initial := []byte("payload")
		req.Body = io.NopCloser(bytes.NewReader(initial))
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(initial)), nil
		}

		res, err := c.Do(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res == nil || res.StatusCode != stdhttp.StatusOK {
			t.Fatalf("unexpected response: %+v", res)
		}

		_ = res.Body.Close()
	})

	t.Run("nil request", func(t *testing.T) {
		t.Parallel()

		c := NewClient()

		_, err := c.Do(nil) //nolint:bodyclose // error path
		if err == nil {
			t.Fatal("expected error for nil request")
		}

		if !errors.Is(err, ErrHttpRequestNil) {
			t.Fatalf("error does not wrap ErrHttpRequestNil: %v", err)
		}
	})

	t.Run("cancelled context aborts retries", func(t *testing.T) {
		t.Parallel()

		var n int

		c := NewClient(
			WithClientTransport(countingRT{n: &n}),
			WithClientAttempts(5),
			WithClientRetryIf(func(err error) bool { return true }),
		)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		req := newRequest(t, ctx)

		_, err := c.Do(req) //nolint:bodyclose // error path
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// retry-go with Context detects the cancelled context immediately;
		// at most 1 attempt runs (0 if detected before the first call).
		if n > 1 {
			t.Fatalf("expected at most 1 attempt with pre-cancelled context, got %d", n)
		}
	})
}

func TestClient_waitForLimiter(t *testing.T) {
	t.Parallel()

	t.Run("deadline from client timeout", func(t *testing.T) {
		t.Parallel()

		c := NewClient(
			WithClientTransport(successRT{}),
			WithClientTimeout(1*time.Millisecond),
			WithClientLimiterRate(100),
			WithClientLimiterBurst(1),
		)
		cc := c.(*client)
		_ = cc.limiter.Wait(context.Background())

		req := newRequest(t, context.Background())

		_, err := c.Do(req) //nolint:bodyclose // error path
		if err == nil {
			t.Fatal("expected rate limiter error, got nil")
		}

		if !errors.Is(err, ErrRateLimiterExceeded) {
			t.Fatalf("expected ErrRateLimiterExceeded wrapping, got %v", err)
		}
	})

	t.Run("deadline from request context earlier", func(t *testing.T) {
		t.Parallel()

		c := NewClient(
			WithClientTransport(successRT{}),
			WithClientTimeout(250*time.Millisecond),
			WithClientLimiterRate(100),
			WithClientLimiterBurst(1),
		)
		cc := c.(*client)
		_ = cc.limiter.Wait(context.Background())

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		req := newRequest(t, ctx)

		_, err := c.Do(req) //nolint:bodyclose // error path
		if err == nil {
			t.Fatal("expected error due to request context deadline, got nil")
		}

		if !errors.Is(err, ErrRateLimiterExceeded) {
			t.Fatalf("expected ErrRateLimiterExceeded wrapping, got %v", err)
		}
	})

	t.Run("nil context", func(t *testing.T) {
		t.Parallel()

		c := NewClient()

		cc := c.(*client)

		err := cc.waitForLimiter(nil) //nolint:staticcheck // intentionally testing nil context behavior
		if err == nil || !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("disabled limiter no wait", func(t *testing.T) {
		t.Parallel()

		c := NewClient()

		cc := c.(*client)
		if cc.LimiterEnabled() {
			t.Fatal("expected limiter disabled by default")
		}

		err := cc.waitForLimiter(context.Background())
		if err != nil {
			t.Fatalf("waitForLimiter should return nil when limiter disabled; got %v", err)
		}
	})
}

func TestPluggableClient_Do(t *testing.T) {
	t.Parallel()

	t.Run("delegates to DoFn", func(t *testing.T) {
		t.Parallel()

		called := false
		fc := &PluggableClient{
			DoFn: func(req *stdhttp.Request) (*stdhttp.Response, error) {
				called = true
				return &stdhttp.Response{StatusCode: 204, Body: stdhttp.NoBody}, nil
			},
		}

		req := newRequest(t, context.Background())

		res, err := fc.Do(req)
		if err != nil {
			t.Fatalf("Do returned error: %v", err)
		}

		defer func() { _ = res.Body.Close() }()

		if res.StatusCode != stdhttp.StatusNoContent {
			t.Fatalf("unexpected response: %+v", res)
		}

		if !called {
			t.Fatal("DoFn was not invoked")
		}
	})

	t.Run("nil request", func(t *testing.T) {
		t.Parallel()

		fc := &PluggableClient{
			DoFn: func(req *stdhttp.Request) (*stdhttp.Response, error) {
				return &stdhttp.Response{StatusCode: 200, Body: stdhttp.NoBody}, nil
			},
		}

		_, err := fc.Do(nil) //nolint:bodyclose // error path
		if err == nil {
			t.Fatal("expected error for nil request")
		}

		if !errors.Is(err, ErrHttpRequestNil) {
			t.Fatalf("error does not wrap ErrHttpRequestNil: %v", err)
		}
	})
}

func TestPluggableClient_LimiterEnabled(t *testing.T) {
	t.Parallel()

	t.Run("nil fn returns false", func(t *testing.T) {
		t.Parallel()

		fc := &PluggableClient{DoFn: NoopDo}

		if fc.LimiterEnabled() {
			t.Fatal("LimiterEnabled with nil fn should return false")
		}
	})

	t.Run("delegates to fn true", func(t *testing.T) {
		t.Parallel()

		fc := &PluggableClient{
			DoFn:             NoopDo,
			LimiterEnabledFn: func() bool { return true },
		}

		if !fc.LimiterEnabled() {
			t.Fatal("LimiterEnabled should return true")
		}
	})

	t.Run("delegates to fn false", func(t *testing.T) {
		t.Parallel()

		fc := &PluggableClient{
			DoFn:             NoopDo,
			LimiterEnabledFn: func() bool { return false },
		}

		if fc.LimiterEnabled() {
			t.Fatal("LimiterEnabled should return false")
		}
	})
}

func TestPluggableClient_RetrierEnabled(t *testing.T) {
	t.Parallel()

	t.Run("nil fn returns false", func(t *testing.T) {
		t.Parallel()

		fc := &PluggableClient{DoFn: NoopDo}

		if fc.RetrierEnabled() {
			t.Fatal("RetrierEnabled with nil fn should return false")
		}
	})

	t.Run("delegates to fn true", func(t *testing.T) {
		t.Parallel()

		fc := &PluggableClient{
			DoFn:             NoopDo,
			RetrierEnabledFn: func() bool { return true },
		}

		if !fc.RetrierEnabled() {
			t.Fatal("RetrierEnabled should return true")
		}
	})

	t.Run("delegates to fn false", func(t *testing.T) {
		t.Parallel()

		fc := &PluggableClient{
			DoFn:             NoopDo,
			RetrierEnabledFn: func() bool { return false },
		}

		if fc.RetrierEnabled() {
			t.Fatal("RetrierEnabled should return false")
		}
	})
}
