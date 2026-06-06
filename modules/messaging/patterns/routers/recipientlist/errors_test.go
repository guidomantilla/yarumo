package recipientlist

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("includes type prefix and joined causes", func(t *testing.T) {
		t.Parallel()

		err := ErrRecipientList(ErrNoRoute)

		msg := err.Error()
		if !strings.HasPrefix(msg, "recipientlist "+RecipientListType) {
			t.Fatalf("expected prefix %q, got %q", "recipientlist "+RecipientListType, msg)
		}

		if !strings.Contains(msg, ErrNoRoute.Error()) {
			t.Fatalf("expected cause %q in message, got %q", ErrNoRoute.Error(), msg)
		}

		if !strings.Contains(msg, ErrRecipientListFailed.Error()) {
			t.Fatalf("expected sentinel %q in message, got %q", ErrRecipientListFailed.Error(), msg)
		}
	})

	t.Run("ErrRecipientList joins all causes with ErrRecipientListFailed", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("custom failure")

		err := ErrRecipientList(ErrForwardFailed, boom)
		if !errors.Is(err, ErrRecipientListFailed) {
			t.Fatal("expected ErrRecipientListFailed in chain")
		}

		if !errors.Is(err, ErrForwardFailed) {
			t.Fatal("expected ErrForwardFailed in chain")
		}

		if !errors.Is(err, boom) {
			t.Fatal("expected origin error in chain")
		}
	})
}
