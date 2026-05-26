package rbac

import (
	"context"
	"reflect"
	"testing"

	"github.com/guidomantilla/yarumo/security/authz"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.store == nil {
			t.Fatal("expected default RolesStore installed")
		}

		if opts.principalIDResolver == nil {
			t.Fatal("expected default principal resolver installed")
		}

		if opts.auditHook == nil {
			t.Fatal("expected default audit hook installed")
		}
	})

	t.Run("default principal resolver returns false", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()

		_, ok := opts.principalIDResolver("alice")
		if ok {
			t.Fatal("expected default resolver to return false")
		}
	})

	t.Run("default audit hook is authz.DefaultAuditHook", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		want := reflect.ValueOf(authz.DefaultAuditHook).Pointer()
		got := reflect.ValueOf(opts.auditHook).Pointer()

		if got != want {
			t.Fatal("expected authz.DefaultAuditHook installed")
		}
	})
}

func TestWithRolePermissions(t *testing.T) {
	t.Parallel()

	t.Run("registers permissions", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRolePermissions("viewer", "orders.read"))

		perms := opts.rolePermissions["viewer"]
		if len(perms) != 1 || perms[0] != "orders.read" {
			t.Fatalf("expected [orders.read], got %v", perms)
		}
	})

	t.Run("accumulates across calls", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(
			WithRolePermissions("viewer", "orders.read"),
			WithRolePermissions("viewer", "products.read"),
		)

		perms := opts.rolePermissions["viewer"]
		if len(perms) != 2 {
			t.Fatalf("expected 2 perms, got %v", perms)
		}
	})

	t.Run("empty role ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRolePermissions("", "orders.read"))

		if len(opts.rolePermissions) != 0 {
			t.Fatalf("expected empty map, got %v", opts.rolePermissions)
		}
	})

	t.Run("empty permissions filtered", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRolePermissions("viewer", "", "orders.read", ""))

		perms := opts.rolePermissions["viewer"]
		if len(perms) != 1 || perms[0] != "orders.read" {
			t.Fatalf("expected [orders.read], got %v", perms)
		}
	})

	t.Run("all-empty permissions ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRolePermissions("viewer", "", ""))

		_, ok := opts.rolePermissions["viewer"]
		if ok {
			t.Fatal("expected no entry for all-empty input")
		}
	})
}

func TestWithInheritance(t *testing.T) {
	t.Parallel()

	t.Run("registers parent", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithInheritance("admin", "editor"))

		parents := opts.hierarchy["admin"]
		if len(parents) != 1 || parents[0] != "editor" {
			t.Fatalf("expected [editor], got %v", parents)
		}
	})

	t.Run("multiple parents", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithInheritance("manager", "audit", "billing"))

		parents := opts.hierarchy["manager"]
		if len(parents) != 2 {
			t.Fatalf("expected 2 parents, got %v", parents)
		}
	})

	t.Run("empty child ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithInheritance("", "editor"))
		if len(opts.hierarchy) != 0 {
			t.Fatalf("expected empty hierarchy, got %v", opts.hierarchy)
		}
	})

	t.Run("empty parents filtered", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithInheritance("admin", "", "editor", ""))

		parents := opts.hierarchy["admin"]
		if len(parents) != 1 || parents[0] != "editor" {
			t.Fatalf("expected [editor], got %v", parents)
		}
	})

	t.Run("all-empty parents ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithInheritance("admin", "", ""))

		_, ok := opts.hierarchy["admin"]
		if ok {
			t.Fatal("expected no entry for all-empty input")
		}
	})
}

func TestWithRolesStore(t *testing.T) {
	t.Parallel()

	t.Run("overrides default store", func(t *testing.T) {
		t.Parallel()

		custom := NewInMemoryRolesStore()
		custom.Assign("alice", "viewer")

		opts := NewOptions(WithRolesStore(custom))

		roles, _ := opts.store.Roles(context.Background(), "alice")
		if len(roles) != 1 || roles[0] != "viewer" {
			t.Fatalf("expected custom store used, got %v", roles)
		}
	})

	t.Run("nil store ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRolesStore(nil))
		if opts.store == nil {
			t.Fatal("expected default store preserved when nil passed")
		}
	})
}

func TestWithPrincipalIDResolver(t *testing.T) {
	t.Parallel()

	t.Run("installs custom resolver", func(t *testing.T) {
		t.Parallel()

		resolver := PrincipalIDResolverFn(func(p any) (string, bool) {
			s, ok := p.(string)
			return s, ok
		})

		opts := NewOptions(WithPrincipalIDResolver(resolver))

		id, ok := opts.principalIDResolver("alice")
		if !ok || id != "alice" {
			t.Fatalf("expected alice/true, got %s/%v", id, ok)
		}
	})

	t.Run("nil resolver ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPrincipalIDResolver(nil))

		// Default resolver returns ("", false).
		_, ok := opts.principalIDResolver("alice")
		if ok {
			t.Fatal("expected default (false-returning) resolver preserved")
		}
	})
}

func TestWithAuditHook(t *testing.T) {
	t.Parallel()

	t.Run("installs custom hook", func(t *testing.T) {
		t.Parallel()

		var called bool

		custom := authz.AuditHookFn(func(_ context.Context, _ authz.Request, _ authz.Decision) {
			called = true
		})

		opts := NewOptions(WithAuditHook(custom))
		opts.auditHook(context.Background(), authz.Request{}, authz.Decision{})

		if !called {
			t.Fatal("expected custom hook called")
		}
	})

	t.Run("nil hook preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithAuditHook(nil))
		want := reflect.ValueOf(authz.DefaultAuditHook).Pointer()
		got := reflect.ValueOf(opts.auditHook).Pointer()

		if got != want {
			t.Fatal("expected DefaultAuditHook preserved")
		}
	})

	t.Run("silent hook installable", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithAuditHook(authz.SilentAuditHook))
		want := reflect.ValueOf(authz.SilentAuditHook).Pointer()
		got := reflect.ValueOf(opts.auditHook).Pointer()

		if got != want {
			t.Fatal("expected SilentAuditHook installed")
		}
	})
}
