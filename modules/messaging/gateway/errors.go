package gateway

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// GatewayType is the error domain identifier for gateway operations.
const GatewayType = "gateway"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for gateway operations.
var (
	// ErrGatewayFailed is the top-level sentinel embedded in every
	// gateway-domain Error returned by ErrGateway.
	ErrGatewayFailed = errors.New("gateway failed")
	// ErrRequestTimeout indicates that no reply arrived before the
	// effective per-Request deadline (whichever is tighter between
	// WithRequestTimeout and the caller's ctx deadline) expired.
	ErrRequestTimeout = errors.New("request timed out")
	// ErrRequestCancelled indicates that the caller's ctx was
	// cancelled before a reply arrived. The original ctx.Err() is
	// joined alongside this sentinel.
	ErrRequestCancelled = errors.New("request cancelled by caller")
	// ErrGatewayShuttingDown indicates that the Gateway entered Stop
	// while at least one Request was still waiting on a reply. All
	// in-flight Requests receive this error and the zero Res.
	ErrGatewayShuttingDown = errors.New("gateway shutting down")
	// ErrGatewayNotStarted indicates that Request was invoked before
	// Start. The Gateway cannot correlate replies without an active
	// reply-channel subscription.
	ErrGatewayNotStarted = errors.New("gateway not started")
	// ErrCorrelationIDFailed indicates that the configured uid
	// generator failed to produce a correlation ID for a Request.
	ErrCorrelationIDFailed = errors.New("correlation id generation failed")
	// ErrRequestSendFailed indicates that the request Channel.Send
	// returned a non-nil error. The originating error is joined
	// alongside this sentinel.
	ErrRequestSendFailed = errors.New("request channel send failed")
	// ErrUnknownCorrelationID indicates that a reply arrived bearing
	// a CorrelationID not present in the pending-request map (either
	// already returned to the caller or never registered). The reply
	// is dropped and the ErrorHandler is invoked for visibility.
	ErrUnknownCorrelationID = errors.New("unknown correlation id")
)

// Error is the domain error type for gateway operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("gateway %s error: %s", e.Type, e.Err)
}

// ErrGateway wraps the given causes into a domain Error joined with
// ErrGatewayFailed.
func ErrGateway(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: GatewayType,
			Err:  errors.Join(append(causes, ErrGatewayFailed)...),
		},
	}
}
