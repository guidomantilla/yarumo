package adapters

import (
	"errors"
	"testing"

	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

func TestAdaptDeductiveRules(t *testing.T) {
	t.Parallel()

	t.Run("valid rules", func(t *testing.T) {
		t.Parallel()

		config := &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{
					Name:       "r1",
					Condition:  "a and b",
					Conclusion: map[string]bool{"c": true},
					Priority:   1,
				},
				{
					Name:       "r2",
					Condition:  "c",
					Conclusion: map[string]bool{"d": true},
				},
			},
		}

		rules, parsed, err := AdaptDeductiveRules(config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(rules) != 2 {
			t.Fatalf("expected 2 rules, got %d", len(rules))
		}

		if len(parsed) != 2 {
			t.Fatalf("expected 2 parsed, got %d", len(parsed))
		}

		if rules[0].Name() != "r1" {
			t.Fatalf("expected r1, got %s", rules[0].Name())
		}

		if rules[0].Priority() != 1 {
			t.Fatalf("expected priority 1, got %d", rules[0].Priority())
		}
	})

	t.Run("invalid condition", func(t *testing.T) {
		t.Parallel()

		config := &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{
					Name:       "bad",
					Condition:  "(((",
					Conclusion: map[string]bool{"x": true},
				},
			},
		}

		_, _, err := AdaptDeductiveRules(config)
		if err == nil {
			t.Fatal("expected parse error")
		}

		if !errors.Is(err, ErrAdaptRulesFailed) {
			t.Fatalf("expected ErrAdaptRulesFailed, got: %v", err)
		}
	})
}

func TestAdaptDeductiveOpts(t *testing.T) {
	t.Parallel()

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		opts := AdaptDeductiveOpts(&schema.DeductiveConfig{})
		if len(opts) != 0 {
			t.Fatalf("expected 0 opts, got %d", len(opts))
		}
	})

	t.Run("with max iterations and first match", func(t *testing.T) {
		t.Parallel()

		opts := AdaptDeductiveOpts(&schema.DeductiveConfig{
			MaxIterations: 500,
			Strategy:      "first_match",
		})

		if len(opts) != 2 {
			t.Fatalf("expected 2 opts, got %d", len(opts))
		}
	})
}
