package history

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("includes type prefix and joined causes", func(t *testing.T) {
		t.Parallel()

		err := ErrHistory(ErrForwardFailed)

		msg := err.Error()
		if !strings.HasPrefix(msg, "history "+HistoryType) {
			t.Fatalf("expected prefix %q, got %q", "history "+HistoryType, msg)
		}

		if !strings.Contains(msg, ErrForwardFailed.Error()) {
			t.Fatalf("expected cause %q in message, got %q", ErrForwardFailed.Error(), msg)
		}

		if !strings.Contains(msg, ErrHistoryFailed.Error()) {
			t.Fatalf("expected sentinel %q in message, got %q", ErrHistoryFailed.Error(), msg)
		}
	})

	t.Run("ErrHistory joins all causes with ErrHistoryFailed", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("custom failure")

		err := ErrHistory(ErrForwardFailed, boom)
		if !errors.Is(err, ErrHistoryFailed) {
			t.Fatal("expected ErrHistoryFailed in chain")
		}

		if !errors.Is(err, ErrForwardFailed) {
			t.Fatal("expected ErrForwardFailed in chain")
		}

		if !errors.Is(err, boom) {
			t.Fatal("expected origin error in chain")
		}
	})
}
