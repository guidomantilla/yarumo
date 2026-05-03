package validate

import (
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cexpressions "github.com/guidomantilla/yarumo/common/expressions"
	"github.com/guidomantilla/yarumo/compute/math/logic"
	"github.com/guidomantilla/yarumo/compute/math/logic/entailment"

	"github.com/guidomantilla/yarumo/decisions/core/adapters"
	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

// validator performs pre-deploy validation of rulesets.
type validator struct{}

// NewValidator creates a new Validator and registers the DPLL SAT solver (once).
func NewValidator(solver logic.SATSolverFn) Validator {
	cassert.NotNil(solver, "solver is nil")

	logic.RegisterSATSolver(solver)

	return &validator{}
}

// ValidateDeductive validates a deductive ruleset.
func (v *validator) ValidateDeductive(config *schema.DeductiveConfig) Report {
	cassert.NotNil(v, "validator is nil")

	report := Report{}

	_, parsed, err := adapters.AdaptDeductiveRules(config)
	if err != nil {
		report.Errors = append(report.Errors, err.Error())

		return report
	}

	v.checkContradictions(parsed, &report)
	v.checkRedundancy(parsed, &report)
	v.checkCoverage(parsed, &report)
	v.simplifyRules(parsed, &report)

	report.Valid = len(report.Contradictions) == 0
	report.Parsed = len(parsed)

	return report
}

// ValidateBayesian validates a Bayesian network configuration.
func (v *validator) ValidateBayesian(config *schema.BayesianConfig) Report {
	cassert.NotNil(v, "validator is nil")

	report := Report{}

	_, err := adapters.AdaptBayesianNetwork(config)
	if err != nil {
		report.Errors = append(report.Errors, err.Error())
	}

	report.Parsed = len(config.Nodes)
	report.Valid = len(report.Errors) == 0

	return report
}

// ValidateFuzzy validates a fuzzy inference configuration.
func (v *validator) ValidateFuzzy(config *schema.FuzzyConfig) Report {
	cassert.NotNil(v, "validator is nil")

	report := Report{}

	allVars := make(map[string]bool)
	varTerms := make(map[string]map[string]bool)
	outputVars := make(map[string]bool)

	v.validateFuzzyVars(config.InputVars, "input", allVars, varTerms, &report)
	v.validateFuzzyVars(config.OutputVars, "output", allVars, varTerms, &report)

	for _, def := range config.OutputVars {
		if def.Name != "" {
			outputVars[def.Name] = true
		}
	}

	v.validateFuzzyRules(config.Rules, allVars, varTerms, outputVars, &report)

	if config.Method == "sugeno" {
		v.validateSugenoOutputs(config, &report)
	}

	report.Parsed = len(config.InputVars) + len(config.OutputVars) + len(config.Rules)
	report.Valid = len(report.Errors) == 0

	return report
}

// validHitPolicies lists the valid hit policies for decision tables.
var validHitPolicies = map[string]bool{ //nolint:gochecknoglobals // constant map
	"first": true, "unique": true, "collect": true, "priority": true, "": true,
}

// ValidateTable validates a decision table configuration.
func (v *validator) ValidateTable(config *schema.TableConfig) Report {
	cassert.NotNil(v, "validator is nil")

	report := Report{}

	if !validHitPolicies[config.HitPolicy] {
		report.Errors = append(report.Errors,
			fmt.Sprintf("invalid hit policy: %q", config.HitPolicy))
	}

	if len(config.Rules) == 0 {
		report.Errors = append(report.Errors, "no rules defined")
	}

	for _, rule := range config.Rules {
		if rule.Name == "" {
			report.Errors = append(report.Errors, "rule has empty name")

			continue
		}

		if len(rule.Conditions) == 0 {
			report.Errors = append(report.Errors, "rule "+rule.Name+": no conditions defined")
		}

		for _, cond := range rule.Conditions {
			_, err := cexpressions.Parse(cond)
			if err != nil {
				report.Errors = append(report.Errors,
					fmt.Sprintf("rule %s: condition %q: %v", rule.Name, cond, err))
			}
		}

		if len(rule.Outputs) == 0 {
			report.Errors = append(report.Errors, "rule "+rule.Name+": no outputs defined")
		}
	}

	report.Parsed = len(config.Rules)
	report.Valid = len(report.Errors) == 0

	return report
}

// ValidateScorecard validates a scorecard configuration.
func (v *validator) ValidateScorecard(config *schema.ScorecardConfig) Report {
	cassert.NotNil(v, "validator is nil")

	report := Report{}

	if len(config.Attributes) == 0 {
		report.Errors = append(report.Errors, "no attributes defined")
	}

	for _, attr := range config.Attributes {
		if attr.Name == "" {
			report.Errors = append(report.Errors, "attribute has empty name")

			continue
		}

		if attr.Weight <= 0 {
			report.Errors = append(report.Errors,
				fmt.Sprintf("attribute %s: weight must be positive, got %g", attr.Name, attr.Weight))
		}

		if len(attr.Bins) == 0 {
			report.Errors = append(report.Errors, "attribute "+attr.Name+": no bins defined")
		}

		for _, bin := range attr.Bins {
			_, err := cexpressions.Parse(bin.Condition)
			if err != nil {
				report.Errors = append(report.Errors,
					fmt.Sprintf("attribute %s: bin condition %q: %v", attr.Name, bin.Condition, err))
			}
		}
	}

	report.Parsed = len(config.Attributes)
	report.Valid = len(report.Errors) == 0

	return report
}

const maxTreeDepth = 100

// ValidateTree validates a decision tree configuration.
func (v *validator) ValidateTree(config *schema.TreeConfig) Report {
	cassert.NotNil(v, "validator is nil")

	report := Report{}

	v.validateTreeNode(&config.Root, &report, 0)

	report.Valid = len(report.Errors) == 0

	return report
}

// --- private methods ---

func (v *validator) validateFuzzyVars(defs []schema.FuzzyVarDef, kind string, allVars map[string]bool, varTerms map[string]map[string]bool, report *Report) {
	for _, def := range defs {
		if def.Name == "" {
			report.Errors = append(report.Errors, kind+" variable has empty name")

			continue
		}

		if allVars[def.Name] {
			report.Errors = append(report.Errors, "duplicate variable name: "+def.Name)
		}

		allVars[def.Name] = true

		if def.Min >= def.Max {
			report.Errors = append(report.Errors,
				fmt.Sprintf("variable %s: min (%g) must be less than max (%g)", def.Name, def.Min, def.Max))
		}

		if len(def.Terms) == 0 {
			report.Errors = append(report.Errors, "variable "+def.Name+": no terms defined")
		}

		terms := v.validateFuzzyTerms(def.Name, def.Terms, report)
		varTerms[def.Name] = terms
	}
}

func (v *validator) validateFuzzyTerms(varName string, defs []schema.FuzzyTermDef, report *Report) map[string]bool {
	terms := make(map[string]bool)

	for _, tdef := range defs {
		if tdef.Name == "" {
			report.Errors = append(report.Errors, "variable "+varName+": term has empty name")

			continue
		}

		if terms[tdef.Name] {
			report.Errors = append(report.Errors,
				"variable "+varName+": duplicate term name: "+tdef.Name)
		}

		terms[tdef.Name] = true

		expected, ok := adapters.MembershipParamCounts[tdef.Type]
		if !ok {
			report.Errors = append(report.Errors,
				fmt.Sprintf("variable %s term %s: unknown type %q", varName, tdef.Name, tdef.Type))

			continue
		}

		if len(tdef.Params) != expected {
			report.Errors = append(report.Errors,
				fmt.Sprintf("variable %s term %s: %s requires %d params, got %d",
					varName, tdef.Name, tdef.Type, expected, len(tdef.Params)))
		}
	}

	return terms
}

func (v *validator) validateSugenoOutputs(config *schema.FuzzyConfig, report *Report) {
	for _, def := range config.OutputVars {
		if def.Name == "" {
			continue
		}

		_, exists := config.SugenoOutputs[def.Name]
		if !exists {
			report.Errors = append(report.Errors,
				"sugeno method requires output value for variable "+def.Name)
		}
	}
}

func (v *validator) validateFuzzyRules(defs []schema.FuzzyRuleDef, allVars map[string]bool, varTerms map[string]map[string]bool, outputVars map[string]bool, report *Report) {
	for _, def := range defs {
		if def.Name == "" {
			report.Errors = append(report.Errors, "rule has empty name")

			continue
		}

		if len(def.Conditions) == 0 {
			report.Errors = append(report.Errors, "rule "+def.Name+": no conditions defined")
		}

		v.validateFuzzyConditions(def.Name, def.Conditions, allVars, varTerms, report)
		v.validateFuzzyConsequent(def.Name, def.Consequent, allVars, varTerms, outputVars, report)

		if def.Weight < 0 || def.Weight > 1 {
			report.Errors = append(report.Errors,
				fmt.Sprintf("rule %s: weight must be in [0,1], got %g", def.Name, def.Weight))
		}
	}
}

func (v *validator) validateFuzzyConditions(ruleName string, conds []schema.FuzzyConditionDef, allVars map[string]bool, varTerms map[string]map[string]bool, report *Report) {
	for _, cond := range conds {
		if !allVars[cond.Variable] {
			report.Errors = append(report.Errors,
				fmt.Sprintf("rule %s: condition references unknown variable %q", ruleName, cond.Variable))

			continue
		}

		terms := varTerms[cond.Variable]
		if terms != nil && !terms[cond.Term] {
			report.Errors = append(report.Errors,
				fmt.Sprintf("rule %s: condition references unknown term %q for variable %q", ruleName, cond.Term, cond.Variable))
		}
	}
}

func (v *validator) validateFuzzyConsequent(ruleName string, cons schema.FuzzyConsequentDef, allVars map[string]bool, varTerms map[string]map[string]bool, outputVars map[string]bool, report *Report) {
	switch {
	case !allVars[cons.Variable]:
		report.Errors = append(report.Errors,
			fmt.Sprintf("rule %s: consequent references unknown variable %q", ruleName, cons.Variable))
	case !outputVars[cons.Variable]:
		report.Errors = append(report.Errors,
			fmt.Sprintf("rule %s: consequent variable %q is not an output variable", ruleName, cons.Variable))
	default:
		terms := varTerms[cons.Variable]
		if terms != nil && !terms[cons.Term] {
			report.Errors = append(report.Errors,
				fmt.Sprintf("rule %s: consequent references unknown term %q for variable %q",
					ruleName, cons.Term, cons.Variable))
		}
	}
}

func (v *validator) checkContradictions(rules []adapters.ParsedRule, report *Report) {
	for i := range rules {
		for j := i + 1; j < len(rules); j++ {
			for varA, valA := range rules[i].Conclusion {
				valB, exists := rules[j].Conclusion[varA]
				if !exists || valA == valB {
					continue
				}

				combined := logic.AndF{L: rules[i].Formula, R: rules[j].Formula}

				if logic.IsSatisfiable(combined) {
					report.Contradictions = append(report.Contradictions, ConflictPair{
						RuleA:  rules[i].Name,
						RuleB:  rules[j].Name,
						Detail: "conflicting conclusions for " + string(varA),
					})
				}
			}
		}
	}
}

func (v *validator) checkRedundancy(rules []adapters.ParsedRule, report *Report) {
	for i, rule := range rules {
		var premises []logic.Formula
		var premiseNames []string

		for j, other := range rules {
			if j == i {
				continue
			}

			premises = append(premises, ruleImplication(other))
			premiseNames = append(premiseNames, other.Name)
		}

		ruleFormula := ruleImplication(rule)

		if len(premises) > 0 && entailment.Entails(premises, ruleFormula) {
			report.Redundant = append(report.Redundant, RedundantRule{
				Rule:      rule.Name,
				ImpliedBy: premiseNames,
			})
		}
	}
}

func (v *validator) checkCoverage(rules []adapters.ParsedRule, report *Report) {
	if len(rules) == 0 {
		return
	}

	var disjunction logic.Formula

	for i, rule := range rules {
		if i == 0 {
			disjunction = rule.Formula
		} else {
			disjunction = logic.OrF{L: disjunction, R: rule.Formula}
		}
	}

	report.Gaps = logic.FailCases(disjunction)
}

func (v *validator) simplifyRules(rules []adapters.ParsedRule, report *Report) {
	for _, rule := range rules {
		simplified := logic.Simplify(rule.Formula)
		originalStr := logic.Format(rule.Formula)
		simplifiedStr := logic.Format(simplified)

		if originalStr != simplifiedStr && logic.Equivalent(rule.Formula, simplified) {
			report.Simplified = append(report.Simplified, SimplifiedRule{
				RuleName:   rule.Name,
				Original:   originalStr,
				Simplified: simplifiedStr,
			})
		}
	}
}

func (v *validator) validateTreeNode(node *schema.TreeNodeDef, report *Report, depth int) {
	if depth > maxTreeDepth {
		report.Errors = append(report.Errors,
			fmt.Sprintf("tree exceeds maximum depth of %d", maxTreeDepth))

		return
	}

	// Leaf node.
	if node.Output != nil {
		report.Parsed++

		return
	}

	// Internal node.
	if node.Condition == "" {
		report.Errors = append(report.Errors, "internal node has no condition")

		return
	}

	_, err := cexpressions.Parse(node.Condition)
	if err != nil {
		report.Errors = append(report.Errors,
			fmt.Sprintf("condition %q: %v", node.Condition, err))
	}

	report.Parsed++

	if node.True == nil {
		report.Errors = append(report.Errors,
			fmt.Sprintf("condition %q: missing true branch", node.Condition))
	} else {
		v.validateTreeNode(node.True, report, depth+1)
	}

	if node.False == nil {
		report.Errors = append(report.Errors,
			fmt.Sprintf("condition %q: missing false branch", node.Condition))
	} else {
		v.validateTreeNode(node.False, report, depth+1)
	}
}

// ruleImplication converts a parsed rule to an implication formula: condition → conclusion.
func ruleImplication(r adapters.ParsedRule) logic.Formula {
	parts := make([]logic.Formula, 0, len(r.Conclusion))

	for v, val := range r.Conclusion {
		if val {
			parts = append(parts, v)
		} else {
			parts = append(parts, logic.NotF{F: v})
		}
	}

	conclusionFormula := conjoinFormulas(parts)

	return logic.ImplF{L: r.Formula, R: conclusionFormula}
}

// conjoinFormulas builds a conjunction (AND) of multiple formulas.
func conjoinFormulas(parts []logic.Formula) logic.Formula {
	if len(parts) == 0 {
		return logic.TrueF{}
	}

	result := parts[0]

	for i := 1; i < len(parts); i++ {
		result = logic.AndF{L: result, R: parts[i]}
	}

	return result
}
