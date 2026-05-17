package validation

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
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
	ErrFieldLookupFailed = errors.New("field lookup failed")
	ErrEngineNil         = errors.New("engine is nil")
	ErrReaderNil         = errors.New("reader is nil")
	ErrDataNil           = errors.New("data is nil")
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
