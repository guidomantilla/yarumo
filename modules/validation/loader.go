package validation

import (
	"encoding/json"
	"io"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	yaml "go.yaml.in/yaml/v3"
)

// LoadYAML parses a YAML document into a Ruleset. The document may either be
// a sequence of rule nodes (the canonical sketch form) or a mapping with a
// top-level "rules" key.
func LoadYAML(data []byte) (Ruleset, error) {
	if data == nil {
		return Ruleset{}, ErrLoad(ErrDataNil)
	}

	nodes, ok := tryParseYAMLList(data)
	if ok {
		return Ruleset{Rules: nodes}, nil
	}

	var doc Ruleset

	err := yaml.Unmarshal(data, &doc)
	if err != nil {
		return Ruleset{}, ErrLoad(cerrs.Wrap(ErrLoadFailed, err))
	}

	return doc, nil
}

// LoadJSON parses a JSON document into a Ruleset. The document may either be
// an array of rule nodes or an object with a top-level "rules" key.
func LoadJSON(data []byte) (Ruleset, error) {
	if data == nil {
		return Ruleset{}, ErrLoad(ErrDataNil)
	}

	nodes, ok := tryParseJSONList(data)
	if ok {
		return Ruleset{Rules: nodes}, nil
	}

	var doc Ruleset

	err := json.Unmarshal(data, &doc)
	if err != nil {
		return Ruleset{}, ErrLoad(cerrs.Wrap(ErrLoadFailed, err))
	}

	return doc, nil
}

// LoadYAMLReader reads from r and parses YAML.
func LoadYAMLReader(r io.Reader) (Ruleset, error) {
	if r == nil {
		return Ruleset{}, ErrLoad(ErrReaderNil)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return Ruleset{}, ErrLoad(cerrs.Wrap(ErrLoadFailed, err))
	}

	return LoadYAML(data)
}

// LoadJSONReader reads from r and parses JSON.
func LoadJSONReader(r io.Reader) (Ruleset, error) {
	if r == nil {
		return Ruleset{}, ErrLoad(ErrReaderNil)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return Ruleset{}, ErrLoad(cerrs.Wrap(ErrLoadFailed, err))
	}

	return LoadJSON(data)
}

// tryParseYAMLList attempts to decode data as a sequence of rule nodes.
// Returns the parsed nodes and true on success.
func tryParseYAMLList(data []byte) ([]RuleNode, bool) {
	var nodes []RuleNode

	err := yaml.Unmarshal(data, &nodes)
	if err != nil {
		return nil, false
	}

	if len(nodes) == 0 {
		return nil, false
	}

	return nodes, true
}

// tryParseJSONList attempts to decode data as a JSON array of rule nodes.
// Returns the parsed nodes and true on success.
func tryParseJSONList(data []byte) ([]RuleNode, bool) {
	var nodes []RuleNode

	err := json.Unmarshal(data, &nodes)
	if err != nil {
		return nil, false
	}

	if len(nodes) == 0 {
		return nil, false
	}

	return nodes, true
}
