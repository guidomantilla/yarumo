package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

func TestNewMemoryRepository(t *testing.T) {
	t.Parallel()

	repo := NewMemoryRepository()
	if repo == nil {
		t.Fatal("expected non-nil repository")
	}
}

func TestMemoryRepository_Save(t *testing.T) {
	t.Parallel()

	t.Run("save a ruleset", func(t *testing.T) {
		t.Parallel()

		repo := NewMemoryRepository()

		rs := &schema.RuleSet{Name: "test", Version: "1.0"}
		err := repo.Save(context.Background(), rs)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("save overwrites existing ruleset", func(t *testing.T) {
		t.Parallel()

		repo := NewMemoryRepository()

		err := repo.Save(context.Background(), &schema.RuleSet{
			Name:    "test",
			Version: "1.0",
			Deductive: &schema.DeductiveConfig{
				Rules: []schema.DeductiveRuleDef{{Name: "old"}},
			},
		})
		if err != nil {
			t.Fatalf("expected no error on first save, got %v", err)
		}

		err = repo.Save(context.Background(), &schema.RuleSet{
			Name:    "test",
			Version: "1.0",
			Deductive: &schema.DeductiveConfig{
				Rules: []schema.DeductiveRuleDef{{Name: "new"}},
			},
		})
		if err != nil {
			t.Fatalf("expected no error on second save, got %v", err)
		}

		got, err := repo.Get(context.Background(), "test", "1.0")
		if err != nil {
			t.Fatalf("expected no error on get, got %v", err)
		}

		if got.Deductive.Rules[0].Name != "new" {
			t.Fatalf("expected overwritten rule name 'new', got %q", got.Deductive.Rules[0].Name)
		}
	})
}

func TestMemoryRepository_Get(t *testing.T) {
	t.Parallel()

	t.Run("get existing ruleset", func(t *testing.T) {
		t.Parallel()

		repo := NewMemoryRepository()

		rs := &schema.RuleSet{Name: "test", Version: "1.0"}
		err := repo.Save(context.Background(), rs)
		if err != nil {
			t.Fatalf("expected no error on save, got %v", err)
		}

		got, err := repo.Get(context.Background(), "test", "1.0")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got.Name != "test" {
			t.Fatalf("expected name 'test', got %q", got.Name)
		}

		if got.Version != "1.0" {
			t.Fatalf("expected version '1.0', got %q", got.Version)
		}
	})

	t.Run("get non-existing ruleset returns ErrNotFound", func(t *testing.T) {
		t.Parallel()

		repo := NewMemoryRepository()

		_, err := repo.Get(context.Background(), "missing", "1.0")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}

		if !errors.Is(err, ErrGetFailed) {
			t.Fatalf("expected ErrGetFailed, got %v", err)
		}
	})
}

func TestMemoryRepository_List(t *testing.T) {
	t.Parallel()

	t.Run("list empty repository", func(t *testing.T) {
		t.Parallel()

		repo := NewMemoryRepository()

		result, err := repo.List(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(result) != 0 {
			t.Fatalf("expected 0 rulesets, got %d", len(result))
		}
	})

	t.Run("list returns all saved rulesets", func(t *testing.T) {
		t.Parallel()

		repo := NewMemoryRepository()

		err := repo.Save(context.Background(), &schema.RuleSet{Name: "a", Version: "1.0"})
		if err != nil {
			t.Fatalf("expected no error on save, got %v", err)
		}

		err = repo.Save(context.Background(), &schema.RuleSet{Name: "b", Version: "2.0"})
		if err != nil {
			t.Fatalf("expected no error on save, got %v", err)
		}

		result, err := repo.List(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(result) != 2 {
			t.Fatalf("expected 2 rulesets, got %d", len(result))
		}
	})
}

func TestMemoryRepository_Delete(t *testing.T) {
	t.Parallel()

	t.Run("delete existing ruleset", func(t *testing.T) {
		t.Parallel()

		repo := NewMemoryRepository()

		err := repo.Save(context.Background(), &schema.RuleSet{Name: "test", Version: "1.0"})
		if err != nil {
			t.Fatalf("expected no error on save, got %v", err)
		}

		err = repo.Delete(context.Background(), "test", "1.0")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		_, err = repo.Get(context.Background(), "test", "1.0")
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("delete non-existing ruleset does not error", func(t *testing.T) {
		t.Parallel()

		repo := NewMemoryRepository()

		err := repo.Delete(context.Background(), "missing", "1.0")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestRulesetKey(t *testing.T) {
	t.Parallel()

	got := rulesetKey("name", "v1")
	if got != "name:v1" {
		t.Fatalf("expected 'name:v1', got %q", got)
	}
}
