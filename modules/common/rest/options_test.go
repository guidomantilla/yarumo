package rest

import (
	"net/http"
	"testing"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("applies default DoFn", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.doFn == nil {
			t.Fatal("default doFn should not be nil")
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

func TestWithDoFn(t *testing.T) {
	t.Parallel()

	t.Run("overrides DoFn", func(t *testing.T) {
		t.Parallel()

		called := false
		do := func(_ *http.Request) (*http.Response, error) {
			called = true

			return makeResp(200, []byte("{}"), map[string]string{"Content-Type": applicationJSON}), nil
		}

		opts := NewOptions(WithDoFn(do))
		resp, err := opts.doFn(&http.Request{})

		if resp != nil {
			_ = resp.Body.Close()
		}

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !called {
			t.Fatal("WithDoFn did not override doFn")
		}
	})

	t.Run("ignores nil DoFn", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDoFn(nil))
		if opts.doFn == nil {
			t.Fatal("doFn should remain default when nil is passed")
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
