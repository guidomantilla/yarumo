package transformer

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("includes type prefix and joined causes", func(t *testing.T) {
		t.Parallel()

		err := ErrTransformer(ErrForwardFailed)

		msg := err.Error()
		if !strings.HasPrefix(msg, "transformer "+TransformerType) {
			t.Fatalf("expected prefix %q, got %q", "transformer "+TransformerType, msg)
		}

		if !strings.Contains(msg, ErrForwardFailed.Error()) {
			t.Fatalf("expected cause %q in message, got %q", ErrForwardFailed.Error(), msg)
		}

		if !strings.Contains(msg, ErrTransformerFailed.Error()) {
			t.Fatalf("expected sentinel %q in message, got %q", ErrTransformerFailed.Error(), msg)
		}
	})

	t.Run("ErrTransformer joins all causes with ErrTransformerFailed", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("custom failure")

		err := ErrTransformer(ErrTransformFailed, boom)
		if !errors.Is(err, ErrTransformerFailed) {
			t.Fatal("expected ErrTransformerFailed in chain")
		}

		if !errors.Is(err, ErrTransformFailed) {
			t.Fatal("expected ErrTransformFailed in chain")
		}

		if !errors.Is(err, boom) {
			t.Fatal("expected origin error in chain")
		}
	})
}
