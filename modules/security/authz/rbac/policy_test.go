package rbac

import (
	"context"
	"errors"
	"testing"

	"github.com/guidomantilla/yarumo/security/authz"
)

// stringIDResolver extracts the id from a principal that's already a string.
func stringIDResolver(p any) (string, bool) {
	s, ok := p.(string)
	if !ok {
		return "", false
	}

	return s, true
}

func TestNewPolicy_AllowSingleRole(t *testing.T) {
	t.Parallel()

	t.Run("role with matching permission grants access", func(t *testing.T) {
		t.Parallel()

		store := NewInMemoryRolesStore()
		store.Assign("alice", "viewer")

		p := NewPolicy(
			WithRolePermissions("viewer", "orders.read"),
			WithRolesStore(store),
			WithPrincipalIDResolver(stringIDResolver),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := authz.NewRequest("alice", "read", authz.Resource{Type: "orders", ID: "1"}, authz.Environment{})
		dec := p.Evaluate(context.Background(), req)

		if dec.Effect != authz.EffectAllow {
			t.Fatalf("expected EffectAllow, got %q (reason=%q)", dec.Effect, dec.Reason)
		}
	})
}

func TestNewPolicy_DenyNoMatchingPermission(t *testing.T) {
	t.Parallel()

	t.Run("role without permission denied", func(t *testing.T) {
		t.Parallel()

		store := NewInMemoryRolesStore()
		store.Assign("alice", "viewer")

		p := NewPolicy(
			WithRolePermissions("viewer", "orders.read"),
			WithRolesStore(store),
			WithPrincipalIDResolver(stringIDResolver),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := authz.NewRequest("alice", "write", authz.Resource{Type: "orders"}, authz.Environment{})
		dec := p.Evaluate(context.Background(), req)

		if dec.Effect != authz.EffectDeny {
			t.Fatalf("expected EffectDeny, got %q", dec.Effect)
		}
	})
}

func TestNewPolicy_NoPrincipalIDResolverDenies(t *testing.T) {
	t.Parallel()

	t.Run("default resolver denies", func(t *testing.T) {
		t.Parallel()

		p := NewPolicy(
			WithRolePermissions("admin", "*"),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := authz.NewRequest("alice", "read", authz.Resource{Type: "orders"}, authz.Environment{})
		dec := p.Evaluate(context.Background(), req)

		if dec.Effect != authz.EffectDeny {
			t.Fatalf("expected EffectDeny, got %q", dec.Effect)
		}
	})

	t.Run("empty id denies", func(t *testing.T) {
		t.Parallel()

		p := NewPolicy(
			WithRolePermissions("admin", "*"),
			WithPrincipalIDResolver(func(_ any) (string, bool) {
				return "", true
			}),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := authz.NewRequest("alice", "read", authz.Resource{Type: "orders"}, authz.Environment{})
		dec := p.Evaluate(context.Background(), req)

		if dec.Effect != authz.EffectDeny {
			t.Fatalf("expected EffectDeny on empty id, got %q", dec.Effect)
		}
	})
}

func TestNewPolicy_PrincipalWithNoRolesDenies(t *testing.T) {
	t.Parallel()

	store := NewInMemoryRolesStore()

	p := NewPolicy(
		WithRolePermissions("admin", "*"),
		WithRolesStore(store),
		WithPrincipalIDResolver(stringIDResolver),
		WithAuditHook(authz.SilentAuditHook),
	)

	req := authz.NewRequest("nobody", "read", authz.Resource{Type: "orders"}, authz.Environment{})
	dec := p.Evaluate(context.Background(), req)

	if dec.Effect != authz.EffectDeny {
		t.Fatalf("expected EffectDeny, got %q", dec.Effect)
	}
}

func TestNewPolicy_RoleInheritance(t *testing.T) {
	t.Parallel()

	t.Run("admin inherits editor permissions", func(t *testing.T) {
		t.Parallel()

		store := NewInMemoryRolesStore()
		store.Assign("alice", "admin")

		p := NewPolicy(
			WithRolePermissions("viewer", "orders.read"),
			WithRolePermissions("editor", "orders.write"),
			WithRolePermissions("admin", "orders.delete"),
			WithInheritance("editor", "viewer"),
			WithInheritance("admin", "editor"),
			WithRolesStore(store),
			WithPrincipalIDResolver(stringIDResolver),
			WithAuditHook(authz.SilentAuditHook),
		)

		// admin should inherit read (from viewer via editor) and write (from editor).
		req := authz.NewRequest("alice", "read", authz.Resource{Type: "orders"}, authz.Environment{})
		dec := p.Evaluate(context.Background(), req)

		if dec.Effect != authz.EffectAllow {
			t.Fatalf("expected EffectAllow for inherited read, got %q (reason=%q)", dec.Effect, dec.Reason)
		}
	})

	t.Run("transitive inheritance works", func(t *testing.T) {
		t.Parallel()

		store := NewInMemoryRolesStore()
		store.Assign("bob", "admin")

		p := NewPolicy(
			WithRolePermissions("viewer", "orders.read"),
			WithRolePermissions("editor", "orders.write"),
			WithRolePermissions("admin", "orders.delete"),
			WithInheritance("editor", "viewer"),
			WithInheritance("admin", "editor"),
			WithRolesStore(store),
			WithPrincipalIDResolver(stringIDResolver),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := authz.NewRequest("bob", "delete", authz.Resource{Type: "orders"}, authz.Environment{})
		dec := p.Evaluate(context.Background(), req)

		if dec.Effect != authz.EffectAllow {
			t.Fatalf("expected EffectAllow on own permission, got %q", dec.Effect)
		}
	})

	t.Run("multiple parents", func(t *testing.T) {
		t.Parallel()

		store := NewInMemoryRolesStore()
		store.Assign("carol", "manager")

		p := NewPolicy(
			WithRolePermissions("audit", "logs.read"),
			WithRolePermissions("billing", "invoices.read"),
			WithInheritance("manager", "audit", "billing"),
			WithRolesStore(store),
			WithPrincipalIDResolver(stringIDResolver),
			WithAuditHook(authz.SilentAuditHook),
		)

		req1 := authz.NewRequest("carol", "read", authz.Resource{Type: "logs"}, authz.Environment{})
		dec1 := p.Evaluate(context.Background(), req1)

		if dec1.Effect != authz.EffectAllow {
			t.Fatalf("expected manager inherits from audit, got %q", dec1.Effect)
		}

		req2 := authz.NewRequest("carol", "read", authz.Resource{Type: "invoices"}, authz.Environment{})
		dec2 := p.Evaluate(context.Background(), req2)

		if dec2.Effect != authz.EffectAllow {
			t.Fatalf("expected manager inherits from billing, got %q", dec2.Effect)
		}
	})
}

func TestNewPolicy_Wildcards(t *testing.T) {
	t.Parallel()

	t.Run("star matches everything", func(t *testing.T) {
		t.Parallel()

		store := NewInMemoryRolesStore()
		store.Assign("root", "superuser")

		p := NewPolicy(
			WithRolePermissions("superuser", "*"),
			WithRolesStore(store),
			WithPrincipalIDResolver(stringIDResolver),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := authz.NewRequest("root", "destroy", authz.Resource{Type: "anything"}, authz.Environment{})
		dec := p.Evaluate(context.Background(), req)

		if dec.Effect != authz.EffectAllow {
			t.Fatalf("expected EffectAllow for *, got %q", dec.Effect)
		}
	})

	t.Run("resource wildcard", func(t *testing.T) {
		t.Parallel()

		store := NewInMemoryRolesStore()
		store.Assign("alice", "manager")

		p := NewPolicy(
			WithRolePermissions("manager", "orders.*"),
			WithRolesStore(store),
			WithPrincipalIDResolver(stringIDResolver),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := authz.NewRequest("alice", "delete", authz.Resource{Type: "orders"}, authz.Environment{})
		dec := p.Evaluate(context.Background(), req)

		if dec.Effect != authz.EffectAllow {
			t.Fatalf("expected EffectAllow for orders.*, got %q", dec.Effect)
		}
	})

	t.Run("action wildcard", func(t *testing.T) {
		t.Parallel()

		store := NewInMemoryRolesStore()
		store.Assign("alice", "reader")

		p := NewPolicy(
			WithRolePermissions("reader", "*.read"),
			WithRolesStore(store),
			WithPrincipalIDResolver(stringIDResolver),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := authz.NewRequest("alice", "read", authz.Resource{Type: "logs"}, authz.Environment{})
		dec := p.Evaluate(context.Background(), req)

		if dec.Effect != authz.EffectAllow {
			t.Fatalf("expected EffectAllow for *.read, got %q", dec.Effect)
		}
	})

	t.Run("wildcard does not match wrong segment", func(t *testing.T) {
		t.Parallel()

		store := NewInMemoryRolesStore()
		store.Assign("alice", "reader")

		p := NewPolicy(
			WithRolePermissions("reader", "*.read"),
			WithRolesStore(store),
			WithPrincipalIDResolver(stringIDResolver),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := authz.NewRequest("alice", "write", authz.Resource{Type: "logs"}, authz.Environment{})
		dec := p.Evaluate(context.Background(), req)

		if dec.Effect != authz.EffectDeny {
			t.Fatalf("expected EffectDeny for write on *.read role, got %q", dec.Effect)
		}
	})
}

func TestNewPolicy_DecisionMetadata(t *testing.T) {
	t.Parallel()

	t.Run("allow includes matched role", func(t *testing.T) {
		t.Parallel()

		store := NewInMemoryRolesStore()
		store.Assign("alice", "viewer")

		p := NewPolicy(
			WithRolePermissions("viewer", "orders.read"),
			WithRolesStore(store),
			WithPrincipalIDResolver(stringIDResolver),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := authz.NewRequest("alice", "read", authz.Resource{Type: "orders"}, authz.Environment{})
		dec := p.Evaluate(context.Background(), req)

		role, ok := dec.Metadata["role"]
		if !ok {
			t.Fatal("expected metadata.role")
		}

		s, isStr := role.(string)
		if !isStr || s != "viewer" {
			t.Fatalf("expected 'viewer', got %#v", role)
		}
	})

	t.Run("deny includes attempted permission", func(t *testing.T) {
		t.Parallel()

		store := NewInMemoryRolesStore()
		store.Assign("alice", "viewer")

		p := NewPolicy(
			WithRolePermissions("viewer", "orders.read"),
			WithRolesStore(store),
			WithPrincipalIDResolver(stringIDResolver),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := authz.NewRequest("alice", "delete", authz.Resource{Type: "orders"}, authz.Environment{})
		dec := p.Evaluate(context.Background(), req)

		perm, ok := dec.Metadata["permission"]
		if !ok {
			t.Fatal("expected metadata.permission")
		}

		s, isStr := perm.(string)
		if !isStr || s != "orders.delete" {
			t.Fatalf("expected 'orders.delete', got %#v", perm)
		}
	})
}

// failingStore returns an error from Roles for cycle / error-path
// coverage.
type failingStore struct {
	err error
}

func (s failingStore) Roles(_ context.Context, _ string) ([]string, error) {
	return nil, s.err
}

func TestNewPolicy_StoreError(t *testing.T) {
	t.Parallel()

	t.Run("denies on store error", func(t *testing.T) {
		t.Parallel()

		p := NewPolicy(
			WithRolesStore(failingStore{err: errors.New("connection lost")}),
			WithPrincipalIDResolver(stringIDResolver),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := authz.NewRequest("alice", "read", authz.Resource{Type: "orders"}, authz.Environment{})
		dec := p.Evaluate(context.Background(), req)

		if dec.Effect != authz.EffectDeny {
			t.Fatalf("expected EffectDeny on store error, got %q", dec.Effect)
		}
	})
}

func TestNewPolicy_InheritanceCycleDeniesAll(t *testing.T) {
	t.Parallel()

	t.Run("cyclic config denies", func(t *testing.T) {
		t.Parallel()

		store := NewInMemoryRolesStore()
		store.Assign("alice", "a")

		p := NewPolicy(
			WithRolePermissions("a", "orders.read"),
			WithInheritance("a", "b"),
			WithInheritance("b", "a"),
			WithRolesStore(store),
			WithPrincipalIDResolver(stringIDResolver),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := authz.NewRequest("alice", "read", authz.Resource{Type: "orders"}, authz.Environment{})
		dec := p.Evaluate(context.Background(), req)

		if dec.Effect != authz.EffectDeny {
			t.Fatalf("expected EffectDeny, got %q (reason=%q)", dec.Effect, dec.Reason)
		}
	})
}

func TestNewPolicy_AuditHookFires(t *testing.T) {
	t.Parallel()

	t.Run("hook receives decision", func(t *testing.T) {
		t.Parallel()

		store := NewInMemoryRolesStore()
		store.Assign("alice", "viewer")

		var calls int
		var lastDec authz.Decision

		hook := authz.AuditHookFn(func(_ context.Context, _ authz.Request, d authz.Decision) {
			calls++
			lastDec = d
		})

		p := NewPolicy(
			WithRolePermissions("viewer", "orders.read"),
			WithRolesStore(store),
			WithPrincipalIDResolver(stringIDResolver),
			WithAuditHook(hook),
		)

		req := authz.NewRequest("alice", "read", authz.Resource{Type: "orders"}, authz.Environment{})
		_ = p.Evaluate(context.Background(), req)

		if calls != 1 {
			t.Fatalf("expected 1 hook call, got %d", calls)
		}

		if lastDec.Effect != authz.EffectAllow {
			t.Fatalf("expected EffectAllow recorded, got %q", lastDec.Effect)
		}
	})
}

func TestNewPolicy_PermissionDeduplication(t *testing.T) {
	t.Parallel()

	t.Run("duplicate permissions still allow", func(t *testing.T) {
		t.Parallel()

		store := NewInMemoryRolesStore()
		store.Assign("alice", "viewer")

		p := NewPolicy(
			WithRolePermissions("viewer", "orders.read", "orders.read"),
			WithRolePermissions("viewer", "orders.read"),
			WithRolesStore(store),
			WithPrincipalIDResolver(stringIDResolver),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := authz.NewRequest("alice", "read", authz.Resource{Type: "orders"}, authz.Environment{})
		dec := p.Evaluate(context.Background(), req)

		if dec.Effect != authz.EffectAllow {
			t.Fatalf("expected EffectAllow, got %q", dec.Effect)
		}
	})
}
