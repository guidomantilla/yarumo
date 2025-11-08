package parser

import (
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/props"
)

// ParseWith parses input using the provided options.
func ParseWith(input string, opts ParseOptions) (props.Formula, error) {
	p := newParserWithOptions(input, opts)
	f, err := p.parse()
	if err != nil {
		return nil, err
	}
	if p.cur.typ != tEOF {
		return nil, newParseError(p.cur.pos, "extra input after expression")
	}
	return props.Simplify(f), nil
}

// Parse converts a string into a propositional formula using the default (non-strict) grammar.
func Parse(input string) (props.Formula, error) {
	return ParseWith(input, ParseOptions{})
}

// MustParse is a helper that panics if Parse fails.
func MustParse(input string) props.Formula {
	f, err := Parse(input)
	if err != nil {
		panic(err)
	}
	return f
}
