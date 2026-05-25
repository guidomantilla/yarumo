package rbac

import (
	"context"
	"sort"
	"sync"
	"testing"
)

func TestNewInMemoryRolesStore(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil store", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryRolesStore()
		if s == nil {
			t.Fatal("expected non-nil store")
		}
	})

	t.Run("empty store returns nil roles", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryRolesStore()

		roles, err := s.Roles(context.Background(), "alice")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if roles != nil {
			t.Fatalf("expected nil roles, got %v", roles)
		}
	})
}

func TestInMemoryRolesStore_Assign(t *testing.T) {
	t.Parallel()

	t.Run("assigns single role", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryRolesStore()
		s.Assign("alice", "viewer")

		roles, err := s.Roles(context.Background(), "alice")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(roles) != 1 || roles[0] != "viewer" {
			t.Fatalf("expected [viewer], got %v", roles)
		}
	})

	t.Run("assigns multiple roles", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryRolesStore()
		s.Assign("alice", "viewer", "editor")

		roles, _ := s.Roles(context.Background(), "alice")
		sort.Strings(roles)

		if len(roles) != 2 || roles[0] != "editor" || roles[1] != "viewer" {
			t.Fatalf("expected [editor viewer], got %v", roles)
		}
	})

	t.Run("re-assign replaces", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryRolesStore()
		s.Assign("alice", "viewer")
		s.Assign("alice", "admin")

		roles, _ := s.Roles(context.Background(), "alice")
		if len(roles) != 1 || roles[0] != "admin" {
			t.Fatalf("expected [admin], got %v", roles)
		}
	})

	t.Run("empty principal id ignored", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryRolesStore()
		s.Assign("", "viewer")

		roles, _ := s.Roles(context.Background(), "")
		if roles != nil {
			t.Fatalf("expected nil, got %v", roles)
		}
	})

	t.Run("empty roles filtered", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryRolesStore()
		s.Assign("alice", "", "viewer", "")

		roles, _ := s.Roles(context.Background(), "alice")
		if len(roles) != 1 || roles[0] != "viewer" {
			t.Fatalf("expected [viewer], got %v", roles)
		}
	})

	t.Run("all-empty role list removes assignment", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryRolesStore()
		s.Assign("alice", "viewer")
		s.Assign("alice", "", "")

		roles, _ := s.Roles(context.Background(), "alice")
		if roles != nil {
			t.Fatalf("expected nil after empty assign, got %v", roles)
		}
	})

	t.Run("returned slice is defensive copy", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryRolesStore()
		s.Assign("alice", "viewer")

		roles, _ := s.Roles(context.Background(), "alice")
		roles[0] = "tampered"

		// Re-fetch and confirm the store was not mutated.
		again, _ := s.Roles(context.Background(), "alice")
		if again[0] != "viewer" {
			t.Fatalf("store was mutated through returned slice: got %v", again)
		}
	})
}

func TestInMemoryRolesStore_Unassign(t *testing.T) {
	t.Parallel()

	t.Run("removes assignment", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryRolesStore()
		s.Assign("alice", "viewer")
		s.Unassign("alice")

		roles, _ := s.Roles(context.Background(), "alice")
		if roles != nil {
			t.Fatalf("expected nil after Unassign, got %v", roles)
		}
	})

	t.Run("empty id ignored", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryRolesStore()
		s.Unassign("") // must not panic
	})

	t.Run("missing id is a no-op", func(t *testing.T) {
		t.Parallel()

		s := NewInMemoryRolesStore()
		s.Unassign("ghost") // must not panic
	})
}

func TestInMemoryRolesStore_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	s := NewInMemoryRolesStore()
	s.Assign("alice", "viewer")

	var wg sync.WaitGroup

	for range 32 {
		wg.Add(2)

		go func() {
			defer wg.Done()

			s.Assign("alice", "viewer", "editor")
		}()

		go func() {
			defer wg.Done()

			_, _ = s.Roles(context.Background(), "alice")
		}()
	}

	wg.Wait()
}
