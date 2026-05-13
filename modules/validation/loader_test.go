package validation_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/guidomantilla/yarumo/validation"
)

func TestLoadYAML(t *testing.T) {
	t.Parallel()

	t.Run("list shape", func(t *testing.T) {
		t.Parallel()

		rs, err := validation.LoadYAML([]byte(`
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

		rs, err := validation.LoadYAML([]byte(`
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

		rs, err := validation.LoadYAML([]byte(`
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

		rs, err := validation.LoadYAML([]byte(`
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

		_, err := validation.LoadYAML(nil)
		if !errors.Is(err, validation.ErrDataNil) {
			t.Fatalf("expected ErrDataNil, got %v", err)
		}
	})

	t.Run("error malformed", func(t *testing.T) {
		t.Parallel()

		_, err := validation.LoadYAML([]byte(`{`))
		if !errors.Is(err, validation.ErrLoadFailed) {
			t.Fatalf("expected ErrLoadFailed, got %v", err)
		}
	})
}

func TestLoadJSON(t *testing.T) {
	t.Parallel()

	t.Run("list shape", func(t *testing.T) {
		t.Parallel()

		rs, err := validation.LoadJSON([]byte(`[{"field":"Name","rules":["required"]}]`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(rs.Rules) != 1 {
			t.Fatalf("expected 1 rule, got %d", len(rs.Rules))
		}
	})

	t.Run("scalar leaf string", func(t *testing.T) {
		t.Parallel()

		rs, err := validation.LoadJSON([]byte(`["required"]`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rs.Rules[0].Name != "required" {
			t.Fatalf("expected Name=required, got %q", rs.Rules[0].Name)
		}
	})

	t.Run("sugar map leaf array", func(t *testing.T) {
		t.Parallel()

		rs, err := validation.LoadJSON([]byte(`[{"in_range":[1,100]}]`))
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

		rs, err := validation.LoadJSON([]byte(`[{"min_len":3}]`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(rs.Rules[0].Params) != 1 {
			t.Fatalf("expected 1 param, got %d", len(rs.Rules[0].Params))
		}
	})

	t.Run("mapping shape", func(t *testing.T) {
		t.Parallel()

		rs, err := validation.LoadJSON([]byte(`{"rules":[{"field":"Name","rules":["required"]}]}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(rs.Rules) != 1 {
			t.Fatalf("expected 1 rule, got %d", len(rs.Rules))
		}
	})

	t.Run("error nil data", func(t *testing.T) {
		t.Parallel()

		_, err := validation.LoadJSON(nil)
		if !errors.Is(err, validation.ErrDataNil) {
			t.Fatalf("expected ErrDataNil, got %v", err)
		}
	})

	t.Run("error malformed", func(t *testing.T) {
		t.Parallel()

		_, err := validation.LoadJSON([]byte(`{`))
		if !errors.Is(err, validation.ErrLoadFailed) {
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

		rs, err := validation.LoadYAMLReader(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(rs.Rules) != 1 {
			t.Fatalf("expected 1 rule, got %d", len(rs.Rules))
		}
	})

	t.Run("error nil reader", func(t *testing.T) {
		t.Parallel()

		_, err := validation.LoadYAMLReader(nil)
		if !errors.Is(err, validation.ErrReaderNil) {
			t.Fatalf("expected ErrReaderNil, got %v", err)
		}
	})
}

func TestLoadJSONReader(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		r := strings.NewReader(`[{"field":"Name","rules":["required"]}]`)

		rs, err := validation.LoadJSONReader(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(rs.Rules) != 1 {
			t.Fatalf("expected 1 rule, got %d", len(rs.Rules))
		}
	})

	t.Run("error nil reader", func(t *testing.T) {
		t.Parallel()

		_, err := validation.LoadJSONReader(nil)
		if !errors.Is(err, validation.ErrReaderNil) {
			t.Fatalf("expected ErrReaderNil, got %v", err)
		}
	})
}

func TestLoadFromReader(t *testing.T) {
	t.Parallel()

	t.Run("happy path yaml", func(t *testing.T) {
		t.Parallel()

		r := strings.NewReader(`- required`)

		rs, err := validation.LoadFromReader(r, validation.LoadYAML)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rs.Rules[0].Name != "required" {
			t.Fatalf("expected Name=required, got %q", rs.Rules[0].Name)
		}
	})

	t.Run("error nil reader", func(t *testing.T) {
		t.Parallel()

		_, err := validation.LoadFromReader(nil, validation.LoadYAML)
		if !errors.Is(err, validation.ErrReaderNil) {
			t.Fatalf("expected ErrReaderNil, got %v", err)
		}
	})

	t.Run("error nil loader", func(t *testing.T) {
		t.Parallel()

		_, err := validation.LoadFromReader(strings.NewReader(""), nil)
		if !errors.Is(err, validation.ErrLoadFailed) {
			t.Fatalf("expected ErrLoadFailed, got %v", err)
		}
	})
}
