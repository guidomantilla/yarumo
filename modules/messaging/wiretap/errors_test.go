package wiretap

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("includes type prefix and joined causes", func(t *testing.T) {
		t.Parallel()

		err := ErrWiretap(ErrTapSendFailed)

		msg := err.Error()
		if !strings.HasPrefix(msg, "wiretap "+WiretapType) {
			t.Fatalf("expected prefix %q, got %q", "wiretap "+WiretapType, msg)
		}

		if !strings.Contains(msg, ErrTapSendFailed.Error()) {
			t.Fatalf("expected cause %q in message, got %q", ErrTapSendFailed.Error(), msg)
		}

		if !strings.Contains(msg, ErrWiretapFailed.Error()) {
			t.Fatalf("expected sentinel %q in message, got %q", ErrWiretapFailed.Error(), msg)
		}
	})

	t.Run("ErrWiretap joins all causes with ErrWiretapFailed", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("custom failure")

		err := ErrWiretap(ErrForwardFailed, boom)
		if !errors.Is(err, ErrWiretapFailed) {
			t.Fatal("expected ErrWiretapFailed in chain")
		}

		if !errors.Is(err, ErrForwardFailed) {
			t.Fatal("expected ErrForwardFailed in chain")
		}

		if !errors.Is(err, boom) {
			t.Fatal("expected origin error in chain")
		}
	})
}
