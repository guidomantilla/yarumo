package cache

import "testing"

func TestResolveKeyPrefix(t *testing.T) {
	t.Parallel()

	t.Run("returns configured prefix when non-empty", func(t *testing.T) {
		t.Parallel()

		got := ResolveKeyPrefix("ignored-name", "explicit::")
		if got != "explicit::" {
			t.Fatalf("got %q, want %q", got, "explicit::")
		}
	})

	t.Run("falls back to name+: when configured prefix is empty", func(t *testing.T) {
		t.Parallel()

		got := ResolveKeyPrefix("alpha", "")
		if got != "alpha:" {
			t.Fatalf("got %q, want %q", got, "alpha:")
		}
	})
}
