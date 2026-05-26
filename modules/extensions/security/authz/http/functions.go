package http

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/guidomantilla/yarumo/security/authz"
)

// RequireHTTP returns an HTTP middleware that evaluates policy against
// the inbound request for the given action and either calls the next
// handler (Allow) or short-circuits with 403 Forbidden (Deny / Abstain).
//
// The middleware reads the authenticated principal from ctx via the
// configured PrincipalReader (WithPrincipalReader). If no
// PrincipalReader is configured or it returns no principal, the
// request is denied without invoking the policy.
//
// On deny, the response body is the Decision.Reason wrapped in a small
// JSON envelope ({"error": "...", "reason": "..."}) and Content-Type
// is set to application/json. The Reason is also surfaced in the
// X-Authz-Reason response header so callers that consume the response
// body as plain text still see it.
//
// Action must be non-empty; an empty action panics at construction
// (fail closed loud, not at request time). policy must be non-nil;
// a nil policy panics at construction.
func RequireHTTP(policy authz.Policy, action string, opts ...Option) func(http.Handler) http.Handler {
	if policy == nil {
		panic(authz.ErrAuthz(authz.ErrPolicyNil))
	}

	if action == "" {
		panic(authz.ErrAuthz(authz.ErrActionEmpty))
	}

	options := NewOptions(opts...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			principal, ok := readPrincipal(ctx, options.principalReader)
			if !ok {
				dec := authz.Deny("principal not present in context")
				options.auditHook(ctx, authz.Request{Action: action}, dec)
				writeAuthzDeny(w, dec)

				return
			}

			req := buildRequest(r, principal, action, options.resourceFn)
			dec := policy.Evaluate(ctx, req)
			options.auditHook(ctx, req, dec)

			if dec.Effect != authz.EffectAllow {
				writeAuthzDeny(w, dec)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// readPrincipal delegates to the configured PrincipalReader. A nil
// reader returns ok=false (caller treats as deny).
func readPrincipal(ctx context.Context, reader authz.PrincipalReader) (any, bool) {
	if reader == nil {
		return nil, false
	}

	return reader.Read(ctx)
}

// buildRequest assembles a Request from the HTTP request, the
// principal pulled from ctx, the configured action, and the resource
// resolver (if any).
func buildRequest(r *http.Request, principal any, action string, resolver HTTPResourceResolverFn) authz.Request {
	resource := authz.Resource{}
	if resolver != nil {
		resource = resolver(r)
	}

	env := authz.Environment{IP: clientIP(r)}

	return authz.NewRequest(principal, action, resource, env)
}

// clientIP extracts the caller IP from r. It honors X-Forwarded-For (
// first hop) when present and falls back to RemoteAddr. Returns nil if
// no IP can be resolved.
func clientIP(r *http.Request) net.IP {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		first := strings.TrimSpace(parts[0])
		ip := net.ParseIP(first)
		if ip != nil {
			return ip
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	return net.ParseIP(host)
}

// writeAuthzDeny serializes a Deny / Abstain Decision as a 403 response
// with the Reason in both the body envelope and an X-Authz-Reason
// header for plain-text consumers.
//
// Marshaling a map[string]string with only string keys/values is
// infallible per encoding/json semantics, so the Marshal error path
// is unreachable at runtime; the linter directive below silences
// errchkjson for that reason.
func writeAuthzDeny(w http.ResponseWriter, dec authz.Decision) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Authz-Reason", dec.Reason)
	w.WriteHeader(http.StatusForbidden)

	body := map[string]string{
		"error":  "forbidden",
		"reason": dec.Reason,
	}

	payload, _ := json.Marshal(body) //nolint:errchkjson // map[string]string is infallible

	_, _ = w.Write(payload)
}
