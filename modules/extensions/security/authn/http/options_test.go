package http

import (
	"net/http"
	"testing"

)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})

	t.Run("with overrides", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(
			WithHeaderName("X-Auth"),
			WithScheme("Token"),
			WithErrorHandler(func(_ http.ResponseWriter, _ *http.Request, _ error) {}),
		)
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}

func TestWithHeaderName(t *testing.T) {
	t.Parallel()

	t.Run("empty ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithHeaderName(""))
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}

func TestWithScheme(t *testing.T) {
	t.Parallel()

	t.Run("empty ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithScheme(""))
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}

func TestWithErrorHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithErrorHandler(nil))
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}
