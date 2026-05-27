package validation

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// EngineType is the error type classification used when the engine itself
// fails (bad ruleset, unknown rule name, expression failure, …). Leaf
// violations keep the "validation" type inherited from common/validation/.
const EngineType = "validation-engine"

var _ error = (*Error)(nil)

// Error is the domain error for engine-level failures.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted engine error message.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("validation engine %s error: %s", e.Type, e.Err)
}

// Sentinel errors for the engine.
var (
	ErrLoadFailed        = errors.New("ruleset load failed")
	ErrUnknownRule       = errors.New("unknown rule name")
	ErrBadRule           = errors.New("rule is malformed")
	ErrBadParams         = errors.New("rule parameters are invalid")
	ErrWhenEvalFailed    = errors.New("when expression evaluation failed")
	ErrWhenNotBoolean    = errors.New("when expression must evaluate to a boolean")
	ErrWhenParseFailed   = errors.New("when expression parse failed")
	ErrFieldLookupFailed = errors.New("field lookup failed")
	ErrEngineNil         = errors.New("engine is nil")
	ErrReaderNil         = errors.New("reader is nil")
	ErrDataNil           = errors.New("data is nil")
	ErrEmptyGroup        = errors.New("group node has no rules and no leaf name")
	ErrMixedShape        = errors.New("node mixes group fields and leaf name")
	ErrUnknownVersion    = errors.New("ruleset version is not supported by this engine")
	ErrLintFailed        = errors.New("ruleset lint failed")
	ErrCycleDetected     = errors.New("ruleset define references itself transitively")
	ErrUndefinedUse      = errors.New("use references a name not declared in defines")
	ErrUnknownField      = errors.New("field path does not resolve against the bound type")
)

// ErrLoad creates an engine domain error joining the given causes with
// ErrLoadFailed.
func ErrLoad(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: EngineType,
			Err:  errors.Join(append(causes, ErrLoadFailed)...),
		},
	}
}

// ErrEngine creates an engine domain error joining the given causes.
func ErrEngine(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: EngineType,
			Err:  errors.Join(causes...),
		},
	}
}

// pathError carries a field path in error form so AsErrorInfo lists it as a
// leaf message under the "validation" type.
type pathError struct {
	path string
}

// Error returns the path-prefixed marker.
func (p *pathError) Error() string {
	return "field=" + p.path
}

// errPathPrefix builds a sentinel-shaped error that carries the field path
// so AsErrorInfo aggregates it next to the violation.
func errPathPrefix(path string) error {
	return &pathError{path: path}
}

// unknownRuleError carries the offending rule name as a leaf error so it
// shows up under AsErrorInfo.
type unknownRuleError struct {
	name string
}

// Error returns the formatted unknown-rule message.
func (u *unknownRuleError) Error() string {
	return "unknown rule: " + u.name
}

// errUnknownRuleName creates a leaf error tagged with the offending rule name.
func errUnknownRuleName(name string) error {
	return &unknownRuleError{name: name}
}

// messageError carries a node's custom Message so AsErrorInfo surfaces the
// caller-supplied wording above the underlying sentinel.
type messageError struct {
	message string
}

// Error returns the custom message verbatim.
func (m *messageError) Error() string {
	return m.message
}

// errMessage wraps a custom message into an error suitable for prepending
// to a violation chain via cerrs.Wrap.
func errMessage(msg string) error {
	return &messageError{message: msg}
}
