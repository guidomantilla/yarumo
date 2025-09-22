package parser

import (
	"errors"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic2/props"
)

// Parse converts a string into a propositional formula.
// Phase 0: stub that returns an error. Real implementation in Phase 1.
func Parse(input string) (props.Formula, error) {
	return nil, errors.New("logic2/parser: Parse not implemented (Phase 0)")
}

// MustParse is a helper that panics if Parse fails.
// Phase 0: always panics; it will be implemented in Phase 1.
func MustParse(input string) props.Formula {
	panic("logic2/parser: MustParse not implemented (Phase 0)")
}
