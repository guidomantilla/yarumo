package messaging

import (
	"time"

	cuids "github.com/guidomantilla/yarumo/core/common/uids"
)

// Message is the typed envelope dispatched through a Channel[T]. It pairs
// the strongly-typed Payload with a Headers value carrying routing and
// provenance metadata.
type Message[T any] struct {
	// Payload is the typed event payload.
	Payload T
	// Headers carries routing and provenance metadata for the message.
	Headers Headers
}

// Headers groups the metadata fields that travel alongside a Message
// payload. Headers are propagated end-to-end through every Channel
// handler.
//
// The current set covers the minimum required for in-process pub/sub
// (correlation, time-of-publish, publishing-module identity, plus a
// caller-controlled key/value bag). Future tickets extend this struct
// with the headers a broker driver or outbox layer would need
// (MessageID, ReplyTo, Priority, Expiration, ContentType, …).
type Headers struct {
	// CorrelationID is a unique identifier used to correlate the
	// message across handlers and downstream emissions. It is
	// auto-populated by NewMessage when a uid generator is supplied.
	CorrelationID string
	// Timestamp records when the message was created.
	Timestamp time.Time
	// Source identifies the publishing module. Empty by default;
	// callers set it after construction or with a future Option.
	Source string
	// Custom carries optional caller-provided metadata as a key/value
	// bag for ad-hoc propagation. Nil unless the caller populates it.
	Custom map[string]any
}

// NewMessage creates a Message[T] with the given payload. Headers.
// Timestamp is set to time.Now(). When uid is non-nil, Headers.
// CorrelationID is generated via uid.Generate(); when uid is nil or
// Generate returns an error, Headers.CorrelationID is left empty.
//
// The function takes a uid generator as a parameter (instead of pulling
// a default from common/uids) so the messaging module stays free of any
// hardcoded UID algorithm — callers inject their preferred generator.
func NewMessage[T any](payload T, uid cuids.UID) Message[T] {
	correlationID := ""
	if uid != nil {
		id, err := uid.Generate()
		if err == nil {
			correlationID = id
		}
	}

	return Message[T]{
		Payload: payload,
		Headers: Headers{
			CorrelationID: correlationID,
			Timestamp:     time.Now(),
		},
	}
}
