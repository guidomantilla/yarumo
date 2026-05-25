package validation

import (
	"errors"
	"testing"

	cvalidation "github.com/guidomantilla/yarumo/common/validation"
)

func TestValidate_HappyPath(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{Name: "required"}}}

	err := Validate(rs)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidate_UnknownRule(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{Name: "no_such_rule"}}}

	err := Validate(rs)
	if !errors.Is(err, ErrUnknownRule) {
		t.Fatalf("expected ErrUnknownRule wrapped, got %v", err)
	}
}

func TestValidate_MixedShape(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{Name: "required", Field: "X"}}}

	err := Validate(rs)
	if !errors.Is(err, ErrMixedShape) {
		t.Fatalf("expected ErrMixedShape, got %v", err)
	}
}

func TestValidate_EmptyNode(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{}}}

	err := Validate(rs)
	if !errors.Is(err, ErrEmptyGroup) {
		t.Fatalf("expected ErrEmptyGroup, got %v", err)
	}
}

func TestValidate_BadWhen(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{When: "(((", Rules: []RuleNode{{Name: "required"}}}}}

	err := Validate(rs)
	if !errors.Is(err, ErrWhenParseFailed) {
		t.Fatalf("expected ErrWhenParseFailed, got %v", err)
	}
}

func TestValidate_VersionMismatch(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Version: "9.99", Rules: []RuleNode{{Name: "required"}}}

	err := Validate(rs)
	if !errors.Is(err, ErrUnknownVersion) {
		t.Fatalf("expected ErrUnknownVersion, got %v", err)
	}
}

func TestValidate_StrictVersionRequiresField(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{Name: "required"}}}

	err := Validate(rs, WithStrictVersion())
	if !errors.Is(err, ErrUnknownVersion) {
		t.Fatalf("expected ErrUnknownVersion under strict, got %v", err)
	}
}

func TestBuildEngine_LintOnLoadFailsFast(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{Name: "no_such_rule"}}}

	_, err := BuildEngine(rs, WithLintOnLoad())
	if !errors.Is(err, ErrUnknownRule) {
		t.Fatalf("expected ErrUnknownRule, got %v", err)
	}
}

func TestBuildEngine_LintOnLoadPasses(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{Name: "required"}}}

	eng, err := BuildEngine(rs, WithLintOnLoad())
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	if eng == nil {
		t.Fatalf("expected engine, got nil")
	}
}

func TestEngine_Run_NoViolations(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{Field: "Name", Rules: []RuleNode{{Name: "required"}}}}}
	eng := NewEngine(rs)

	type sample struct{ Name string }

	violations := eng.Run(sample{Name: "ok"}, nil)
	if violations != nil {
		t.Fatalf("expected nil, got %v", violations)
	}
}

func TestEngine_Run_ReturnsViolations(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{Field: "Name", Rules: []RuleNode{{Name: "required"}}}}}
	eng := NewEngine(rs)

	type sample struct{ Name string }

	violations := eng.Run(sample{Name: ""}, nil)
	if len(violations) == 0 {
		t.Fatalf("expected violations, got none")
	}

	found := false
	for _, v := range violations {
		if errors.Is(v.Cause, cvalidation.ErrFieldRequired) {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected violation referencing ErrFieldRequired, got %#v", violations)
	}
}
