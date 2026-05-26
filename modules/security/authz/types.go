// Copyright 2026 Guido Mauricio Mantilla Tarazona
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package authz provides authorization primitives: a Policy interface,
// a Decision type carrying Allow/Deny/Abstain with a reason, a Request
// envelope (Principal/Resource/Action/Environment), and helpers to
// compose policies (ChainPolicies) and emit audit observations
// (DefaultAuditHook / SilentAuditHook).
//
// Sub-package rbac implements role-based access control with role
// inheritance and permission wildcards on top of the Policy interface.
//
// Transport adapters live in their own top-level modules under
// modules/extensions/security/authz/ so a consumer of the contract
// never pulls google.golang.org/grpc unless it imports the grpc
// adapter explicitly:
//
//   - extensions/security/authz/http: Require middleware for net/http.
//   - extensions/security/authz/grpc: Require unary + stream
//     interceptors for gRPC.
//
// # Design notes
//
// Principal is typed as any inside Request so authz does NOT take a
// dependency on authn. Consumers cast inside the policy to whatever
// shape their authentication layer produces (typically an
// authn.Principal). This keeps the two modules independent and lets
// the Policy be reused with custom principal shapes (machine identities,
// API keys, service accounts, etc.).
//
// The Require HTTP and gRPC middleware do not pull the Principal from
// a hardcoded context key. Consumers wire a PrincipalReader option
// (WithPrincipalReader) which knows how their authn layer stashes the
// principal in ctx. This keeps coupling explicit and lets the same
// middleware run on top of any authn implementation.
//
// Concurrency: all public types in this package are safe for concurrent
// use by multiple goroutines.
package authz

import (
	"context"
	"net"
	"time"
)

var (
	_ PrincipalReader = (PrincipalReaderFn)(nil)

	_ NewRequestFn    = NewRequest
	_ AllowFn         = Allow
	_ DenyFn          = Deny
	_ AbstainFn       = Abstain
	_ ChainPoliciesFn = ChainPolicies
	_ AuditHookFn     = DefaultAuditHook
	_ AuditHookFn     = SilentAuditHook
	_ LocalIPFn       = LocalIP
	_ ErrAuthzFn      = ErrAuthz
)

// NewRequestFn is the function type for NewRequest.
type NewRequestFn func(principal any, action string, resource Resource, env Environment) Request

// AllowFn is the function type for Allow.
type AllowFn func(reason string) Decision

// DenyFn is the function type for Deny.
type DenyFn func(reason string) Decision

// AbstainFn is the function type for Abstain.
type AbstainFn func(reason string) Decision

// ChainPoliciesFn is the function type for ChainPolicies.
type ChainPoliciesFn func(policies ...Policy) Policy

// LocalIPFn is the function type for LocalIP.
type LocalIPFn func(ip string) net.IP

// ErrAuthzFn is the function type for ErrAuthz.
type ErrAuthzFn func(causes ...error) error

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
