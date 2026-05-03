package examples

import (
	"context"
	"fmt"

	"github.com/guidomantilla/yarumo/decisions/core/schema"
	"github.com/guidomantilla/yarumo/decisions/core/repository"
)

// memoryRepo is an in-memory repository used by examples.
type memoryRepo struct {
	rulesets map[string]*schema.RuleSet
}

func (r *memoryRepo) Get(_ context.Context, name string, version string) (*schema.RuleSet, error) {
	key := name + ":" + version
	rs, ok := r.rulesets[key]

	if !ok {
		return nil, fmt.Errorf("ruleset %s:%s not found", name, version)
	}

	return rs, nil
}

func (r *memoryRepo) List(_ context.Context) ([]schema.RuleSet, error) {
	result := make([]schema.RuleSet, 0, len(r.rulesets))

	for _, rs := range r.rulesets {
		result = append(result, *rs)
	}

	return result, nil
}

func (r *memoryRepo) Save(_ context.Context, rs *schema.RuleSet) error {
	if r.rulesets == nil {
		r.rulesets = make(map[string]*schema.RuleSet)
	}

	r.rulesets[rs.Name+":"+rs.Version] = rs

	return nil
}

func (r *memoryRepo) Delete(_ context.Context, name string, version string) error {
	delete(r.rulesets, name+":"+version)

	return nil
}

// Verify interface compliance.
var _ repository.Repository = (*memoryRepo)(nil)
