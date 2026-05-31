package barrier

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("includes type prefix and joined causes", func(t *testing.T) {
		t.Parallel()

		err := ErrBarrier(ErrForwardFailed)

		msg := err.Error()
		if !strings.HasPrefix(msg, "barrier "+BarrierType) {
			t.Fatalf("expected prefix %q, got %q", "barrier "+BarrierType, msg)
		}

		if !strings.Contains(msg, ErrForwardFailed.Error()) {
			t.Fatalf("expected cause %q in message, got %q", ErrForwardFailed.Error(), msg)
		}

		if !strings.Contains(msg, ErrBarrierFailed.Error()) {
			t.Fatalf("expected sentinel %q in message, got %q", ErrBarrierFailed.Error(), msg)
		}
	})

	t.Run("ErrBarrier joins all causes with ErrBarrierFailed", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("custom failure")

		err := ErrBarrier(ErrForwardFailed, boom)
		if !errors.Is(err, ErrBarrierFailed) {
			t.Fatal("expected ErrBarrierFailed in chain")
		}

		if !errors.Is(err, ErrForwardFailed) {
			t.Fatal("expected ErrForwardFailed in chain")
		}

		if !errors.Is(err, boom) {
			t.Fatal("expected origin error in chain")
		}
	})
}
