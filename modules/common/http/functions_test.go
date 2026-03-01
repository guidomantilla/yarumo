package http

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestEnabledLimiter(t *testing.T) {
	t.Parallel()

	t.Run("returns true", func(t *testing.T) {
		t.Parallel()

		got := EnabledLimiter()
		if !got {
			t.Fatalf("EnabledLimiter() = %v, want true", got)
		}
	})
}

func TestDisabledLimiter(t *testing.T) {
	t.Parallel()

	t.Run("returns false", func(t *testing.T) {
		t.Parallel()

		got := DisabledLimiter()
		if got {
			t.Fatalf("DisabledLimiter() = %v, want false", got)
		}
	})
}

func TestEnabledRetrier(t *testing.T) {
	t.Parallel()

	t.Run("returns true", func(t *testing.T) {
		t.Parallel()

		got := EnabledRetrier()
		if !got {
			t.Fatalf("EnabledRetrier() = %v, want true", got)
		}
	})
}

func TestDisabledRetrier(t *testing.T) {
	t.Parallel()

	t.Run("returns false", func(t *testing.T) {
		t.Parallel()

		got := DisabledRetrier()
		if got {
			t.Fatalf("DisabledRetrier() = %v, want false", got)
		}
	})
}

func TestNoopRetryOnResponse(t *testing.T) {
	t.Parallel()

	t.Run("nil response returns false", func(t *testing.T) {
		t.Parallel()

		got := NoopRetryOnResponse(nil)
		if got {
			t.Fatalf("NoopRetryOnResponse(nil) = %v, want false", got)
		}
	})

	t.Run("non-nil response returns false", func(t *testing.T) {
		t.Parallel()

		got := NoopRetryOnResponse(&http.Response{StatusCode: http.StatusOK})
		if got {
			t.Fatalf("NoopRetryOnResponse(res) = %v, want false", got)
		}
	})
}

func TestNoopRetryIf(t *testing.T) {
	t.Parallel()

	t.Run("nil error returns false", func(t *testing.T) {
		t.Parallel()

		got := NoopRetryIf(nil)
		if got {
			t.Fatalf("NoopRetryIf(nil) = %v, want false", got)
		}
	})

	t.Run("non-nil error returns false", func(t *testing.T) {
		t.Parallel()

		got := NoopRetryIf(errors.New("boom"))
		if got {
			t.Fatalf("NoopRetryIf(non-nil) = %v, want false", got)
		}
	})
}

func TestNoopRetryHook(t *testing.T) {
	t.Parallel()

	t.Run("does not panic", func(t *testing.T) {
		t.Parallel()

		NoopRetryHook(0, nil)
		NoopRetryHook(3, errors.New("x"))
	})
}

func TestRetryOn5xxAnd429Response(t *testing.T) {
	t.Parallel()

	t.Run("nil response returns false", func(t *testing.T) {
		t.Parallel()

		got := RetryOn5xxAnd429Response(nil)
		if got {
			t.Fatalf("RetryOn5xxAnd429Response(nil) = %v, want false", got)
		}
	})

	t.Run("200 returns false", func(t *testing.T) {
		t.Parallel()

		got := RetryOn5xxAnd429Response(&http.Response{StatusCode: http.StatusOK})
		if got {
			t.Fatalf("RetryOn5xxAnd429Response(200) = %v, want false", got)
		}
	})

	t.Run("400 returns false", func(t *testing.T) {
		t.Parallel()

		got := RetryOn5xxAnd429Response(&http.Response{StatusCode: http.StatusBadRequest})
		if got {
			t.Fatalf("RetryOn5xxAnd429Response(400) = %v, want false", got)
		}
	})

	t.Run("499 returns false", func(t *testing.T) {
		t.Parallel()

		got := RetryOn5xxAnd429Response(&http.Response{StatusCode: 499})
		if got {
			t.Fatalf("RetryOn5xxAnd429Response(499) = %v, want false", got)
		}
	})

	t.Run("429 returns true", func(t *testing.T) {
		t.Parallel()

		got := RetryOn5xxAnd429Response(&http.Response{StatusCode: http.StatusTooManyRequests})
		if !got {
			t.Fatalf("RetryOn5xxAnd429Response(429) = %v, want true", got)
		}
	})

	t.Run("500 returns true", func(t *testing.T) {
		t.Parallel()

		got := RetryOn5xxAnd429Response(&http.Response{StatusCode: http.StatusInternalServerError})
		if !got {
			t.Fatalf("RetryOn5xxAnd429Response(500) = %v, want true", got)
		}
	})

	t.Run("503 returns true", func(t *testing.T) {
		t.Parallel()

		got := RetryOn5xxAnd429Response(&http.Response{StatusCode: http.StatusServiceUnavailable})
		if !got {
			t.Fatalf("RetryOn5xxAnd429Response(503) = %v, want true", got)
		}
	})

	t.Run("599 returns true", func(t *testing.T) {
		t.Parallel()

		got := RetryOn5xxAnd429Response(&http.Response{StatusCode: 599})
		if !got {
			t.Fatalf("RetryOn5xxAnd429Response(599) = %v, want true", got)
		}
	})
}

// testNetErr is a custom net.Error implementation for testing.
type testNetErr struct{ timeout bool }

func (e testNetErr) Error() string { return "net" }

func (e testNetErr) Timeout() bool { return e.timeout }

func (e testNetErr) Temporary() bool { return false }

func TestRetryIfHttpError(t *testing.T) {
	t.Parallel()

	t.Run("nil error returns false", func(t *testing.T) {
		t.Parallel()

		got := RetryIfHttpError(nil)
		if got {
			t.Fatalf("RetryIfHttpError(nil) = %v, want false", got)
		}
	})

	t.Run("StatusCodeError returns true", func(t *testing.T) {
		t.Parallel()

		got := RetryIfHttpError(&StatusCodeError{StatusCode: 503})
		if !got {
			t.Fatalf("RetryIfHttpError(StatusCodeError) = %v, want true", got)
		}
	})

	t.Run("net error with timeout returns true", func(t *testing.T) {
		t.Parallel()

		got := RetryIfHttpError(testNetErr{timeout: true})
		if !got {
			t.Fatalf("RetryIfHttpError(net.Error timeout) = %v, want true", got)
		}
	})

	t.Run("net error without timeout returns false", func(t *testing.T) {
		t.Parallel()

		got := RetryIfHttpError(testNetErr{timeout: false})
		if got {
			t.Fatalf("RetryIfHttpError(net.Error no-timeout) = %v, want false", got)
		}
	})

	t.Run("plain error returns false", func(t *testing.T) {
		t.Parallel()

		got := RetryIfHttpError(errors.New("plain"))
		if got {
			t.Fatalf("RetryIfHttpError(plain error) = %v, want false", got)
		}
	})
}

func TestNoopDo(t *testing.T) {
	t.Parallel()

	t.Run("returns 204 no content", func(t *testing.T) {
		t.Parallel()

		res, err := NoopDo(nil)
		if err != nil {
			t.Fatalf("NoopDo returned error: %v", err)
		}

		if res == nil || res.StatusCode != http.StatusNoContent {
			t.Fatalf("expected 204 No Content, got %+v", res)
		}

		_ = res.Body.Close()
	})
}

func TestErrorDo(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrHttpRequestFailed", func(t *testing.T) {
		t.Parallel()

		res, err := ErrorDo(nil) //nolint:bodyclose // ErrorDo always returns nil response
		if res != nil {
			t.Fatalf("ErrorDo returned non-nil response: %+v", res)
		}

		if err == nil || !errors.Is(err, ErrHttpRequestFailed) {
			t.Fatalf("ErrorDo error = %v, want wrapping ErrHttpRequestFailed", err)
		}
	})
}

func TestDo(t *testing.T) {
	t.Parallel()

	t.Run("delegates to DefaultClient", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost:1/nonexistent", nil)
		if err != nil {
			t.Fatalf("new request: %v", err)
		}

		_, doErr := Do(req) //nolint:bodyclose,gosec // error path; test uses hardcoded URL
		if doErr == nil {
			t.Fatal("expected error from Do with unreachable host")
		}

		if !errors.Is(doErr, ErrHttpRequestFailed) {
			t.Fatalf("expected ErrHttpRequestFailed wrapping, got %v", doErr)
		}
	})
}
