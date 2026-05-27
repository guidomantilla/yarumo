package validation

import (
	"errors"
	"net"
	"reflect"
	"regexp"
	"time"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// errBadParam is the package-internal sentinel for parameter conversion
// failures.
var errBadParam = errors.New("parameter conversion failed")

// asString coerces value into a string, otherwise returns an engine error.
func asString(value any) (string, error) {
	s, ok := value.(string)
	if !ok {
		return "", ErrEngine(cerrs.Wrap(ErrBadParams, errBadParam))
	}

	return s, nil
}

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

// asTime coerces value into time.Time. RFC 3339 strings are parsed; values
// that already are time.Time pass through unchanged.
func asTime(value any) (time.Time, error) {
	t, ok := value.(time.Time)
	if ok {
		return t, nil
	}

	s, ok := value.(string)
	if !ok {
		return time.Time{}, ErrEngine(cerrs.Wrap(ErrBadParams, errBadParam))
	}

	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, ErrEngine(cerrs.Wrap(ErrBadParams, err))
	}

	return parsed, nil
}

// asTimeParam coerces params[i] into time.Time.
func asTimeParam(params []any, i int) (time.Time, error) {
	if i >= len(params) {
		return time.Time{}, ErrEngine(ErrBadParams)
	}

	return asTime(params[i])
}

// asDuration coerces value into time.Duration. String inputs are parsed via
// time.ParseDuration (e.g. "5s", "100ms"); numbers are treated as
// nanoseconds for consistency with time.Duration's underlying int64.
func asDuration(value any) (time.Duration, error) {
	d, ok := value.(time.Duration)
	if ok {
		return d, nil
	}

	s, ok := value.(string)
	if ok {
		parsed, err := time.ParseDuration(s)
		if err != nil {
			return 0, ErrEngine(cerrs.Wrap(ErrBadParams, err))
		}

		return parsed, nil
	}

	f, err := asFloat(value)
	if err != nil {
		return 0, err
	}

	return time.Duration(f), nil
}

// asRegex coerces value into *regexp.Regexp. Strings are compiled on the
// fly; pre-compiled regexps pass through.
func asRegex(value any) (*regexp.Regexp, error) {
	re, ok := value.(*regexp.Regexp)
	if ok {
		return re, nil
	}

	s, ok := value.(string)
	if !ok {
		return nil, ErrEngine(cerrs.Wrap(ErrBadParams, errBadParam))
	}

	compiled, err := regexp.Compile(s)
	if err != nil {
		return nil, ErrEngine(cerrs.Wrap(ErrBadParams, err))
	}

	return compiled, nil
}

// asIP coerces value into net.IP. Strings are parsed via net.ParseIP;
// net.IP values pass through.
func asIP(value any) (net.IP, error) {
	ip, ok := value.(net.IP)
	if ok {
		return ip, nil
	}

	s, ok := value.(string)
	if !ok {
		return nil, ErrEngine(cerrs.Wrap(ErrBadParams, errBadParam))
	}

	parsed := net.ParseIP(s)
	if parsed == nil {
		return nil, ErrEngine(cerrs.Wrap(ErrBadParams, errBadParam))
	}

	return parsed, nil
}

// asCIDR coerces value into the (ip, network) pair returned by
// net.ParseCIDR. Only string inputs are accepted.
func asCIDR(value any) (net.IP, *net.IPNet, error) {
	s, ok := value.(string)
	if !ok {
		return nil, nil, ErrEngine(cerrs.Wrap(ErrBadParams, errBadParam))
	}

	ip, network, err := net.ParseCIDR(s)
	if err != nil {
		return nil, nil, ErrEngine(cerrs.Wrap(ErrBadParams, err))
	}

	return ip, network, nil
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
