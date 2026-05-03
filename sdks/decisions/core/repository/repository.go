package repository

import (
	"context"
	"sync"

	cassert "github.com/guidomantilla/yarumo/common/assert"

	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

var _ Repository = (*memoryRepository)(nil)

// memoryRepository is a thread-safe in-memory implementation of Repository.
type memoryRepository struct {
	mu       sync.RWMutex
	rulesets map[string]*schema.RuleSet
}

// NewMemoryRepository creates a new empty in-memory Repository.
func NewMemoryRepository() Repository {
	return &memoryRepository{
		rulesets: make(map[string]*schema.RuleSet),
	}
}

// rulesetKey builds a lookup key from name and version.
func rulesetKey(name, version string) string {
	return name + ":" + version
}

// Get retrieves a ruleset by name and version.
func (r *memoryRepository) Get(_ context.Context, name string, version string) (*schema.RuleSet, error) {
	cassert.NotNil(r, "repository is nil")

	r.mu.RLock()
	defer r.mu.RUnlock()

	rs, ok := r.rulesets[rulesetKey(name, version)]
	if !ok {
		return nil, ErrGet(ErrNotFound)
	}

	return rs, nil
}

// List returns all available rulesets.
func (r *memoryRepository) List(_ context.Context) ([]schema.RuleSet, error) {
	cassert.NotNil(r, "repository is nil")

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]schema.RuleSet, 0, len(r.rulesets))
	for _, rs := range r.rulesets {
		result = append(result, *rs)
	}

	return result, nil
}

// Save persists a ruleset.
func (r *memoryRepository) Save(_ context.Context, ruleSet *schema.RuleSet) error {
	cassert.NotNil(r, "repository is nil")
	cassert.NotNil(ruleSet, "ruleSet is nil")

	r.mu.Lock()
	defer r.mu.Unlock()

	r.rulesets[rulesetKey(ruleSet.Name, ruleSet.Version)] = ruleSet

	return nil
}

// Delete removes a ruleset by name and version.
func (r *memoryRepository) Delete(_ context.Context, name string, version string) error {
	cassert.NotNil(r, "repository is nil")

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.rulesets, rulesetKey(name, version))

	return nil
}
