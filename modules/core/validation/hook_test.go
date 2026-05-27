package validation

import (
	"context"
	"sync"
	"testing"
	"time"
)

type hookSpy struct {
	mu    sync.Mutex
	calls []hookCall
}

type hookCall struct {
	stage string
	path  string
	rule  string
	err   error
}

func (s *hookSpy) BeforeRule(_ context.Context, path, rule string, _ []any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls = append(s.calls, hookCall{stage: "before", path: path, rule: rule})
}

func (s *hookSpy) AfterRule(_ context.Context, path, rule string, _ []any, err error, _ time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls = append(s.calls, hookCall{stage: "after", path: path, rule: rule, err: err})
}

func TestHook_FiresAroundLeaf(t *testing.T) {
	t.Parallel()

	spy := &hookSpy{}
	rs := Ruleset{Rules: []RuleNode{{Field: "Name", Rules: []RuleNode{{Name: "required"}}}}}
	eng := NewEngine(rs, WithHook(spy))

	type sample struct{ Name string }

	_ = eng.Validate(sample{Name: "ok"}, nil)

	if len(spy.calls) != 2 {
		t.Fatalf("expected 2 hook calls (before/after), got %d", len(spy.calls))
	}

	if spy.calls[0].stage != "before" || spy.calls[1].stage != "after" {
		t.Fatalf("expected before then after, got %+v", spy.calls)
	}

	if spy.calls[1].err != nil {
		t.Fatalf("expected nil err for passing rule, got %v", spy.calls[1].err)
	}
}

func TestHook_AfterReceivesError(t *testing.T) {
	t.Parallel()

	spy := &hookSpy{}
	rs := Ruleset{Rules: []RuleNode{{Field: "Name", Rules: []RuleNode{{Name: "required"}}}}}
	eng := NewEngine(rs, WithHook(spy))

	type sample struct{ Name string }

	_ = eng.Validate(sample{Name: ""}, nil)

	if spy.calls[1].err == nil {
		t.Fatalf("expected non-nil err for failing rule")
	}
}

func TestMultiHook_FansOut(t *testing.T) {
	t.Parallel()

	a := &hookSpy{}
	b := &hookSpy{}
	rs := Ruleset{Rules: []RuleNode{{Field: "Name", Rules: []RuleNode{{Name: "required"}}}}}
	eng := NewEngine(rs, WithHook(MultiHook{a, nil, b}))

	type sample struct{ Name string }

	_ = eng.Validate(sample{Name: "ok"}, nil)

	if len(a.calls) != 2 || len(b.calls) != 2 {
		t.Fatalf("expected both hooks to receive 2 calls, got a=%d b=%d", len(a.calls), len(b.calls))
	}
}
