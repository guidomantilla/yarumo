package grpc

import (
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
			WithMetadataKey("X-Auth"),
			WithScheme("Token"),
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

		opts := NewOptions(WithMetadataKey(""))
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
