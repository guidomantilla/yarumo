// Package parser provides a recursive descent parser for propositional logic formulas.
package parser

import (
	"errors"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error domain type for parser errors.
const (
	ParserType = "math-parser"
)

var _ error = (*ParseError)(nil)

// ParseError represents a parsing error with position information.
type ParseError struct {
	cerrs.TypedError

	Pos int
	Col int
	Msg string
}

// Error returns the error message.
func (e *ParseError) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return e.Msg
}

// Error sentinels for the parser package.
var (
	ErrUnexpectedEnd   = errors.New("unexpected end of input")
	ErrUnexpectedToken = errors.New("unexpected token")
	ErrUnclosedParen   = errors.New("unclosed parenthesis")
	ErrEmptyInput      = errors.New("empty input")
	ErrParseFailed     = errors.New("parser operation failed")
)

// ErrParse creates a parse error with position information.
func ErrParse(pos, col int, msg string, causes ...error) *ParseError {
	return &ParseError{
		TypedError: cerrs.TypedError{
			Type: ParserType,
			Err:  errors.Join(append(causes, ErrParseFailed)...),
		},
		Pos: pos,
		Col: col,
		Msg: msg,
	}
}
