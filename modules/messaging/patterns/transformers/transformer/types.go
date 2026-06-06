// Package transformer provides a Message Translator pattern over
// messaging.Channel.
//
// A Transformer subscribes to a source Channel[T] and republishes a
// derived Message[U] (produced by a user-supplied TransformFn) to a
// single destination Channel[U]. It is the only pattern in this module
// that crosses type parameters — input and output payload types may
// differ.
//
// Typical use cases:
//
//   - **Wire-format adaptation**: decode raw bytes into a domain struct
//     before downstream handlers, or re-encode a domain struct into a
//     transport-friendly envelope before forwarding to a broker bridge.
//   - **Schema upgrade**: translate a legacy event shape into the
//     current canonical shape so subscribers do not carry version
//     branches.
//   - **Enrichment**: combine an incoming Message[T] with side data
//     (cached lookups, headers) and emit a Message[U] that carries the
//     enriched payload.
//
// The TransformFn returns a fully-formed Message[U] — the caller owns
// the Headers translation policy (preserve CorrelationID, mutate Type,
// stamp Source, etc.). The transformer never silently rewrites headers.
//
// # Lifecycle
//
// Transformer implements common/lifecycle.Component (worker-style):
// Start registers the subscription on the source channel and returns
// immediately; Stop cancels the subscription and closes Done. The
// transformer does not spawn goroutines of its own — dispatch
// concurrency is inherited from the source channel implementation. Wire
// it via lifecycle.Build for the standard daemon CloseFn pattern.
//
// # Error handling
//
// The handler installed on the source channel always returns nil.
// Transformation failures (TransformFn error, TransformFn panic,
// forward Send failure) are surfaced via the Transformer's own
// ErrorHandler (installed with WithErrorHandler, defaulting to
// messaging.DefaultErrorHandler which logs via common/log). This keeps
// transformation concerns out of the source channel's caller error
// path — consistent with the package-wide policy in
// modules/messaging/CODING_STANDARDS.md.
package transformer

import (
	"context"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

var (
	_ Transformer[any, any] = (*transformer[any, any])(nil)

	_ ErrTransformerFn = ErrTransformer
)

// Transformer is the public interface for a Message Translator. It
// embeds lifecycle.Component so callers wire it up with
// lifecycle.Build. The interface exists (rather than returning
// lifecycle.Component directly) so the consumer's API surface
// preserves "this is a Transformer" semantics and the type stays open
// to future transformer-specific methods without breaking callers.
type Transformer[T, U any] interface {
	lifecycle.Component
}

// TransformFn maps an incoming Message[T] to an outgoing Message[U].
// The function owns the Headers translation policy — callers typically
// preserve CorrelationID, mutate Type and ContentType, and stamp Source
// before returning. An error returned by TransformFn is wrapped in
// ErrTransformer(ErrTransformFailed, err) and forwarded to the
// ErrorHandler; the message is dropped from the flow. A panic in
// TransformFn is recovered and wrapped in
// ErrTransformer(ErrTransformerPanic, ...) so it cannot kill the source
// channel's dispatcher.
type TransformFn[T, U any] func(ctx context.Context, msg messaging.Message[T]) (messaging.Message[U], error)

// ErrTransformerFn is the function type for ErrTransformer.
type ErrTransformerFn func(causes ...error) error
