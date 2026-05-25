package authn_test

import (
	"context"
	"testing"

	"github.com/guidomantilla/yarumo/security/authn"
)

func TestPrincipal_Fields(t *testing.T) {
	t.Parallel()

	p := &authn.Principal{
		ID:         "user-1",
		Name:       "Alice",
		Roles:      []string{"admin"},
		Attributes: map[string]any{"tenant": "acme"},
	}

	if p.ID != "user-1" {
		t.Fatalf("ID = %q, want %q", p.ID, "user-1")
	}

	if p.Name != "Alice" {
		t.Fatalf("Name = %q, want %q", p.Name, "Alice")
	}

	if len(p.Roles) != 1 || p.Roles[0] != "admin" {
		t.Fatalf("Roles = %v", p.Roles)
	}

	tenant, ok := p.Attributes["tenant"]
	if !ok || tenant != "acme" {
		t.Fatalf("Attributes[tenant] = %v, ok = %v", tenant, ok)
	}
}

func TestWithPrincipal(t *testing.T) {
	t.Parallel()

	t.Run("happy path stores principal", func(t *testing.T) {
		t.Parallel()

		p := &authn.Principal{ID: "u-1"}
		ctx := authn.WithPrincipal(context.Background(), p)

		got, ok := authn.FromContext(ctx)
		if !ok {
			t.Fatal("FromContext = ok=false, want ok=true")
		}

		if got != p {
			t.Fatalf("FromContext returned %v, want %v", got, p)
		}
	})

	t.Run("nil principal returns ctx unchanged", func(t *testing.T) {
		t.Parallel()

		parent := context.Background()
		ctx := authn.WithPrincipal(parent, nil)

		_, ok := authn.FromContext(ctx)
		if ok {
			t.Fatal("FromContext = ok=true after WithPrincipal(nil), want ok=false")
		}
	})

	t.Run("nil ctx returns nil ctx", func(t *testing.T) {
		t.Parallel()

		//nolint:staticcheck // intentional nil ctx to validate guard
		ctx := authn.WithPrincipal(nil, &authn.Principal{ID: "u"})
		if ctx != nil {
			t.Fatalf("WithPrincipal(nil, _) = %v, want nil", ctx)
		}
	})
}

func TestFromContext(t *testing.T) {
	t.Parallel()

	t.Run("missing principal returns false", func(t *testing.T) {
		t.Parallel()

		got, ok := authn.FromContext(context.Background())
		if ok {
			t.Fatal("FromContext on empty ctx = ok=true, want ok=false")
		}

		if got != nil {
			t.Fatalf("FromContext on empty ctx returned %v, want nil", got)
		}
	})

	t.Run("nil ctx returns false", func(t *testing.T) {
		t.Parallel()

		//nolint:staticcheck // intentional nil ctx to validate guard
		got, ok := authn.FromContext(nil)
		if ok {
			t.Fatal("FromContext(nil) = ok=true, want ok=false")
		}

		if got != nil {
			t.Fatalf("FromContext(nil) returned %v, want nil", got)
		}
	})

	t.Run("wrong value type returns false", func(t *testing.T) {
		t.Parallel()

		// Manually inject a value under the same key shape via a
		// different package is impossible (key type is unexported);
		// emulate by storing through WithValue with a string key. The
		// principalCtxKey path remains the only entry point, so
		// FromContext correctly reports absence.
		ctx := context.WithValue(context.Background(), struct{ name string }{name: "principal"}, "not-a-principal")

		got, ok := authn.FromContext(ctx)
		if ok {
			t.Fatal("FromContext with collateral key = ok=true, want ok=false")
		}

		if got != nil {
			t.Fatalf("FromContext returned %v, want nil", got)
		}
	})
}
