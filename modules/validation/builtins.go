package validation

import (
	extuids "github.com/guidomantilla/yarumo/extensions/common/uids"
	cvalidation "github.com/guidomantilla/yarumo/common/validation"
)

// builtins is the default leaf catalogue. Each entry adapts a typed function
// from common/validation/ into the RuleFn shape expected by the engine.
var builtins = map[string]RuleFn{
	"required":          ruleRequired,
	"must_be_undefined": ruleMustBeUndefined,
	"min_len":           ruleMinLen,
	"max_len":           ruleMaxLen,
	"regex":             ruleRegex,
	"email":             ruleEmail,
	"url":               ruleURL,
	"min":               ruleMin,
	"max":               ruleMax,
	"in_range":          ruleInRange,
	"uuid":              ruleUUID,
	"ulid":              ruleULID,
	"non_empty":         ruleNonEmpty,
}

// ruleRequired delegates to common/validation/.IsRequired.
func ruleRequired(value any, _ []any) error {
	return cvalidation.IsRequired(value)
}

// ruleMustBeUndefined delegates to common/validation/.MustBeUndefined.
func ruleMustBeUndefined(value any, _ []any) error {
	return cvalidation.MustBeUndefined(value)
}

// ruleMinLen reads min_len from params[0] and delegates to common/validation/.MinLen.
func ruleMinLen(value any, params []any) error {
	s, err := asString(value)
	if err != nil {
		return err
	}

	n, err := asInt(params, 0)
	if err != nil {
		return err
	}

	return cvalidation.MinLen(s, n)
}

// ruleMaxLen reads max_len from params[0] and delegates to common/validation/.MaxLen.
func ruleMaxLen(value any, params []any) error {
	s, err := asString(value)
	if err != nil {
		return err
	}

	n, err := asInt(params, 0)
	if err != nil {
		return err
	}

	return cvalidation.MaxLen(s, n)
}

// ruleRegex reads the pattern from params[0] and delegates to common/validation/.MatchesRegex.
func ruleRegex(value any, params []any) error {
	s, err := asString(value)
	if err != nil {
		return err
	}

	pattern, err := asStringParam(params, 0)
	if err != nil {
		return err
	}

	return cvalidation.MatchesRegex(s, pattern)
}

// ruleEmail delegates to common/validation/.IsEmail.
func ruleEmail(value any, _ []any) error {
	s, err := asString(value)
	if err != nil {
		return err
	}

	return cvalidation.IsEmail(s)
}

// ruleURL delegates to common/validation/.IsURL.
func ruleURL(value any, _ []any) error {
	s, err := asString(value)
	if err != nil {
		return err
	}

	return cvalidation.IsURL(s)
}

// ruleMin delegates to common/validation/.Min on float64-coerced inputs.
func ruleMin(value any, params []any) error {
	v, err := asFloat(value)
	if err != nil {
		return err
	}

	lo, err := asFloatParam(params, 0)
	if err != nil {
		return err
	}

	return cvalidation.Min(v, lo)
}

// ruleMax delegates to common/validation/.Max on float64-coerced inputs.
func ruleMax(value any, params []any) error {
	v, err := asFloat(value)
	if err != nil {
		return err
	}

	hi, err := asFloatParam(params, 0)
	if err != nil {
		return err
	}

	return cvalidation.Max(v, hi)
}

// ruleInRange delegates to common/validation/.InRange on float64-coerced inputs.
func ruleInRange(value any, params []any) error {
	v, err := asFloat(value)
	if err != nil {
		return err
	}

	lo, err := asFloatParam(params, 0)
	if err != nil {
		return err
	}

	hi, err := asFloatParam(params, 1)
	if err != nil {
		return err
	}

	return cvalidation.InRange(v, lo, hi)
}

// ruleUUID delegates to cvalidation.IsUID with the UUID predicate from
// extensions/common/uids — the engine module owns the choice of algorithm
// since the validation leaves are algorithm-agnostic.
func ruleUUID(value any, _ []any) error {
	s, err := asString(value)
	if err != nil {
		return err
	}

	return cvalidation.IsUID(s, extuids.IsUUID)
}

// ruleULID delegates to cvalidation.IsUID with the ULID predicate from
// extensions/common/uids.
func ruleULID(value any, _ []any) error {
	s, err := asString(value)
	if err != nil {
		return err
	}

	return cvalidation.IsUID(s, extuids.IsULID)
}

// ruleNonEmpty rejects nil collections via reflection-free interface
// inspection. It accepts []any and any underlying slice/array/map through
// asSlice.
func ruleNonEmpty(value any, _ []any) error {
	xs, err := asSlice(value)
	if err != nil {
		return err
	}

	return cvalidation.NonEmpty(xs)
}

