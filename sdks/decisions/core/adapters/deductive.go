package adapters

import (
	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cengine "github.com/guidomantilla/yarumo/compute/engine/deductive/engine"
	crules "github.com/guidomantilla/yarumo/compute/engine/deductive/rules"
	"github.com/guidomantilla/yarumo/compute/math/logic"
	"github.com/guidomantilla/yarumo/compute/math/logic/parser"

	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

// AdaptDeductiveRules converts serialized rule definitions to inference engine rules.
func AdaptDeductiveRules(config *schema.DeductiveConfig) ([]crules.Rule, []ParsedRule, error) {
	cassert.NotNil(config, "config is nil")

	engineRules := make([]crules.Rule, 0, len(config.Rules))
	parsed := make([]ParsedRule, 0, len(config.Rules))

	for _, def := range config.Rules {
		formula, err := parser.Parse(def.Condition)
		if err != nil {
			return nil, nil, ErrAdaptRules(err)
		}

		conclusion := make(map[logic.Var]bool, len(def.Conclusion))
		for k, v := range def.Conclusion {
			conclusion[logic.Var(k)] = v
		}

		var opts []crules.Option
		if def.Priority != 0 {
			opts = append(opts, crules.WithPriority(def.Priority))
		}

		r := crules.NewRule(def.Name, formula, conclusion, opts...)
		engineRules = append(engineRules, r)

		parsed = append(parsed, ParsedRule{
			Name:       def.Name,
			Formula:    formula,
			Conclusion: conclusion,
			Priority:   def.Priority,
		})
	}

	return engineRules, parsed, nil
}

// AdaptDeductiveOpts converts serialized strategy config to engine options.
func AdaptDeductiveOpts(config *schema.DeductiveConfig) []cengine.Option {
	cassert.NotNil(config, "config is nil")

	var opts []cengine.Option

	if config.MaxIterations > 0 {
		opts = append(opts, cengine.WithMaxIterations(config.MaxIterations))
	}

	if config.Strategy == "first_match" {
		opts = append(opts, cengine.WithStrategy(cengine.FirstMatch))
	}

	return opts
}
