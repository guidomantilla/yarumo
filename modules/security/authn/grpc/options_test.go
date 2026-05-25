package grpc_test

import (
	"testing"

	authngrpc "github.com/guidomantilla/yarumo/security/authn/grpc"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()

		opts := authngrpc.NewOptions()
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})

	t.Run("with overrides", func(t *testing.T) {
		t.Parallel()

		opts := authngrpc.NewOptions(
			authngrpc.WithMetadataKey("X-Auth"),
			authngrpc.WithScheme("Token"),
		)
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}

func TestWithMetadataKey(t *testing.T) {
	t.Parallel()

	t.Run("empty ignored", func(t *testing.T) {
		t.Parallel()

		opts := authngrpc.NewOptions(authngrpc.WithMetadataKey(""))
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}

func TestWithScheme(t *testing.T) {
	t.Parallel()

	t.Run("empty ignored", func(t *testing.T) {
		t.Parallel()

		opts := authngrpc.NewOptions(authngrpc.WithScheme(""))
		if opts == nil {
			t.Fatal("NewOptions returned nil")
		}
	})
}
