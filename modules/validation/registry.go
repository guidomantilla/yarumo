package validation

import (
	"sync"

	cassert "github.com/guidomantilla/yarumo/common/assert"
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
