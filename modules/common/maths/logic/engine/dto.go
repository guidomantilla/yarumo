package engine

// DTOs and helpers for JSON/YAML serialization (v1).
// Note: We do NOT serialize internal AST structs. Formulas are represented
// as strings using the logic/parser as the stable contract.

import (
	"encoding/json"
	"fmt"

	"github.com/guidomantilla/yarumo/modules/common/maths/logic/parser"
	p "github.com/guidomantilla/yarumo/modules/common/maths/logic/props"
)

// Version string for current DTOs.
const dtoVersionV1 = "v1"

// RuleDTO is the wire format for a rule. When is a parseable formula string,
// Then is the variable name.
//
// Example JSON:
//
//	{
//	  "version": "v1",
//	  "id": "r1",
//	  "when": "A & B",
//	  "then": "C"
//	}
//
// Versioning allows future breaking changes while keeping backward compatibility.
type RuleDTO struct {
	Version string `json:"version"`
	ID      string `json:"id"`
	When    string `json:"when"`
	Then    string `json:"then"`
}

// RuleSetDTO is a simple container for a list of rules.
type RuleSetDTO struct {
	Version string    `json:"version"`
	Rules   []RuleDTO `json:"rules"`
}

// ExplainDTO is an acyclic, JSON-friendly explanation tree.
// Expr is the formatted expression for the node, Value is its truth value,
// Why provides a short reason, and Kids is the list of child nodes.
type ExplainDTO struct {
	Expr  string       `json:"expr"`
	Value bool         `json:"value"`
	Why   string       `json:"why,omitempty"`
	Kids  []ExplainDTO `json:"kids,omitempty"`
}

// ToDTO converts a runtime Explain to its DTO form (acyclic; no pointers).
func ToDTO(e *Explain) ExplainDTO {
	if e == nil {
		return ExplainDTO{}
	}
	d := ExplainDTO{Expr: e.Expr, Value: e.Value, Why: e.Why}
	if len(e.Kids) > 0 {
		d.Kids = make([]ExplainDTO, len(e.Kids))
		for i, k := range e.Kids {
			d.Kids[i] = ToDTO(k)
		}
	}
	return d
}

// FromDTO reconstructs a runtime Explain tree from a DTO.
func FromDTO(d ExplainDTO) *Explain {
	e := &Explain{Expr: d.Expr, Value: d.Value, Why: d.Why}
	if len(d.Kids) > 0 {
		e.Kids = make([]*Explain, len(d.Kids))
		for i, kd := range d.Kids {
			e.Kids[i] = FromDTO(kd)
		}
	}
	return e
}

// RuleToDTO converts an Engine Rule to RuleDTO (v1).
func RuleToDTO(r Rule) RuleDTO {
	return RuleDTO{Version: dtoVersionV1, ID: r.id, When: r.when.String(), Then: string(r.then)}
}

// RuleFromDTO converts a RuleDTO to an Engine Rule by parsing the When string.
func RuleFromDTO(d RuleDTO) (Rule, error) {
	if d.Version != dtoVersionV1 && d.Version != "" {
		return Rule{}, fmt.Errorf("unsupported rule version: %s", d.Version)
	}
	f, err := parser.Parse(d.When)
	if err != nil {
		return Rule{}, err
	}
	if d.Then == "" {
		return Rule{}, fmt.Errorf("missing 'then' var name")
	}
	return Rule{id: d.ID, when: f, then: p.Var(d.Then)}, nil
}

// RulesToDTO converts a slice of rules to a RuleSetDTO (v1).
func RulesToDTO(rules []Rule) RuleSetDTO {
	out := RuleSetDTO{Version: dtoVersionV1, Rules: make([]RuleDTO, len(rules))}
	for i, r := range rules {
		out.Rules[i] = RuleToDTO(r)
	}
	return out
}

// RulesFromDTO converts a RuleSetDTO to a slice of rules.
func RulesFromDTO(set RuleSetDTO) ([]Rule, error) {
	if set.Version != dtoVersionV1 && set.Version != "" {
		return nil, fmt.Errorf("unsupported ruleset version: %s", set.Version)
	}
	out := make([]Rule, len(set.Rules))
	for i, d := range set.Rules {
		var err error
		out[i], err = RuleFromDTO(d)
		if err != nil {
			return nil, fmt.Errorf("rule %d: %w", i, err)
		}
	}
	return out, nil
}

// normalizeJSON produces a canonical, indented JSON string for DTOs so that
// round-trip comparisons are stable modulo whitespace.
func normalizeJSON(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
