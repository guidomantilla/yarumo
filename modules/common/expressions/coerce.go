package expressions

import (
	"fmt"
	"strings"
)

// toFloat64 attempts to convert a value to float64.
func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case float32:
		return float64(n), true
	default:
		return 0, false
	}
}

// toBool attempts to convert a value to bool.
func toBool(v any) (bool, bool) {
	b, ok := v.(bool)
	return b, ok
}

// toString attempts to convert a value to string.
func toString(v any) (string, bool) {
	s, ok := v.(string)
	return s, ok
}

// toSlice attempts to convert a value to []any.
func toSlice(v any) ([]any, bool) {
	s, ok := v.([]any)
	return s, ok
}

// formatValue formats a value for display in error messages.
func formatValue(v any) string {
	if v == nil {
		return "nil"
	}
	return fmt.Sprint(v)
}

// resolveProperty navigates nested map[string]any by dot-separated field.
func resolveProperty(obj any, field string) (any, error) {
	if obj == nil {
		return nil, ErrEval("cannot access field "+field+" on nil", ErrNilAccess)
	}

	m, ok := obj.(map[string]any)
	if !ok {
		ctx, isCtx := obj.(Context)
		if !isCtx {
			return nil, ErrEval("cannot access field "+field+" on "+formatValue(obj), ErrTypeMismatch)
		}
		m = map[string]any(ctx)
	}

	parts := strings.SplitN(field, ".", 2)
	val, exists := m[parts[0]]
	if !exists {
		return nil, ErrEval("unknown field "+parts[0], ErrUnknownField)
	}

	if len(parts) == 1 {
		return val, nil
	}

	return resolveProperty(val, parts[1])
}
