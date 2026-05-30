package bridge

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("includes type prefix and joined causes", func(t *testing.T) {
		t.Parallel()

		err := ErrBridge(ErrForwardFailed)

		msg := err.Error()
		if !strings.HasPrefix(msg, "bridge "+BridgeType) {
			t.Fatalf("expected prefix %q, got %q", "bridge "+BridgeType, msg)
		}

		if !strings.Contains(msg, ErrForwardFailed.Error()) {
			t.Fatalf("expected cause %q in message, got %q", ErrForwardFailed.Error(), msg)
		}

		if !strings.Contains(msg, ErrBridgeFailed.Error()) {
			t.Fatalf("expected sentinel %q in message, got %q", ErrBridgeFailed.Error(), msg)
		}
	})

	t.Run("ErrBridge joins all causes with ErrBridgeFailed", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("custom failure")

		err := ErrBridge(ErrForwardFailed, boom)
		if !errors.Is(err, ErrBridgeFailed) {
			t.Fatal("expected ErrBridgeFailed in chain")
		}

		if !errors.Is(err, ErrForwardFailed) {
			t.Fatal("expected ErrForwardFailed in chain")
		}

		if !errors.Is(err, boom) {
			t.Fatal("expected origin error in chain")
		}
	})
}
