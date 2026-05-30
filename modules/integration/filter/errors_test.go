package filter

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("includes type prefix and joined causes", func(t *testing.T) {
		t.Parallel()

		err := ErrFilter(ErrPredicateFailed)

		msg := err.Error()
		if !strings.HasPrefix(msg, "filter "+FilterType) {
			t.Fatalf("expected prefix %q, got %q", "filter "+FilterType, msg)
		}

		if !strings.Contains(msg, ErrPredicateFailed.Error()) {
			t.Fatalf("expected cause %q in message, got %q", ErrPredicateFailed.Error(), msg)
		}

		if !strings.Contains(msg, ErrFilterFailed.Error()) {
			t.Fatalf("expected sentinel %q in message, got %q", ErrFilterFailed.Error(), msg)
		}
	})

	t.Run("ErrFilter joins all causes with ErrFilterFailed", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("custom failure")

		err := ErrFilter(ErrPredicateFailed, boom)
		if !errors.Is(err, ErrFilterFailed) {
			t.Fatal("expected ErrFilterFailed in chain")
		}

		if !errors.Is(err, ErrPredicateFailed) {
			t.Fatal("expected ErrPredicateFailed in chain")
		}

		if !errors.Is(err, boom) {
			t.Fatal("expected origin error in chain")
		}
	})
}
