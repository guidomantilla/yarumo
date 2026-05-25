package rbac

import (
	"context"
	"sort"
	"strings"

	"github.com/guidomantilla/yarumo/security/authz"
)

// policy implements authz.Policy as an RBAC engine. It is built once
// (NewPolicy) from a configuration of role permissions + inheritance
// edges + a RolesStore, then evaluated repeatedly against
// authz.Request values.
//
// The Evaluate method:
//   1. Resolves a principal id via PrincipalIDResolverFn.
//   2. Looks up roles via RolesStore.
//   3. Expands roles through the inheritance closure (cached).
//   4. Builds the canonical permission "<resource_type>.<action>"
//      from the Request and matches against every effective role's
//      permissions with wildcard support.
//   5. Invokes the audit hook with the final Decision.
type policy struct {
	rolePermissions     map[string][]string
	closure             map[string][]string
	store               RolesStore
	principalIDResolver PrincipalIDResolverFn
	auditHook           authz.AuditHookFn
	configError         error
}

// NewPolicy returns an authz.Policy configured via the given Options.
//
// Inheritance cycles cause NewPolicy to install a deny-only fallback
// policy: every Evaluate returns Deny("rbac configuration invalid:
// inheritance cycle") and the audit hook surfaces the failure. This
// keeps the constructor non-fallible (no error return) while still
// failing closed loudly at evaluation time.
//
// Direct consumers of authz.Policy use this constructor; the engine
// implements authz.Policy so it can be chained via
// authz.ChainPolicies with other policies.
func NewPolicy(opts ...Option) authz.Policy {
	options := NewOptions(opts...)

	closure, err := buildClosure(options.hierarchy)

	return &policy{
		rolePermissions:     normalizePermissions(options.rolePermissions),
		closure:             closure,
		store:               options.store,
		principalIDResolver: options.principalIDResolver,
		auditHook:           options.auditHook,
		configError:         err,
	}
}

// Evaluate runs the RBAC check described in the type doc-comment.
func (p *policy) Evaluate(ctx context.Context, req authz.Request) authz.Decision {
	if p.configError != nil {
		dec := authz.Deny("rbac configuration invalid: " + p.configError.Error())
		p.auditHook(ctx, req, dec)

		return dec
	}

	principalID, ok := p.principalIDResolver(req.Principal)
	if !ok || principalID == "" {
		dec := authz.Deny("rbac: principal id not resolvable")
		p.auditHook(ctx, req, dec)

		return dec
	}

	assigned, storeErr := p.store.Roles(ctx, principalID)
	if storeErr != nil {
		dec := authz.Deny("rbac: roles store error: " + storeErr.Error())
		p.auditHook(ctx, req, dec)

		return dec
	}

	if len(assigned) == 0 {
		dec := authz.Deny("rbac: principal has no roles")
		p.auditHook(ctx, req, dec)

		return dec
	}

	effective := p.effectiveRoles(assigned)
	wanted := canonicalPermission(req.Resource.Type, req.Action)

	matched, role := p.match(effective, wanted)
	if matched {
		dec := authz.Allow("rbac: granted by role " + role)
		dec.Metadata = map[string]any{
			"role":       role,
			"permission": wanted,
		}
		p.auditHook(ctx, req, dec)

		return dec
	}

	dec := authz.Deny("rbac: no role grants " + wanted)
	dec.Metadata = map[string]any{
		"permission":      wanted,
		"effective_roles": effective,
	}
	p.auditHook(ctx, req, dec)

	return dec
}

// effectiveRoles expands the assigned roles through the inheritance
// closure and returns the deduplicated, alphabetized set.
func (p *policy) effectiveRoles(assigned []string) []string {
	set := map[string]struct{}{}

	for _, r := range assigned {
		set[r] = struct{}{}

		for _, ancestor := range p.closure[r] {
			set[ancestor] = struct{}{}
		}
	}

	out := make([]string, 0, len(set))
	for r := range set {
		out = append(out, r)
	}

	sort.Strings(out)

	return out
}

// match returns (true, role) for the first role whose permission set
// covers wanted under wildcard rules, or (false, "") when no role
// matches.
func (p *policy) match(effective []string, wanted string) (bool, string) {
	for _, role := range effective {
		perms := p.rolePermissions[role]
		for _, perm := range perms {
			if permissionMatches(perm, wanted) {
				return true, role
			}
		}
	}

	return false, ""
}

// canonicalPermission encodes the resource type and action into the
// matching string "<resource_type>.<action>". Empty resource type is
// encoded as "*"; empty action is encoded as "*".
func canonicalPermission(resourceType, action string) string {
	if resourceType == "" {
		resourceType = "*"
	}

	if action == "" {
		action = "*"
	}

	return resourceType + "." + action
}

// permissionMatches reports whether grant covers wanted under the
// wildcard rules. "*" alone matches everything; segment-level "*"
// matches every value of that segment ("orders.*" matches
// "orders.read", "orders.write", etc.). Both grant and wanted are
// expected to be canonical "<segment>.<segment>" or the special "*".
func permissionMatches(grant, wanted string) bool {
	if grant == wanted {
		return true
	}

	if grant == "*" {
		return true
	}

	grantParts := strings.SplitN(grant, ".", 2)
	wantedParts := strings.SplitN(wanted, ".", 2)

	if len(grantParts) != 2 || len(wantedParts) != 2 {
		return false
	}

	return segmentMatches(grantParts[0], wantedParts[0]) &&
		segmentMatches(grantParts[1], wantedParts[1])
}

// segmentMatches reports whether grant covers wanted at the segment
// level. "*" matches every value; otherwise the comparison is exact.
func segmentMatches(grant, wanted string) bool {
	return grant == "*" || grant == wanted
}

// normalizePermissions deduplicates and lowercases nothing — the
// engine treats permission strings as opaque labels — but it does
// collapse duplicate entries to keep match() loops tight.
func normalizePermissions(in map[string][]string) map[string][]string {
	out := make(map[string][]string, len(in))

	for role, perms := range in {
		seen := map[string]struct{}{}
		kept := make([]string, 0, len(perms))

		for _, p := range perms {
			_, dup := seen[p]
			if dup {
				continue
			}

			seen[p] = struct{}{}

			kept = append(kept, p)
		}

		out[role] = kept
	}

	return out
}
