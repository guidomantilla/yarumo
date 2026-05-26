package authz

import (
	"context"
	"net"
	"time"

	clog "github.com/guidomantilla/yarumo/common/log"
)

var (
	_ Policy = (*chain)(nil)
)

// NewRequest constructs a Request. When env.Time is zero, NewRequest
// defaults it to time.Now() so policies can compare without a nil
// guard. The principal is taken as-is; callers pass whatever shape
// their authn layer produces.
func NewRequest(principal any, action string, resource Resource, env Environment) Request {
	if env.Time.IsZero() {
		env.Time = time.Now()
	}

	return Request{
		Principal:   principal,
		Resource:    resource,
		Action:      action,
		Environment: env,
	}
}

// Allow returns a Decision with EffectAllow and the given Reason.
func Allow(reason string) Decision {
	return Decision{Effect: EffectAllow, Reason: reason}
}

// Deny returns a Decision with EffectDeny and the given Reason.
func Deny(reason string) Decision {
	return Decision{Effect: EffectDeny, Reason: reason}
}

// Abstain returns a Decision with EffectAbstain and the given Reason.
func Abstain(reason string) Decision {
	return Decision{Effect: EffectAbstain, Reason: reason}
}

// ChainPolicies composes N policies into a single Policy that evaluates
// them in order. The first Allow or Deny wins; Abstain falls through to
// the next policy. If every policy abstains, the chain returns the last
// Abstain Decision (callers can treat that as deny-by-default in their
// Require middleware).
//
// Nil policies in the input are skipped. An empty or all-nil chain
// returns an Abstain-only Policy with reason "no policies configured".
func ChainPolicies(policies ...Policy) Policy {
	filtered := make([]Policy, 0, len(policies))
	for _, p := range policies {
		if p != nil {
			filtered = append(filtered, p)
		}
	}

	return &chain{policies: filtered}
}

// chain implements Policy as a fail-fast chain over a slice of
// sub-policies. Defined here (instead of a dedicated file) because the
// type has a single responsibility and no companion methods beyond
// Evaluate.
type chain struct {
	policies []Policy
}

// Evaluate walks the configured policies in order. The first non-
// Abstain decision wins; Abstain falls through. An empty chain
// abstains.
func (c *chain) Evaluate(ctx context.Context, req Request) Decision {
	last := Abstain("no policies configured")

	for _, p := range c.policies {
		dec := p.Evaluate(ctx, req)
		if dec.Effect != EffectAbstain {
			return dec
		}

		last = dec
	}

	return last
}

// DefaultAuditHook logs the Decision via common/log. Allow goes to Info,
// Deny and Abstain go to Warn. The hook is installed by NewOptions when
// the caller does not pass WithAuditHook.
func DefaultAuditHook(ctx context.Context, req Request, dec Decision) {
	args := []any{
		"effect", string(dec.Effect),
		"reason", dec.Reason,
		"action", req.Action,
		"resource_type", req.Resource.Type,
		"resource_id", req.Resource.ID,
	}

	if !req.Environment.Time.IsZero() {
		args = append(args, "time", req.Environment.Time.Format(time.RFC3339))
	}

	if req.Environment.IP != nil {
		args = append(args, "ip", req.Environment.IP.String())
	}

	switch dec.Effect {
	case EffectAllow:
		clog.Info(ctx, "authz decision", args...)
	case EffectDeny, EffectAbstain:
		clog.Warn(ctx, "authz decision", args...)
	default:
		clog.Warn(ctx, "authz decision unknown effect", args...)
	}
}

// SilentAuditHook is a no-op AuditHookFn. Use it when the caller wants
// to suppress audit logging entirely (typically in tests).
func SilentAuditHook(_ context.Context, _ Request, _ Decision) {}

// LocalIP is a convenience helper for constructing the Environment.IP
// field from a string. Invalid input returns nil.
func LocalIP(ip string) net.IP {
	if ip == "" {
		return nil
	}

	return net.ParseIP(ip)
}
