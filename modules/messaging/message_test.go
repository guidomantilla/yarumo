package messaging

import (
	"errors"
	"fmt"
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

	t.Run("message id empty when uid is nil", func(t *testing.T) {
		t.Parallel()

		msg := NewMessage[int](1, nil)
		if msg.Headers.MessageID != "" {
			t.Fatalf("expected empty message id, got %q", msg.Headers.MessageID)
		}
	})

	t.Run("message id set when uid generates", func(t *testing.T) {
		t.Parallel()

		// Counter ensures each Generate call returns a distinct id so
		// MessageID and CorrelationID get different values.
		count := 0
		uid := cuids.NewUID("test", func() (string, error) {
			count++

			return fmt.Sprintf("id-%d", count), nil
		})

		msg := NewMessage[int](1, uid)
		if msg.Headers.MessageID == "" {
			t.Fatal("expected non-empty message id")
		}

		if msg.Headers.CorrelationID == "" {
			t.Fatal("expected non-empty correlation id")
		}

		if msg.Headers.MessageID == msg.Headers.CorrelationID {
			t.Fatalf("MessageID and CorrelationID must be independent, both = %q", msg.Headers.MessageID)
		}
	})

	t.Run("message id empty when uid Generate errors", func(t *testing.T) {
		t.Parallel()

		uid := cuids.NewUID("test", func() (string, error) {
			return "", errors.New("boom")
		})

		msg := NewMessage[int](1, uid)
		if msg.Headers.MessageID != "" {
			t.Fatalf("expected empty message id on error, got %q", msg.Headers.MessageID)
		}
	})

	t.Run("extended headers zero-valued by default", func(t *testing.T) {
		t.Parallel()

		msg := NewMessage[int](1, nil)
		h := msg.Headers

		if h.CausationID != "" {
			t.Fatalf("expected empty CausationID, got %q", h.CausationID)
		}

		if h.ReplyTo != "" {
			t.Fatalf("expected empty ReplyTo, got %q", h.ReplyTo)
		}

		if h.Type != "" {
			t.Fatalf("expected empty Type, got %q", h.Type)
		}

		if h.Priority != 0 {
			t.Fatalf("expected zero Priority, got %d", h.Priority)
		}

		if h.ContentType != "" {
			t.Fatalf("expected empty ContentType, got %q", h.ContentType)
		}

		if !h.ExpirationTime.IsZero() {
			t.Fatalf("expected zero ExpirationTime, got %v", h.ExpirationTime)
		}

		if h.SequenceNumber != 0 {
			t.Fatalf("expected zero SequenceNumber, got %d", h.SequenceNumber)
		}

		if h.SequenceSize != 0 {
			t.Fatalf("expected zero SequenceSize, got %d", h.SequenceSize)
		}
	})

	t.Run("extended headers writable by caller", func(t *testing.T) {
		t.Parallel()

		msg := NewMessage[int](1, nil)
		expires := time.Now().Add(time.Hour)

		msg.Headers.CausationID = "cause-1"
		msg.Headers.ReplyTo = "reply.q"
		msg.Headers.Type = "OrderPlaced"
		msg.Headers.Priority = 7
		msg.Headers.ContentType = "application/json"
		msg.Headers.ExpirationTime = expires
		msg.Headers.SequenceNumber = 3
		msg.Headers.SequenceSize = 10

		if msg.Headers.CausationID != "cause-1" {
			t.Fatalf("CausationID = %q", msg.Headers.CausationID)
		}

		if msg.Headers.ReplyTo != "reply.q" {
			t.Fatalf("ReplyTo = %q", msg.Headers.ReplyTo)
		}

		if msg.Headers.Type != "OrderPlaced" {
			t.Fatalf("Type = %q", msg.Headers.Type)
		}

		if msg.Headers.Priority != 7 {
			t.Fatalf("Priority = %d", msg.Headers.Priority)
		}

		if msg.Headers.ContentType != "application/json" {
			t.Fatalf("ContentType = %q", msg.Headers.ContentType)
		}

		if !msg.Headers.ExpirationTime.Equal(expires) {
			t.Fatalf("ExpirationTime = %v, want %v", msg.Headers.ExpirationTime, expires)
		}

		if msg.Headers.SequenceNumber != 3 {
			t.Fatalf("SequenceNumber = %d", msg.Headers.SequenceNumber)
		}

		if msg.Headers.SequenceSize != 10 {
			t.Fatalf("SequenceSize = %d", msg.Headers.SequenceSize)
		}
	})
}

func TestNewErrorMessage(t *testing.T) {
	t.Parallel()

	t.Run("wraps original and cause into Message envelope", func(t *testing.T) {
		t.Parallel()

		orig := NewMessage[int](42, nil)
		cause := errors.New("boom")

		out := NewErrorMessage(orig, cause)

		if out.Payload.Original.Payload != 42 {
			t.Fatalf("expected original payload 42, got %d", out.Payload.Original.Payload)
		}

		if !errors.Is(out.Payload.Cause, cause) {
			t.Fatalf("expected cause %v, got %v", cause, out.Payload.Cause)
		}
	})

	t.Run("envelope has a recent timestamp", func(t *testing.T) {
		t.Parallel()

		before := time.Now()
		out := NewErrorMessage(NewMessage[int](1, nil), errors.New("x"))
		after := time.Now()

		if out.Headers.Timestamp.Before(before) || out.Headers.Timestamp.After(after) {
			t.Fatalf("timestamp %v not within [%v, %v]", out.Headers.Timestamp, before, after)
		}
	})

	t.Run("preserves nil cause", func(t *testing.T) {
		t.Parallel()

		out := NewErrorMessage(NewMessage[int](7, nil), nil)
		if out.Payload.Cause != nil {
			t.Fatalf("expected nil cause, got %v", out.Payload.Cause)
		}
	})
}
