package validation

import "testing"

func TestRegistryFrom_HappyPath(t *testing.T) {
	t.Parallel()

	noop := func(_ any, _ []any) error { return nil }
	r := RegistryFrom(map[string]RuleFn{"a": noop, "b": noop})

	_, ok := r.Get("a")
	if !ok {
		t.Fatalf("expected a, got missing")
	}

	_, ok = r.Get("b")
	if !ok {
		t.Fatalf("expected b, got missing")
	}
}

func TestRegistryFrom_Nil(t *testing.T) {
	t.Parallel()

	r := RegistryFrom(nil)
	if r == nil {
		t.Fatalf("expected non-nil registry, got nil")
	}

	if len(r.Names()) != 0 {
		t.Fatalf("expected empty registry, got %d entries", len(r.Names()))
	}
}

func TestMergeRegistries_OverlayWins(t *testing.T) {
	t.Parallel()

	mark := func(s string) RuleFn {
		return func(_ any, _ []any) error { return errMarker(s) }
	}

	base := RegistryFrom(map[string]RuleFn{"a": mark("base-a"), "b": mark("base-b")})
	overlay := RegistryFrom(map[string]RuleFn{"a": mark("overlay-a"), "c": mark("overlay-c")})

	merged := MergeRegistries(base, overlay)

	fn, ok := merged.Get("a")
	if !ok {
		t.Fatalf("expected a, got missing")
	}

	err := fn(nil, nil)
	if err.Error() != "overlay-a" {
		t.Fatalf("expected overlay-a to win, got %v", err)
	}

	_, ok = merged.Get("b")
	if !ok {
		t.Fatalf("expected b from base, got missing")
	}

	_, ok = merged.Get("c")
	if !ok {
		t.Fatalf("expected c from overlay, got missing")
	}
}

func TestRegistry_Clone_Independent(t *testing.T) {
	t.Parallel()

	noop := func(_ any, _ []any) error { return nil }
	original := RegistryFrom(map[string]RuleFn{"a": noop})

	clone := original.Clone()
	clone.Register("b", noop)

	_, ok := original.Get("b")
	if ok {
		t.Fatalf("clone mutation leaked into original")
	}
}

type errMarker string

func (e errMarker) Error() string { return string(e) }
