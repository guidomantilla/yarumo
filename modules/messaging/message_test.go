package messaging

import (
	"errors"
	"testing"
	"time"

	cuids "github.com/guidomantilla/yarumo/core/common/uids"
)

func TestNewMessage(t *testing.T) {
	t.Parallel()

	t.Run("carries payload", func(t *testing.T) {
		t.Parallel()

		msg := NewMessage[string]("hello", nil)
		if msg.Payload != "hello" {
			t.Fatalf("expected payload %q, got %q", "hello", msg.Payload)
		}
	})

	t.Run("timestamp is recent", func(t *testing.T) {
		t.Parallel()

		before := time.Now()
		msg := NewMessage[int](42, nil)
		after := time.Now()

		if msg.Headers.Timestamp.Before(before) || msg.Headers.Timestamp.After(after) {
			t.Fatalf("timestamp %v not within [%v, %v]", msg.Headers.Timestamp, before, after)
		}
	})

	t.Run("correlation id empty when uid is nil", func(t *testing.T) {
		t.Parallel()

		msg := NewMessage[int](7, nil)
		if msg.Headers.CorrelationID != "" {
			t.Fatalf("expected empty correlation id, got %q", msg.Headers.CorrelationID)
		}
	})

	t.Run("correlation id set when uid generates", func(t *testing.T) {
		t.Parallel()

		uid := cuids.NewUID("test", func() (string, error) {
			return "id-1", nil
		})

		msg := NewMessage[int](7, uid)
		if msg.Headers.CorrelationID != "id-1" {
			t.Fatalf("expected correlation id %q, got %q", "id-1", msg.Headers.CorrelationID)
		}
	})

	t.Run("correlation id empty when uid Generate errors", func(t *testing.T) {
		t.Parallel()

		uid := cuids.NewUID("test", func() (string, error) {
			return "", errors.New("boom")
		})

		msg := NewMessage[int](7, uid)
		if msg.Headers.CorrelationID != "" {
			t.Fatalf("expected empty correlation id on error, got %q", msg.Headers.CorrelationID)
		}
	})

	t.Run("source empty by default", func(t *testing.T) {
		t.Parallel()

		msg := NewMessage[int](1, nil)
		if msg.Headers.Source != "" {
			t.Fatalf("expected empty source, got %q", msg.Headers.Source)
		}
	})

	t.Run("custom nil by default", func(t *testing.T) {
		t.Parallel()

		msg := NewMessage[int](1, nil)
		if msg.Headers.Custom != nil {
			t.Fatalf("expected nil Custom, got %v", msg.Headers.Custom)
		}
	})
}
