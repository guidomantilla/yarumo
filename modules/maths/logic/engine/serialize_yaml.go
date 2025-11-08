package engine

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// LoadRulesYAML reads a RuleSetDTO from r (YAML) and parses it into engine Rules.
func LoadRulesYAML(r io.Reader) ([]Rule, error) {
	var set RuleSetDTO
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&set); err != nil {
		return nil, fmt.Errorf("decode rules YAML: %w", err)
	}
	return RulesFromDTO(set)
}

// SaveRulesYAML writes rules to w as an indented YAML RuleSetDTO (v1).
func SaveRulesYAML(w io.Writer, rules []Rule) error {
	set := RulesToDTO(rules)
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(set)
}
