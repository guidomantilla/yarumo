package bayesian

import (
	"maps"
	"slices"
	"strings"

	"github.com/guidomantilla/yarumo/compute/math/stats"
)

// NewCPT creates an empty conditional probability table for the given variable and parents.
func NewCPT(variable stats.Var, parents []stats.Var) CPT {
	parentsCopy := make([]stats.Var, len(parents))
	copy(parentsCopy, parents)

	return CPT{
		Variable: variable,
		Parents:  parentsCopy,
		Entries:  make(map[string]stats.Distribution),
	}
}

// Set assigns a distribution for a given parent configuration.
func (c CPT) Set(parentConfig stats.Assignment, dist stats.Distribution) {
	key := serializeAssignment(parentConfig, c.Parents)
	copied := make(stats.Distribution, len(dist))
	maps.Copy(copied, dist)

	c.Entries[key] = copied
}

// Lookup returns the distribution for a given parent configuration.
func (c CPT) Lookup(parentConfig stats.Assignment) (stats.Distribution, error) {
	key := serializeAssignment(parentConfig, c.Parents)

	dist, ok := c.Entries[key]
	if !ok {
		return nil, stats.ErrOutcomeNotFound
	}

	copied := make(stats.Distribution, len(dist))
	maps.Copy(copied, dist)

	return copied, nil
}

// Validate checks that all entries in the CPT are valid distributions.
func (c CPT) Validate() error {
	if len(c.Entries) == 0 {
		return stats.ErrEmptyDist
	}

	for _, dist := range c.Entries {
		if !stats.IsValid(dist) {
			return stats.ErrNotNormalized
		}
	}

	return nil
}

// serializeAssignment produces a deterministic string key for a parent configuration.
func serializeAssignment(config stats.Assignment, order []stats.Var) string {
	if len(order) == 0 {
		return ""
	}

	parts := make([]string, 0, len(order))

	for _, v := range order {
		parts = append(parts, string(v)+"="+string(config[v]))
	}

	return strings.Join(parts, ",")
}

// SerializeAssignmentSorted produces a deterministic string key sorted by variable name.
func SerializeAssignmentSorted(config stats.Assignment) string {
	keys := make([]stats.Var, 0, len(config))

	for k := range config {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	parts := make([]string, 0, len(keys))

	for _, v := range keys {
		parts = append(parts, string(v)+"="+string(config[v]))
	}

	return strings.Join(parts, ",")
}
