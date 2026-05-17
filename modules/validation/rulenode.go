package validation

import (
	"encoding/json"
	"fmt"

	yaml "go.yaml.in/yaml/v3"
)

// Ruleset is the in-memory shape of a validation configuration. A Ruleset is
// a flat list of top-level rule nodes; each node may itself be a group with
// nested children.
type Ruleset struct {
	Rules []RuleNode `json:"rules,omitempty" yaml:"rules,omitempty"`
}

// RuleNode is a node in the rule tree.
//
// Exactly one of the following shapes is meaningful per node:
//   - Group: Field or When (or both) set and Rules is non-empty.
//   - Leaf:  Name set; Params optional.
//
// Mixing a Group-shaped node with a Name field is not supported and produces
// a configuration error at load time.
type RuleNode struct {
	// Field selects a value out of the target object via dotted path. When
	// empty the group operates on the current value (root or the field of an
	// outer group).
	Field string `json:"field,omitempty" yaml:"field,omitempty"`

	// When is an optional boolean expression. The group only runs when the
	// expression evaluates to a truthy value against the engine context.
	When string `json:"when,omitempty" yaml:"when,omitempty"`

	// Name selects a leaf validator registered with the engine.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Params carries optional positional arguments for the leaf validator.
	Params []any `json:"params,omitempty" yaml:"params,omitempty"`

	// Rules holds nested rule nodes. Only meaningful when this node is a
	// group.
	Rules []RuleNode `json:"rules,omitempty" yaml:"rules,omitempty"`
}

// reservedKeys lists the structural keys that, when present, force a node to
// be parsed as a full RuleNode rather than as a sugar-style leaf.
var reservedKeys = map[string]struct{}{
	"field":  {},
	"when":   {},
	"name":   {},
	"params": {},
	"rules":  {},
}

// ruleNodeRaw mirrors RuleNode for YAML/JSON decoding without inheriting its
// custom unmarshaller, avoiding infinite recursion.
type ruleNodeRaw struct {
	Field  string     `json:"field,omitempty"  yaml:"field,omitempty"`
	When   string     `json:"when,omitempty"   yaml:"when,omitempty"`
	Name   string     `json:"name,omitempty"   yaml:"name,omitempty"`
	Params []any      `json:"params,omitempty" yaml:"params,omitempty"`
	Rules  []RuleNode `json:"rules,omitempty"  yaml:"rules,omitempty"`
}

// UnmarshalYAML supports three shapes for a rule entry:
//   - A scalar string ("required") — a leaf with no params.
//   - A single-key mapping ({min_len: 5}) — a leaf with positional params.
//   - A full mapping with field/when/name/params/rules — a structural node.
func (n *RuleNode) UnmarshalYAML(node *yaml.Node) error {
	if node == nil {
		return ErrLoad(ErrBadRule)
	}

	if node.Kind == yaml.ScalarNode && node.Tag == "!!str" {
		n.Name = node.Value

		return nil
	}

	if node.Kind != yaml.MappingNode {
		return ErrLoad(fmt.Errorf("unsupported YAML node kind: %v", node.Kind))
	}

	keys := mapKeys(node)
	if len(keys) == 1 && !isReserved(keys[0]) {
		return n.unmarshalSugarYAML(node, keys[0])
	}

	var raw ruleNodeRaw

	err := node.Decode(&raw)
	if err != nil {
		return ErrLoad(err)
	}

	*n = RuleNode(raw)

	return nil
}

// unmarshalSugarYAML handles `{<leaf_name>: <params>}` entries.
func (n *RuleNode) unmarshalSugarYAML(node *yaml.Node, leaf string) error {
	n.Name = leaf

	// Locate the value node paired with the leaf key.
	for i := 0; i < len(node.Content)-1; i += 2 {
		if node.Content[i].Value != leaf {
			continue
		}

		params, err := decodeParams(node.Content[i+1])
		if err != nil {
			return err
		}

		n.Params = params

		return nil
	}

	return nil
}

// mapKeys returns the keys of a mapping yaml.Node in document order.
func mapKeys(node *yaml.Node) []string {
	out := make([]string, 0, len(node.Content)/2)
	for i := 0; i < len(node.Content)-1; i += 2 {
		out = append(out, node.Content[i].Value)
	}

	return out
}

// isReserved reports whether key is a structural RuleNode field name.
func isReserved(key string) bool {
	_, ok := reservedKeys[key]

	return ok
}

// decodeParams flattens a YAML value node into the []any params slice. A
// sequence is decoded element-by-element; any other shape becomes a single
// param.
func decodeParams(value *yaml.Node) ([]any, error) {
	if value.Kind == yaml.SequenceNode {
		params := make([]any, 0, len(value.Content))
		for _, item := range value.Content {
			var v any

			err := item.Decode(&v)
			if err != nil {
				return nil, ErrLoad(err)
			}

			params = append(params, v)
		}

		return params, nil
	}

	var single any

	err := value.Decode(&single)
	if err != nil {
		return nil, ErrLoad(err)
	}

	return []any{single}, nil
}

// UnmarshalJSON mirrors the YAML loader: strings become leaves, single-key
// objects become parameterized leaves, full objects map onto the struct.
func (n *RuleNode) UnmarshalJSON(data []byte) error {
	var asString string

	err := json.Unmarshal(data, &asString)
	if err == nil {
		n.Name = asString

		return nil
	}

	var asMap map[string]any

	err = json.Unmarshal(data, &asMap)
	if err != nil {
		return ErrLoad(err)
	}

	if len(asMap) == 1 {
		for key, val := range asMap {
			if !isReserved(key) {
				n.Name = key
				n.Params = paramsFromJSON(val)

				return nil
			}
		}
	}

	var raw ruleNodeRaw

	err = json.Unmarshal(data, &raw)
	if err != nil {
		return ErrLoad(err)
	}

	*n = RuleNode(raw)

	return nil
}

// paramsFromJSON flattens a JSON value into []any params: arrays expand,
// everything else becomes a single-element slice.
func paramsFromJSON(v any) []any {
	xs, ok := v.([]any)
	if ok {
		return xs
	}

	return []any{v}
}
