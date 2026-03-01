package rsapss

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

		if opts.signFn == nil {
			t.Fatal("expected default signFn")
		}

		if opts.verifyFn == nil {
			t.Fatal("expected default verifyFn")
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

func TestWithSignFn(t *testing.T) {
	t.Parallel()

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithSignFn(nil))

		if opts.signFn == nil {
			t.Fatal("expected default signFn preserved")
		}
	})
}

func TestWithVerifyFn(t *testing.T) {
	t.Parallel()

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithVerifyFn(nil))

		if opts.verifyFn == nil {
			t.Fatal("expected default verifyFn preserved")
		}
	})
}
