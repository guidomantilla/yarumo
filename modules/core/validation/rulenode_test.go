package validation

import (
	"errors"
	"testing"

	yaml "go.yaml.in/yaml/v3"
)

func TestRuleNode_UnmarshalYAML(t *testing.T) {
	t.Parallel()

	t.Run("scalar string becomes leaf", func(t *testing.T) {
		t.Parallel()

		var n RuleNode

		err := yaml.Unmarshal([]byte(`required`), &n)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}

		if n.Name != "required" {
			t.Fatalf("expected Name=required, got %q", n.Name)
		}
		if len(n.Params) != 0 {
			t.Fatalf("expected empty Params, got %v", n.Params)
		}
	})

	t.Run("single-key mapping with scalar param", func(t *testing.T) {
		t.Parallel()

		var n RuleNode

		err := yaml.Unmarshal([]byte(`min_len: 5`), &n)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}

		if n.Name != "min_len" {
			t.Fatalf("expected Name=min_len, got %q", n.Name)
		}
		if len(n.Params) != 1 {
			t.Fatalf("expected 1 param, got %d", len(n.Params))
		}
	})

	t.Run("single-key mapping with sequence params", func(t *testing.T) {
		t.Parallel()

		var n RuleNode

		err := yaml.Unmarshal([]byte(`in_range: [1, 10]`), &n)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}

		if n.Name != "in_range" {
			t.Fatalf("expected Name=in_range, got %q", n.Name)
		}
		if len(n.Params) != 2 {
			t.Fatalf("expected 2 params, got %d", len(n.Params))
		}
	})

	t.Run("full mapping with structural keys", func(t *testing.T) {
		t.Parallel()

		var n RuleNode

		err := yaml.Unmarshal([]byte(`
field: Email
rules:
  - email
`), &n)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}

		if n.Field != "Email" {
			t.Fatalf("expected Field=Email, got %q", n.Field)
		}
		if len(n.Rules) != 1 || n.Rules[0].Name != "email" {
			t.Fatalf("expected nested rules with email, got %+v", n.Rules)
		}
	})

	t.Run("unsupported node kind errors", func(t *testing.T) {
		t.Parallel()

		// A YAML sequence at the RuleNode position is not a supported shape.
		var n RuleNode

		err := yaml.Unmarshal([]byte(`[1, 2, 3]`), &n)
		if !errors.Is(err, ErrLoadFailed) {
			t.Fatalf("expected ErrLoadFailed, got %v", err)
		}
	})
}

func TestRuleNode_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("string becomes leaf", func(t *testing.T) {
		t.Parallel()

		var n RuleNode

		err := n.UnmarshalJSON([]byte(`"required"`))
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}

		if n.Name != "required" {
			t.Fatalf("expected Name=required, got %q", n.Name)
		}
	})

	t.Run("single-key object with scalar param", func(t *testing.T) {
		t.Parallel()

		var n RuleNode

		err := n.UnmarshalJSON([]byte(`{"min_len": 5}`))
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}

		if n.Name != "min_len" {
			t.Fatalf("expected Name=min_len, got %q", n.Name)
		}
		if len(n.Params) != 1 {
			t.Fatalf("expected 1 param, got %d", len(n.Params))
		}
	})

	t.Run("single-key object with array params", func(t *testing.T) {
		t.Parallel()

		var n RuleNode

		err := n.UnmarshalJSON([]byte(`{"in_range": [1, 10]}`))
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}

		if n.Name != "in_range" {
			t.Fatalf("expected Name=in_range, got %q", n.Name)
		}
		if len(n.Params) != 2 {
			t.Fatalf("expected 2 params, got %d", len(n.Params))
		}
	})

	t.Run("full object with structural keys", func(t *testing.T) {
		t.Parallel()

		var n RuleNode

		err := n.UnmarshalJSON([]byte(`{"field": "Email", "rules": [{"name": "email"}]}`))
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}

		if n.Field != "Email" {
			t.Fatalf("expected Field=Email, got %q", n.Field)
		}
		if len(n.Rules) != 1 || n.Rules[0].Name != "email" {
			t.Fatalf("expected nested email rule, got %+v", n.Rules)
		}
	})

	t.Run("malformed JSON errors", func(t *testing.T) {
		t.Parallel()

		var n RuleNode

		err := n.UnmarshalJSON([]byte(`not json`))
		if !errors.Is(err, ErrLoadFailed) {
			t.Fatalf("expected ErrLoadFailed, got %v", err)
		}
	})
}
