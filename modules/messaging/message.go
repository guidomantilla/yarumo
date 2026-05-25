package messaging

import (
	"time"

	cuids "github.com/guidomantilla/yarumo/common/uids"
)

// Message is the typed envelope dispatched through a Channel[T].
//
// Headers carry routing and provenance metadata: CorrelationID for
// causal tracing across handlers, Timestamp for time-of-publish,
// Source for the publishing module (caller-set, may be empty), and
// Custom for ad-hoc key/value pairs the caller wants to propagate
// alongside the payload.
type Message[T any] struct {
	// Payload is the typed event payload.
	Payload T
	// CorrelationID is a unique identifier used to correlate the
	// message across handlers and downstream emissions. It is
	// auto-populated by NewMessage when a generator is available.
	CorrelationID string
	// Timestamp records when the message was created.
	Timestamp time.Time
	// Source identifies the publishing module. Empty by default;
	// callers set it after construction or with a future Option.
	Source string
	// Custom carries optional caller-provided metadata.
	Custom map[string]any
}

// NewMessage creates a Message[T] with the given payload. Timestamp is
// set to time.Now(). When uid is non-nil, CorrelationID is generated
// via uid.Generate(); when uid is nil or its Generate returns an
// error, CorrelationID is left empty.
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
		Payload:       payload,
		CorrelationID: correlationID,
		Timestamp:     time.Now(),
	}
}
