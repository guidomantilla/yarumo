package validation

import (
	"errors"
	"testing"

	cvalidation "github.com/guidomantilla/yarumo/common/validation"
)

type address struct {
	Street string
	Zip    string
}

type ordercompose struct {
	Shipping address
	Billing  address
}

func TestExpand_HappyPath(t *testing.T) {
	t.Parallel()

	rs := Ruleset{
		Defines: map[string][]RuleNode{
			"address_rules": {
				{Field: "Street", Rules: []RuleNode{{Name: "required"}}},
				{Field: "Zip", Rules: []RuleNode{{Name: "required"}}},
			},
		},
		Rules: []RuleNode{
			{Field: "Shipping", Use: "address_rules"},
			{Field: "Billing", Use: "address_rules"},
		},
	}

	expanded, err := Expand(rs)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	eng := NewEngine(expanded)

	err = eng.Validate(ordercompose{Shipping: address{Street: "x", Zip: "00000"}, Billing: address{}}, nil)
	if !errors.Is(err, cvalidation.ErrFieldRequired) {
		t.Fatalf("expected ErrFieldRequired from billing, got %v", err)
	}
}

func TestExpand_UndefinedReference(t *testing.T) {
	t.Parallel()

	rs := Ruleset{
		Rules: []RuleNode{{Use: "missing"}},
	}

	_, err := Expand(rs)
	if !errors.Is(err, ErrUndefinedUse) {
		t.Fatalf("expected ErrUndefinedUse, got %v", err)
	}
}

func TestExpand_Cycle(t *testing.T) {
	t.Parallel()

	rs := Ruleset{
		Defines: map[string][]RuleNode{
			"a": {{Use: "b"}},
			"b": {{Use: "a"}},
		},
		Rules: []RuleNode{{Use: "a"}},
	}

	_, err := Expand(rs)
	if !errors.Is(err, ErrCycleDetected) {
		t.Fatalf("expected ErrCycleDetected, got %v", err)
	}
}

type bindTarget struct {
	Name  string
	Owner struct {
		Email string
	}
}

func TestRulesetFor_HappyPath(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{
		{Field: "Name", Rules: []RuleNode{{Name: "required"}}},
		{Field: "Owner", Rules: []RuleNode{{Field: "Email", Rules: []RuleNode{{Name: "email"}}}}},
	}}

	_, err := RulesetFor[bindTarget](rs)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestRulesetFor_UnknownField(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{Field: "Missing", Rules: []RuleNode{{Name: "required"}}}}}

	_, err := RulesetFor[bindTarget](rs)
	if !errors.Is(err, ErrUnknownField) {
		t.Fatalf("expected ErrUnknownField, got %v", err)
	}
}

func TestBind_DynamicSample(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{Field: "Name", Rules: []RuleNode{{Name: "required"}}}}}

	_, err := Bind(rs, bindTarget{})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}
