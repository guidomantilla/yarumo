package authn

import (
	"context"
)

// principalCtxKeyType is an unexported type used as the context key for
// the Principal value. Using a dedicated type (instead of a string)
// prevents accidental collisions with other packages that store values
// under the same name in the same context.
type principalCtxKeyType struct{}

// principalCtxKey is the singleton context key under which authn stores
// the *Principal value. Exported access goes through WithPrincipal and
// FromContext.
//
//nolint:gochecknoglobals // dedicated unexported ctx-key sentinel; the canonical pattern for context.WithValue keys
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
