package http

import (
	"net/http"
	"testing"
	"time"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults are safe for production", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()

		if opts.timeout != 30*time.Second {
			t.Fatalf("timeout = %v, want %v", opts.timeout, 30*time.Second)
		}

		if opts.transport == nil {
			t.Fatal("expected non-nil transport")
		}
	})

	t.Run("applies given options", func(t *testing.T) {
		t.Parallel()

		custom := &http.Transport{TLSHandshakeTimeout: 10 * time.Second}

		opts := NewOptions(
			WithTimeout(45*time.Second),
			WithTransport(custom),
		)

		if opts.timeout != 45*time.Second {
			t.Fatalf("timeout = %v, want %v", opts.timeout, 45*time.Second)
		}
	})

	t.Run("caps *http.Transport internal timeouts to total timeout", func(t *testing.T) {
		t.Parallel()

		oversized := &http.Transport{
			TLSHandshakeTimeout:   60 * time.Second,
			ResponseHeaderTimeout: 90 * time.Second,
			ExpectContinueTimeout: 45 * time.Second,
		}

		opts := NewOptions(WithTimeout(30*time.Second), WithTransport(oversized))

		got, ok := opts.transport.(*http.Transport)
		if !ok {
			t.Fatalf("expected *http.Transport, got %T", opts.transport)
		}

		if got.TLSHandshakeTimeout != 30*time.Second {
			t.Fatalf("TLSHandshakeTimeout = %v, want %v (capped)", got.TLSHandshakeTimeout, 30*time.Second)
		}

		if got.ResponseHeaderTimeout != 30*time.Second {
			t.Fatalf("ResponseHeaderTimeout = %v, want %v (capped)", got.ResponseHeaderTimeout, 30*time.Second)
		}

		if got.ExpectContinueTimeout != 30*time.Second {
			t.Fatalf("ExpectContinueTimeout = %v, want %v (capped)", got.ExpectContinueTimeout, 30*time.Second)
		}

		// Original instance must not be mutated.
		if oversized.TLSHandshakeTimeout != 60*time.Second {
			t.Fatal("original transport was mutated")
		}
	})

	t.Run("does not cap zero internal timeouts", func(t *testing.T) {
		t.Parallel()

		zeros := &http.Transport{}

		opts := NewOptions(WithTimeout(10*time.Second), WithTransport(zeros))

		got, ok := opts.transport.(*http.Transport)
		if !ok {
			t.Fatalf("expected *http.Transport, got %T", opts.transport)
		}

		if got.TLSHandshakeTimeout != 0 {
			t.Fatalf("TLSHandshakeTimeout = %v, want 0 (unchanged)", got.TLSHandshakeTimeout)
		}
	})
}

func TestWithTimeout(t *testing.T) {
	t.Parallel()

	t.Run("sets positive timeout", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTimeout(60 * time.Second))
		if opts.timeout != 60*time.Second {
			t.Fatalf("timeout = %v, want %v", opts.timeout, 60*time.Second)
		}
	})

	t.Run("ignores zero", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTimeout(0))
		if opts.timeout != 30*time.Second {
			t.Fatalf("timeout = %v, want default %v", opts.timeout, 30*time.Second)
		}
	})

	t.Run("ignores negative", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTimeout(-5 * time.Second))
		if opts.timeout != 30*time.Second {
			t.Fatalf("timeout = %v, want default %v", opts.timeout, 30*time.Second)
		}
	})
}

func TestWithTransport(t *testing.T) {
	t.Parallel()

	t.Run("sets non-nil transport", func(t *testing.T) {
		t.Parallel()

		custom := &http.Transport{}
		opts := NewOptions(WithTransport(custom))

		if opts.transport == nil {
			t.Fatal("expected non-nil transport")
		}
	})

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTransport(nil))
		if opts.transport == nil {
			t.Fatal("expected default transport, got nil")
		}
	})
}
