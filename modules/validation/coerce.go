package validation

import (
	"reflect"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// asInt coerces params[i] into an int. YAML numerics arrive as int or
// float64 depending on whether the literal has a decimal point.
func asInt(params []any, i int) (int, error) {
	if i >= len(params) {
		return 0, ErrEngine(ErrBadParams)
	}

	f, err := asFloat(params[i])
	if err != nil {
		return 0, err
	}

	return int(f), nil
}

// asFloat coerces value into a float64. Used by numeric rules so the engine
// can accept ints, int64, float32, and float64 transparently.
func asFloat(value any) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, ErrEngine(cerrs.Wrap(ErrBadParams, errBadParam))
	}
}

// asFloatParam coerces params[i] into a float64.
func asFloatParam(params []any, i int) (float64, error) {
	if i >= len(params) {
		return 0, ErrEngine(ErrBadParams)
	}

	return asFloat(params[i])
}

// asStringParam coerces params[i] into a string.
func asStringParam(params []any, i int) (string, error) {
	if i >= len(params) {
		return "", ErrEngine(ErrBadParams)
	}

	s, ok := params[i].(string)
	if !ok {
		return "", ErrEngine(cerrs.Wrap(ErrBadParams, errBadParam))
	}

	return s, nil
}

// asSlice coerces value into []any. Plain []any inputs pass through;
// strongly-typed slices and arrays are unpacked via reflection.
func asSlice(value any) ([]any, error) {
	xs, ok := value.([]any)
	if ok {
		return xs, nil
	}

	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil, ErrEngine(cerrs.Wrap(ErrBadParams, errBadParam))
	}

	kind := v.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return nil, ErrEngine(cerrs.Wrap(ErrBadParams, errBadParam))
	}

	out := make([]any, v.Len())
	for i := range v.Len() {
		out[i] = v.Index(i).Interface()
	}

	return out, nil
}
