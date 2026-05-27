package validation

import (
	"reflect"
	"strconv"
	"strings"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// pathSegment is a parsed step from a dotted path.
type pathSegment struct {
	name    string
	indices []int
}

// parsePath splits "A.B[0][1].C" into [{A}, {B,[0,1]}, {C}].
func parsePath(path string) ([]pathSegment, error) {

	rawParts := strings.Split(path, ".")

	segments := make([]pathSegment, 0, len(rawParts))
	for _, part := range rawParts {
		seg, err := parseSegment(part)
		if err != nil {
			return nil, err
		}

		segments = append(segments, seg)
	}

	return segments, nil
}

// parseSegment parses one dotted-path component, e.g. "Items[0][1]".
func parseSegment(part string) (pathSegment, error) {
	if part == "" {
		return pathSegment{}, ErrValidation(ErrPathInvalid)
	}

	bracket := strings.Index(part, "[")
	if bracket < 0 {
		return pathSegment{name: part}, nil
	}

	name := part[:bracket]
	if name == "" {
		return pathSegment{}, ErrValidation(ErrPathInvalid)
	}

	indices, err := parseIndices(part[bracket:])
	if err != nil {
		return pathSegment{}, err
	}

	return pathSegment{name: name, indices: indices}, nil
}

// parseIndices parses one or more "[N]" suffixes; returns ErrPathInvalid on
// malformed input.
func parseIndices(s string) ([]int, error) {
	var indices []int

	for s != "" {
		if s[0] != '[' {
			return nil, ErrValidation(ErrPathInvalid)
		}

		end := strings.Index(s, "]")
		if end < 0 {
			return nil, ErrValidation(ErrPathInvalid)
		}

		idxStr := s[1:end]
		if idxStr == "" {
			return nil, ErrValidation(ErrPathInvalid)
		}

		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			return nil, ErrValidation(cerrs.Wrap(ErrPathInvalid, err))
		}

		if idx < 0 {
			return nil, ErrValidation(ErrPathInvalid)
		}

		indices = append(indices, idx)
		s = s[end+1:]
	}

	return indices, nil
}

// walkSegment resolves a name (struct field / map key) then drills down
// through any trailing slice indices.
func walkSegment(current reflect.Value, seg pathSegment) (reflect.Value, error) {
	next, err := resolveName(current, seg.name)
	if err != nil {
		return reflect.Value{}, err
	}

	for _, idx := range seg.indices {
		next, err = resolveIndex(next, idx)
		if err != nil {
			return reflect.Value{}, err
		}
	}

	return next, nil
}

// resolveName looks up a struct field by name or a map key (string-keyed).
// Pointers and interfaces are auto-dereferenced.
func resolveName(v reflect.Value, name string) (reflect.Value, error) {
	v = deref(v)

	if !v.IsValid() {
		return reflect.Value{}, ErrValidation(ErrPathNotFound)
	}

	switch v.Kind() {
	case reflect.Struct:
		field := v.FieldByName(name)
		if !field.IsValid() {
			return reflect.Value{}, ErrValidation(ErrPathNotFound)
		}

		return field, nil
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return reflect.Value{}, ErrValidation(ErrPathTypeMismatch)
		}

		val := v.MapIndex(reflect.ValueOf(name))
		if !val.IsValid() {
			return reflect.Value{}, ErrValidation(ErrPathNotFound)
		}

		return val, nil
	default:
		return reflect.Value{}, ErrValidation(ErrPathTypeMismatch)
	}
}

// resolveIndex applies one [idx] step to a slice or array value.
func resolveIndex(v reflect.Value, idx int) (reflect.Value, error) {
	v = deref(v)

	if !v.IsValid() {
		return reflect.Value{}, ErrValidation(ErrPathNotFound)
	}

	kind := v.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return reflect.Value{}, ErrValidation(ErrPathTypeMismatch)
	}

	if idx >= v.Len() {
		return reflect.Value{}, ErrValidation(ErrIndexOutOfRange)
	}

	return v.Index(idx), nil
}

// deref unwraps pointer and interface values until it reaches a concrete kind
// or an invalid value.
func deref(v reflect.Value) reflect.Value {
	for v.IsValid() && (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) {
		if v.IsNil() {
			return reflect.Value{}
		}

		v = v.Elem()
	}

	return v
}
