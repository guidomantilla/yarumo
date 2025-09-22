package parser

import (
	"fmt"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic2/props"
)

// Parse converts a string into a propositional formula using the minimal grammar.
func Parse(input string) (props.Formula, error) {
	p := newParser(input)
	f, err := p.parse()
	if err != nil {
		return nil, err
	}
	if p.cur.typ != tEOF {
		return nil, fmt.Errorf("parse error at %d: extra input after expression", p.cur.pos)
	}
	return props.Simplify(f), nil
}

// MustParse is a helper that panics if Parse fails.
func MustParse(input string) props.Formula {
	f, err := Parse(input)
	if err != nil {
		panic(err)
	}
	return f
}
