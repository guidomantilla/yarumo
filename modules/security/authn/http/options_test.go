package http_test

import (
	"net/http"
	"testing"

	authnhttp "github.com/guidomantilla/yarumo/security/authn/http"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()

		opts := authnhttp.NewOptions()
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})

	t.Run("with overrides", func(t *testing.T) {
		t.Parallel()

		opts := authnhttp.NewOptions(
			authnhttp.WithHeaderName("X-Auth"),
			authnhttp.WithScheme("Token"),
			authnhttp.WithErrorHandler(func(_ http.ResponseWriter, _ *http.Request, _ error) {}),
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

		opts := authnhttp.NewOptions(authnhttp.WithHeaderName(""))
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}

func TestWithScheme(t *testing.T) {
	t.Parallel()

	t.Run("empty ignored", func(t *testing.T) {
		t.Parallel()

		opts := authnhttp.NewOptions(authnhttp.WithScheme(""))
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}

func TestWithErrorHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil ignored", func(t *testing.T) {
		t.Parallel()

		opts := authnhttp.NewOptions(authnhttp.WithErrorHandler(nil))
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}
