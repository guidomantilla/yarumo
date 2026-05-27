package validation

import (
	"encoding/json"
	"errors"
	"io"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
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

// LoadFromReader parses a ruleset from r using the given Load function.
func LoadFromReader(r io.Reader, load LoadFn) (Ruleset, error) {
	if r == nil {
		return Ruleset{}, ErrLoad(ErrReaderNil)
	}

	if load == nil {
		return Ruleset{}, ErrLoad(ErrLoadFailed)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return Ruleset{}, ErrLoad(err)
	}

	return load(data)
}

// PathOf extracts the field path from a violation produced by the engine,
// or returns an empty string when the violation does not carry a path. It is
// a small convenience for consumers building UI feedback maps.
func PathOf(err error) string {
	for _, leaf := range cerrs.Unwrap(err) {
		var pe *pathError

		ok := errors.As(leaf, &pe)
		if !ok {
			continue
		}

		return pe.path
	}

	return ""
}
