package breaker

import (
	"net/http"
	"testing"
)

func TestFailOn5xxAnd429(t *testing.T) {
	t.Parallel()

	t.Run("returns true for 5xx", func(t *testing.T) {
		t.Parallel()

		for _, code := range []int{500, 502, 503, 504, 599} {
			if !FailOn5xxAnd429(&http.Response{StatusCode: code}) {
				t.Fatalf("expected fail for %d", code)
			}
		}
	})

	t.Run("returns true for 429", func(t *testing.T) {
		t.Parallel()

		if !FailOn5xxAnd429(&http.Response{StatusCode: 429}) {
			t.Fatal("expected fail for 429")
		}
	})

	t.Run("returns false for 2xx, 3xx, 4xx (except 429)", func(t *testing.T) {
		t.Parallel()

		for _, code := range []int{200, 201, 204, 301, 304, 400, 401, 404, 418} {
			if FailOn5xxAnd429(&http.Response{StatusCode: code}) {
				t.Fatalf("did not expect fail for %d", code)
			}
		}
	})

	t.Run("returns false for nil response", func(t *testing.T) {
		t.Parallel()

		if FailOn5xxAnd429(nil) {
			t.Fatal("expected false for nil response")
		}
	})
}

func TestNoopFailOnResponse(t *testing.T) {
	t.Parallel()

	if NoopFailOnResponse(&http.Response{StatusCode: 500}) {
		t.Fatal("expected false")
	}
}
