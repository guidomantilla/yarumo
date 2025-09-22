package engine

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic2/parser"
	p "github.com/guidomantilla/yarumo/pkg/common/maths/logic2/props"
)

// LoadRulesJSON reads a RuleSetDTO from r and parses it into engine Rules.
func LoadRulesJSON(r io.Reader) ([]Rule, error) {
	var set RuleSetDTO
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&set); err != nil {
		return nil, fmt.Errorf("decode rules JSON: %w", err)
	}
	return RulesFromDTO(set)
}

// SaveRulesJSON writes rules to w as a canonical, indented JSON RuleSetDTO (v1).
func SaveRulesJSON(w io.Writer, rules []Rule) error {
	set := RulesToDTO(rules)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(set)
}

// --- Convenience helpers for single Rule DTO ---

// ParseRule parses a single RuleDTO into a Rule (utility for ad-hoc uses).
func ParseRule(d RuleDTO) (Rule, error) { return RuleFromDTO(d) }

// EncodeRuleJSON serializes a single Rule as JSON RuleDTO (indentation matches SaveRulesJSON style).
func EncodeRuleJSON(w io.Writer, r Rule) error {
	d := RuleToDTO(r)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(d)
}

// DecodeRuleJSON decodes a single RuleDTO and parses it to a Rule.
func DecodeRuleJSON(rdr io.Reader) (Rule, error) {
	var d RuleDTO
	dec := json.NewDecoder(rdr)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&d); err != nil {
		return Rule{}, fmt.Errorf("decode rule JSON: %w", err)
	}
	return RuleFromDTO(d)
}

// --- Utility: Facts JSON (flat map) ---

// LoadFactsJSON reads a flat JSON object {"A":true,...} into a FactBase.
func LoadFactsJSON(r io.Reader) (FactBase, error) {
	m := map[string]bool{}
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("decode facts JSON: %w", err)
	}
	fb := FactBase{}
	for k, v := range m {
		fb[p.Var(k)] = v
	}
	return fb, nil
}

// SaveFactsJSON writes a FactBase as a flat JSON object.
func SaveFactsJSON(w io.Writer, facts FactBase) error {
	m := map[string]bool{}
	for k, v := range facts {
		m[string(k)] = v
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(m)
}

// --- Optional helpers: compile rules from strings ---

// NewRulesFromStrings is a small helper to build rules directly from strings without DTOs.
func NewRulesFromStrings(specs []struct{ ID, When, Then string }) ([]Rule, error) {
	out := make([]Rule, len(specs))
	for i, s := range specs {
		f, err := parser.Parse(s.When)
		if err != nil {
			return nil, fmt.Errorf("parse rule %d: %w", i, err)
		}
		if s.Then == "" {
			return nil, fmt.Errorf("rule %d: empty Then", i)
		}
		out[i] = Rule{id: s.ID, when: f, then: p.Var(s.Then)}
	}
	return out, nil
}
