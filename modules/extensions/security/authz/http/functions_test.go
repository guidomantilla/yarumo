package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/guidomantilla/yarumo/security/authz"
)

// allowPolicy returns Allow for any request.
type allowPolicy struct{}

func (allowPolicy) Evaluate(_ context.Context, _ authz.Request) authz.Decision {
	return authz.Allow("ok")
}

// denyPolicy returns Deny with a fixed reason.
type denyPolicy struct {
	reason string
}

func (p denyPolicy) Evaluate(_ context.Context, _ authz.Request) authz.Decision {
	return authz.Deny(p.reason)
}

// captureRequestPolicy stores the last Request passed to Evaluate and
// returns Allow.
type captureRequestPolicy struct {
	last authz.Request
}

func (p *captureRequestPolicy) Evaluate(_ context.Context, req authz.Request) authz.Decision {
	p.last = req

	return authz.Allow("captured")
}

func TestRequireHTTP_Allow(t *testing.T) {
	t.Parallel()

	t.Run("invokes next handler on allow", func(t *testing.T) {
		t.Parallel()

		reader := authz.PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return "alice", true
		})

		var called bool

		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		mw := RequireHTTP(allowPolicy{}, "read",
			WithPrincipalReader(reader),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()

		mw(next).ServeHTTP(rec, req)

		if !called {
			t.Fatal("expected next handler called")
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})
}

func TestRequireHTTP_Deny(t *testing.T) {
	t.Parallel()

	t.Run("403 on deny", func(t *testing.T) {
		t.Parallel()

		reader := authz.PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return "alice", true
		})

		var called bool

		next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			called = true
		})

		mw := RequireHTTP(denyPolicy{reason: "not allowed"}, "read",
			WithPrincipalReader(reader),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()

		mw(next).ServeHTTP(rec, req)

		if called {
			t.Fatal("expected next handler NOT called")
		}

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rec.Code)
		}

		gotReason := rec.Header().Get("X-Authz-Reason")
		if gotReason != "not allowed" {
			t.Fatalf("expected X-Authz-Reason 'not allowed', got %q", gotReason)
		}

		body, _ := io.ReadAll(rec.Body)
		bodyDecoded := map[string]string{}
		_ = json.Unmarshal(body, &bodyDecoded)

		if bodyDecoded["reason"] != "not allowed" {
			t.Fatalf("expected reason 'not allowed' in body, got %q", bodyDecoded["reason"])
		}
	})

	t.Run("403 on missing principal", func(t *testing.T) {
		t.Parallel()

		mw := RequireHTTP(allowPolicy{}, "read",
			WithAuditHook(authz.SilentAuditHook),
		)

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()

		mw(http.NotFoundHandler()).ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("403 when reader returns false", func(t *testing.T) {
		t.Parallel()

		reader := authz.PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return nil, false
		})

		mw := RequireHTTP(allowPolicy{}, "read",
			WithPrincipalReader(reader),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()

		mw(http.NotFoundHandler()).ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rec.Code)
		}
	})
}

func TestRequireHTTP_AuditHookCalled(t *testing.T) {
	t.Parallel()

	t.Run("hook fires once on allow", func(t *testing.T) {
		t.Parallel()

		reader := authz.PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return "alice", true
		})

		var calls int

		hook := authz.AuditHookFn(func(_ context.Context, _ authz.Request, _ authz.Decision) {
			calls++
		})

		mw := RequireHTTP(allowPolicy{}, "read",
			WithPrincipalReader(reader),
			WithAuditHook(hook),
		)

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()

		mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(rec, req)

		if calls != 1 {
			t.Fatalf("expected 1 hook call, got %d", calls)
		}
	})

	t.Run("hook fires on missing principal", func(t *testing.T) {
		t.Parallel()

		var calls int
		var lastDec authz.Decision

		hook := authz.AuditHookFn(func(_ context.Context, _ authz.Request, d authz.Decision) {
			calls++
			lastDec = d
		})

		mw := RequireHTTP(allowPolicy{}, "read",
			WithAuditHook(hook),
		)

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()

		mw(http.NotFoundHandler()).ServeHTTP(rec, req)

		if calls != 1 {
			t.Fatalf("expected 1 hook call, got %d", calls)
		}

		if lastDec.Effect != authz.EffectDeny {
			t.Fatalf("expected EffectDeny, got %q", lastDec.Effect)
		}
	})
}

func TestRequireHTTP_ResourceResolver(t *testing.T) {
	t.Parallel()

	t.Run("resolver populates resource", func(t *testing.T) {
		t.Parallel()

		reader := authz.PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return "alice", true
		})

		policy := &captureRequestPolicy{}

		resolver := HTTPResourceResolverFn(func(r *http.Request) authz.Resource {
			return authz.Resource{Type: "orders", ID: strings.TrimPrefix(r.URL.Path, "/orders/")}
		})

		mw := RequireHTTP(policy, "read",
			WithPrincipalReader(reader),
			WithResourceResolver(resolver),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := httptest.NewRequest(http.MethodGet, "/orders/42", http.NoBody)
		rec := httptest.NewRecorder()

		mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(rec, req)

		if policy.last.Resource.Type != "orders" {
			t.Fatalf("expected 'orders', got %q", policy.last.Resource.Type)
		}

		if policy.last.Resource.ID != "42" {
			t.Fatalf("expected '42', got %q", policy.last.Resource.ID)
		}
	})
}

func TestRequireHTTP_ConstructorFailsClosed(t *testing.T) {
	t.Parallel()

	t.Run("panics on nil policy", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic")
			}
		}()

		_ = RequireHTTP(nil, "read")
	})

	t.Run("panics on empty action", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic")
			}
		}()

		_ = RequireHTTP(allowPolicy{}, "")
	})
}

func TestRequireHTTP_XForwardedFor(t *testing.T) {
	t.Parallel()

	t.Run("uses first X-Forwarded-For hop", func(t *testing.T) {
		t.Parallel()

		reader := authz.PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return "alice", true
		})

		policy := &captureRequestPolicy{}

		mw := RequireHTTP(policy, "read",
			WithPrincipalReader(reader),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("X-Forwarded-For", "10.0.0.5, 10.0.0.6")

		rec := httptest.NewRecorder()
		mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(rec, req)

		if policy.last.Environment.IP == nil {
			t.Fatal("expected non-nil IP")
		}

		if policy.last.Environment.IP.String() != "10.0.0.5" {
			t.Fatalf("expected 10.0.0.5, got %s", policy.last.Environment.IP)
		}
	})

	t.Run("falls back to RemoteAddr", func(t *testing.T) {
		t.Parallel()

		reader := authz.PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return "alice", true
		})

		policy := &captureRequestPolicy{}

		mw := RequireHTTP(policy, "read",
			WithPrincipalReader(reader),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.RemoteAddr = "192.168.0.1:12345"

		rec := httptest.NewRecorder()
		mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(rec, req)

		if policy.last.Environment.IP == nil {
			t.Fatal("expected non-nil IP")
		}

		if policy.last.Environment.IP.String() != "192.168.0.1" {
			t.Fatalf("expected 192.168.0.1, got %s", policy.last.Environment.IP)
		}
	})

	t.Run("ignores invalid X-Forwarded-For", func(t *testing.T) {
		t.Parallel()

		reader := authz.PrincipalReaderFn(func(_ context.Context) (any, bool) {
			return "alice", true
		})

		policy := &captureRequestPolicy{}

		mw := RequireHTTP(policy, "read",
			WithPrincipalReader(reader),
			WithAuditHook(authz.SilentAuditHook),
		)

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("X-Forwarded-For", "garbage")
		req.RemoteAddr = "192.168.0.2:9999"

		rec := httptest.NewRecorder()
		mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(rec, req)

		if policy.last.Environment.IP.String() != "192.168.0.2" {
			t.Fatalf("expected 192.168.0.2, got %s", policy.last.Environment.IP)
		}
	})
}
