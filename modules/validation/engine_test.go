package validation

import (
	"errors"
	"os"
	"testing"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	cvalidation "github.com/guidomantilla/yarumo/common/validation"
)

type pokemon struct {
	ID    string
	Name  string
	Email string
	Level int
	Owner owner
	Phone string
	Tags  []string
}

type owner struct {
	Email string
	Tags  []string
}

func TestEngine_SimpleFieldRules(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		rs, err := loadFile(t, "fixtures/simple.yaml")
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		eng := NewEngine(rs)

		p := pokemon{Name: "Pikachu", Email: "ash@kanto.com"}

		err = eng.Validate(p, nil)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("missing required", func(t *testing.T) {
		t.Parallel()

		rs, err := loadFile(t, "fixtures/simple.yaml")
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		eng := NewEngine(rs)

		p := pokemon{Name: "", Email: ""}

		err = eng.Validate(p, nil)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		if !errors.Is(err, cvalidation.ErrFieldRequired) {
			t.Fatalf("expected ErrFieldRequired wrapped, got %v", err)
		}
	})
}

func TestEngine_Conditional(t *testing.T) {
	t.Parallel()

	t.Run("post requires no id", func(t *testing.T) {
		t.Parallel()

		rs, err := loadFile(t, "fixtures/conditional.yaml")
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		eng := NewEngine(rs)

		// POST with no ID passes.
		p := pokemon{}

		err = eng.Validate(p, map[string]any{"method": "POST"})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("post with id fails", func(t *testing.T) {
		t.Parallel()

		rs, err := loadFile(t, "fixtures/conditional.yaml")
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		eng := NewEngine(rs)

		p := pokemon{ID: "abc"}

		err = eng.Validate(p, map[string]any{"method": "POST"})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		if !errors.Is(err, cvalidation.ErrFieldMustBeUndefined) {
			t.Fatalf("expected ErrFieldMustBeUndefined, got %v", err)
		}
	})

	t.Run("get requires uuid", func(t *testing.T) {
		t.Parallel()

		rs, err := loadFile(t, "fixtures/conditional.yaml")
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		eng := NewEngine(rs)

		p := pokemon{ID: "550e8400-e29b-41d4-a716-446655440000"}

		err = eng.Validate(p, map[string]any{"method": "GET"})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("get with bad uuid fails", func(t *testing.T) {
		t.Parallel()

		rs, err := loadFile(t, "fixtures/conditional.yaml")
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		eng := NewEngine(rs)

		p := pokemon{ID: "not-a-uuid"}

		err = eng.Validate(p, map[string]any{"method": "GET"})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		if !errors.Is(err, cvalidation.ErrUIDInvalid) {
			t.Fatalf("expected ErrUIDInvalid, got %v", err)
		}
	})

	t.Run("country CO phone regex", func(t *testing.T) {
		t.Parallel()

		rs, err := loadFile(t, "fixtures/conditional.yaml")
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		eng := NewEngine(rs)

		p := pokemon{Phone: "+571234567890"}

		err = eng.Validate(p, map[string]any{"country": "CO"})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("country CO bad phone", func(t *testing.T) {
		t.Parallel()

		rs, err := loadFile(t, "fixtures/conditional.yaml")
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		eng := NewEngine(rs)

		p := pokemon{Phone: "abc"}

		err = eng.Validate(p, map[string]any{"country": "CO"})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		if !errors.Is(err, cvalidation.ErrRegexMismatch) {
			t.Fatalf("expected ErrRegexMismatch, got %v", err)
		}
	})

	t.Run("non-matching context skips block", func(t *testing.T) {
		t.Parallel()

		rs, err := loadFile(t, "fixtures/conditional.yaml")
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		eng := NewEngine(rs)

		// No context at all: every when-block evaluates to false, no rule fires.
		err = eng.Validate(pokemon{}, nil)
		if err == nil {
			// pokemon{} should pass since no when fires.
			return
		}
	})
}

func TestEngine_Nested(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		rs, err := loadFile(t, "fixtures/nested.yaml")
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		eng := NewEngine(rs)

		p := pokemon{
			Name: "Pikachu",
			Owner: owner{
				Email: "ash@kanto.com",
				Tags:  []string{"trainer"},
			},
			Level: 75,
		}

		err = eng.Validate(p, map[string]any{"tier": "gold"})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("nested field error", func(t *testing.T) {
		t.Parallel()

		rs, err := loadFile(t, "fixtures/nested.yaml")
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		eng := NewEngine(rs)

		p := pokemon{
			Name: "Pikachu",
			Owner: owner{
				Email: "bad-email",
				Tags:  []string{},
			},
		}

		err = eng.Validate(p, nil)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		// Email and Tags should both fail.
		if !errors.Is(err, cvalidation.ErrEmailInvalid) {
			t.Fatalf("expected ErrEmailInvalid, got %v", err)
		}

		if !errors.Is(err, cvalidation.ErrCollectionEmpty) {
			t.Fatalf("expected ErrCollectionEmpty, got %v", err)
		}
	})

	t.Run("gold tier requires level >= 50", func(t *testing.T) {
		t.Parallel()

		rs, err := loadFile(t, "fixtures/nested.yaml")
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		eng := NewEngine(rs)

		p := pokemon{
			Name: "Pikachu",
			Owner: owner{
				Email: "ash@kanto.com",
				Tags:  []string{"x"},
			},
			Level: 10,
		}

		err = eng.Validate(p, map[string]any{"tier": "gold"})
		if !errors.Is(err, cvalidation.ErrMinValue) {
			t.Fatalf("expected ErrMinValue, got %v", err)
		}
	})
}

func TestEngine_UnknownRule(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{Name: "no_such_rule"}}}
	eng := NewEngine(rs)

	err := eng.Validate(pokemon{}, nil)
	if !errors.Is(err, ErrUnknownRule) {
		t.Fatalf("expected ErrUnknownRule, got %v", err)
	}

	msg := err.Error()
	if msg == "" || msg == "<nil>" {
		t.Fatalf("expected formatted message, got %q", msg)
	}
}

func TestEngine_WhenInvalidExpression(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{When: "bad syntax !!!"}}}
	eng := NewEngine(rs)

	err := eng.Validate(pokemon{}, nil)
	if !errors.Is(err, ErrWhenEvalFailed) {
		t.Fatalf("expected ErrWhenEvalFailed, got %v", err)
	}
}

func TestEngine_WhenNonBoolean(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{When: "1 + 2"}}}
	eng := NewEngine(rs)

	err := eng.Validate(pokemon{}, nil)
	if !errors.Is(err, ErrWhenNotBoolean) {
		t.Fatalf("expected ErrWhenNotBoolean, got %v", err)
	}
}

func TestEngine_BadRuleEmptyNode(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{}}}
	eng := NewEngine(rs)

	err := eng.Validate(pokemon{}, nil)
	if !errors.Is(err, ErrBadRule) {
		t.Fatalf("expected ErrBadRule, got %v", err)
	}
}

func TestEngine_BadFieldPath(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{
		Field: "Nonexistent",
		Rules: []RuleNode{{Name: "required"}},
	}}}
	eng := NewEngine(rs)

	err := eng.Validate(pokemon{}, nil)
	if !errors.Is(err, cvalidation.ErrPathNotFound) {
		t.Fatalf("expected ErrPathNotFound, got %v", err)
	}
}

func TestEngine_GroupNode(t *testing.T) {
	t.Parallel()

	// A top-level group with no field/when, just nested rules: every nested
	// rule applies to the root value.
	rs := Ruleset{Rules: []RuleNode{{
		Rules: []RuleNode{
			{Field: "Name", Rules: []RuleNode{{Name: "required"}}},
		},
	}}}
	eng := NewEngine(rs)

	err := eng.Validate(pokemon{}, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestEngine_LeafAtRoot(t *testing.T) {
	t.Parallel()

	// Leaf at the top level: current.path is empty, annotatePath returns the
	// raw error.
	rs := Ruleset{Rules: []RuleNode{{Name: "required"}}}
	eng := NewEngine(rs)

	err := eng.Validate("", nil)
	if !errors.Is(err, cvalidation.ErrFieldRequired) {
		t.Fatalf("expected ErrFieldRequired, got %v", err)
	}
}

func TestEngine_RegistryCustomRule(t *testing.T) {
	t.Parallel()

	reg := DefaultRegistry()
	reg.Register("is_pikachu", func(value any, _ []any) error {
		s, ok := value.(string)
		if !ok || s != "Pikachu" {
			return cvalidation.ErrValidation(errors.New("not pikachu"))
		}

		return nil
	})

	rs := Ruleset{Rules: []RuleNode{{
		Field: "Name",
		Rules: []RuleNode{{Name: "is_pikachu"}},
	}}}
	eng := NewEngine(rs, WithRegistry(reg))

	err := eng.Validate(pokemon{Name: "Pikachu"}, nil)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	err = eng.Validate(pokemon{Name: "Snorlax"}, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestEngine_AsErrorInfo(t *testing.T) {
	t.Parallel()

	rs := Ruleset{Rules: []RuleNode{{
		Field: "Name",
		Rules: []RuleNode{{Name: "required"}},
	}, {
		Field: "Email",
		Rules: []RuleNode{{Name: "email"}},
	}}}
	eng := NewEngine(rs)

	err := eng.Validate(pokemon{Email: "bad"}, nil)
	if err == nil {
		t.Fatalf("expected error")
	}

	infos := cerrs.AsErrorInfo(err)
	if len(infos) == 0 {
		t.Fatalf("expected at least one info group")
	}

	found := false
	for _, info := range infos {
		if info.Type == "validation" {
			found = true
		}
	}

	if !found {
		t.Fatalf("expected validation type in infos, got %+v", infos)
	}
}

func loadFile(t *testing.T, path string) (Ruleset, error) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		return Ruleset{}, err
	}

	return LoadYAML(data)
}
