package authn

import (
	"context"
)

// Authenticator defines the contract for validating an authentication
// token and extracting the resulting Principal.
//
// Implementations are stateless validators: they take an opaque token
// string (typically a JWT, an opaque AEAD-encrypted token, an API key,
// etc.) and produce the *Principal it represents on success. The token
// shape and verification semantics are implementation-defined.
//
// Implementations must be safe for concurrent use by multiple goroutines.
// On failure they MUST return an error wrapping a sentinel from the
// authn package (ErrTokenEmpty, ErrTokenInvalid) so transport middleware
// can translate verification failures to a uniform 401 response without
// inspecting concrete impl-specific error types.
type Authenticator interface {
	// Validate verifies token and returns the *Principal it carries.
	// Implementations must return ErrTokenEmpty when token is the empty
	// string and an ErrTokenInvalid-wrapped error for any verification
	// failure (signature, expiry, issuer mismatch, malformed claims).
	// ctx is propagated so implementations that talk to an external
	// service can honor caller cancellation.
	Validate(ctx context.Context, token string) (*Principal, error)
}

// Principal is the immutable identity of an authenticated caller.
//
// It carries the minimal data needed by downstream authorization layers:
// an opaque ID (typically the JWT "sub"), a human-readable name (the
// "name" claim or equivalent), the roles assigned to the caller, and a
// free-form attributes bag for impl-specific claims (tenant, scopes,
// permissions, custom JWT claims, etc.).
//
// Principal is constructed once by an Authenticator and propagated
// read-only through ctx via WithPrincipal / FromContext. Mutating a
// Principal after construction is undefined behavior; treat the value as
// frozen.
type Principal struct {
	// ID is the stable identifier of the caller (e.g. the JWT "sub"
	// claim, a user ID, a service account name). Never empty for a
	// successfully validated Principal.
	ID string
	// Name is the human-readable display name of the caller, when the
	// underlying token carries one. May be empty.
	Name string
	// Roles is the list of roles granted to the caller. The slice may
	// be empty but is never nil for a Principal produced by a
	// well-formed Authenticator.
	Roles []string
	// Attributes is a free-form bag for impl-specific claims (tenant,
	// scopes, permissions, arbitrary custom JWT claims). The map may
	// be empty but is never nil for a Principal produced by a
	// well-formed Authenticator. Values are typed as any to mirror the
	// JSON shape of JWT claims.
	Attributes map[string]any
}

// principalCtxKeyType is an unexported type used as the context key for
// the Principal value. Using a dedicated type (instead of a string)
// prevents accidental collisions with other packages that store values
// under the same name in the same context.
type principalCtxKeyType struct{}

// principalCtxKey is the singleton context key under which authn stores
// the *Principal value. Exported access goes through WithPrincipal and
// FromContext.
var principalCtxKey = principalCtxKeyType{}

// WithPrincipal returns a copy of ctx that carries the given Principal.
// It is invoked by transport middleware after successful validation. A
// nil principal is rejected: the returned ctx is the original ctx
// unchanged so downstream FromContext calls keep reporting "not
// authenticated".
func WithPrincipal(ctx context.Context, principal *Principal) context.Context {
	if ctx == nil || principal == nil {
		return ctx
	}

	return context.WithValue(ctx, principalCtxKey, principal)
}

// FromContext retrieves the Principal previously stored on ctx by
// WithPrincipal. It returns (nil, false) when ctx is nil, when the value
// is absent, or when the stored value is not a *Principal.
//
// Authorization layers should treat (nil, false) as "anonymous request"
// and reject access where authentication is required.
func FromContext(ctx context.Context) (*Principal, bool) {
	if ctx == nil {
		return nil, false
	}

	value := ctx.Value(principalCtxKey)
	principal, ok := value.(*Principal)

	if !ok {
		return nil, false
	}

	return principal, true
}
