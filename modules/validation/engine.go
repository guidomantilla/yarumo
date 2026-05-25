package validation

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
	cexpressions "github.com/guidomantilla/yarumo/common/expressions"
	cutils "github.com/guidomantilla/yarumo/common/utils"
	cvalidation "github.com/guidomantilla/yarumo/common/validation"
)

// engine is the default Engine implementation.
type engine struct {
	ruleset   Ruleset
	options   *Options
	whenCache sync.Map // map[string]cexpressions.Expr
}

// nodeContext is the internal context threaded through ruleset walking.
type nodeContext struct {
	value any
	path  string
}

// NewEngine creates an Engine bound to the given ruleset. The constructor
// pre-parses every when: expression in the ruleset and caches the AST so
// subsequent Validate / Run calls do not re-parse. Parse failures are
// stored as a sentinel and surface as ErrWhenParseFailed at evaluation
// time; if you want a fail-at-boot check use BuildEngine with
// WithLintOnLoad.
func NewEngine(rs Ruleset, opts ...Option) Engine {
	e := &engine{
		ruleset: rs,
		options: NewOptions(opts...),
	}

	primeWhenCache(rs.Rules, &e.whenCache)

	return e
}

// BuildEngine creates an Engine and, when WithLintOnLoad is set, runs
// Validate against the ruleset before returning so callers fail fast on a
// broken configuration. Without WithLintOnLoad it behaves like NewEngine
// and always returns a nil error.
func BuildEngine(rs Ruleset, opts ...Option) (Engine, error) {
	options := NewOptions(opts...)
	if options.lintOnLoad {
		err := Validate(rs, opts...)
		if err != nil {
			return nil, err
		}
	}

	return NewEngine(rs, opts...), nil
}

// primeWhenCache walks the ruleset tree and pre-parses every when:
// expression so Engine.Validate / Engine.Run avoid the parser on the hot
// path. Parse errors are not stored — evalWhen will discover and report
// them lazily so the cache remains tidy.
func primeWhenCache(nodes []RuleNode, cache *sync.Map) {
	for i := range nodes {
		if nodes[i].When != "" {
			expr, err := cexpressions.Parse(nodes[i].When)
			if err == nil {
				cache.Store(nodes[i].When, expr)
			}
		}

		primeWhenCache(nodes[i].Rules, cache)
	}
}

// Validate runs every rule in the ruleset against obj. ctx exposes variables
// to any "when" expression in the ruleset. The validated object itself is
// exposed under the "obj" key so when: expressions can reference fields
// like `obj.Role == "admin"`.
func (e *engine) Validate(obj any, ctx map[string]any) error {
	cassert.NotNil(e, "engine is nil")

	ctx = enrichContext(ctx, obj)

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

// enrichContext returns a copy of ctx with the validated object exposed
// under the "obj" key so when: expressions can reach into the payload. The
// caller's ctx is never mutated. Existing "obj" entries are preserved if
// the caller already supplied one.
//
// Struct objects are flattened to map[string]any via reflection so the
// expressions engine (which only understands map access) can read fields
// like `obj.Name`. Maps, slices, and primitives pass through unchanged.
func enrichContext(ctx map[string]any, obj any) map[string]any {
	out := make(map[string]any, len(ctx)+1)
	for k, v := range ctx {
		out[k] = v
	}

	_, has := out["obj"]
	if !has {
		out["obj"] = toExpressionView(obj)
	}

	return out
}

// toExpressionView converts a struct (or pointer to struct) into a
// map[string]any so the expressions engine can walk it via dot notation.
// Non-struct inputs are returned unchanged.
func toExpressionView(obj any) any {
	rv := reflect.ValueOf(obj)
	for rv.IsValid() && rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return obj
		}

		rv = rv.Elem()
	}

	if !rv.IsValid() || rv.Kind() != reflect.Struct {
		return obj
	}

	out := make(map[string]any, rv.NumField())

	rt := rv.Type()
	for i := range rv.NumField() {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}

		out[field.Name] = rv.Field(i).Interface()
	}

	return out
}

// Run is the structured counterpart to Validate. It runs the same logic but
// converts each engine-level error into a Violation so callers can branch on
// rule/path/cause without parsing error strings.
func (e *engine) Run(obj any, ctx map[string]any) []Violation {
	cassert.NotNil(e, "engine is nil")

	err := e.Validate(obj, ctx)
	if err == nil {
		return nil
	}

	return toViolations(err)
}

// runNode dispatches a single RuleNode and returns any violations. The
// Optional flag is honoured at this layer: when set and the current value
// is the zero value of its type, the node is skipped entirely.
func (e *engine) runNode(root any, current nodeContext, node RuleNode, ctx map[string]any) []error {
	if node.Optional && cutils.Empty(current.value) {
		return nil
	}

	switch {
	case node.When != "":
		return e.runConditional(root, current, node, ctx)
	case node.Field != "":
		return e.runField(root, node, ctx)
	case isCombinator(node.Name):
		return e.runCombinator(root, current, node, ctx)
	case node.Name != "":
		return e.applyMessage(node, e.runLeaf(current, node))
	case len(node.Rules) > 0:
		return e.runGroup(root, current, node, ctx)
	default:
		return []error{ErrEngine(ErrBadRule)}
	}
}

// runCombinator dispatches the four pseudo-leaves any_of / all_of / not /
// for_each against the current value.
func (e *engine) runCombinator(root any, current nodeContext, node RuleNode, ctx map[string]any) []error {
	switch node.Name {
	case combinatorAnyOf:
		return e.applyMessage(node, e.runAnyOf(root, current, node, ctx))
	case combinatorAllOf:
		return e.applyMessage(node, e.runAllOf(root, current, node, ctx))
	case combinatorNot:
		return e.applyMessage(node, e.runNot(root, current, node, ctx))
	case combinatorForEach:
		return e.applyMessage(node, e.runForEach(root, current, node, ctx))
	default:
		return []error{ErrEngine(ErrBadRule)}
	}
}

// runAnyOf passes when any nested rule passes; reports every attempt only
// when all nested rules fail.
func (e *engine) runAnyOf(root any, current nodeContext, node RuleNode, ctx map[string]any) []error {
	var causes []error

	for _, child := range node.Rules {
		errs := e.runChildOnField(root, current, child, ctx)
		if len(errs) == 0 {
			return nil
		}

		causes = append(causes, errs...)
	}

	return causes
}

// runAllOf passes only when every nested rule passes; aggregates every
// violation.
func (e *engine) runAllOf(root any, current nodeContext, node RuleNode, ctx map[string]any) []error {
	var causes []error

	for _, child := range node.Rules {
		errs := e.runChildOnField(root, current, child, ctx)
		causes = append(causes, errs...)
	}

	return causes
}

// runNot passes when the (single) nested rule fails. The combinator
// expects exactly one nested rule; extra rules are evaluated as AnyOf for
// rejection — any one passing causes the Not to fail.
func (e *engine) runNot(root any, current nodeContext, node RuleNode, ctx map[string]any) []error {
	for _, child := range node.Rules {
		errs := e.runChildOnField(root, current, child, ctx)
		if len(errs) == 0 {
			return []error{cvalidation.ErrValidation(cvalidation.ErrAssertionInverted)}
		}
	}

	return nil
}

// runForEach applies the nested rules against every element of a slice or
// array (or every value of a map). The current value must be enumerable.
func (e *engine) runForEach(root any, current nodeContext, node RuleNode, ctx map[string]any) []error {
	xs, kind, err := enumerate(current.value)
	if err != nil {
		return []error{annotatePath(current.path, err)}
	}

	var causes []error

	for i, x := range xs {
		element := nodeContext{value: x, path: indexPath(current.path, kind, i)}
		for _, child := range node.Rules {
			errs := e.runChildOnField(root, element, child, ctx)
			causes = append(causes, errs...)
		}
	}

	return causes
}

// enumerate returns a slice of the values inside v (elements of a slice or
// array, or values of a map). The boolean kind tracks whether the source
// was a slice/array (true) or a map (false) for path formatting.
func enumerate(v any) ([]any, bool, error) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return nil, false, ErrEngine(ErrBadParams)
	}

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		out := make([]any, rv.Len())
		for i := range rv.Len() {
			out[i] = rv.Index(i).Interface()
		}

		return out, true, nil
	case reflect.Map:
		out := make([]any, 0, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			out = append(out, iter.Value().Interface())
		}

		return out, false, nil
	default:
		return nil, false, ErrEngine(ErrBadParams)
	}
}

// indexPath builds a per-element path like "items[0]" for slices/arrays or
// "items.<i>" for map values where the iteration order is not stable.
func indexPath(parent string, sliceLike bool, i int) string {
	if parent == "" {
		return ""
	}

	if sliceLike {
		return parent + "[" + intToStr(i) + "]"
	}

	return parent + "." + intToStr(i)
}

// intToStr is a tiny allocation-free int -> decimal converter used by
// indexPath to avoid pulling fmt for one call.
func intToStr(i int) string {
	if i == 0 {
		return "0"
	}

	var buf [20]byte

	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}

	return string(buf[pos:])
}

// applyMessage wraps every violation with the node's custom Message when
// set, so callers see the friendly text in the aggregated error. The
// underlying cause is preserved for errors.Is.
func (e *engine) applyMessage(node RuleNode, errs []error) []error {
	if node.Message == "" || len(errs) == 0 {
		return errs
	}

	wrapped := make([]error, 0, len(errs))
	for _, err := range errs {
		wrapped = append(wrapped, cerrs.Wrap(errMessage(node.Message), err))
	}

	return wrapped
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

// runLeaf resolves the leaf and invokes it on the current value. Hook
// callbacks fire around the invocation so consumers can collect metrics or
// emit telemetry without forking the engine.
func (e *engine) runLeaf(current nodeContext, node RuleNode) []error {
	fn, ok := e.options.registry.Get(node.Name)
	if !ok {
		return []error{ErrEngine(cerrs.Wrap(ErrUnknownRule, errUnknownRuleName(node.Name)))}
	}

	hookCtx := context.Background()
	e.options.hook.BeforeRule(hookCtx, current.path, node.Name, node.Params)

	start := time.Now()
	err := fn(current.value, node.Params)
	took := time.Since(start)

	e.options.hook.AfterRule(hookCtx, current.path, node.Name, node.Params, err, took)

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
//
// The expression is parsed only on the first call for any given string —
// subsequent calls hit the engine's whenCache and skip the parser.
func (e *engine) evalWhen(expr string, ctx map[string]any) (bool, error) {
	parsed, hit := e.whenCache.Load(expr)
	if !hit {
		result, err := e.options.evaluator.Evaluate(expr, cexpressions.Context(ctx))

		return interpretWhenResult(result, err)
	}

	ast, ok := parsed.(cexpressions.Expr)
	if !ok {
		return false, ErrEngine(ErrWhenEvalFailed)
	}

	result, err := ast.Eval(cexpressions.Context(ctx), cexpressions.DefaultFuncs())

	return interpretWhenResult(result, err)
}

// interpretWhenResult is the shared post-eval branching used by both the
// cached and uncached when-evaluation paths.
func interpretWhenResult(result any, err error) (bool, error) {
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
