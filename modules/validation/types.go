// Package validation provides a config-driven validation engine on top of
// modules/common/validation/ leaves and modules/common/expressions/ predicates.
//
// A Ruleset is a tree of RuleNodes. Each node may declare:
//   - a "field" that selects a value out of the target object via
//     common/validation/.GetField, or
//   - a "when" boolean expression evaluated against the context, or
//   - a list of nested "rules" applied to the current field (or to the root
//     when no field is set), or
//   - a leaf invocation: a rule name plus optional parameters.
//
// Engine.Validate aggregates every violation into one *Error (from
// common/validation/.ErrValidation) so callers serialize all failures with a
// single errs.AsErrorInfo call.
package validation

import (
	"io"
)

var (
	_ Engine = (*engine)(nil)

	_ LoadFn           = LoadYAML
	_ LoadFn           = LoadJSON
	_ LoadReaderFn     = LoadYAMLReader
	_ LoadReaderFn     = LoadJSONReader
	_ LoadFromReaderFn = LoadFromReader
	_ PathOfFn         = PathOf
	_ NewRegistryFn    = NewRegistry
	_ NewRegistryFn    = DefaultRegistry

	_ ErrLoadFn   = ErrLoad
	_ ErrEngineFn = ErrEngine
)

// Engine is the public abstraction for a config-driven validator.
//
// Implementations must be safe for concurrent use by multiple goroutines:
// callers may share a single Engine across handlers and invoke Validate
// concurrently against different objects. The Engine retains a reference to
// the Ruleset and Options supplied at construction time; the caller must not
// mutate them after the Engine has been built.
type Engine interface {
	// Validate runs the loaded ruleset against obj. ctx exposes variables to
	// any "when" expressions evaluated during the run. The returned error, if
	// non-nil, is the domain *cvalidation.Error joining every violation.
	Validate(obj any, ctx map[string]any) error
}

// RuleFn is the function signature for engine leaves. It receives the
// resolved value, optional parameters, and returns an error on violation.
type RuleFn func(value any, params []any) error

// LoadFn is the function type for loading rulesets from bytes.
type LoadFn func(data []byte) (Ruleset, error)

// LoadReaderFn is the function type for loading rulesets from an io.Reader.
type LoadReaderFn func(r io.Reader) (Ruleset, error)

// LoadFromReaderFn is the function type for LoadFromReader.
type LoadFromReaderFn func(r io.Reader, load LoadFn) (Ruleset, error)

// PathOfFn is the function type for PathOf.
type PathOfFn func(err error) string

// NewRegistryFn is the function type for NewRegistry and DefaultRegistry.
type NewRegistryFn func() *Registry

// ErrLoadFn is the function type for ErrLoad.
type ErrLoadFn func(causes ...error) error

// ErrEngineFn is the function type for ErrEngine.
type ErrEngineFn func(causes ...error) error
