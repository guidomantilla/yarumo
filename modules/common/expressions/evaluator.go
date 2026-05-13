package expressions

import (
	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// evaluator is the default Evaluator implementation backed by Parse + Expr.Eval.
type evaluator struct {
	options Options
}

// NewEvaluator creates a new Evaluator with the given options.
func NewEvaluator(opts ...Option) Evaluator {
	return &evaluator{
		options: NewOptions(opts...),
	}
}

// Evaluate parses and evaluates an expression string against the given context.
func (e *evaluator) Evaluate(input string, ctx Context) (any, error) {
	cassert.NotNil(e, "evaluator is nil")

	expr, err := Parse(input)
	if err != nil {
		return nil, err
	}

	return expr.Eval(ctx, e.options.funcs)
}
