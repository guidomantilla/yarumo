package router

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("includes type prefix and joined causes", func(t *testing.T) {
		t.Parallel()

		err := ErrRoute(ErrNoRoute)

		msg := err.Error()
		if !strings.HasPrefix(msg, "router "+RouterType) {
			t.Fatalf("expected prefix %q, got %q", "router "+RouterType, msg)
		}

		if !strings.Contains(msg, ErrNoRoute.Error()) {
			t.Fatalf("expected cause %q in message, got %q", ErrNoRoute.Error(), msg)
		}

		if !strings.Contains(msg, ErrRouteFailed.Error()) {
			t.Fatalf("expected sentinel %q in message, got %q", ErrRouteFailed.Error(), msg)
		}
	})

	t.Run("ErrRoute joins all causes with ErrRouteFailed", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("custom failure")

		err := ErrRoute(ErrRouteFnFailed, boom)
		if !errors.Is(err, ErrRouteFailed) {
			t.Fatal("expected ErrRouteFailed in chain")
		}

		if !errors.Is(err, ErrRouteFnFailed) {
			t.Fatal("expected ErrRouteFnFailed in chain")
		}

		if !errors.Is(err, boom) {
			t.Fatal("expected origin error in chain")
		}
	})
}
