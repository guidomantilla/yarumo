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
	cvalidation "github.com/guidomantilla/yarumo/common/validation"
)

// Engine defines the public abstraction for a config-driven validator.
type Engine interface {
	// Validate runs the loaded ruleset against obj. ctx exposes variables to
	// any "when" expressions evaluated during the run. The returned error, if
	// non-nil, is the domain *cvalidation.Error joining every violation.
	Validate(obj any, ctx map[string]any) error
}

// Ruleset is the in-memory shape of a validation configuration. A Ruleset is
// a flat list of top-level rule nodes; each node may itself be a group with
// nested children.
type Ruleset struct {
	Rules []RuleNode `json:"rules,omitempty" yaml:"rules,omitempty"`
}

// RuleNode is a node in the rule tree.
//
// Exactly one of the following shapes is meaningful per node:
//   - Group: Field or When (or both) set and Rules is non-empty.
//   - Leaf:  Name set; Params optional.
//
// Mixing a Group-shaped node with a Name field is not supported and produces
// a configuration error at load time.
type RuleNode struct {
	// Field selects a value out of the target object via dotted path. When
	// empty the group operates on the current value (root or the field of an
	// outer group).
	Field string `json:"field,omitempty" yaml:"field,omitempty"`

	// When is an optional boolean expression. The group only runs when the
	// expression evaluates to a truthy value against the engine context.
	When string `json:"when,omitempty" yaml:"when,omitempty"`

	// Name selects a leaf validator registered with the engine.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Params carries optional positional arguments for the leaf validator.
	Params []any `json:"params,omitempty" yaml:"params,omitempty"`

	// Rules holds nested rule nodes. Only meaningful when this node is a
	// group.
	Rules []RuleNode `json:"rules,omitempty" yaml:"rules,omitempty"`
}

// RuleFn is the function signature for engine leaves. It receives the
// resolved value, optional parameters, and returns an error on violation.
type RuleFn func(value any, params []any) error

// LoadFn is the function type for loading rulesets from bytes.
type LoadFn func(data []byte) (Ruleset, error)

// EngineFactoryFn is the function type for constructing an engine.
type EngineFactoryFn func(rs Ruleset, opts ...Option) Engine

var (
	_ Engine          = (*engine)(nil)
	_ EngineFactoryFn = NewEngine
	_ LoadFn          = LoadYAML
	_ LoadFn          = LoadJSON
)

// nodeContext is the internal context threaded through ruleset walking.
type nodeContext struct {
	value any
	path  string
}

// Type-compliance vars: keep cvalidation referenced so reflect.GetField stays
// a compile-time dependency of the engine.
var (
	_ cvalidation.FieldFn = cvalidation.GetField
)
