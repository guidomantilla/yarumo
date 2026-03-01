package probability

import (
	"maps"
	"slices"
	"strings"
)

// NewCPT creates an empty conditional probability table for the given variable and parents.
func NewCPT(variable Var, parents []Var) CPT {
	parentsCopy := make([]Var, len(parents))
	copy(parentsCopy, parents)

	return CPT{
		Variable: variable,
		Parents:  parentsCopy,
		Entries:  make(map[string]Distribution),
	}
}

// Set assigns a distribution for a given parent configuration.
func (c CPT) Set(parentConfig Assignment, dist Distribution) {
	key := serializeAssignment(parentConfig, c.Parents)
	copied := make(Distribution, len(dist))
	maps.Copy(copied, dist)

	c.Entries[key] = copied
}

// Lookup returns the distribution for a given parent configuration.
func (c CPT) Lookup(parentConfig Assignment) (Distribution, error) {
	key := serializeAssignment(parentConfig, c.Parents)

	dist, ok := c.Entries[key]
	if !ok {
		return nil, ErrOutcomeNotFound
	}

	copied := make(Distribution, len(dist))
	maps.Copy(copied, dist)

	return copied, nil
}

// Validate checks that all entries in the CPT are valid distributions.
func (c CPT) Validate() error {
	if len(c.Entries) == 0 {
		return ErrEmptyDist
	}

	for _, dist := range c.Entries {
		if !IsValid(dist) {
			return ErrNotNormalized
		}
	}

	return nil
}

// serializeAssignment produces a deterministic string key for a parent configuration.
func serializeAssignment(config Assignment, order []Var) string {
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
func SerializeAssignmentSorted(config Assignment) string {
	keys := make([]Var, 0, len(config))

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
