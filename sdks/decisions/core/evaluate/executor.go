package evaluate

import (
	"context"
	"maps"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	cexpressions "github.com/guidomantilla/yarumo/common/expressions"
	bengine "github.com/guidomantilla/yarumo/compute/engine/bayesian/engine"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
	cengine "github.com/guidomantilla/yarumo/compute/engine/deductive/engine"
	cexplain "github.com/guidomantilla/yarumo/compute/engine/deductive/explain"
	fengine "github.com/guidomantilla/yarumo/compute/engine/fuzzy/engine"
	"github.com/guidomantilla/yarumo/compute/math/logic"
	"github.com/guidomantilla/yarumo/compute/math/stats"

	"github.com/guidomantilla/yarumo/decisions/core/adapters"
	"github.com/guidomantilla/yarumo/decisions/core/explain"
	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

// dispatchParadigm routes execution to the appropriate inference engine based on paradigm.
// The input type must match the paradigm: logic.Fact for Deductive, evidence.EvidenceBase
// for Bayesian, map[string]float64 for Fuzzy. Errors are returned raw (without wrapping).
func dispatchParadigm(ctx context.Context, paradigm Paradigm, ruleSetAny any, //nolint:cyclop // dispatch function with type assertions per paradigm
	input any, query string, explainers explainerSet) (Result, error) {

	ruleSet, ok := ruleSetAny.(*schema.RuleSet)
	if !ok {
		return Result{}, cerrs.Wrap(ErrTypeMismatch)
	}

	switch paradigm {
	case Deductive:
		facts, fOk := input.(logic.Fact)
		if !fOk {
			return Result{}, cerrs.Wrap(ErrTypeMismatch)
		}

		if ruleSet.Deductive == nil {
			return Result{}, cerrs.Wrap(ErrMissingConfig)
		}

		return runDeductive(ctx, ruleSet.Deductive, facts, explainers.deductive)

	case Bayesian:
		ev, eOk := input.(evidence.EvidenceBase)
		if !eOk {
			return Result{}, cerrs.Wrap(ErrTypeMismatch)
		}

		if ruleSet.Bayesian == nil {
			return Result{}, cerrs.Wrap(ErrMissingConfig)
		}

		return runBayesian(ctx, ruleSet.Bayesian, ev, query, explainers.bayesian)

	case Fuzzy:
		inputs, iOk := input.(map[string]float64)
		if !iOk {
			return Result{}, cerrs.Wrap(ErrTypeMismatch)
		}

		if ruleSet.Fuzzy == nil {
			return Result{}, cerrs.Wrap(ErrMissingConfig)
		}

		return runFuzzy(ctx, ruleSet.Fuzzy, inputs, explainers.fuzzy)

	case Table, Scorecard, Tree:
		return Result{}, cerrs.Wrap(ErrUnsupported)
	default:
		return Result{}, cerrs.Wrap(ErrUnsupported)
	}
}

// dispatchModelParadigm routes execution to the appropriate model engine based on paradigm.
func dispatchModelParadigm(ctx context.Context, paradigm Paradigm, ruleSetAny any,
	exprCtx cexpressions.Context, opts *Options) (Result, error) { //nolint:unparam // Result is used by service.go caller

	ruleSet, ok := ruleSetAny.(*schema.RuleSet)
	if !ok {
		return Result{}, cerrs.Wrap(ErrTypeMismatch)
	}

	switch paradigm {
	case Table:
		if ruleSet.Table == nil {
			return Result{}, cerrs.Wrap(ErrMissingConfig)
		}

		return runTable(ctx, ruleSet.Table, exprCtx, opts)

	case Scorecard:
		if ruleSet.Scorecard == nil {
			return Result{}, cerrs.Wrap(ErrMissingConfig)
		}

		return runScorecard(ctx, ruleSet.Scorecard, exprCtx, opts)

	case Tree:
		if ruleSet.Tree == nil {
			return Result{}, cerrs.Wrap(ErrMissingConfig)
		}

		return runTree(ctx, ruleSet.Tree, exprCtx, opts)

	case Deductive, Bayesian, Fuzzy:
		return Result{}, cerrs.Wrap(ErrUnsupported)
	default:
		return Result{}, cerrs.Wrap(ErrUnsupported)
	}
}

// explainerSet groups segregated explainers for paradigm dispatch.
type explainerSet struct {
	deductive explain.DeductiveExplainer
	bayesian  explain.BayesianExplainer
	fuzzy     explain.FuzzyExplainer
	table     explain.TableExplainer
	scorecard explain.ScorecardExplainer
	tree      explain.TreeExplainer
}

// runDeductive builds and executes a deductive inference, returning a unified Result.
func runDeductive(ctx context.Context, config *schema.DeductiveConfig, facts logic.Fact, explainer explain.DeductiveExplainer) (Result, error) {
	engineRules, _, err := adapters.AdaptDeductiveRules(config)
	if err != nil {
		return Result{}, cerrs.Wrap(ErrExecuteFailed, err)
	}

	opts := adapters.AdaptDeductiveOpts(config)
	eng := cengine.NewEngine(opts...)
	result := eng.Forward(facts, engineRules)

	trace := toDeductiveTrace(result)

	explanation, err := explainer.ExplainDeductive(ctx, trace)
	if err != nil {
		return Result{}, ErrExplain(err)
	}

	return Result{
		Outcome: Outcome{
			Facts: result.Facts.Snapshot(),
		},
		Explanation: explanation,
		Paradigm:    Deductive,
	}, nil
}

// runBayesian builds and executes a Bayesian inference, returning a unified Result.
func runBayesian(ctx context.Context, config *schema.BayesianConfig, ev evidence.EvidenceBase, query string, explainer explain.BayesianExplainer) (Result, error) {
	net, err := adapters.AdaptBayesianNetwork(config)
	if err != nil {
		return Result{}, cerrs.Wrap(ErrExecuteFailed, err)
	}

	opts := adapters.AdaptBayesianOpts(config)
	eng := bengine.NewEngine(opts...)
	result := eng.Query(net, ev, stats.Var(query))

	trace := toBayesianTrace(result, query)

	explanation, err := explainer.ExplainBayesian(ctx, trace)
	if err != nil {
		return Result{}, ErrExplain(err)
	}

	dist := make(stats.Distribution, len(result.Posterior))
	maps.Copy(dist, result.Posterior)

	return Result{
		Outcome: Outcome{
			Distribution: dist,
		},
		Explanation: explanation,
		Paradigm:    Bayesian,
	}, nil
}

// runFuzzy builds and executes a fuzzy inference, returning a unified Result.
func runFuzzy(ctx context.Context, config *schema.FuzzyConfig, inputs map[string]float64, explainer explain.FuzzyExplainer) (Result, error) {
	inputVars, err := adapters.AdaptFuzzyVars(config.InputVars)
	if err != nil {
		return Result{}, cerrs.Wrap(ErrExecuteFailed, err)
	}

	outputVars, err := adapters.AdaptFuzzyVars(config.OutputVars)
	if err != nil {
		return Result{}, cerrs.Wrap(ErrExecuteFailed, err)
	}

	fuzzyRules := adapters.AdaptFuzzyRules(config.Rules)

	opts := adapters.AdaptFuzzyOpts(config)
	eng := fengine.NewEngine(inputVars, outputVars, fuzzyRules, opts...)
	result := eng.Infer(inputs)

	trace := toFuzzyTrace(result)

	explanation, err := explainer.ExplainFuzzy(ctx, trace)
	if err != nil {
		return Result{}, ErrExplain(err)
	}

	return Result{
		Outcome: Outcome{
			Outputs: result.Outputs,
		},
		Explanation: explanation,
		Paradigm:    Fuzzy,
	}, nil
}

func toDeductiveTrace(result cengine.Result) explain.DeductiveTrace {
	provenance := result.Facts.AllProvenance()
	derived := make([]explain.DeductiveReason, 0, len(provenance))

	for _, p := range provenance {
		if p.Origin != cexplain.Derived {
			continue
		}

		derived = append(derived, explain.DeductiveReason{
			Variable: string(p.Variable),
			Value:    p.Value,
			RuleName: p.RuleName,
			Step:     p.Step,
		})
	}

	return explain.DeductiveTrace{
		Steps:   result.Steps,
		Reasons: derived,
	}
}

func toBayesianTrace(result bengine.Result, query string) explain.BayesianTrace {
	factors := make([]explain.BayesianFactor, 0, len(result.Posterior))

	for outcome, prob := range result.Posterior {
		factors = append(factors, explain.BayesianFactor{
			Outcome:     string(outcome),
			Probability: float64(prob),
		})
	}

	return explain.BayesianTrace{
		Query:   query,
		Factors: factors,
	}
}

func toFuzzyTrace(result fengine.Result) explain.FuzzyTrace {
	outputs := make([]explain.FuzzyOutput, 0, len(result.Outputs))

	for variable, value := range result.Outputs {
		outputs = append(outputs, explain.FuzzyOutput{
			Variable: variable,
			Value:    value,
		})
	}

	memberships := make([]explain.FuzzyMembership, 0)

	for _, step := range result.Trace.Steps {
		for _, m := range step.Memberships {
			memberships = append(memberships, explain.FuzzyMembership{
				Variable: m.Variable,
				Term:     m.Term,
				Degree:   float64(m.Degree),
			})
		}
	}

	return explain.FuzzyTrace{
		Outputs:     outputs,
		Memberships: memberships,
	}
}
