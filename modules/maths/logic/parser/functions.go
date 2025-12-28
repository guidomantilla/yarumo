package parser

import (
	p "github.com/guidomantilla/yarumo/maths/logic/props"
)

// ParseWith parses input using the provided options.
func ParseWith(input string, opts ParseOptions) (p.Formula, error) {
	parser := newParserWithOptions(input, opts)

	f, err := parser.parse()
	if err != nil {
		return nil, err
	}

	if parser.cur.typ != tEOF {
		return nil, newParseError(parser.cur.pos, "extra input after expression")
	}

	return p.Simplify(f), nil
}

// Parse converts a string into a propositional formula using the default (non-strict) grammar.
func Parse(input string) (p.Formula, error) {
	return ParseWith(input, ParseOptions{})
}

// MustParse is a helper that panics if Parse fails.
func MustParse(input string) p.Formula {
	f, err := Parse(input)
	if err != nil {
		panic(err)
	}

	return f
}
