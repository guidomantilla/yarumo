package stores

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("includes type prefix and joined causes", func(t *testing.T) {
		t.Parallel()

		err := ErrStore(ErrInvalidTTL)

		msg := err.Error()
		if !strings.HasPrefix(msg, "messaging-store "+StoreType) {
			t.Fatalf("expected prefix %q, got %q", "messaging-store "+StoreType, msg)
		}

		if !strings.Contains(msg, ErrInvalidTTL.Error()) {
			t.Fatalf("expected cause %q in message, got %q", ErrInvalidTTL.Error(), msg)
		}

		if !strings.Contains(msg, ErrStoreFailed.Error()) {
			t.Fatalf("expected sentinel %q in message, got %q", ErrStoreFailed.Error(), msg)
		}
	})
}

func TestErrStore(t *testing.T) {
	t.Parallel()

	t.Run("joins all causes with ErrStoreFailed", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("custom failure")

		err := ErrStore(ErrInvalidTTL, boom)
		if !errors.Is(err, ErrStoreFailed) {
			t.Fatal("expected ErrStoreFailed in chain")
		}

		if !errors.Is(err, ErrInvalidTTL) {
			t.Fatal("expected ErrInvalidTTL in chain")
		}

		if !errors.Is(err, boom) {
			t.Fatal("expected origin error in chain")
		}
	})

	t.Run("ErrStoreClosed surfaces through ErrStore", func(t *testing.T) {
		t.Parallel()

		err := ErrStore(ErrStoreClosed)
		if !errors.Is(err, ErrStoreClosed) {
			t.Fatal("expected ErrStoreClosed in chain")
		}

		if !errors.Is(err, ErrStoreFailed) {
			t.Fatal("expected ErrStoreFailed in chain")
		}
	})
}

func TestErrNotFound(t *testing.T) {
	t.Parallel()

	t.Run("joins ErrStoreNotFound into chain", func(t *testing.T) {
		t.Parallel()

		err := ErrNotFound()
		if !errors.Is(err, ErrStoreNotFound) {
			t.Fatal("expected ErrStoreNotFound in chain")
		}
	})

	t.Run("preserves caller causes", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("backend down")

		err := ErrNotFound(boom)
		if !errors.Is(err, boom) {
			t.Fatal("expected origin error in chain")
		}

		if !errors.Is(err, ErrStoreNotFound) {
			t.Fatal("expected ErrStoreNotFound in chain")
		}
	})
}
