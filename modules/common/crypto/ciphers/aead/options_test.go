package aead

import (
	"testing"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("returns defaults", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()

		if opts.keyFn == nil {
			t.Fatal("expected default keyFn")
		}

		if opts.encryptFn == nil {
			t.Fatal("expected default encryptFn")
		}

		if opts.decryptFn == nil {
			t.Fatal("expected default decryptFn")
		}
	})
}

func TestWithKeyFn(t *testing.T) {
	t.Parallel()

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithKeyFn(nil))

		if opts.keyFn == nil {
			t.Fatal("expected default keyFn preserved")
		}
	})
}

func TestWithEncryptFn(t *testing.T) {
	t.Parallel()

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithEncryptFn(nil))

		if opts.encryptFn == nil {
			t.Fatal("expected default encryptFn preserved")
		}
	})
}

func TestWithDecryptFn(t *testing.T) {
	t.Parallel()

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDecryptFn(nil))

		if opts.decryptFn == nil {
			t.Fatal("expected default decryptFn preserved")
		}
	})
}
