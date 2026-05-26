// Package rbac provides a role-based access control engine that
// implements authz.Policy.
//
// Roles map to permission strings; permissions match against the
// authz.Request action and resource type via the canonical encoding
// "<resource_type>.<action>" with wildcard support ("*" matches every
// segment).
//
// Role inheritance is configured via a hierarchy: declaring "admin >
// editor > viewer" lets admin inherit every permission editor and
// viewer hold.
//
// The roles a Principal carries are produced by the configured
// RolesStore implementation. The default impl (NewInMemoryRolesStore)
// resolves roles by principal id string, leaving the choice of how to
// derive that id from a Principal value entirely to the consumer (the
// authz.Request.Principal field is typed as any specifically to keep
// this module free of an authn dependency).
package rbac

import (
	"context"

	"github.com/guidomantilla/yarumo/security/authz"
)

var (
	_ authz.Policy = (*policy)(nil)
)

// PrincipalIDResolverFn is the function type for extracting a principal
// id from authz.Request.Principal. The id is then used to look up
// roles via the RolesStore. The resolver returns (id, true) when the
// principal carries a usable id; (_, false) signals "no id available"
// and the engine treats the request as a deny.
type PrincipalIDResolverFn func(principal any) (string, bool)

// RolesStore defines the interface for resolving the roles a principal
// holds. Implementations must be safe for concurrent use by multiple
// goroutines.
type RolesStore interface {
	// Roles returns the role names assigned to the principal id. The
	// returned slice MUST NOT be retained or mutated by the caller.
	// The error path is reserved for transport-bound stores; the
	// in-memory impl never returns a non-nil error.
	Roles(ctx context.Context, principalID string) ([]string, error)
}
