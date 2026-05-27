package errs

import "reflect"

// leafTypeName resolves the type name for a leaf error. If a currentType was propagated
// from a TypedError ancestor, it is used. Otherwise, falls back to the reflect type name.
func leafTypeName(err error, currentType string) string {
	if currentType != "" {
		return currentType
	}

	t := reflect.TypeOf(err)
	if t != nil {
		return t.String()
	}

	return "<nil>"
}
