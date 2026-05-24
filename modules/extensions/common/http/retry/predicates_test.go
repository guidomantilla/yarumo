package retry

import (
	"errors"
	"net/http"
	"testing"
)

func TestRetryOn5xxAnd429(t *testing.T) {
	t.Parallel()

	t.Run("returns true for 5xx", func(t *testing.T) {
		t.Parallel()

		for _, code := range []int{500, 502, 503, 504, 599} {
			if !RetryOn5xxAnd429(&http.Response{StatusCode: code}) {
				t.Fatalf("expected retry for %d", code)
			}
		}
	})

	t.Run("returns true for 429", func(t *testing.T) {
		t.Parallel()

		if !RetryOn5xxAnd429(&http.Response{StatusCode: 429}) {
			t.Fatal("expected retry for 429")
		}
	})

	t.Run("returns false for 2xx, 3xx, 4xx (except 429)", func(t *testing.T) {
		t.Parallel()

		for _, code := range []int{200, 201, 204, 301, 304, 400, 401, 404, 418} {
			if RetryOn5xxAnd429(&http.Response{StatusCode: code}) {
				t.Fatalf("did not expect retry for %d", code)
			}
		}
	})

	t.Run("returns false for nil response", func(t *testing.T) {
		t.Parallel()

		if RetryOn5xxAnd429(nil) {
			t.Fatal("expected false for nil response")
		}
	})
}

func TestRetryIfHttpError(t *testing.T) {
	t.Parallel()

	t.Run("returns true for *StatusCodeError", func(t *testing.T) {
		t.Parallel()

		err := &StatusCodeError{StatusCode: 500}
		if !RetryIfHttpError(err) {
			t.Fatal("expected true for *StatusCodeError")
		}
	})

	t.Run("returns true for wrapped *StatusCodeError", func(t *testing.T) {
		t.Parallel()

		err := errors.Join(errors.New("boom"), &StatusCodeError{StatusCode: 502})
		if !RetryIfHttpError(err) {
			t.Fatal("expected true for wrapped *StatusCodeError")
		}
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		t.Parallel()

		err := errors.New("network failed")
		if RetryIfHttpError(err) {
			t.Fatal("expected false for non-StatusCodeError")
		}
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		t.Parallel()

		if RetryIfHttpError(nil) {
			t.Fatal("expected false for nil error")
		}
	})
}

func TestNoopHelpers(t *testing.T) {
	t.Parallel()

	t.Run("NoopRetryOnResponse returns false", func(t *testing.T) {
		t.Parallel()

		if NoopRetryOnResponse(&http.Response{StatusCode: 500}) {
			t.Fatal("expected false")
		}
	})

	t.Run("NoopRetryIf returns false", func(t *testing.T) {
		t.Parallel()

		if NoopRetryIf(errors.New("anything")) {
			t.Fatal("expected false")
		}
	})

	t.Run("NoopRetryHook does not panic", func(t *testing.T) {
		t.Parallel()

		NoopRetryHook(0, errors.New("anything"))
	})
}
