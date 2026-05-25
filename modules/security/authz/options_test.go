package authz

import (
	"context"
	"net/http"
	"reflect"
	"testing"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.auditHook == nil {
			t.Fatal("expected DefaultAuditHook installed")
		}

		if opts.principalReader != nil {
			t.Fatal("expected principalReader nil by default")
		}

		if opts.httpResourceFn != nil {
			t.Fatal("expected httpResourceFn nil by default")
		}

		if opts.grpcResourceFn != nil {
			t.Fatal("expected grpcResourceFn nil by default")
		}
	})

	t.Run("default audit hook is DefaultAuditHook", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		want := reflect.ValueOf(DefaultAuditHook).Pointer()
		got := reflect.ValueOf(opts.auditHook).Pointer()

		if got != want {
			t.Fatal("expected DefaultAuditHook installed by default")
		}
	})
}

func TestWithPrincipalReader(t *testing.T) {
	t.Parallel()

	t.Run("installs custom reader", func(t *testing.T) {
		t.Parallel()

		reader := PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return "alice", true
		})

		opts := NewOptions(WithPrincipalReader(reader))
		if opts.principalReader == nil {
			t.Fatal("expected reader installed")
		}

		got, ok := opts.principalReader.Read(context.Background())
		if !ok || got != "alice" {
			t.Fatalf("expected alice/true, got %v/%v", got, ok)
		}
	})

	t.Run("nil reader ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPrincipalReader(nil))
		if opts.principalReader != nil {
			t.Fatal("expected nil reader preserved as nil")
		}
	})
}

func TestWithAuditHook(t *testing.T) {
	t.Parallel()

	t.Run("installs custom hook", func(t *testing.T) {
		t.Parallel()

		var called bool

		custom := AuditHookFn(func(_ context.Context, _ Request, _ Decision) {
			called = true
		})

		opts := NewOptions(WithAuditHook(custom))
		opts.auditHook(context.Background(), Request{}, Decision{})

		if !called {
			t.Fatal("expected custom hook called")
		}
	})

	t.Run("nil hook preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithAuditHook(nil))
		want := reflect.ValueOf(DefaultAuditHook).Pointer()
		got := reflect.ValueOf(opts.auditHook).Pointer()

		if got != want {
			t.Fatal("expected DefaultAuditHook preserved when nil passed")
		}
	})

	t.Run("silent hook installable", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithAuditHook(SilentAuditHook))
		want := reflect.ValueOf(SilentAuditHook).Pointer()
		got := reflect.ValueOf(opts.auditHook).Pointer()

		if got != want {
			t.Fatal("expected SilentAuditHook installed")
		}
	})
}

func TestWithHTTPResourceResolver(t *testing.T) {
	t.Parallel()

	t.Run("installs custom resolver", func(t *testing.T) {
		t.Parallel()

		resolver := HTTPResourceResolverFn(func(_ *http.Request) Resource {
			return Resource{Type: "orders"}
		})

		opts := NewOptions(WithHTTPResourceResolver(resolver))
		if opts.httpResourceFn == nil {
			t.Fatal("expected resolver installed")
		}

		r, _ := http.NewRequest(http.MethodGet, "http://x/", http.NoBody)
		out := opts.httpResourceFn(r)
		if out.Type != "orders" {
			t.Fatalf("expected 'orders', got %q", out.Type)
		}
	})

	t.Run("nil resolver ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithHTTPResourceResolver(nil))
		if opts.httpResourceFn != nil {
			t.Fatal("expected nil resolver preserved")
		}
	})
}

func TestWithGRPCResourceResolver(t *testing.T) {
	t.Parallel()

	t.Run("installs custom resolver", func(t *testing.T) {
		t.Parallel()

		resolver := GRPCResourceResolverFn(func(_ context.Context, _ string, _ any) Resource {
			return Resource{Type: "rpc"}
		})

		opts := NewOptions(WithGRPCResourceResolver(resolver))
		if opts.grpcResourceFn == nil {
			t.Fatal("expected resolver installed")
		}

		out := opts.grpcResourceFn(context.Background(), "/svc/Method", nil)
		if out.Type != "rpc" {
			t.Fatalf("expected 'rpc', got %q", out.Type)
		}
	})

	t.Run("nil resolver ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithGRPCResourceResolver(nil))
		if opts.grpcResourceFn != nil {
			t.Fatal("expected nil resolver preserved")
		}
	})
}
