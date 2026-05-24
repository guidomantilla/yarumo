package retry

import (
	"errors"
	"testing"
)

func TestAlwaysRetry(t *testing.T) {
	t.Parallel()

	t.Run("returns true for a non-nil error", func(t *testing.T) {
		t.Parallel()

		if !AlwaysRetry(errors.New("any")) {
			t.Fatal("AlwaysRetry should return true for any error")
		}
	})

	t.Run("returns true for nil (callers never invoke it with nil)", func(t *testing.T) {
		t.Parallel()

		if !AlwaysRetry(nil) {
			t.Fatal("AlwaysRetry should return true even for nil")
		}
	})
}

func TestNoopOnRetry(t *testing.T) {
	t.Parallel()

	t.Run("does not panic on any input", func(t *testing.T) {
		t.Parallel()

		NoopOnRetry(0, errors.New("any"))
		NoopOnRetry(99, nil)
	})
}
