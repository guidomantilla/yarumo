package rules

import (
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"
)

type Rule[T any] struct {
	Label       string
	Formula     propositions.Formula
	Consequence *propositions.Var
	tree        *logic.EvalNode
}
