package evaluate

import (
	"context"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
	cexpressions "github.com/guidomantilla/yarumo/core/common/expressions"

	"github.com/guidomantilla/yarumo/decisions/core/explain"
	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

func runScorecard(ctx context.Context, config *schema.ScorecardConfig, exprCtx cexpressions.Context, opts *Options) (Result, error) {
	evaluator := cexpressions.NewEvaluator(opts.expressionOpts...)
	total := config.BaseScore
	breakdown := make([]explain.ScoreEntry, 0, len(config.Attributes))
	breakdownMap := make(map[string]float64, len(config.Attributes))

	for _, attr := range config.Attributes {
		points, matched, err := evalBins(evaluator, attr.Bins, exprCtx)
		if err != nil {
			return Result{}, cerrs.Wrap(ErrConditionEval, err)
		}

		if !matched {
			continue
		}

		weighted := points * attr.Weight
		total += weighted

		breakdown = append(breakdown, explain.ScoreEntry{
			Attribute: attr.Name,
			Points:    points,
			Weight:    attr.Weight,
			Weighted:  weighted,
		})

		breakdownMap[attr.Name] = weighted
	}

	trace := explain.ScoreTrace{
		BaseScore:  config.BaseScore,
		TotalScore: total,
		Breakdown:  breakdown,
	}

	explanation, err := opts.scorecardExplainer.ExplainScorecard(ctx, trace)
	if err != nil {
		return Result{}, ErrExplain(err)
	}

	return Result{
		Outcome: Outcome{
			Score: &ScoreOutcome{
				TotalScore: total,
				Breakdown:  breakdownMap,
			},
		},
		Explanation: explanation,
		Paradigm:    Scorecard,
	}, nil
}

func evalBins(evaluator cexpressions.Evaluator, bins []schema.ScorecardBinDef, exprCtx cexpressions.Context) (float64, bool, error) {
	for _, bin := range bins {
		val, err := evaluator.Evaluate(bin.Condition, exprCtx)
		if err != nil {
			return 0, false, err
		}

		boolVal, ok := val.(bool)
		if !ok {
			return 0, false, cerrs.Wrap(ErrConditionEval)
		}

		if boolVal {
			return bin.Points, true, nil
		}
	}

	return 0, false, nil
}
