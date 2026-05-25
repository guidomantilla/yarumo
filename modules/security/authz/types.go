package authz

import (
	"context"
	"net"
	"time"
)

var (
	_ PrincipalReader = (PrincipalReaderFn)(nil)
)

// Effect classifies the outcome of a Policy evaluation. Allow grants
// the request, Deny rejects it, Abstain signals that the policy does
// not have an opinion (the caller can chain to another policy or fall
// back to a default).
type Effect string

const (
	// EffectAllow grants access for the evaluated Request.
	EffectAllow Effect = "allow"
	// EffectDeny rejects access for the evaluated Request.
	EffectDeny Effect = "deny"
	// EffectAbstain signals that the policy has no opinion on the
	// Request; downstream code should chain to another policy or apply
	// a deny-by-default fallback.
	EffectAbstain Effect = "abstain"
)

// Decision is the result of evaluating a Policy against a Request.
//
// Effect is the only field that drives access; Reason and Metadata are
// observability payload (audit logs, debug headers, error responses).
// Decisions are values, not pointers — Decision is small and cheap to
// pass around.
type Decision struct {
	// Effect is Allow, Deny, or Abstain.
	Effect Effect
	// Reason is a human-readable explanation, used in audit logs and
	// in the response body (HTTP) or status message (gRPC) on deny.
	// Empty Reason is allowed but not recommended.
	Reason string
	// Metadata carries policy-specific debug context (matched role,
	// rule id, evaluation path, etc.). Nil unless the policy populates
	// it. Consumers MUST treat Metadata as read-only.
	Metadata map[string]any
}

// Resource describes the object being accessed. Type identifies the
// kind (e.g. "orders", "documents"), ID is the concrete instance, Owner
// is the principal id that owns the resource (used by RBAC/ABAC
// "owner-or-admin" patterns), Attrs is an arbitrary attribute bag for
// attribute-based policies.
type Resource struct {
	// Type is the resource kind (e.g. "orders", "documents").
	Type string
	// ID identifies the concrete instance (empty for collection-level
	// actions like "list").
	ID string
	// Owner is the principal id that owns the resource (empty when
	// ownership does not apply or is unknown).
	Owner string
	// Attrs carries arbitrary resource attributes for ABAC. Nil unless
	// the caller populates it.
	Attrs map[string]any
}

// Environment carries request-scoped context that is not part of the
// principal or the resource: caller IP, evaluation time, custom
// attributes. The Time field is populated by NewRequest with time.Now
// when zero.
type Environment struct {
	// IP is the caller IP (empty when not known).
	IP net.IP
	// Time is the evaluation time. NewRequest defaults to time.Now()
	// when the caller leaves it zero.
	Time time.Time
	// Attrs carries arbitrary environment attributes for ABAC. Nil
	// unless the caller populates it.
	Attrs map[string]any
}

// Request is the input to a Policy. Principal is typed as any to avoid
// coupling this module to a specific authentication library; consumers
// cast inside the policy to whatever Principal shape their authn layer
// produces.
type Request struct {
	// Principal is the authenticated identity (typed as any to keep
	// this module independent from authn). May be nil for
	// unauthenticated requests — policies decide how to react.
	Principal any
	// Resource is the object being accessed.
	Resource Resource
	// Action is the operation being attempted (e.g. "read", "write",
	// "delete"). Empty Action is allowed but most policies will deny.
	Action string
	// Environment carries request-scoped context (IP, time, attrs).
	Environment Environment
}

// Policy defines the interface for an authorization policy. Evaluate
// inspects the Request and returns a Decision. Implementations must be
// safe for concurrent use by multiple goroutines.
type Policy interface {
	// Evaluate decides whether the Request is allowed, denied, or
	// abstained. ctx propagates deadlines and cancellation to any
	// out-of-process call the policy makes (db lookups, OPA, etc.).
	Evaluate(ctx context.Context, req Request) Decision
}

// PrincipalReader defines the interface for extracting the authenticated
// Principal from ctx. Consumers wire a reader on the Require middleware
// via WithPrincipalReader so authz stays independent from authn.
//
// Implementations return (nil, false) when no principal is bound to ctx.
type PrincipalReader interface {
	// Read returns the principal bound to ctx (if any). The boolean
	// ok is false when ctx carries no principal — Require treats that
	// as deny.
	Read(ctx context.Context) (any, bool)
}

// PrincipalReaderFn is the function-typed adapter implementing
// PrincipalReader. It lets callers pass a closure where an interface is
// expected.
type PrincipalReaderFn func(ctx context.Context) (any, bool)

// Read invokes the wrapped function, satisfying PrincipalReader.
func (f PrincipalReaderFn) Read(ctx context.Context) (any, bool) {
	if f == nil {
		return nil, false
	}

	return f(ctx)
}

// AuditHookFn is the function type for the audit hook invoked once per
// Decision. The hook fires after the Policy has evaluated the Request
// and BEFORE Require translates the Decision into an HTTP status / gRPC
// code. ctx is the request ctx, req is the evaluated Request, dec is
// the Decision returned by the policy.
//
// The hook must not block on long observability work; dispatch
// asynchronously if needed.
type AuditHookFn func(ctx context.Context, req Request, dec Decision)
