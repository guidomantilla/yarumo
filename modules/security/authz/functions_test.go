package authz

import (
	"context"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestNewRequest(t *testing.T) {
	t.Parallel()

	t.Run("populates time when zero", func(t *testing.T) {
		t.Parallel()

		before := time.Now()
		req := NewRequest("alice", "read", Resource{}, Environment{})
		after := time.Now()

		if req.Environment.Time.Before(before) || req.Environment.Time.After(after) {
			t.Fatalf("expected time between %v and %v, got %v", before, after, req.Environment.Time)
		}
	})

	t.Run("preserves non-zero time", func(t *testing.T) {
		t.Parallel()

		want := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

		req := NewRequest(nil, "read", Resource{}, Environment{Time: want})
		if !req.Environment.Time.Equal(want) {
			t.Fatalf("expected %v, got %v", want, req.Environment.Time)
		}
	})

	t.Run("passes principal as-is", func(t *testing.T) {
		t.Parallel()

		principal := struct{ Name string }{Name: "bob"}

		req := NewRequest(principal, "read", Resource{}, Environment{})

		got, ok := req.Principal.(struct{ Name string })
		if !ok {
			t.Fatalf("expected struct{ Name string }, got %T", req.Principal)
		}

		if got.Name != "bob" {
			t.Fatalf("expected 'bob', got %q", got.Name)
		}
	})

	t.Run("nil principal allowed", func(t *testing.T) {
		t.Parallel()

		req := NewRequest(nil, "read", Resource{}, Environment{})
		if req.Principal != nil {
			t.Fatalf("expected nil principal, got %#v", req.Principal)
		}
	})
}

func TestAllow(t *testing.T) {
	t.Parallel()

	t.Run("returns allow decision", func(t *testing.T) {
		t.Parallel()

		d := Allow("admin")
		if d.Effect != EffectAllow {
			t.Fatalf("expected EffectAllow, got %q", d.Effect)
		}

		if d.Reason != "admin" {
			t.Fatalf("expected 'admin', got %q", d.Reason)
		}
	})
}

func TestDeny(t *testing.T) {
	t.Parallel()

	t.Run("returns deny decision", func(t *testing.T) {
		t.Parallel()

		d := Deny("no role")
		if d.Effect != EffectDeny {
			t.Fatalf("expected EffectDeny, got %q", d.Effect)
		}
	})
}

func TestAbstain(t *testing.T) {
	t.Parallel()

	t.Run("returns abstain decision", func(t *testing.T) {
		t.Parallel()

		d := Abstain("not applicable")
		if d.Effect != EffectAbstain {
			t.Fatalf("expected EffectAbstain, got %q", d.Effect)
		}
	})
}

// stubPolicy is a Policy that returns a fixed Decision and records the
// number of times Evaluate was called.
type stubPolicy struct {
	decision Decision
	calls    int
}

func (p *stubPolicy) Evaluate(_ context.Context, _ Request) Decision {
	p.calls++

	return p.decision
}

func TestChainPolicies(t *testing.T) {
	t.Parallel()

	t.Run("first allow wins", func(t *testing.T) {
		t.Parallel()

		p1 := &stubPolicy{decision: Allow("p1")}
		p2 := &stubPolicy{decision: Deny("p2")}

		chained := ChainPolicies(p1, p2)
		dec := chained.Evaluate(context.Background(), Request{})

		if dec.Effect != EffectAllow {
			t.Fatalf("expected EffectAllow, got %q", dec.Effect)
		}

		if p2.calls != 0 {
			t.Fatalf("expected p2 not invoked, got %d calls", p2.calls)
		}
	})

	t.Run("first deny wins", func(t *testing.T) {
		t.Parallel()

		p1 := &stubPolicy{decision: Deny("p1")}
		p2 := &stubPolicy{decision: Allow("p2")}

		chained := ChainPolicies(p1, p2)
		dec := chained.Evaluate(context.Background(), Request{})

		if dec.Effect != EffectDeny {
			t.Fatalf("expected EffectDeny, got %q", dec.Effect)
		}

		if p2.calls != 0 {
			t.Fatalf("expected p2 not invoked, got %d calls", p2.calls)
		}
	})

	t.Run("abstain falls through", func(t *testing.T) {
		t.Parallel()

		p1 := &stubPolicy{decision: Abstain("p1")}
		p2 := &stubPolicy{decision: Allow("p2")}

		chained := ChainPolicies(p1, p2)
		dec := chained.Evaluate(context.Background(), Request{})

		if dec.Effect != EffectAllow {
			t.Fatalf("expected EffectAllow, got %q", dec.Effect)
		}
	})

	t.Run("all abstain returns last abstain", func(t *testing.T) {
		t.Parallel()

		p1 := &stubPolicy{decision: Abstain("p1")}
		p2 := &stubPolicy{decision: Abstain("p2")}

		chained := ChainPolicies(p1, p2)
		dec := chained.Evaluate(context.Background(), Request{})

		if dec.Effect != EffectAbstain {
			t.Fatalf("expected EffectAbstain, got %q", dec.Effect)
		}

		if dec.Reason != "p2" {
			t.Fatalf("expected last abstain reason 'p2', got %q", dec.Reason)
		}
	})

	t.Run("empty chain abstains", func(t *testing.T) {
		t.Parallel()

		chained := ChainPolicies()
		dec := chained.Evaluate(context.Background(), Request{})

		if dec.Effect != EffectAbstain {
			t.Fatalf("expected EffectAbstain, got %q", dec.Effect)
		}
	})

	t.Run("nil policies skipped", func(t *testing.T) {
		t.Parallel()

		p1 := &stubPolicy{decision: Allow("p1")}

		chained := ChainPolicies(nil, p1, nil)
		dec := chained.Evaluate(context.Background(), Request{})

		if dec.Effect != EffectAllow {
			t.Fatalf("expected EffectAllow, got %q", dec.Effect)
		}
	})
}

func TestDefaultAuditHook(t *testing.T) {
	t.Parallel()

	t.Run("does not panic on allow", func(t *testing.T) {
		t.Parallel()

		// Smoke test: just confirm the hook runs without panic.
		// The output goes through common/log's slot — invisible here.
		DefaultAuditHook(context.Background(), Request{Action: "read"}, Allow("ok"))
	})

	t.Run("does not panic on deny", func(t *testing.T) {
		t.Parallel()

		DefaultAuditHook(context.Background(), Request{Action: "read"}, Deny("nope"))
	})

	t.Run("does not panic on abstain", func(t *testing.T) {
		t.Parallel()

		DefaultAuditHook(context.Background(), Request{Action: "read"}, Abstain("?"))
	})

	t.Run("handles populated environment", func(t *testing.T) {
		t.Parallel()

		req := Request{
			Action:      "read",
			Resource:    Resource{Type: "orders", ID: "1"},
			Environment: Environment{IP: net.ParseIP("10.0.0.1"), Time: time.Now()},
		}

		DefaultAuditHook(context.Background(), req, Allow("ok"))
	})

	t.Run("handles unknown effect", func(t *testing.T) {
		t.Parallel()

		DefaultAuditHook(context.Background(), Request{Action: "read"}, Decision{Effect: "weird"})
	})
}

func TestSilentAuditHook(t *testing.T) {
	t.Parallel()

	t.Run("does nothing", func(t *testing.T) {
		t.Parallel()

		// Confirm SilentAuditHook is not the same function pointer as
		// DefaultAuditHook (would indicate a copy-paste bug).
		want := reflect.ValueOf(SilentAuditHook).Pointer()
		def := reflect.ValueOf(DefaultAuditHook).Pointer()

		if want == def {
			t.Fatal("expected SilentAuditHook and DefaultAuditHook to be different functions")
		}

		// And it should not panic.
		SilentAuditHook(context.Background(), Request{}, Decision{})
	})
}

func TestLocalIP(t *testing.T) {
	t.Parallel()

	t.Run("parses ipv4", func(t *testing.T) {
		t.Parallel()

		ip := LocalIP("10.0.0.1")
		if ip == nil {
			t.Fatal("expected non-nil IP")
		}

		if ip.String() != "10.0.0.1" {
			t.Fatalf("expected 10.0.0.1, got %s", ip)
		}
	})

	t.Run("parses ipv6", func(t *testing.T) {
		t.Parallel()

		ip := LocalIP("::1")
		if ip == nil {
			t.Fatal("expected non-nil IP")
		}
	})

	t.Run("empty returns nil", func(t *testing.T) {
		t.Parallel()

		ip := LocalIP("")
		if ip != nil {
			t.Fatalf("expected nil, got %s", ip)
		}
	})

	t.Run("invalid returns nil", func(t *testing.T) {
		t.Parallel()

		ip := LocalIP("not-an-ip")
		if ip != nil {
			t.Fatalf("expected nil, got %s", ip)
		}
	})
}
