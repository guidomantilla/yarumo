package validation

import (
	"encoding/json"

	yaml "go.yaml.in/yaml/v3"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
	cvalidation "github.com/guidomantilla/yarumo/core/common/validation"
)

// annotatePath prefixes a violation message with the field path so callers
// can identify the offending field in aggregated output. When path is empty
// the error is returned unchanged. Callers must not pass a nil err.
func annotatePath(path string, err error) error {
	if path == "" {
		return err
	}

	return cvalidation.ErrValidation(cerrs.Wrap(errPathPrefix(path), err))
}

// joinPath concatenates a parent path with a child field name.
func joinPath(parent, child string) string {
	if parent == "" {
		return child
	}

	return parent + "." + child
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
