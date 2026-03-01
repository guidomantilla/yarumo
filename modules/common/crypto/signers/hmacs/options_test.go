package hmacs

import (
	"testing"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("returns defaults when no options provided", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()

		if opts.keyFn == nil {
			t.Fatal("expected default keyFn")
		}

		if opts.digestFn == nil {
			t.Fatal("expected default digestFn")
		}

		if opts.validateFn == nil {
			t.Fatal("expected default validateFn")
		}
	})
}

func TestWithKeyFn(t *testing.T) {
	t.Parallel()

	t.Run("sets custom key function", func(t *testing.T) {
		t.Parallel()

		called := false
		custom := func(method *Method) (ctypes.Bytes, error) {
			called = true
			return nil, nil
		}

		opts := NewOptions(WithKeyFn(custom))

		_, _ = opts.keyFn(nil)

		if !called {
			t.Fatal("expected custom keyFn")
		}
	})

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithKeyFn(nil))

		if opts.keyFn == nil {
			t.Fatal("expected default keyFn preserved")
		}
	})
}

func TestWithDigestFn(t *testing.T) {
	t.Parallel()

	t.Run("sets custom digest function", func(t *testing.T) {
		t.Parallel()

		called := false
		custom := func(method *Method, key ctypes.Bytes, data ctypes.Bytes) (ctypes.Bytes, error) {
			called = true
			return nil, nil
		}

		opts := NewOptions(WithDigestFn(custom))

		_, _ = opts.digestFn(nil, nil, nil)

		if !called {
			t.Fatal("expected custom digestFn")
		}
	})

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDigestFn(nil))

		if opts.digestFn == nil {
			t.Fatal("expected default digestFn preserved")
		}
	})
}

func TestWithValidateFn(t *testing.T) {
	t.Parallel()

	t.Run("sets custom validate function", func(t *testing.T) {
		t.Parallel()

		called := false
		custom := func(method *Method, key ctypes.Bytes, sig ctypes.Bytes, data ctypes.Bytes) (bool, error) {
			called = true
			return true, nil
		}

		opts := NewOptions(WithValidateFn(custom))

		_, _ = opts.validateFn(nil, nil, nil, nil)

		if !called {
			t.Fatal("expected custom validateFn")
		}
	})

	t.Run("ignores nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithValidateFn(nil))

		if opts.validateFn == nil {
			t.Fatal("expected default validateFn preserved")
		}
	})
}
