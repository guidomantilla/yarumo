package validation

import (
	"errors"
	"strings"
	"testing"
)

func TestLoadYAML(t *testing.T) {
	t.Parallel()

	t.Run("list shape", func(t *testing.T) {
		t.Parallel()

		rs, err := LoadYAML([]byte(`
- field: Name
  rules:
    - required
- field: Email
  rules:
    - required
    - email
`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(rs.Rules) != 2 {
			t.Fatalf("expected 2 rules, got %d", len(rs.Rules))
		}

		if rs.Rules[0].Field != "Name" {
			t.Fatalf("expected first rule field=Name, got %q", rs.Rules[0].Field)
		}
	})

	t.Run("mapping shape", func(t *testing.T) {
		t.Parallel()

		rs, err := LoadYAML([]byte(`
rules:
  - field: Name
    rules: [required]
`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(rs.Rules) != 1 {
			t.Fatalf("expected 1 rule, got %d", len(rs.Rules))
		}
	})

	t.Run("scalar leaf", func(t *testing.T) {
		t.Parallel()

		rs, err := LoadYAML([]byte(`
- required
`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rs.Rules[0].Name != "required" {
			t.Fatalf("expected Name=required, got %q", rs.Rules[0].Name)
		}
	})

	t.Run("sugar map leaf", func(t *testing.T) {
		t.Parallel()

		rs, err := LoadYAML([]byte(`
- { min_len: 3 }
- { in_range: [1, 100] }
`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rs.Rules[0].Name != "min_len" {
			t.Fatalf("expected Name=min_len, got %q", rs.Rules[0].Name)
		}

		if len(rs.Rules[0].Params) != 1 {
			t.Fatalf("expected 1 param, got %d", len(rs.Rules[0].Params))
		}

		if rs.Rules[1].Name != "in_range" {
			t.Fatalf("expected Name=in_range, got %q", rs.Rules[1].Name)
		}

		if len(rs.Rules[1].Params) != 2 {
			t.Fatalf("expected 2 params, got %d", len(rs.Rules[1].Params))
		}
	})

	t.Run("error nil data", func(t *testing.T) {
		t.Parallel()

		_, err := LoadYAML(nil)
		if !errors.Is(err, ErrDataNil) {
			t.Fatalf("expected ErrDataNil, got %v", err)
		}
	})

	t.Run("error malformed", func(t *testing.T) {
		t.Parallel()

		_, err := LoadYAML([]byte(`{`))
		if !errors.Is(err, ErrLoadFailed) {
			t.Fatalf("expected ErrLoadFailed, got %v", err)
		}
	})
}

func TestLoadJSON(t *testing.T) {
	t.Parallel()

	t.Run("list shape", func(t *testing.T) {
		t.Parallel()

		rs, err := LoadJSON([]byte(`[{"field":"Name","rules":["required"]}]`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(rs.Rules) != 1 {
			t.Fatalf("expected 1 rule, got %d", len(rs.Rules))
		}
	})

	t.Run("scalar leaf string", func(t *testing.T) {
		t.Parallel()

		rs, err := LoadJSON([]byte(`["required"]`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rs.Rules[0].Name != "required" {
			t.Fatalf("expected Name=required, got %q", rs.Rules[0].Name)
		}
	})

	t.Run("sugar map leaf array", func(t *testing.T) {
		t.Parallel()

		rs, err := LoadJSON([]byte(`[{"in_range":[1,100]}]`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rs.Rules[0].Name != "in_range" {
			t.Fatalf("expected Name=in_range, got %q", rs.Rules[0].Name)
		}

		if len(rs.Rules[0].Params) != 2 {
			t.Fatalf("expected 2 params, got %d", len(rs.Rules[0].Params))
		}
	})

	t.Run("sugar map leaf single", func(t *testing.T) {
		t.Parallel()

		rs, err := LoadJSON([]byte(`[{"min_len":3}]`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(rs.Rules[0].Params) != 1 {
			t.Fatalf("expected 1 param, got %d", len(rs.Rules[0].Params))
		}
	})

	t.Run("mapping shape", func(t *testing.T) {
		t.Parallel()

		rs, err := LoadJSON([]byte(`{"rules":[{"field":"Name","rules":["required"]}]}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(rs.Rules) != 1 {
			t.Fatalf("expected 1 rule, got %d", len(rs.Rules))
		}
	})

	t.Run("error nil data", func(t *testing.T) {
		t.Parallel()

		_, err := LoadJSON(nil)
		if !errors.Is(err, ErrDataNil) {
			t.Fatalf("expected ErrDataNil, got %v", err)
		}
	})

	t.Run("error malformed", func(t *testing.T) {
		t.Parallel()

		_, err := LoadJSON([]byte(`{`))
		if !errors.Is(err, ErrLoadFailed) {
			t.Fatalf("expected ErrLoadFailed, got %v", err)
		}
	})
}

func TestLoadYAMLReader(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		r := strings.NewReader(`
- field: Name
  rules: [required]
`)

		rs, err := LoadYAMLReader(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(rs.Rules) != 1 {
			t.Fatalf("expected 1 rule, got %d", len(rs.Rules))
		}
	})

	t.Run("error nil reader", func(t *testing.T) {
		t.Parallel()

		_, err := LoadYAMLReader(nil)
		if !errors.Is(err, ErrReaderNil) {
			t.Fatalf("expected ErrReaderNil, got %v", err)
		}
	})
}

func TestLoadJSONReader(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		r := strings.NewReader(`[{"field":"Name","rules":["required"]}]`)

		rs, err := LoadJSONReader(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(rs.Rules) != 1 {
			t.Fatalf("expected 1 rule, got %d", len(rs.Rules))
		}
	})

	t.Run("error nil reader", func(t *testing.T) {
		t.Parallel()

		_, err := LoadJSONReader(nil)
		if !errors.Is(err, ErrReaderNil) {
			t.Fatalf("expected ErrReaderNil, got %v", err)
		}
	})
}

func TestLoadFromReader(t *testing.T) {
	t.Parallel()

	t.Run("happy path yaml", func(t *testing.T) {
		t.Parallel()

		r := strings.NewReader(`- required`)

		rs, err := LoadFromReader(r, LoadYAML)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rs.Rules[0].Name != "required" {
			t.Fatalf("expected Name=required, got %q", rs.Rules[0].Name)
		}
	})

	t.Run("error nil reader", func(t *testing.T) {
		t.Parallel()

		_, err := LoadFromReader(nil, LoadYAML)
		if !errors.Is(err, ErrReaderNil) {
			t.Fatalf("expected ErrReaderNil, got %v", err)
		}
	})

	t.Run("error nil loader", func(t *testing.T) {
		t.Parallel()

		_, err := LoadFromReader(strings.NewReader(""), nil)
		if !errors.Is(err, ErrLoadFailed) {
			t.Fatalf("expected ErrLoadFailed, got %v", err)
		}
	})
}

func TestPathOf_NoPath(t *testing.T) {
	t.Parallel()

	// A leaf that fails without a field annotation should yield empty PathOf.
	rs := Ruleset{Rules: []RuleNode{{Name: "required"}}}
	eng := NewEngine(rs)

	err := eng.Validate("", nil)

	path := PathOf(err)
	if path != "" {
		t.Fatalf("expected empty path, got %q", path)
	}
}

func TestPathOf_Nil(t *testing.T) {
	t.Parallel()

	path := PathOf(nil)
	if path != "" {
		t.Fatalf("expected empty path, got %q", path)
	}
}

func TestPathOf(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{
		Field: "Name",
		Rules: []RuleNode{{Name: "required"}},
	}}}
	eng := NewEngine(rs)

	err := eng.Validate(pokemon{}, nil)
	if err == nil {
		t.Fatalf("expected error")
	}

	path := PathOf(err)
	if path != "Name" {
		t.Fatalf("expected Name, got %q", path)
	}
}
