package parser

import "fmt"

// ParseError provides a structured parsing error with position and message.
// Pos is the byte index where the error was detected (0-based), and Col is a
// human-friendly 1-based column number for single-line inputs.
// Msg is a clear diagnostic about what was expected or found.
type ParseError struct {
	Pos int
	Col int
	Msg string
}

func (e *ParseError) Error() string {
	if e == nil {
		return "<nil>"
	}

	return fmt.Sprintf("parse error at byte %d (col %d): %s", e.Pos, e.Col, e.Msg)
}

func newParseError(pos int, msg string) *ParseError {
	// For our single-line grammar, column is byte offset + 1.
	return &ParseError{Pos: pos, Col: pos + 1, Msg: msg}
}
