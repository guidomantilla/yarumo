package validation

import (
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
)

// Registry maps rule names to RuleFn implementations. It is safe for
// concurrent use.
type Registry struct {
	lock  *sync.RWMutex
	rules map[string]RuleFn
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		lock:  new(sync.RWMutex),
		rules: make(map[string]RuleFn),
	}
}

// Register associates name with fn. Calling Register with an already
// registered name overwrites the previous binding.
func (r *Registry) Register(name string, fn RuleFn) {
	cassert.NotNil(r, "registry is nil")
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(fn, "fn is nil")

	r.lock.Lock()
	defer r.lock.Unlock()

	r.rules[name] = fn
}

// Get returns the RuleFn registered under name. The bool reports whether the
// name was found.
func (r *Registry) Get(name string) (RuleFn, bool) {
	cassert.NotNil(r, "registry is nil")

	r.lock.RLock()
	defer r.lock.RUnlock()

	fn, ok := r.rules[name]

	return fn, ok
}

// Names returns the sorted-by-insertion set of registered names. The result
// is a snapshot; callers may mutate it freely.
func (r *Registry) Names() []string {
	cassert.NotNil(r, "registry is nil")

	r.lock.RLock()
	defer r.lock.RUnlock()

	names := make([]string, 0, len(r.rules))
	for name := range r.rules {
		names = append(names, name)
	}

	return names
}

// DefaultRegistry returns a fresh registry preloaded with every leaf shipped
// by common/validation/. Each call returns a new instance so consumers can
// freely Register custom rules without affecting other engines.
func DefaultRegistry() *Registry {
	r := NewRegistry()

	for name, fn := range builtins {
		r.Register(name, fn)
	}

	return r
}

// RegistryFrom builds a registry from a literal map of rule names to
// RuleFn implementations. Useful for callers shipping their own catalogue.
// A nil map produces an empty registry — never panics.
func RegistryFrom(rules map[string]RuleFn) *Registry {
	r := NewRegistry()

	for name, fn := range rules {
		r.Register(name, fn)
	}

	return r
}

// MergeRegistries returns a new registry containing every entry from base
// merged with each overlay in order. Overlay entries take precedence on
// name collision; the original registries are not mutated.
func MergeRegistries(base *Registry, overlays ...*Registry) *Registry {
	r := NewRegistry()

	if base != nil {
		copyInto(base, r)
	}

	for _, overlay := range overlays {
		if overlay == nil {
			continue
		}

		copyInto(overlay, r)
	}

	return r
}

// Clone returns an independent copy of r so callers can mutate the copy
// without affecting the original (typical use: start from DefaultRegistry
// and add custom rules without touching the package-default).
func (r *Registry) Clone() *Registry {
	cassert.NotNil(r, "registry is nil")

	out := NewRegistry()
	copyInto(r, out)

	return out
}

// copyInto copies every entry from src into dst, overwriting collisions.
func copyInto(src, dst *Registry) {
	src.lock.RLock()
	defer src.lock.RUnlock()

	for name, fn := range src.rules {
		dst.Register(name, fn)
	}
}
