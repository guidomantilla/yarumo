package expressions

// Evaluator integrates parsing and evaluation with configurable options.
type Evaluator struct {
	options Options
}

// NewEvaluator creates a new Evaluator with the given options.
func NewEvaluator(opts ...Option) *Evaluator {
	return &Evaluator{
		options: NewOptions(opts...),
	}
}

// Evaluate parses and evaluates an expression string against the given context.
func (e *Evaluator) Evaluate(input string, ctx Context) (any, error) {
	expr, err := Parse(input)
	if err != nil {
		return nil, err
	}

	return expr.Eval(ctx, e.options.funcs)
}
