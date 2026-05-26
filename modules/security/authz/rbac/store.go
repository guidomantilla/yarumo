package rbac

import (
	"context"
	"sync"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

var (
	_ RolesStore = (*InMemoryRolesStore)(nil)
)

// InMemoryRolesStore implements RolesStore as a concurrent-safe map.
// It is the default store wired by NewOptions when the caller does
// not pass WithRolesStore.
//
// The struct is exported (instead of returned through the RolesStore
// interface) because callers need direct access to the mutator
// methods (Assign / Unassign) — there is no abstraction worth hiding
// here. Consumers that want a different store (DB-backed, remote
// service, etc.) implement RolesStore themselves and wire it via
// WithRolesStore.
type InMemoryRolesStore struct {
	mu    sync.RWMutex
	roles map[string][]string
}

// NewInMemoryRolesStore creates a fresh in-memory RolesStore. The
// returned store is empty; populate it via Assign before evaluating
// requests against it.
func NewInMemoryRolesStore() *InMemoryRolesStore {
	return &InMemoryRolesStore{
		roles: map[string][]string{},
	}
}

// Roles returns the role list assigned to principalID. The returned
// slice is a defensive copy so callers cannot mutate the store's
// internal state.
func (s *InMemoryRolesStore) Roles(_ context.Context, principalID string) ([]string, error) {
	cassert.NotNil(s, "store is nil")

	s.mu.RLock()
	defer s.mu.RUnlock()

	roles := s.roles[principalID]
	if len(roles) == 0 {
		return nil, nil
	}

	out := make([]string, len(roles))
	copy(out, roles)

	return out, nil
}

// Assign sets the role list for principalID, replacing any previous
// assignment. Empty principal ids are silently ignored. Empty role
// names inside the list are filtered out.
func (s *InMemoryRolesStore) Assign(principalID string, roles ...string) {
	cassert.NotNil(s, "store is nil")

	if principalID == "" {
		return
	}

	filtered := make([]string, 0, len(roles))
	for _, r := range roles {
		if r != "" {
			filtered = append(filtered, r)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(filtered) == 0 {
		delete(s.roles, principalID)

		return
	}

	s.roles[principalID] = filtered
}

// Unassign removes every role assignment for principalID. Empty
// principal ids are silently ignored. Calling Unassign on a principal
// that has no assignment is a no-op.
func (s *InMemoryRolesStore) Unassign(principalID string) {
	cassert.NotNil(s, "store is nil")

	if principalID == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.roles, principalID)
}
