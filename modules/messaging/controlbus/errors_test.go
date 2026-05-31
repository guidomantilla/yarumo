package controlbus

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("includes type prefix and joined causes", func(t *testing.T) {
		t.Parallel()

		err := ErrControlBus(ErrHandlerPanic)

		msg := err.Error()
		if !strings.HasPrefix(msg, "controlbus "+ControlBusType) {
			t.Fatalf("expected prefix %q, got %q", "controlbus "+ControlBusType, msg)
		}

		if !strings.Contains(msg, ErrHandlerPanic.Error()) {
			t.Fatalf("expected cause %q in message, got %q", ErrHandlerPanic.Error(), msg)
		}

		if !strings.Contains(msg, ErrControlBusFailed.Error()) {
			t.Fatalf("expected sentinel %q in message, got %q", ErrControlBusFailed.Error(), msg)
		}
	})

	t.Run("ErrControlBus joins all causes with ErrControlBusFailed", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("custom failure")

		err := ErrControlBus(ErrForwardFailed, boom)
		if !errors.Is(err, ErrControlBusFailed) {
			t.Fatal("expected ErrControlBusFailed in chain")
		}

		if !errors.Is(err, ErrForwardFailed) {
			t.Fatal("expected ErrForwardFailed in chain")
		}

		if !errors.Is(err, boom) {
			t.Fatal("expected origin error in chain")
		}
	})
}
