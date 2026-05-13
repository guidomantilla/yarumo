package validation

import (
	"io"
)

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
