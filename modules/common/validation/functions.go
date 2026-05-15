package validation

import (
	"net/mail"
	"net/url"
	"reflect"
	"regexp"

	"github.com/google/uuid"
	ulid "github.com/oklog/ulid/v2"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	cutils "github.com/guidomantilla/yarumo/common/utils"
)

// IsRequired returns ErrFieldRequired wrapped in a domain Error when value is
// the zero value of T (nil pointer, empty string, empty slice/map, …). It is
// the canonical presence check used by both imperative callers and the
// config-driven engine.
func IsRequired[T any](value T) error {
	if cutils.Empty(value) {
		return ErrValidation(ErrFieldRequired)
	}

	return nil
}

// MustBeUndefined returns ErrFieldMustBeUndefined wrapped in a domain Error
// when value is anything other than the zero value of T. It is the
// counterpart to IsRequired and is used by conditional rulesets that forbid
// a field in certain contexts (e.g. POST must not carry an ID).
func MustBeUndefined[T any](value T) error {
	if cutils.NotEmpty(value) {
		return ErrValidation(ErrFieldMustBeUndefined)
	}

	return nil
}

// MinLen returns ErrMinLen when len(s) is below n. Negative thresholds are
// treated as zero — the check trivially passes.
func MinLen(s string, n int) error {
	if n < 0 {
		n = 0
	}

	if len(s) < n {
		return ErrValidation(ErrMinLen)
	}

	return nil
}

// MaxLen returns ErrMaxLen when len(s) is above n. Negative thresholds are
// treated as zero — only the empty string passes.
func MaxLen(s string, n int) error {
	if n < 0 {
		n = 0
	}

	if len(s) > n {
		return ErrValidation(ErrMaxLen)
	}

	return nil
}

// MatchesRegex returns ErrRegexMismatch when s does not match the given
// pattern, and ErrRegexInvalid when the pattern itself cannot be compiled.
func MatchesRegex(s string, pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ErrValidation(cerrs.Wrap(ErrRegexInvalid, err))
	}

	if !re.MatchString(s) {
		return ErrValidation(ErrRegexMismatch)
	}

	return nil
}

// IsEmail returns ErrEmailInvalid when s is not a syntactically valid email
// address per RFC 5322. The check delegates to net/mail.ParseAddress.
func IsEmail(s string) error {
	if s == "" {
		return ErrValidation(ErrEmailInvalid)
	}

	addr, err := mail.ParseAddress(s)
	if err != nil {
		return ErrValidation(cerrs.Wrap(ErrEmailInvalid, err))
	}

	// mail.ParseAddress accepts inputs like "Name <a@b.com>"; the leaf must
	// reject anything that is not the bare address form.
	if addr.Address != s {
		return ErrValidation(ErrEmailInvalid)
	}

	return nil
}

// IsURL returns ErrURLInvalid when s is not a parseable absolute URL.
func IsURL(s string) error {
	if s == "" {
		return ErrValidation(ErrURLInvalid)
	}

	u, err := url.Parse(s)
	if err != nil {
		return ErrValidation(cerrs.Wrap(ErrURLInvalid, err))
	}

	if u.Scheme == "" || u.Host == "" {
		return ErrValidation(ErrURLInvalid)
	}

	return nil
}

// Min returns ErrMinValue when v is strictly below lo.
func Min[T Numeric](v, lo T) error {
	if v < lo {
		return ErrValidation(ErrMinValue)
	}

	return nil
}

// Max returns ErrMaxValue when v is strictly above hi.
func Max[T Numeric](v, hi T) error {
	if v > hi {
		return ErrValidation(ErrMaxValue)
	}

	return nil
}

// InRange returns ErrOutOfRange when v is outside [lo, hi], and
// ErrInvalidRange when the caller provided lo > hi.
func InRange[T Numeric](v, lo, hi T) error {
	if lo > hi {
		return ErrValidation(ErrInvalidRange)
	}

	if v < lo || v > hi {
		return ErrValidation(ErrOutOfRange)
	}

	return nil
}

// IsUUID returns ErrUUIDInvalid when s is not a syntactically valid UUID per
// RFC 4122 / RFC 9562. Empty strings are rejected.
func IsUUID(s string) error {
	if s == "" {
		return ErrValidation(ErrUUIDInvalid)
	}

	_, err := uuid.Parse(s)
	if err != nil {
		return ErrValidation(cerrs.Wrap(ErrUUIDInvalid, err))
	}

	return nil
}

// IsULID returns ErrULIDInvalid when s is not a syntactically valid ULID.
// Empty strings are rejected.
func IsULID(s string) error {
	if s == "" {
		return ErrValidation(ErrULIDInvalid)
	}

	_, err := ulid.Parse(s)
	if err != nil {
		return ErrValidation(cerrs.Wrap(ErrULIDInvalid, err))
	}

	return nil
}

// NonEmpty returns ErrCollectionEmpty when xs has no elements.
func NonEmpty[T any](xs []T) error {
	if len(xs) == 0 {
		return ErrValidation(ErrCollectionEmpty)
	}

	return nil
}

// Each applies check to every element of xs and aggregates all violations
// into a single domain error. A nil check is treated as the no-op validator;
// an empty xs trivially passes.
func Each[T any](xs []T, check func(T) error) error {
	if check == nil {
		return nil
	}

	var causes []error
	for _, x := range xs {
		err := check(x)
		if err != nil {
			causes = append(causes, err)
		}
	}

	if len(causes) == 0 {
		return nil
	}

	causes = append(causes, ErrEachFailed)

	return ErrValidation(causes...)
}

// GetField walks obj along a dotted path and returns the resolved value.
// Path segments may be plain identifiers ("Owner.Email") or include slice
// indices ("Items[0].Name"). The traversal supports structs, maps keyed by
// string, slices, arrays, and pointers (auto-dereferenced).
func GetField(obj any, path string) (any, error) {
	if cutils.Nil(obj) {
		return nil, ErrValidation(ErrObjectNil)
	}

	if path == "" {
		return nil, ErrValidation(ErrPathInvalid)
	}

	segments, err := parsePath(path)
	if err != nil {
		return nil, err
	}

	current := reflect.ValueOf(obj)
	for _, seg := range segments {
		current, err = walkSegment(current, seg)
		if err != nil {
			return nil, err
		}
	}

	return current.Interface(), nil
}
