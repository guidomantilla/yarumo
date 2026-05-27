package evaluate

import (
	"context"
	"maps"
	"sort"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
	cexpressions "github.com/guidomantilla/yarumo/core/common/expressions"

	"github.com/guidomantilla/yarumo/decisions/core/explain"
	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

func runTable(ctx context.Context, config *schema.TableConfig, exprCtx cexpressions.Context, opts *Options) (Result, error) {
	evaluator := cexpressions.NewEvaluator(opts.expressionOpts...)
	policy := resolveHitPolicy(config.HitPolicy)

	var matched []matchedRule

	for _, rule := range config.Rules {
		allTrue, err := evalAllConditions(evaluator, rule.Conditions, exprCtx)
		if err != nil {
			return Result{}, cerrs.Wrap(ErrConditionEval, err)
		}

		if allTrue {
			matched = append(matched, matchedRule{name: rule.Name, priority: rule.Priority, outputs: rule.Outputs})
		}
	}

	outputs, matchNames, err := applyHitPolicy(policy, matched)
	if err != nil {
		return Result{}, err
	}

	traceEntries := make([]explain.TableMatchEntry, len(matched))
	for i, m := range matched {
		traceEntries[i] = explain.TableMatchEntry{RuleName: m.name, Outputs: m.outputs}
	}

	trace := explain.TableTrace{
		HitPolicy:    policy,
		MatchedRules: traceEntries,
		Outputs:      outputs,
	}

	explanation, err := opts.tableExplainer.ExplainTable(ctx, trace)
	if err != nil {
		return Result{}, ErrExplain(err)
	}

	return Result{
		Outcome: Outcome{
			Table: &TableOutcome{
				MatchedRules: matchNames,
				Outputs:      outputs,
			},
		},
		Explanation: explanation,
		Paradigm:    Table,
	}, nil
}

type matchedRule struct {
	name     string
	priority int
	outputs  map[string]any
}

func resolveHitPolicy(policy string) string {
	switch policy {
	case HitPolicyFirst, HitPolicyUnique, HitPolicyCollect, HitPolicyPriority:
		return policy
	case "":
		return HitPolicyFirst
	default:
		return policy
	}
}

func evalAllConditions(evaluator cexpressions.Evaluator, conditions []string, exprCtx cexpressions.Context) (bool, error) {
	for _, cond := range conditions {
		val, err := evaluator.Evaluate(cond, exprCtx)
		if err != nil {
			return false, err
		}

		boolVal, ok := val.(bool)
		if !ok {
			return false, cerrs.Wrap(ErrConditionEval)
		}

		if !boolVal {
			return false, nil
		}
	}

	return true, nil
}

func applyHitPolicy(policy string, matched []matchedRule) (map[string]any, []string, error) {
	switch policy {
	case HitPolicyFirst:
		return applyFirstPolicy(matched)
	case HitPolicyUnique:
		return applyUniquePolicy(matched)
	case HitPolicyCollect:
		return applyCollectPolicy(matched)
	case HitPolicyPriority:
		return applyPriorityPolicy(matched)
	default:
		return nil, nil, cerrs.Wrap(ErrInvalidHitPolicy)
	}
}

func applyFirstPolicy(matched []matchedRule) (map[string]any, []string, error) {
	if len(matched) == 0 {
		return nil, nil, cerrs.Wrap(ErrNoMatch)
	}

	return matched[0].outputs, []string{matched[0].name}, nil
}

func applyUniquePolicy(matched []matchedRule) (map[string]any, []string, error) {
	if len(matched) == 0 {
		return nil, nil, cerrs.Wrap(ErrNoMatch)
	}

	if len(matched) > 1 {
		return nil, nil, cerrs.Wrap(ErrMultipleMatches)
	}

	return matched[0].outputs, []string{matched[0].name}, nil
}

func applyCollectPolicy(matched []matchedRule) (map[string]any, []string, error) {
	merged := make(map[string]any)
	names := make([]string, len(matched))

	for i, m := range matched {
		names[i] = m.name
		maps.Copy(merged, m.outputs)
	}

	return merged, names, nil
}

func applyPriorityPolicy(matched []matchedRule) (map[string]any, []string, error) {
	if len(matched) == 0 {
		return nil, nil, cerrs.Wrap(ErrNoMatch)
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].priority > matched[j].priority
	})

	return matched[0].outputs, []string{matched[0].name}, nil
}
