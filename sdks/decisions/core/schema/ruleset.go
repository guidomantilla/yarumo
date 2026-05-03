// Package schema defines the serializable configuration types for all decision engines.
package schema

// RuleSet is the serializable definition of a decision ruleset.
type RuleSet struct {
	Name     string `json:"name" yaml:"name"`
	Version  string `json:"version" yaml:"version"`
	Paradigm string `json:"paradigm" yaml:"paradigm"`

	Deductive *DeductiveConfig `json:"deductive,omitempty" yaml:"deductive,omitempty"`
	Bayesian  *BayesianConfig  `json:"bayesian,omitempty" yaml:"bayesian,omitempty"`
	Fuzzy     *FuzzyConfig     `json:"fuzzy,omitempty" yaml:"fuzzy,omitempty"`

	Table     *TableConfig     `json:"table,omitempty" yaml:"table,omitempty"`
	Scorecard *ScorecardConfig `json:"scorecard,omitempty" yaml:"scorecard,omitempty"`
	Tree      *TreeConfig      `json:"tree,omitempty" yaml:"tree,omitempty"`
}

// DeductiveConfig defines a deductive (propositional) rule set.
type DeductiveConfig struct {
	Rules         []DeductiveRuleDef `json:"rules" yaml:"rules"`
	MaxIterations int                `json:"max_iterations,omitempty" yaml:"max_iterations,omitempty"`
	Strategy      string             `json:"strategy,omitempty" yaml:"strategy,omitempty"`
}

// DeductiveRuleDef is the serializable form of a deductive rule.
type DeductiveRuleDef struct {
	Name       string          `json:"name" yaml:"name"`
	Priority   int             `json:"priority,omitempty" yaml:"priority,omitempty"`
	Condition  string          `json:"condition" yaml:"condition"`
	Conclusion map[string]bool `json:"conclusion" yaml:"conclusion"`
}

// BayesianConfig defines a Bayesian network configuration.
type BayesianConfig struct {
	Nodes     []BayesianNodeDef `json:"nodes" yaml:"nodes"`
	Algorithm string            `json:"algorithm,omitempty" yaml:"algorithm,omitempty"`
}

// BayesianNodeDef is the serializable form of a Bayesian network node.
type BayesianNodeDef struct {
	Variable string   `json:"variable" yaml:"variable"`
	Parents  []string `json:"parents,omitempty" yaml:"parents,omitempty"`
	Outcomes []string `json:"outcomes" yaml:"outcomes"`
	CPT      []CPTRow `json:"cpt" yaml:"cpt"`
}

// CPTRow defines one row of a conditional probability table.
type CPTRow struct {
	Given         map[string]string  `json:"given,omitempty" yaml:"given,omitempty"`
	Probabilities map[string]float64 `json:"probabilities" yaml:"probabilities"`
}

// FuzzyConfig defines a fuzzy inference configuration.
type FuzzyConfig struct {
	InputVars     []FuzzyVarDef      `json:"input_vars" yaml:"input_vars"`
	OutputVars    []FuzzyVarDef      `json:"output_vars" yaml:"output_vars"`
	Rules         []FuzzyRuleDef     `json:"rules" yaml:"rules"`
	Method        string             `json:"method,omitempty" yaml:"method,omitempty"`
	SugenoOutputs map[string]float64 `json:"sugeno_outputs,omitempty" yaml:"sugeno_outputs,omitempty"`
}

// FuzzyVarDef is the serializable form of a fuzzy variable.
type FuzzyVarDef struct {
	Name  string         `json:"name" yaml:"name"`
	Min   float64        `json:"min" yaml:"min"`
	Max   float64        `json:"max" yaml:"max"`
	Terms []FuzzyTermDef `json:"terms" yaml:"terms"`
}

// FuzzyTermDef is the serializable form of a fuzzy term.
type FuzzyTermDef struct {
	Name   string    `json:"name" yaml:"name"`
	Type   string    `json:"type" yaml:"type"`
	Params []float64 `json:"params" yaml:"params"`
}

// FuzzyRuleDef is the serializable form of a fuzzy rule.
type FuzzyRuleDef struct {
	Name       string              `json:"name" yaml:"name"`
	Conditions []FuzzyConditionDef `json:"conditions" yaml:"conditions"`
	Consequent FuzzyConsequentDef  `json:"consequent" yaml:"consequent"`
	Operator   string              `json:"operator,omitempty" yaml:"operator,omitempty"`
	Weight     float64             `json:"weight,omitempty" yaml:"weight,omitempty"`
}

// FuzzyConditionDef is the serializable form of a fuzzy condition.
type FuzzyConditionDef struct {
	Variable string `json:"variable" yaml:"variable"`
	Term     string `json:"term" yaml:"term"`
}

// FuzzyConsequentDef is the serializable form of a fuzzy consequent.
type FuzzyConsequentDef struct {
	Variable string `json:"variable" yaml:"variable"`
	Term     string `json:"term" yaml:"term"`
}

// TableConfig defines a decision table configuration.
type TableConfig struct {
	Rules     []TableRuleDef `json:"rules" yaml:"rules"`
	HitPolicy string         `json:"hit_policy,omitempty" yaml:"hit_policy,omitempty"`
}

// TableRuleDef is the serializable form of a decision table rule.
type TableRuleDef struct {
	Name       string         `json:"name" yaml:"name"`
	Priority   int            `json:"priority,omitempty" yaml:"priority,omitempty"`
	Conditions []string       `json:"conditions" yaml:"conditions"`
	Outputs    map[string]any `json:"outputs" yaml:"outputs"`
}

// ScorecardConfig defines a scorecard configuration.
type ScorecardConfig struct {
	Attributes []ScorecardAttributeDef `json:"attributes" yaml:"attributes"`
	BaseScore  float64                 `json:"base_score,omitempty" yaml:"base_score,omitempty"`
}

// ScorecardAttributeDef is the serializable form of a scorecard attribute.
type ScorecardAttributeDef struct {
	Name   string            `json:"name" yaml:"name"`
	Weight float64           `json:"weight" yaml:"weight"`
	Bins   []ScorecardBinDef `json:"bins" yaml:"bins"`
}

// ScorecardBinDef is the serializable form of a scorecard bin.
type ScorecardBinDef struct {
	Condition string  `json:"condition" yaml:"condition"`
	Points    float64 `json:"points" yaml:"points"`
}

// TreeConfig defines a decision tree configuration.
type TreeConfig struct {
	Root TreeNodeDef `json:"root" yaml:"root"`
}

// TreeNodeDef is the serializable form of a decision tree node.
type TreeNodeDef struct {
	Condition string         `json:"condition,omitempty" yaml:"condition,omitempty"`
	True      *TreeNodeDef   `json:"true,omitempty" yaml:"true,omitempty"`
	False     *TreeNodeDef   `json:"false,omitempty" yaml:"false,omitempty"`
	Output    map[string]any `json:"output,omitempty" yaml:"output,omitempty"`
}
