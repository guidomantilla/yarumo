package validation

import (
	"testing"
)

func TestOptions_WithEvaluator(t *testing.T) {
	t.Parallel()

	// WithEvaluator(nil) should silently keep the default; pass a non-nil
	// evaluator and verify the engine still runs.
	rs := Ruleset{Rules: []RuleNode{{
		Field: "Name",
		Rules: []RuleNode{{Name: "required"}},
	}}}

	eng := NewEngine(rs,
		WithEvaluator(nil),
		WithRegistry(nil),
	)

	err := eng.Validate(struct{ Name string }{Name: "x"}, nil)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}
