package headerfilter

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("includes type prefix and joined causes", func(t *testing.T) {
		t.Parallel()

		err := ErrHeaderFilter(ErrForwardFailed)

		msg := err.Error()
		if !strings.HasPrefix(msg, "headerfilter "+HeaderFilterType) {
			t.Fatalf("expected prefix %q, got %q", "headerfilter "+HeaderFilterType, msg)
		}

		if !strings.Contains(msg, ErrForwardFailed.Error()) {
			t.Fatalf("expected cause %q in message, got %q", ErrForwardFailed.Error(), msg)
		}

		if !strings.Contains(msg, ErrHeaderFilterFailed.Error()) {
			t.Fatalf("expected sentinel %q in message, got %q", ErrHeaderFilterFailed.Error(), msg)
		}
	})

	t.Run("ErrHeaderFilter joins all causes with ErrHeaderFilterFailed", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("custom failure")

		err := ErrHeaderFilter(ErrForwardFailed, boom)
		if !errors.Is(err, ErrHeaderFilterFailed) {
			t.Fatal("expected ErrHeaderFilterFailed in chain")
		}

		if !errors.Is(err, ErrForwardFailed) {
			t.Fatal("expected ErrForwardFailed in chain")
		}

		if !errors.Is(err, boom) {
			t.Fatal("expected origin error in chain")
		}
	})
}
