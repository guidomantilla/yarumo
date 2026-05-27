package http

import (
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestRoundTripperFn_RoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("delegates to the wrapped function", func(t *testing.T) {
		t.Parallel()

		want := &http.Response{StatusCode: http.StatusTeapot, Body: http.NoBody}
		rt := RoundTripperFn(func(_ *http.Request) (*http.Response, error) {
			return want, nil
		})

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		got, err := rt.RoundTrip(req)
		if err != nil {
			t.Fatalf("RoundTrip: %v", err)
		}

		if got != want {
			t.Fatal("expected wrapped function's response to be returned")
		}
	})

	t.Run("propagates the function's error", func(t *testing.T) {
		t.Parallel()

		want := errors.New("boom")
		rt := RoundTripperFn(func(_ *http.Request) (*http.Response, error) {
			return nil, want
		})

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		_, err := rt.RoundTrip(req)
		if !errors.Is(err, want) {
			t.Fatalf("expected wrapped function's error, got %v", err)
		}
	})

	t.Run("passes the original request through unchanged", func(t *testing.T) {
		t.Parallel()

		var seen atomic.Pointer[http.Request]
		rt := RoundTripperFn(func(req *http.Request) (*http.Response, error) {
			seen.Store(req)
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		})

		req, _ := http.NewRequest(http.MethodPost, "http://example.com/x", nil)
		_, _ = rt.RoundTrip(req)

		if seen.Load() != req {
			t.Fatal("expected the same request to be passed to the wrapped function")
		}
	})

	t.Run("composes inside an *http.Client", func(t *testing.T) {
		t.Parallel()

		client := &http.Client{Transport: RoundTripperFn(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody}, nil
		})}

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("client.Do: %v", err)
		}

		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("StatusCode = %d, want %d", res.StatusCode, http.StatusNoContent)
		}
	})
}
