package rules

import (
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/predicates"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"
)

type Rule[T any] struct {
	Label       string
	Formula     propositions.Formula
	Predicate   predicates.Predicate[T]
	Consequence *propositions.Var
}
