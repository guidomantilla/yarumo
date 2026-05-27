package adapters

import (
	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	fengine "github.com/guidomantilla/yarumo/compute/engine/fuzzy/engine"
	frules "github.com/guidomantilla/yarumo/compute/engine/fuzzy/rules"
	"github.com/guidomantilla/yarumo/compute/engine/fuzzy/variable"
	fuzzym "github.com/guidomantilla/yarumo/compute/math/fuzzy"

	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

// AdaptFuzzyVars converts serialized variable definitions to inference variables.
func AdaptFuzzyVars(defs []schema.FuzzyVarDef) ([]variable.Variable, error) {
	vars := make([]variable.Variable, 0, len(defs))

	for _, def := range defs {
		terms := make([]variable.Term, 0, len(def.Terms))

		for _, tdef := range def.Terms {
			fn, err := AdaptMembershipFn(tdef)
			if err != nil {
				return nil, ErrAdaptVariables(err)
			}

			terms = append(terms, variable.Term{Name: tdef.Name, Fn: fn})
		}

		v := variable.NewVariable(def.Name, def.Min, def.Max, terms)
		vars = append(vars, v)
	}

	return vars, nil
}

// AdaptMembershipFn creates a membership function from a serialized term definition.
func AdaptMembershipFn(def schema.FuzzyTermDef) (fuzzym.MembershipFn, error) {
	switch def.Type {
	case "triangular":
		if len(def.Params) != 3 {
			return nil, ErrAdaptMembership(ErrInvalidParamCount)
		}

		return fuzzym.Triangular(def.Params[0], def.Params[1], def.Params[2])
	case "trapezoidal":
		if len(def.Params) != 4 {
			return nil, ErrAdaptMembership(ErrInvalidParamCount)
		}

		return fuzzym.Trapezoidal(def.Params[0], def.Params[1], def.Params[2], def.Params[3])
	case "gaussian":
		if len(def.Params) != 2 {
			return nil, ErrAdaptMembership(ErrInvalidParamCount)
		}

		return fuzzym.Gaussian(def.Params[0], def.Params[1])
	default:
		return nil, ErrAdaptMembership(ErrUnknownMembershipType)
	}
}

// AdaptFuzzyRules converts serialized rule definitions to inference engine rules.
func AdaptFuzzyRules(defs []schema.FuzzyRuleDef) []frules.Rule {
	engineRules := make([]frules.Rule, 0, len(defs))

	for _, def := range defs {
		conditions := make([]frules.Condition, len(def.Conditions))
		for i, c := range def.Conditions {
			conditions[i] = frules.Condition{Variable: c.Variable, Term: c.Term}
		}

		consequent := frules.Consequent{Variable: def.Consequent.Variable, Term: def.Consequent.Term}

		var opts []frules.Option

		if def.Operator == "or" {
			opts = append(opts, frules.WithOperator(frules.Or))
		}

		if def.Weight > 0 && def.Weight <= 1 {
			opts = append(opts, frules.WithWeight(def.Weight))
		}

		r := frules.NewRule(def.Name, conditions, consequent, opts...)
		engineRules = append(engineRules, r)
	}

	return engineRules
}

// AdaptFuzzyOpts converts serialized fuzzy config to engine options.
func AdaptFuzzyOpts(config *schema.FuzzyConfig) []fengine.Option {
	cassert.NotNil(config, "config is nil")

	var opts []fengine.Option

	if config.Method == "sugeno" {
		opts = append(opts, fengine.WithMethod(fengine.Sugeno))

		if len(config.SugenoOutputs) > 0 {
			opts = append(opts, fengine.WithSugenoOutputs(config.SugenoOutputs))
		}
	}

	return opts
}
