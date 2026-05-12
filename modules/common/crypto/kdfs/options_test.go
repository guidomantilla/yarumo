package kdfs

import (
	"testing"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("returns defaults when no options provided", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()

		if opts.deriveFn == nil {
			t.Fatal("expected default deriveFn (hkdfDerive)")
		}

		if opts.pbkdf2Params != nil {
			t.Fatal("expected pbkdf2Params to be nil by default")
		}

		if opts.scryptParams != nil {
			t.Fatal("expected scryptParams to be nil by default")
		}
	})
}

func TestWithDeriveFn(t *testing.T) {
	t.Parallel()

	t.Run("sets custom derive function", func(t *testing.T) {
		t.Parallel()

		called := false

		custom := func(method *Method, secret, salt, info ctypes.Bytes, length int) (ctypes.Bytes, error) {
			called = true

			return nil, nil
		}

		opts := NewOptions(WithDeriveFn(custom))

		_, _ = opts.deriveFn(nil, nil, nil, nil, 0)

		if !called {
			t.Fatal("expected custom deriveFn to be called")
		}
	})

	t.Run("ignores nil derive function", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDeriveFn(nil))

		if opts.deriveFn == nil {
			t.Fatal("expected default deriveFn preserved")
		}
	})
}

func TestWithPbkdf2Iterations(t *testing.T) {
	t.Parallel()

	t.Run("installs pbkdf2 params and derive function", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPbkdf2Iterations(1000))

		if opts.pbkdf2Params == nil {
			t.Fatal("expected pbkdf2Params to be set")
		}

		if opts.pbkdf2Params.iterations != 1000 {
			t.Fatalf("expected iterations=1000, got %d", opts.pbkdf2Params.iterations)
		}
	})

	t.Run("ignores zero iterations", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPbkdf2Iterations(0))

		if opts.pbkdf2Params != nil {
			t.Fatal("expected pbkdf2Params to remain nil")
		}
	})

	t.Run("ignores negative iterations", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPbkdf2Iterations(-1))

		if opts.pbkdf2Params != nil {
			t.Fatal("expected pbkdf2Params to remain nil")
		}
	})
}

func TestWithScryptParams(t *testing.T) {
	t.Parallel()

	t.Run("installs scrypt params and derive function", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithScryptParams(1024, 8, 1))

		if opts.scryptParams == nil {
			t.Fatal("expected scryptParams to be set")
		}

		if opts.scryptParams.n != 1024 {
			t.Fatalf("expected n=1024, got %d", opts.scryptParams.n)
		}

		if opts.scryptParams.r != 8 {
			t.Fatalf("expected r=8, got %d", opts.scryptParams.r)
		}

		if opts.scryptParams.p != 1 {
			t.Fatalf("expected p=1, got %d", opts.scryptParams.p)
		}
	})

	t.Run("ignores invalid n", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithScryptParams(1, 8, 1))

		if opts.scryptParams != nil {
			t.Fatal("expected scryptParams to remain nil")
		}
	})

	t.Run("ignores invalid r", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithScryptParams(1024, 0, 1))

		if opts.scryptParams != nil {
			t.Fatal("expected scryptParams to remain nil")
		}
	})

	t.Run("ignores invalid p", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithScryptParams(1024, 8, 0))

		if opts.scryptParams != nil {
			t.Fatal("expected scryptParams to remain nil")
		}
	})
}
