// Package parser provides a recursive descent parser for propositional logic formulas.
package parser

import "errors"

// ParseError represents a parsing error with position information.
type ParseError struct {
	Pos int
	Col int
	Msg string
}

// Error returns the error message with position information.
func (e *ParseError) Error() string {
	return e.Msg
}

// Error sentinels for the parser package.
var (
	ErrUnexpectedEnd   = errors.New("unexpected end of input")
	ErrUnexpectedToken = errors.New("unexpected token")
	ErrUnclosedParen   = errors.New("unclosed parenthesis")
	ErrEmptyInput      = errors.New("empty input")
)
