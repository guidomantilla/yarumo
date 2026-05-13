package expressions

import (
	"errors"
	"strconv"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// ExpressionType is the error type for expression errors.
const ExpressionType = "expression"

var _ error = (*ParseError)(nil)

var _ error = (*EvalError)(nil)

// ParseError represents a parsing error with position information.
type ParseError struct {
	cerrs.TypedError

	Pos int
	End int
	Msg string
}

// Error returns the error message with position.
func (e *ParseError) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return "parse error at " + strconv.Itoa(e.Pos) + ": " + e.Msg
}

// EvalError represents an evaluation error.
type EvalError struct {
	cerrs.TypedError

	Msg string
}

// Error returns the error message.
func (e *EvalError) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return "eval error: " + e.Msg
}

// Sentinel errors for the expressions package.
var (
	ErrEmptyInput      = errors.New("empty input")
	ErrUnexpectedToken = errors.New("unexpected token")
	ErrUnexpectedEnd   = errors.New("unexpected end of input")
	ErrUnclosedParen   = errors.New("unclosed parenthesis")
	ErrUnclosedBracket = errors.New("unclosed bracket")
	ErrUnclosedString  = errors.New("unclosed string")
	ErrInvalidNumber   = errors.New("invalid number")
	ErrTypeMismatch    = errors.New("type mismatch")
	ErrDivisionByZero  = errors.New("division by zero")
	ErrUnknownField    = errors.New("unknown field")
	ErrUnknownFunc     = errors.New("unknown function")
	ErrArgCount        = errors.New("wrong argument count")
	ErrNilAccess       = errors.New("nil access")
	ErrParseFailed     = errors.New("expression parse failed")
	ErrEvalFailed      = errors.New("expression eval failed")
)

// ErrParse creates a parse error with position information.
func ErrParse(pos, end int, msg string, causes ...error) *ParseError {
	return &ParseError{
		TypedError: cerrs.TypedError{
			Type: ExpressionType,
			Err:  errors.Join(append(causes, ErrParseFailed)...),
		},
		Pos: pos,
		End: end,
		Msg: msg,
	}
}

// ErrEval creates an evaluation error.
func ErrEval(msg string, causes ...error) *EvalError {
	return &EvalError{
		TypedError: cerrs.TypedError{
			Type: ExpressionType,
			Err:  errors.Join(append(causes, ErrEvalFailed)...),
		},
		Msg: msg,
	}
}
