package evaluate

import (
	"context"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	cexpressions "github.com/guidomantilla/yarumo/common/expressions"

	"github.com/guidomantilla/yarumo/decisions/core/explain"
	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

func runTree(ctx context.Context, config *schema.TreeConfig, exprCtx cexpressions.Context, opts Options) (Result, error) {
	evaluator := cexpressions.NewEvaluator(opts.expressionOpts...)

	path, outputs, err := walkTree(evaluator, &config.Root, exprCtx)
	if err != nil {
		return Result{}, err
	}

	traceSteps := make([]explain.TreeStep, len(path))
	pathStrings := make([]string, len(path))

	for i, step := range path {
		traceSteps[i] = step
		pathStrings[i] = step.Condition
	}

	trace := explain.TreeTrace{
		Path:    traceSteps,
		Outputs: outputs,
	}

	explanation, err := opts.treeExplainer.ExplainTree(ctx, trace)
	if err != nil {
		return Result{}, ErrExplain(err)
	}

	return Result{
		Outcome: Outcome{
			Tree: &TreeOutcome{
				Path:    pathStrings,
				Outputs: outputs,
			},
		},
		Explanation: explanation,
		Paradigm:    Tree,
	}, nil
}

func walkTree(evaluator cexpressions.Evaluator, node *schema.TreeNodeDef, exprCtx cexpressions.Context) ([]explain.TreeStep, map[string]any, error) {
	if node.Output != nil {
		return nil, node.Output, nil
	}

	if node.Condition == "" {
		return nil, nil, cerrs.Wrap(ErrMissingConfig)
	}

	val, err := evaluator.Evaluate(node.Condition, exprCtx)
	if err != nil {
		return nil, nil, cerrs.Wrap(ErrConditionEval, err)
	}

	boolVal, ok := val.(bool)
	if !ok {
		return nil, nil, cerrs.Wrap(ErrConditionEval)
	}

	step := explain.TreeStep{Condition: node.Condition, Result: boolVal}

	var next *schema.TreeNodeDef

	if boolVal {
		next = node.True
	} else {
		next = node.False
	}

	if next == nil {
		return nil, nil, cerrs.Wrap(ErrMissingConfig)
	}

	childPath, outputs, err := walkTree(evaluator, next, exprCtx)
	if err != nil {
		return nil, nil, err
	}

	return append([]explain.TreeStep{step}, childPath...), outputs, nil
}
