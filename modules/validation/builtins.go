package validation

import (
	cvalidation "github.com/guidomantilla/yarumo/common/validation"
	extuids "github.com/guidomantilla/yarumo/extensions/common/uids"
)

// builtins is the default leaf catalogue. Each entry adapts a typed function
// from common/validation/ into the RuleFn shape expected by the engine.
// Entries lean on the Adapt* helpers from adapters.go so most rows are a
// single line.
var builtins = map[string]RuleFn{
	// presence / required
	"required":          ruleRequired,
	"must_be_undefined": ruleMustBeUndefined,

	// string length / pattern
	"min_len": AdaptStringWithInt(cvalidation.MinLen),
	"max_len": AdaptStringWithInt(cvalidation.MaxLen),
	"regex":   AdaptStringWithParam(cvalidation.MatchesRegex),

	// string content
	"contains":   AdaptStringWithParam(cvalidation.Contains),
	"has_prefix": AdaptStringWithParam(cvalidation.HasPrefix),
	"has_suffix": AdaptStringWithParam(cvalidation.HasSuffix),

	// string format
	"email":          AdaptString(cvalidation.IsEmail),
	"url":            AdaptString(cvalidation.IsURL),
	"lowercase":      AdaptString(cvalidation.IsLowercase),
	"uppercase":      AdaptString(cvalidation.IsUppercase),
	"alpha":          AdaptString(cvalidation.IsAlpha),
	"alphanumeric":   AdaptString(cvalidation.IsAlphanumeric),
	"numeric_string": AdaptString(cvalidation.IsNumeric),
	"ascii":          AdaptString(cvalidation.IsASCII),
	"hex":            AdaptString(cvalidation.IsHex),
	"base64":         AdaptString(cvalidation.IsBase64),
	"trimmed":        AdaptString(cvalidation.IsTrimmed),
	"jwt":            AdaptString(cvalidation.IsJWT),
	"semver":         AdaptString(cvalidation.IsSemver),
	"integer_string": AdaptString(cvalidation.IsIntegerString),
	"float_string":   AdaptString(cvalidation.IsFloatString),

	// unique identifier formats (algorithm picked by the engine)
	"uuid": ruleUUID,
	"ulid": ruleULID,

	// network / transport
	"ip":       AdaptString(cvalidation.IsIP),
	"ipv4":     AdaptString(cvalidation.IsIPv4),
	"ipv6":     AdaptString(cvalidation.IsIPv6),
	"cidr":     AdaptString(cvalidation.IsCIDR),
	"mac":      AdaptString(cvalidation.IsMAC),
	"hostname": AdaptString(cvalidation.IsHostname),
	"fqdn":     AdaptString(cvalidation.IsFQDN),
	"port":     AdaptNumeric(cvalidation.IsPort[float64]),

	// date / time
	"rfc3339":      AdaptString(cvalidation.IsRFC3339),
	"date_layout":  AdaptStringWithParam(cvalidation.IsDate),
	"before":       ruleBefore,
	"after":        ruleAfter,
	"between_time": ruleBetweenTime,

	// numeric
	"min":         AdaptNumericBinary(cvalidation.Min[float64]),
	"max":         AdaptNumericBinary(cvalidation.Max[float64]),
	"in_range":    AdaptNumericRange(cvalidation.InRange[float64]),
	"positive":    AdaptNumeric(cvalidation.Positive[float64]),
	"negative":    AdaptNumeric(cvalidation.Negative[float64]),
	"nonzero":     AdaptNumeric(cvalidation.NonZero[float64]),
	"multiple_of": ruleMultipleOf,

	// equality / set
	"equal":             AdaptStringWithParam(cvalidation.Equal[string]),
	"not_equal":         AdaptStringWithParam(cvalidation.NotEqual[string]),
	"equal_ignore_case": AdaptStringWithParam(cvalidation.EqualIgnoreCase),
	"one_of":            AdaptStringSet(cvalidation.OneOf[string]),
	"not_in":            AdaptStringSet(cvalidation.NotIn[string]),

	// collection
	"non_empty":      ruleNonEmpty,
	"min_count":      AdaptCollectionWithInt(cvalidation.MinCount[any]),
	"max_count":      AdaptCollectionWithInt(cvalidation.MaxCount[any]),
	"count_in_range": ruleCountInRange,
}

// ruleRequired delegates to common/validation/.IsRequired.
func ruleRequired(value any, _ []any) error {
	return cvalidation.IsRequired(value)
}

// ruleMustBeUndefined delegates to common/validation/.MustBeUndefined.
func ruleMustBeUndefined(value any, _ []any) error {
	return cvalidation.MustBeUndefined(value)
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

// ruleBefore reads ref from params[0] as an RFC 3339 string and delegates
// to common/validation/.Before.
func ruleBefore(value any, params []any) error {
	t, err := asTime(value)
	if err != nil {
		return err
	}

	ref, err := asTimeParam(params, 0)
	if err != nil {
		return err
	}

	return cvalidation.Before(t, ref)
}

// ruleAfter reads ref from params[0] and delegates to
// common/validation/.After.
func ruleAfter(value any, params []any) error {
	t, err := asTime(value)
	if err != nil {
		return err
	}

	ref, err := asTimeParam(params, 0)
	if err != nil {
		return err
	}

	return cvalidation.After(t, ref)
}

// ruleBetweenTime reads lo from params[0] and hi from params[1].
func ruleBetweenTime(value any, params []any) error {
	t, err := asTime(value)
	if err != nil {
		return err
	}

	lo, err := asTimeParam(params, 0)
	if err != nil {
		return err
	}

	hi, err := asTimeParam(params, 1)
	if err != nil {
		return err
	}

	return cvalidation.BetweenTime(t, lo, hi)
}

// ruleCountInRange reads lo from params[0] and hi from params[1].
func ruleCountInRange(value any, params []any) error {
	xs, err := asSlice(value)
	if err != nil {
		return err
	}

	lo, err := asInt(params, 0)
	if err != nil {
		return err
	}

	hi, err := asInt(params, 1)
	if err != nil {
		return err
	}

	return cvalidation.CountInRange(xs, lo, hi)
}

// ruleMultipleOf coerces the runtime value and factor to int64 (since
// common/validation/.MultipleOf is integer-only) and delegates.
func ruleMultipleOf(value any, params []any) error {
	v, err := asFloat(value)
	if err != nil {
		return err
	}

	factor, err := asFloatParam(params, 0)
	if err != nil {
		return err
	}

	return cvalidation.MultipleOf(int64(v), int64(factor))
}

