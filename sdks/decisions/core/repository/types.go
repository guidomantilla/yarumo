// Package repository defines the rule repository interface and a default in-memory implementation.
package repository

import (
	"context"

	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

// Repository defines the interface for ruleset storage and retrieval.
// Implementations must be safe for concurrent use.
type Repository interface {
	// Get retrieves a ruleset by name and version.
	Get(ctx context.Context, name string, version string) (*schema.RuleSet, error)
	// List returns all available rulesets.
	List(ctx context.Context) ([]schema.RuleSet, error)
	// Save persists a ruleset. The ruleSet must not be nil.
	Save(ctx context.Context, ruleSet *schema.RuleSet) error
	// Delete removes a ruleset by name and version.
	Delete(ctx context.Context, name string, version string) error
}
