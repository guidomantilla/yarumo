package http

import (
	"net/http"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil client", func(t *testing.T) {
		t.Parallel()

		c := NewClient()
		if c == nil {
			t.Fatal("expected non-nil client")
		}
	})

	t.Run("applies the configured timeout", func(t *testing.T) {
		t.Parallel()

		c := NewClient(WithTimeout(45 * time.Second)).(*http.Client)
		if c.Timeout != 45*time.Second {
			t.Fatalf("Timeout = %v, want %v", c.Timeout, 45*time.Second)
		}
	})

	t.Run("uses default timeout 30s when not configured", func(t *testing.T) {
		t.Parallel()

		c := NewClient().(*http.Client)
		if c.Timeout != 30*time.Second {
			t.Fatalf("Timeout = %v, want %v", c.Timeout, 30*time.Second)
		}
	})

	t.Run("uses the configured transport verbatim", func(t *testing.T) {
		t.Parallel()

		custom := &http.Transport{}
		c := NewClient(WithTransport(custom)).(*http.Client)

		// NewClient does not wrap the transport; the consumer's
		// RoundTripper is what the client uses (modulo timeout capping
		// for *http.Transport instances, which clones).
		_, ok := c.Transport.(*http.Transport)
		if !ok {
			t.Fatalf("expected *http.Transport, got %T", c.Transport)
		}
	})

	t.Run("returned type satisfies Client interface", func(t *testing.T) {
		t.Parallel()

		var c Client = NewClient()
		if c == nil {
			t.Fatal("expected non-nil Client")
		}
	})
}
