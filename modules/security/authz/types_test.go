package authz

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestEffect_Constants(t *testing.T) {
	t.Parallel()

	t.Run("allow constant", func(t *testing.T) {
		t.Parallel()

		if EffectAllow != "allow" {
			t.Fatalf("expected 'allow', got %q", EffectAllow)
		}
	})

	t.Run("deny constant", func(t *testing.T) {
		t.Parallel()

		if EffectDeny != "deny" {
			t.Fatalf("expected 'deny', got %q", EffectDeny)
		}
	})

	t.Run("abstain constant", func(t *testing.T) {
		t.Parallel()

		if EffectAbstain != "abstain" {
			t.Fatalf("expected 'abstain', got %q", EffectAbstain)
		}
	})
}

func TestPrincipalReaderFn_Read(t *testing.T) {
	t.Parallel()

	t.Run("delegates to wrapped function", func(t *testing.T) {
		t.Parallel()

		want := "alice"

		fn := PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return want, true
		})

		got, ok := fn.Read(context.Background())
		if !ok {
			t.Fatal("expected ok=true")
		}

		s, isStr := got.(string)
		if !isStr {
			t.Fatalf("expected string, got %T", got)
		}

		if s != want {
			t.Fatalf("expected %q, got %q", want, s)
		}
	})

	t.Run("nil receiver returns false", func(t *testing.T) {
		t.Parallel()

		var fn PrincipalReaderFn

		got, ok := fn.Read(context.Background())
		if ok {
			t.Fatal("expected ok=false for nil receiver")
		}

		if got != nil {
			t.Fatalf("expected nil principal, got %#v", got)
		}
	})

	t.Run("wrapper returning false", func(t *testing.T) {
		t.Parallel()

		fn := PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return nil, false
		})

		_, ok := fn.Read(context.Background())
		if ok {
			t.Fatal("expected ok=false")
		}
	})
}

func TestRequest_StructFields(t *testing.T) {
	t.Parallel()

	t.Run("zero value is usable", func(t *testing.T) {
		t.Parallel()

		var r Request
		if r.Principal != nil {
			t.Fatalf("expected nil principal, got %#v", r.Principal)
		}

		if r.Action != "" {
			t.Fatalf("expected empty action, got %q", r.Action)
		}
	})

	t.Run("populated request", func(t *testing.T) {
		t.Parallel()

		now := time.Now()

		r := Request{
			Principal:   "alice",
			Action:      "read",
			Resource:    Resource{Type: "orders", ID: "123"},
			Environment: Environment{IP: net.ParseIP("10.0.0.1"), Time: now},
		}

		if r.Resource.Type != "orders" {
			t.Fatalf("expected 'orders', got %q", r.Resource.Type)
		}

		if !r.Environment.Time.Equal(now) {
			t.Fatalf("expected time %v, got %v", now, r.Environment.Time)
		}
	})
}

func TestDecision_StructFields(t *testing.T) {
	t.Parallel()

	t.Run("allow decision", func(t *testing.T) {
		t.Parallel()

		d := Decision{Effect: EffectAllow, Reason: "ok"}
		if d.Effect != EffectAllow {
			t.Fatalf("expected EffectAllow, got %q", d.Effect)
		}
	})

	t.Run("with metadata", func(t *testing.T) {
		t.Parallel()

		d := Decision{
			Effect:   EffectAllow,
			Metadata: map[string]any{"role": "admin"},
		}

		v, ok := d.Metadata["role"]
		if !ok {
			t.Fatal("expected metadata to contain 'role'")
		}

		s, isStr := v.(string)
		if !isStr || s != "admin" {
			t.Fatalf("expected 'admin', got %#v", v)
		}
	})
}
