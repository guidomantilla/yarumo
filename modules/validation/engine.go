package validation

import (
	"errors"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
	cexpressions "github.com/guidomantilla/yarumo/common/expressions"
	cvalidation "github.com/guidomantilla/yarumo/extensions/common/validation"
)

// engine is the default Engine implementation.
type engine struct {
	ruleset Ruleset
	options *Options
}

// nodeContext is the internal context threaded through ruleset walking.
type nodeContext struct {
	value any
	path  string
}

// NewEngine creates an Engine bound to the given ruleset.
func NewEngine(rs Ruleset, opts ...Option) Engine {
	return &engine{
		ruleset: rs,
		options: NewOptions(opts...),
	}
}

// Validate runs every rule in the ruleset against obj. ctx exposes variables
// to any "when" expression in the ruleset.
func (e *engine) Validate(obj any, ctx map[string]any) error {
	cassert.NotNil(e, "engine is nil")

	if ctx == nil {
		ctx = map[string]any{}
	}

	root := nodeContext{value: obj, path: ""}

	violations := make([]error, 0, len(e.ruleset.Rules))
	for _, node := range e.ruleset.Rules {
		errs := e.runNode(obj, root, node, ctx)
		violations = append(violations, errs...)
	}

	if len(violations) == 0 {
		return nil
	}

	return cvalidation.ErrValidation(violations...)
}

// runNode dispatches a single RuleNode and returns any violations.
func (e *engine) runNode(root any, current nodeContext, node RuleNode, ctx map[string]any) []error {
	switch {
	case node.When != "":
		return e.runConditional(root, current, node, ctx)
	case node.Field != "":
		return e.runField(root, node, ctx)
	case node.Name != "":
		return e.runLeaf(current, node)
	case len(node.Rules) > 0:
		return e.runGroup(root, current, node, ctx)
	default:
		return []error{ErrEngine(ErrBadRule)}
	}
}

// runConditional evaluates the when-expression and, if truthy, descends into
// the nested rules (each evaluated against the root).
func (e *engine) runConditional(root any, current nodeContext, node RuleNode, ctx map[string]any) []error {
	pass, err := e.evalWhen(node.When, ctx)
	if err != nil {
		return []error{err}
	}

	if !pass {
		return nil
	}

	out := make([]error, 0, len(node.Rules))
	for _, child := range node.Rules {
		errs := e.runNode(root, current, child, ctx)
		out = append(out, errs...)
	}

	return out
}

// runField resolves the field path against root and runs nested rules
// against the resolved value.
func (e *engine) runField(root any, node RuleNode, ctx map[string]any) []error {
	value, err := cvalidation.GetField(root, node.Field)
	if err != nil {
		return []error{annotatePath(node.Field, err)}
	}

	current := nodeContext{value: value, path: node.Field}

	out := make([]error, 0, len(node.Rules))
	for _, child := range node.Rules {
		errs := e.runChildOnField(root, current, child, ctx)
		out = append(out, errs...)
	}

	return out
}

// runChildOnField applies a child rule node, but resolves nested field paths
// relative to the parent field instead of the root.
func (e *engine) runChildOnField(root any, current nodeContext, node RuleNode, ctx map[string]any) []error {
	if node.Field != "" {
		nested := joinPath(current.path, node.Field)
		clone := node
		clone.Field = nested

		return e.runNode(root, current, clone, ctx)
	}

	return e.runNode(root, current, node, ctx)
}

// runGroup runs a group node without a field or when clause: it just
// dispatches every child against the current value.
func (e *engine) runGroup(root any, current nodeContext, node RuleNode, ctx map[string]any) []error {
	out := make([]error, 0, len(node.Rules))
	for _, child := range node.Rules {
		errs := e.runChildOnField(root, current, child, ctx)
		out = append(out, errs...)
	}

	return out
}

// runLeaf resolves the leaf and invokes it on the current value.
func (e *engine) runLeaf(current nodeContext, node RuleNode) []error {
	fn, ok := e.options.registry.Get(node.Name)
	if !ok {
		return []error{ErrEngine(cerrs.Wrap(ErrUnknownRule, errUnknownRuleName(node.Name)))}
	}

	err := fn(current.value, node.Params)
	if err != nil {
		return []error{annotatePath(current.path, err)}
	}

	return nil
}

// evalWhen evaluates the predicate against ctx and converts the result to a
// boolean. A reference to an undefined context variable evaluates to false,
// matching the typical web-app pattern where rulesets reference optional
// context flags. Genuine syntax or evaluation failures return
// ErrWhenEvalFailed. Non-boolean truthy results return ErrWhenNotBoolean.
func (e *engine) evalWhen(expr string, ctx map[string]any) (bool, error) {
	result, err := e.options.evaluator.Evaluate(expr, cexpressions.Context(ctx))
	if err != nil {
		if errors.Is(err, cexpressions.ErrUnknownField) {
			return false, nil
		}

		return false, ErrEngine(cerrs.Wrap(ErrWhenEvalFailed, err))
	}

	b, ok := result.(bool)
	if !ok {
		return false, ErrEngine(ErrWhenNotBoolean)
	}

	return b, nil
}
