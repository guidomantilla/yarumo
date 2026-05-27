package validation

import (
	"errors"
	"strings"
	"testing"

	cvalidation "github.com/guidomantilla/yarumo/core/common/validation"
)

type dslTarget struct {
	Name   string
	Email  string
	Tags   []string
	Status string
	Role   string
}

func TestDSL_OptionalSkipsZero(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{
		Field:    "Email",
		Optional: true,
		Rules:    []RuleNode{{Name: "email"}},
	}}}
	eng := NewEngine(rs)

	err := eng.Validate(dslTarget{Email: ""}, nil)
	if err != nil {
		t.Fatalf("expected nil for empty optional, got %v", err)
	}
}

func TestDSL_OptionalRunsWhenPresent(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{
		Field:    "Email",
		Optional: true,
		Rules:    []RuleNode{{Name: "email"}},
	}}}
	eng := NewEngine(rs)

	err := eng.Validate(dslTarget{Email: "bad-email"}, nil)
	if !errors.Is(err, cvalidation.ErrEmailInvalid) {
		t.Fatalf("expected ErrEmailInvalid, got %v", err)
	}
}

func TestDSL_AnyOf_PassesWhenOnePasses(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{
		Field: "Name",
		Rules: []RuleNode{{Name: combinatorAnyOf, Rules: []RuleNode{{Name: "email"}, {Name: "uuid"}, {Name: "required"}}}},
	}}}
	eng := NewEngine(rs)

	err := eng.Validate(dslTarget{Name: "plain"}, nil)
	if err != nil {
		t.Fatalf("expected nil (required passes), got %v", err)
	}
}

func TestDSL_AnyOf_FailsWhenAllFail(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{
		Field: "Name",
		Rules: []RuleNode{{Name: combinatorAnyOf, Rules: []RuleNode{{Name: "email"}, {Name: "uuid"}}}},
	}}}
	eng := NewEngine(rs)

	err := eng.Validate(dslTarget{Name: "neither"}, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestDSL_AllOf_AggregatesFailures(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{
		Field: "Name",
		Rules: []RuleNode{{Name: combinatorAllOf, Rules: []RuleNode{{Name: "alpha"}, {Name: "min_len", Params: []any{20}}}}},
	}}}
	eng := NewEngine(rs)

	err := eng.Validate(dslTarget{Name: "ab1"}, nil)
	if !errors.Is(err, cvalidation.ErrMinLen) {
		t.Fatalf("expected ErrMinLen wrapped, got %v", err)
	}

	if !errors.Is(err, cvalidation.ErrNotAlpha) {
		t.Fatalf("expected ErrNotAlpha wrapped, got %v", err)
	}
}

func TestDSL_Not_PassesWhenInnerFails(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{
		Field: "Name",
		Rules: []RuleNode{{Name: combinatorNot, Rules: []RuleNode{{Name: "email"}}}},
	}}}
	eng := NewEngine(rs)

	err := eng.Validate(dslTarget{Name: "not-an-email"}, nil)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestDSL_Not_FailsWhenInnerPasses(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{
		Field: "Name",
		Rules: []RuleNode{{Name: combinatorNot, Rules: []RuleNode{{Name: "email"}}}},
	}}}
	eng := NewEngine(rs)

	err := eng.Validate(dslTarget{Name: "a@b.com"}, nil)
	if !errors.Is(err, cvalidation.ErrAssertionInverted) {
		t.Fatalf("expected ErrAssertionInverted, got %v", err)
	}
}

func TestDSL_ForEach_AppliesPerElement(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{
		Field: "Tags",
		Rules: []RuleNode{{Name: combinatorForEach, Rules: []RuleNode{{Name: "required"}}}},
	}}}
	eng := NewEngine(rs)

	err := eng.Validate(dslTarget{Tags: []string{"a", "", "b"}}, nil)
	if !errors.Is(err, cvalidation.ErrFieldRequired) {
		t.Fatalf("expected ErrFieldRequired from empty element, got %v", err)
	}
}

func TestDSL_ObjInWhen(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{
		When:  `obj.Role == "admin"`,
		Rules: []RuleNode{{Field: "Name", Rules: []RuleNode{{Name: "required"}}}},
	}}}
	eng := NewEngine(rs)

	err := eng.Validate(dslTarget{Role: "admin", Name: ""}, nil)
	if !errors.Is(err, cvalidation.ErrFieldRequired) {
		t.Fatalf("expected ErrFieldRequired when admin, got %v", err)
	}

	err = eng.Validate(dslTarget{Role: "guest", Name: ""}, nil)
	if err != nil {
		t.Fatalf("expected nil when not admin, got %v", err)
	}
}

func TestDSL_CustomMessage(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{
		Field: "Email",
		Rules: []RuleNode{{Name: "email", Message: "Email must be valid for registration"}},
	}}}
	eng := NewEngine(rs)

	err := eng.Validate(dslTarget{Email: "bad-email"}, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "Email must be valid for registration") {
		t.Fatalf("expected custom message in output, got %q", err.Error())
	}

	if !errors.Is(err, cvalidation.ErrEmailInvalid) {
		t.Fatalf("expected ErrEmailInvalid still reachable, got %v", err)
	}
}
