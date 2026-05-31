package resequencer

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("includes type prefix and joined causes", func(t *testing.T) {
		t.Parallel()

		err := ErrResequencer(ErrForwardFailed)

		msg := err.Error()
		if !strings.HasPrefix(msg, "resequencer "+ResequencerType) {
			t.Fatalf("expected prefix %q, got %q", "resequencer "+ResequencerType, msg)
		}

		if !strings.Contains(msg, ErrForwardFailed.Error()) {
			t.Fatalf("expected cause %q in message, got %q", ErrForwardFailed.Error(), msg)
		}

		if !strings.Contains(msg, ErrResequencerFailed.Error()) {
			t.Fatalf("expected sentinel %q in message, got %q", ErrResequencerFailed.Error(), msg)
		}
	})

	t.Run("ErrResequencer joins all causes with ErrResequencerFailed", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("custom failure")

		err := ErrResequencer(ErrForwardFailed, boom)
		if !errors.Is(err, ErrResequencerFailed) {
			t.Fatal("expected ErrResequencerFailed in chain")
		}

		if !errors.Is(err, ErrForwardFailed) {
			t.Fatal("expected ErrForwardFailed in chain")
		}

		if !errors.Is(err, boom) {
			t.Fatal("expected origin error in chain")
		}
	})
}
