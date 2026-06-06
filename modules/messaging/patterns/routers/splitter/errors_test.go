package splitter

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("includes type prefix and joined causes", func(t *testing.T) {
		t.Parallel()

		err := ErrSplitter(ErrForwardFailed)

		msg := err.Error()
		if !strings.HasPrefix(msg, "splitter "+SplitterType) {
			t.Fatalf("expected prefix %q, got %q", "splitter "+SplitterType, msg)
		}

		if !strings.Contains(msg, ErrForwardFailed.Error()) {
			t.Fatalf("expected cause %q in message, got %q", ErrForwardFailed.Error(), msg)
		}

		if !strings.Contains(msg, ErrSplitterFailed.Error()) {
			t.Fatalf("expected sentinel %q in message, got %q", ErrSplitterFailed.Error(), msg)
		}
	})

	t.Run("ErrSplitter joins all causes with ErrSplitterFailed", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("custom failure")

		err := ErrSplitter(ErrSplitFailed, boom)
		if !errors.Is(err, ErrSplitterFailed) {
			t.Fatal("expected ErrSplitterFailed in chain")
		}

		if !errors.Is(err, ErrSplitFailed) {
			t.Fatal("expected ErrSplitFailed in chain")
		}

		if !errors.Is(err, boom) {
			t.Fatal("expected origin error in chain")
		}
	})
}
