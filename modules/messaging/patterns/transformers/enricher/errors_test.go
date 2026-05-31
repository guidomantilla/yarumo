package enricher

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("includes type prefix and joined causes", func(t *testing.T) {
		t.Parallel()

		err := ErrEnricher(ErrEnrichFnFailed)

		msg := err.Error()
		if !strings.HasPrefix(msg, "enricher "+EnricherType) {
			t.Fatalf("expected prefix %q, got %q", "enricher "+EnricherType, msg)
		}

		if !strings.Contains(msg, ErrEnrichFnFailed.Error()) {
			t.Fatalf("expected cause %q in message, got %q", ErrEnrichFnFailed.Error(), msg)
		}

		if !strings.Contains(msg, ErrEnricherFailed.Error()) {
			t.Fatalf("expected sentinel %q in message, got %q", ErrEnricherFailed.Error(), msg)
		}
	})

	t.Run("ErrEnricher joins all causes with ErrEnricherFailed", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("custom failure")

		err := ErrEnricher(ErrEnrichFnFailed, boom)
		if !errors.Is(err, ErrEnricherFailed) {
			t.Fatal("expected ErrEnricherFailed in chain")
		}

		if !errors.Is(err, ErrEnrichFnFailed) {
			t.Fatal("expected ErrEnrichFnFailed in chain")
		}

		if !errors.Is(err, boom) {
			t.Fatal("expected origin error in chain")
		}
	})
}
