// Package controlbus provides a Control Bus pattern over
// messaging.Channel[T].
//
// A ControlBus subscribes to a command Channel[Command] and dispatches
// each received Command to a registered Handler keyed by Verb. Each
// invocation produces a Result that is published to a separate reply
// Channel[Result]. The pattern lets ops dispatch administrative
// commands — "start", "stop", "stats", "reload-config", or any custom
// verb — to running components via the same messaging fabric used for
// domain events, without exposing a dedicated HTTP/gRPC admin endpoint.
//
// # Handler registry
//
// Handlers are passed at construction as a map[string]Handler. Unknown
// verbs are answered by the configured UnknownVerbHandler (default:
// returns Result{Success: false, Message: "unknown verb"}). Handlers
// run under panic recovery: a panicking handler produces
// Result{Success: false} with the panic value recorded in Message, and
// fires the ErrorHandler with ErrHandlerPanic.
//
// # Lifecycle
//
// ControlBus implements common/lifecycle.Component (worker-style):
// Start registers the subscription on the command channel and returns
// immediately; Stop cancels the subscription and closes Done. ControlBus
// does not spawn goroutines of its own — dispatch concurrency is
// inherited from the command channel implementation.
//
// # Error handling
//
// The handler installed on the command channel always returns nil so
// control-bus concerns never propagate to the command channel's Send
// caller. Real failures (handler panic, reply-channel Send failure) are
// surfaced via the ControlBus's own ErrorHandler (installed with
// WithErrorHandler, defaulting to messaging.DefaultErrorHandler which
// logs via common/log). Consistent with the package-wide policy in
// modules/messaging/CODING_STANDARDS.md.
package controlbus

import (
	"context"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

var (
	_ ControlBus = (*controlBus)(nil)

	_ ErrControlBusFn = ErrControlBus
)

// ControlBus is the public interface for a Control Bus. It embeds
// lifecycle.Component so callers wire it up with lifecycle.Build. The
// interface exists (rather than returning lifecycle.Component directly)
// so the consumer's API surface preserves "this is a ControlBus"
// semantics and the type stays open to future control-bus-specific
// methods without breaking callers.
type ControlBus interface {
	lifecycle.Component
}

// Command is the administrative request envelope dispatched through a
// ControlBus. Verb selects the Handler from the registry; Target names
// the component the verb applies to (the empty string conventionally
// means "all components"); Args carries verb-specific parameters as a
// string/string bag for ad-hoc propagation.
type Command struct {
	// Verb is the action name used as the registry key. Verbs are
	// case-sensitive; ControlBus does not normalise.
	Verb string
	// Target identifies the component the verb applies to. Empty
	// conventionally means "all components"; handlers decide the
	// semantics.
	Target string
	// Args carries verb-specific arguments. May be nil when the verb
	// takes no arguments.
	Args map[string]string
}

// Result is the response envelope published to the reply channel by
// the ControlBus after a Handler returns. The originating Command is
// echoed in full so correlation does not depend on Headers; Success +
// Message are the canonical pass/fail summary; Data carries
// verb-specific structured response data.
type Result struct {
	// Command is the originating Command, echoed back so the caller
	// can correlate responses without inspecting Headers.
	Command Command
	// Success reports whether the handler completed without error.
	// Handler panics, unknown verbs, and handler-reported failures
	// all yield false.
	Success bool
	// Message is a human-readable summary of the outcome. For
	// failures it carries the error description (panic value, "unknown
	// verb", handler-returned message); for successes it is free-form.
	Message string
	// Data carries verb-specific structured response data. May be nil
	// when the handler has no structured payload to return.
	Data map[string]any
}

// Handler is the function type for a Command handler. The handler
// receives the dispatch context and the typed Command and returns a
// Result. Handlers MUST NOT return a separate error: failures are
// encoded into the returned Result (Success=false, Message=...). The
// dispatcher runs Handler under panic recovery, so a Handler may panic
// without crashing the bus.
type Handler func(ctx context.Context, cmd Command) Result

// ErrControlBusFn is the function type for ErrControlBus.
type ErrControlBusFn func(causes ...error) error
