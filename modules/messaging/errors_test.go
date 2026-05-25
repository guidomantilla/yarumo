package messaging

import (
	"errors"
	"strings"
	"testing"
)

func TestErrSend(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrSendFailed", func(t *testing.T) {
		t.Parallel()

		err := ErrSend(ErrClosed)
		if !errors.Is(err, ErrSendFailed) {
			t.Fatalf("expected error to wrap ErrSendFailed, got %v", err)
		}
	})

	t.Run("preserves cause", func(t *testing.T) {
		t.Parallel()

		err := ErrSend(ErrClosed)
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("expected error to wrap ErrClosed, got %v", err)
		}
	})

	t.Run("error message includes type", func(t *testing.T) {
		t.Parallel()

		err := ErrSend(ErrClosed)
		if !strings.Contains(err.Error(), MessagingType) {
			t.Fatalf("expected message to contain %q, got %q", MessagingType, err.Error())
		}
	})
}

func TestErrSubscribe(t *testing.T) {
	t.Parallel()

	t.Run("wraps ErrSubscribeFailed", func(t *testing.T) {
		t.Parallel()

		err := ErrSubscribe(ErrHandlerNil)
		if !errors.Is(err, ErrSubscribeFailed) {
			t.Fatalf("expected error to wrap ErrSubscribeFailed, got %v", err)
		}
	})

	t.Run("preserves cause", func(t *testing.T) {
		t.Parallel()

		err := ErrSubscribe(ErrHandlerNil)
		if !errors.Is(err, ErrHandlerNil) {
			t.Fatalf("expected error to wrap ErrHandlerNil, got %v", err)
		}
	})
}

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("formats messaging type and cause", func(t *testing.T) {
		t.Parallel()

		err := ErrSend(ErrClosed)
		got := err.Error()
		if !strings.Contains(got, "messaging") {
			t.Fatalf("expected message to mention messaging, got %q", got)
		}
		if !strings.Contains(got, "channel closed") {
			t.Fatalf("expected message to mention cause, got %q", got)
		}
	})
}
