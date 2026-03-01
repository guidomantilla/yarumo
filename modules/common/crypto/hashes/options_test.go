package hashes

import (
	"crypto"
	"testing"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("returns defaults when no options provided", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()

		if opts == nil {
			t.Fatal("expected non-nil options")
		}

		if opts.hashFn == nil {
			t.Fatal("expected default hashFn")
		}
	})

	t.Run("applies provided options", func(t *testing.T) {
		t.Parallel()

		called := false
		custom := func(hash crypto.Hash, data ctypes.Bytes) ctypes.Bytes {
			called = true
			return nil
		}

		opts := NewOptions(WithHashFn(custom))

		opts.hashFn(crypto.SHA256, nil)

		if !called {
			t.Fatal("expected custom hashFn to be applied")
		}
	})
}

func TestWithHashFn(t *testing.T) {
	t.Parallel()

	t.Run("sets custom hash function", func(t *testing.T) {
		t.Parallel()

		called := false
		custom := func(hash crypto.Hash, data ctypes.Bytes) ctypes.Bytes {
			called = true
			return nil
		}

		opts := NewOptions(WithHashFn(custom))

		opts.hashFn(crypto.SHA256, nil)

		if !called {
			t.Fatal("expected custom hashFn to be set")
		}
	})

	t.Run("ignores nil hash function", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithHashFn(nil))

		if opts.hashFn == nil {
			t.Fatal("expected default hashFn to be preserved when nil provided")
		}
	})
}
