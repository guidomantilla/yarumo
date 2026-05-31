package gateway

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("includes type prefix and joined causes", func(t *testing.T) {
		t.Parallel()

		err := ErrGateway(ErrRequestTimeout)

		msg := err.Error()
		if !strings.HasPrefix(msg, "gateway "+GatewayType) {
			t.Fatalf("expected prefix %q, got %q", "gateway "+GatewayType, msg)
		}

		if !strings.Contains(msg, ErrRequestTimeout.Error()) {
			t.Fatalf("expected cause %q in message, got %q", ErrRequestTimeout.Error(), msg)
		}

		if !strings.Contains(msg, ErrGatewayFailed.Error()) {
			t.Fatalf("expected sentinel %q in message, got %q", ErrGatewayFailed.Error(), msg)
		}
	})

	t.Run("ErrGateway joins all causes with ErrGatewayFailed", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("custom failure")

		err := ErrGateway(ErrRequestSendFailed, boom)
		if !errors.Is(err, ErrGatewayFailed) {
			t.Fatal("expected ErrGatewayFailed in chain")
		}

		if !errors.Is(err, ErrRequestSendFailed) {
			t.Fatal("expected ErrRequestSendFailed in chain")
		}

		if !errors.Is(err, boom) {
			t.Fatal("expected origin error in chain")
		}
	})
}
