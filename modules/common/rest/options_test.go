package rest

import (
	"net/http"
	"testing"
)

// doClient adapts a plain Do function to chttp.Client for tests.
type doClient func(*http.Request) (*http.Response, error)

func (f doClient) Do(req *http.Request) (*http.Response, error) { return f(req) }

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("applies default client", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.client == nil {
			t.Fatal("default client should not be nil")
		}
	})

	t.Run("applies default maxResponseSize", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.maxResponseSize != defaultMaxResponseSize {
			t.Fatalf("expected default maxResponseSize=%d, got %d", defaultMaxResponseSize, opts.maxResponseSize)
		}
	})
}

func TestWithClient(t *testing.T) {
	t.Parallel()

	t.Run("overrides client", func(t *testing.T) {
		t.Parallel()

		called := false
		do := func(_ *http.Request) (*http.Response, error) {
			called = true

			return makeResp(200, []byte("{}"), map[string]string{"Content-Type": applicationJSON}), nil
		}

		opts := NewOptions(WithClient(doClient(do)))
		resp, err := opts.client.Do(&http.Request{})

		if resp != nil {
			_ = resp.Body.Close()
		}

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !called {
			t.Fatal("WithClient did not override client")
		}
	})

	t.Run("ignores nil client", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithClient(nil))
		if opts.client == nil {
			t.Fatal("client should remain default when nil is passed")
		}
	})
}

func TestWithMaxResponseSize(t *testing.T) {
	t.Parallel()

	t.Run("overrides default max response size", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxResponseSize(1024))
		if opts.maxResponseSize != 1024 {
			t.Fatalf("expected maxResponseSize=1024, got %d", opts.maxResponseSize)
		}
	})

	t.Run("ignores zero value", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxResponseSize(0))
		if opts.maxResponseSize != defaultMaxResponseSize {
			t.Fatalf("expected default maxResponseSize, got %d", opts.maxResponseSize)
		}
	})

	t.Run("ignores negative value", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxResponseSize(-1))
		if opts.maxResponseSize != defaultMaxResponseSize {
			t.Fatalf("expected default maxResponseSize, got %d", opts.maxResponseSize)
		}
	})
}
