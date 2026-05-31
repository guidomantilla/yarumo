package messaging

import (
	"time"

	cuids "github.com/guidomantilla/yarumo/core/common/uids"
)

// Message is the typed envelope dispatched through a Channel[T]. It pairs
// the strongly-typed Payload with a Headers value carrying routing and
// provenance metadata.
//
// For async channels the publisher's ctx is not carried inside Message —
// Headers.CorrelationID is the canonical cross-handler correlation
// mechanism. ctx.Value propagation still happens (see the package doc on
// context propagation) but ctx-based cancellation never crosses the
// async boundary.
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
// The header set follows the Spring Integration Message reference,
// curated to the fields a broker driver or outbox layer needs (dedup,
// request/reply correlation, priority dispatch, TTL, content
// negotiation, splitter/aggregator lineage).
type Headers struct {
	// MessageID is the unique identifier of THIS message — distinct
	// from CorrelationID, which groups related messages. Populated
	// automatically by NewMessage when a uid generator is supplied.
	// Mandatory for dedup and idempotency downstream (broker drivers,
	// outbox tables).
	MessageID string
	// CorrelationID groups messages that belong to the same logical
	// conversation/saga. It is auto-populated by NewMessage when a uid
	// generator is supplied; callers may overwrite it to attach the
	// message to an existing correlation.
	CorrelationID string
	// CausationID is the MessageID of the event that caused THIS
	// event. CorrelationID groups; CausationID chains. Empty for
	// caller-originated messages.
	CausationID string
	// ReplyTo is the channel/queue name (or type discriminator) the
	// responder should publish back to. Empty for fire-and-forget.
	ReplyTo string
	// Type is a string discriminator for the payload. Redundant for
	// in-process generic typing but mandatory once payloads are
	// serialised to bytes (broker wire format, outbox table, audit
	// log).
	Type string
	// Priority is a 0–9 priority hint (higher = more important).
	// Honoured by channels with priority dispatch (none yet); ignored
	// by the in-process PipelineChannel.
	Priority uint8
	// ContentType is a MIME-ish discriminator (e.g. "application/json",
	// "application/protobuf"). Irrelevant in-process; mandatory for
	// wire-format drivers.
	ContentType string
	// ExpirationTime instructs the channel/broker to drop the message
	// if it is dequeued after this instant. Zero value means no
	// expiration.
	ExpirationTime time.Time
	// SequenceNumber is the 0-based position of this message within a
	// logical sequence produced by a Splitter. Zero when not part of a
	// sequence.
	SequenceNumber int
	// SequenceSize is the total number of messages in the logical
	// sequence (size of the Splitter output). Zero when not part of a
	// sequence.
	SequenceSize int
	// Timestamp records when the message was created.
	Timestamp time.Time
	// Source identifies the publishing module. Empty by default;
	// callers set it after construction.
	Source string
	// Custom carries optional caller-provided metadata as a key/value
	// bag for ad-hoc propagation. Nil unless the caller populates it.
	Custom map[string]any
}

// NewMessage creates a Message[T] with the given payload. Headers.
// Timestamp is set to time.Now(). When uid is non-nil, both
// Headers.MessageID and Headers.CorrelationID are populated via
// independent uid.Generate() calls so the message starts as its own
// correlation root; callers can overwrite CorrelationID later to
// attach the message to an existing conversation. When uid is nil or
// Generate returns an error, the affected field is left empty.
//
// The function takes a uid generator as a parameter (instead of pulling
// a default from common/uids) so the messaging module stays free of any
// hardcoded UID algorithm — callers inject their preferred generator.
func NewMessage[T any](payload T, uid cuids.UID) Message[T] {
	messageID := generateID(uid)
	correlationID := generateID(uid)

	return Message[T]{
		Payload: payload,
		Headers: Headers{
			MessageID:     messageID,
			CorrelationID: correlationID,
			Timestamp:     time.Now(),
		},
	}
}

// DeadLetter is the envelope routed to a Dead Letter Channel when a
// handler dispatched by Topic or Queue returns a non-nil error. It
// preserves the original Message[T], the error that caused the
// failure, and the moment the failure was observed so downstream
// consumers (audit, retry, dashboards) can reconstruct the timeline.
//
// DeadLetter is paid into the channel-wide DLQ configured via
// WithDLQChannel. Channels expect a Channel[DeadLetter[T]] as the
// destination; T must match the source channel's payload type.
type DeadLetter[T any] struct {
	// Original is the message the handler failed to process.
	Original Message[T]
	// LastError is the non-nil error returned by the handler.
	LastError error
	// FailedAt is the wall-clock time the failure was observed.
	FailedAt time.Time
}

// ErrorMessage carries a Message that failed processing alongside the
// error that caused the failure. It is the payload of an error
// channel (Channel[ErrorMessage[T]]) — the synchronous counterpart to
// DeadLetter[T], which is the asynchronous DLQ envelope.
//
// Use an error channel when the producer is still in scope and wants
// failures routed back through the same Channel[T] machinery as
// successes — for example, a request/reply pattern where the reply
// channel carries either the result or an ErrorMessage. DeadLetter[T]
// is structurally richer (FailedAt timestamp) because the DLQ flow is
// asynchronous and downstream consumers need provenance; ErrorMessage
// stays minimal because the failure context is already alive on the
// producer side.
type ErrorMessage[T any] struct {
	// Original is the message whose handler returned the failure.
	Original Message[T]
	// Cause is the non-nil error reported by the failing handler.
	Cause error
}

// NewErrorMessage wraps original and cause into a Message[ErrorMessage[T]]
// suitable for publishing on an error channel. Headers are populated
// the usual way (Timestamp now; MessageID/CorrelationID left empty
// because the uid generator is nil).
func NewErrorMessage[T any](original Message[T], cause error) Message[ErrorMessage[T]] {
	return NewMessage(ErrorMessage[T]{Original: original, Cause: cause}, nil)
}
