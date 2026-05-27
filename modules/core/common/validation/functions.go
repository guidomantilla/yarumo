package validation

import (
	"encoding/base64"
	"encoding/json"
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	cconstraints "github.com/guidomantilla/yarumo/core/common/constraints"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
	cuids "github.com/guidomantilla/yarumo/core/common/uids"
	cutils "github.com/guidomantilla/yarumo/core/common/utils"
)

// hostnameLabelRegex matches a single RFC 1123 hostname label: 1-63 chars,
// alphanumeric and hyphen, no leading or trailing hyphen.
var hostnameLabelRegex = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)

// semverRegex implements the official semver 2.0.0 grammar: Major.Minor.Patch
// with an optional pre-release segment (dot-separated identifiers, alphanumeric
// or numeric without leading zeros) and an optional build metadata segment.
var semverRegex = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

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

// Positive returns ErrNotPositive when v is not strictly greater than zero.
func Positive[T Numeric](v T) error {
	if v <= 0 {
		return ErrValidation(ErrNotPositive)
	}

	return nil
}

// Negative returns ErrNotNegative when v is not strictly less than zero.
func Negative[T Numeric](v T) error {
	if v >= 0 {
		return ErrValidation(ErrNotNegative)
	}

	return nil
}

// NonZero returns ErrZero when v equals zero.
func NonZero[T Numeric](v T) error {
	if v == 0 {
		return ErrValidation(ErrZero)
	}

	return nil
}

// MultipleOf returns ErrNotMultipleOf when v is not an integer multiple of
// factor. A zero factor never divides any value, so ErrNotMultipleOf is
// returned (the operator does not panic on div-by-zero). MultipleOf is
// restricted to integer types; floating-point modulo is ambiguous under
// rounding and would need a tolerance parameter the caller cannot safely
// pick at a leaf level.
func MultipleOf[T cconstraints.Integer](v, factor T) error {
	if factor == 0 {
		return ErrValidation(ErrNotMultipleOf)
	}

	if v%factor != 0 {
		return ErrValidation(ErrNotMultipleOf)
	}

	return nil
}

// IsIntegerString returns ErrIntegerStringInvalid when s does not parse as a
// base-10 signed 64-bit integer. The empty string is rejected.
func IsIntegerString(s string) error {
	if s == "" {
		return ErrValidation(ErrIntegerStringInvalid)
	}

	_, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return ErrValidation(cerrs.Wrap(ErrIntegerStringInvalid, err))
	}

	return nil
}

// IsFloatString returns ErrFloatStringInvalid when s does not parse as a
// 64-bit floating-point number. The empty string is rejected.
func IsFloatString(s string) error {
	if s == "" {
		return ErrValidation(ErrFloatStringInvalid)
	}

	_, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return ErrValidation(cerrs.Wrap(ErrFloatStringInvalid, err))
	}

	return nil
}

// Equal returns ErrNotEqual when v does not equal expected. The check uses
// Go's == operator, so the type must be comparable.
func Equal[T comparable](v, expected T) error {
	if v != expected {
		return ErrValidation(ErrNotEqual)
	}

	return nil
}

// NotEqual returns ErrMustNotEqual when v equals forbidden.
func NotEqual[T comparable](v, forbidden T) error {
	if v == forbidden {
		return ErrValidation(ErrMustNotEqual)
	}

	return nil
}

// EqualIgnoreCase returns ErrNotEqual when s and expected differ under
// Unicode case folding (see strings.EqualFold).
func EqualIgnoreCase(s, expected string) error {
	if !strings.EqualFold(s, expected) {
		return ErrValidation(ErrNotEqual)
	}

	return nil
}

// OneOf returns ErrNotInAllowed when v is not present in allowed, and
// ErrEmptyAllowed when allowed is empty (an empty allow-list rejects every
// value, which is almost always a configuration mistake — surfacing it
// explicitly helps the caller spot it).
func OneOf[T comparable](v T, allowed []T) error {
	if len(allowed) == 0 {
		return ErrValidation(ErrEmptyAllowed)
	}

	if !cutils.In(v, allowed...) {
		return ErrValidation(ErrNotInAllowed)
	}

	return nil
}

// NotIn returns ErrInForbidden when v is present in forbidden. An empty
// forbidden list trivially passes.
func NotIn[T comparable](v T, forbidden []T) error {
	if !cutils.NotIn(v, forbidden...) {
		return ErrValidation(ErrInForbidden)
	}

	return nil
}

// IsUID returns ErrUIDInvalid when s is not a syntactically valid unique
// identifier under the format validator f. Empty s and nil f are rejected.
//
// The leaf is algorithm-agnostic: it consumes the cuids.IsUIDFn contract
// declared in common/uids so this package does not depend on any concrete
// UID library. Callers compose with the canonical algorithm of their choice
// (UUID, ULID, XID, NanoID, CUID2, …) by passing the matching predicate
// from modules/extension/common/uids, e.g.
// validation.IsUID(s, extuids.IsUUID).
func IsUID(s string, f cuids.IsUIDFn) error {
	if s == "" {
		return ErrValidation(ErrUIDInvalid)
	}

	if f == nil {
		return ErrValidation(ErrUIDInvalid)
	}

	if !f(s) {
		return ErrValidation(ErrUIDInvalid)
	}

	return nil
}

// IsJWT returns ErrJWTInvalid when s is not a syntactically valid JWT: three
// non-empty base64url segments separated by dots, with header and payload
// segments that decode to JSON objects. The signature segment is not
// verified — that requires the signing key and is out of scope for a
// syntactic leaf.
func IsJWT(s string) error {
	if s == "" {
		return ErrValidation(ErrJWTInvalid)
	}

	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return ErrValidation(ErrJWTInvalid)
	}

	if slices.Contains(parts, "") {
		return ErrValidation(ErrJWTInvalid)
	}

	for _, idx := range []int{0, 1} {
		raw, err := base64.RawURLEncoding.DecodeString(parts[idx])
		if err != nil {
			return ErrValidation(cerrs.Wrap(ErrJWTInvalid, err))
		}

		var obj map[string]any
		err = json.Unmarshal(raw, &obj)
		if err != nil {
			return ErrValidation(cerrs.Wrap(ErrJWTInvalid, err))
		}
	}

	return nil
}

// IsSemver returns ErrSemverInvalid when s is not a syntactically valid
// semver 2.0.0 version string. The leading 'v' prefix is not accepted; pass
// the bare Major.Minor.Patch form.
func IsSemver(s string) error {
	if s == "" {
		return ErrValidation(ErrSemverInvalid)
	}

	if !semverRegex.MatchString(s) {
		return ErrValidation(ErrSemverInvalid)
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

// Contains returns ErrContainsMissing when s does not contain substr. Follows
// strings.Contains semantics, so an empty substr trivially passes.
func Contains(s, substr string) error {
	if !strings.Contains(s, substr) {
		return ErrValidation(ErrContainsMissing)
	}

	return nil
}

// HasPrefix returns ErrPrefixMissing when s does not begin with prefix. An
// empty prefix trivially passes.
func HasPrefix(s, prefix string) error {
	if !strings.HasPrefix(s, prefix) {
		return ErrValidation(ErrPrefixMissing)
	}

	return nil
}

// HasSuffix returns ErrSuffixMissing when s does not end with suffix. An
// empty suffix trivially passes.
func HasSuffix(s, suffix string) error {
	if !strings.HasSuffix(s, suffix) {
		return ErrValidation(ErrSuffixMissing)
	}

	return nil
}

// IsLowercase returns ErrNotLowercase when s contains any uppercase letter.
// Non-letter characters are ignored; the empty string trivially passes.
func IsLowercase(s string) error {
	if strings.ToLower(s) != s {
		return ErrValidation(ErrNotLowercase)
	}

	return nil
}

// IsUppercase returns ErrNotUppercase when s contains any lowercase letter.
// Non-letter characters are ignored; the empty string trivially passes.
func IsUppercase(s string) error {
	if strings.ToUpper(s) != s {
		return ErrValidation(ErrNotUppercase)
	}

	return nil
}

// IsAlpha returns ErrNotAlpha when s contains any rune that is not a unicode
// letter. The empty string is rejected.
func IsAlpha(s string) error {
	if s == "" {
		return ErrValidation(ErrNotAlpha)
	}

	for _, r := range s {
		if !unicode.IsLetter(r) {
			return ErrValidation(ErrNotAlpha)
		}
	}

	return nil
}

// IsAlphanumeric returns ErrNotAlphanumeric when s contains any rune that is
// not a unicode letter or digit. The empty string is rejected.
func IsAlphanumeric(s string) error {
	if s == "" {
		return ErrValidation(ErrNotAlphanumeric)
	}

	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return ErrValidation(ErrNotAlphanumeric)
		}
	}

	return nil
}

// IsNumeric returns ErrNotNumeric when s contains any rune that is not a
// unicode digit. The empty string is rejected. The check is syntactic — it
// does not attempt to parse s as an integer.
func IsNumeric(s string) error {
	if s == "" {
		return ErrValidation(ErrNotNumeric)
	}

	for _, r := range s {
		if !unicode.IsDigit(r) {
			return ErrValidation(ErrNotNumeric)
		}
	}

	return nil
}

// IsASCII returns ErrNotASCII when s contains any byte above 0x7F. The empty
// string is rejected.
func IsASCII(s string) error {
	if s == "" {
		return ErrValidation(ErrNotASCII)
	}

	for i := range len(s) {
		if s[i] > unicode.MaxASCII {
			return ErrValidation(ErrNotASCII)
		}
	}

	return nil
}

// IsHex returns ErrNotHex when s contains any character outside [0-9a-fA-F].
// The empty string is rejected.
func IsHex(s string) error {
	if s == "" {
		return ErrValidation(ErrNotHex)
	}

	for _, r := range s {
		isDigit := r >= '0' && r <= '9'
		isLower := r >= 'a' && r <= 'f'
		isUpper := r >= 'A' && r <= 'F'
		if !isDigit && !isLower && !isUpper {
			return ErrValidation(ErrNotHex)
		}
	}

	return nil
}

// IsBase64 returns ErrBase64Invalid when s is not a valid RFC 4648 standard
// base64 encoding (with padding). The empty string is rejected.
func IsBase64(s string) error {
	if s == "" {
		return ErrValidation(ErrBase64Invalid)
	}

	_, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return ErrValidation(cerrs.Wrap(ErrBase64Invalid, err))
	}

	return nil
}

// IsTrimmed returns ErrNotTrimmed when s has any leading or trailing unicode
// whitespace. The empty string trivially passes.
func IsTrimmed(s string) error {
	if strings.TrimSpace(s) != s {
		return ErrValidation(ErrNotTrimmed)
	}

	return nil
}

// IsIP returns ErrIPInvalid when s is not a syntactically valid IPv4 or
// IPv6 address.
func IsIP(s string) error {
	if net.ParseIP(s) == nil {
		return ErrValidation(ErrIPInvalid)
	}

	return nil
}

// IsIPv4 returns ErrIPv4Invalid when s is not a syntactically valid IPv4
// address. IPv6 inputs (even in IPv4-mapped form like ::ffff:1.2.3.4) are
// rejected; callers that want either form should use IsIP.
func IsIPv4(s string) error {
	ip := net.ParseIP(s)
	if ip == nil || ip.To4() == nil || strings.Contains(s, ":") {
		return ErrValidation(ErrIPv4Invalid)
	}

	return nil
}

// IsIPv6 returns ErrIPv6Invalid when s is not a syntactically valid IPv6
// address. Plain IPv4 dotted-decimal inputs are rejected.
func IsIPv6(s string) error {
	ip := net.ParseIP(s)
	if ip == nil || !strings.Contains(s, ":") {
		return ErrValidation(ErrIPv6Invalid)
	}

	return nil
}

// IsCIDR returns ErrCIDRInvalid when s is not a syntactically valid CIDR
// notation for either IPv4 or IPv6 (e.g. 192.0.2.0/24, 2001:db8::/32).
func IsCIDR(s string) error {
	_, _, err := net.ParseCIDR(s)
	if err != nil {
		return ErrValidation(cerrs.Wrap(ErrCIDRInvalid, err))
	}

	return nil
}

// IsMAC returns ErrMACInvalid when s is not a syntactically valid MAC
// address (48-bit or 64-bit, accepting ':' / '-' separators per
// net.ParseMAC).
func IsMAC(s string) error {
	_, err := net.ParseMAC(s)
	if err != nil {
		return ErrValidation(cerrs.Wrap(ErrMACInvalid, err))
	}

	return nil
}

// IsHostname returns ErrHostnameInvalid when s is not a valid RFC 1123
// hostname: one or more dot-separated labels of 1-63 chars each
// (alphanumeric + hyphen, no leading/trailing hyphen), total length at most
// 253 characters. The empty string is rejected.
func IsHostname(s string) error {
	if s == "" || len(s) > 253 {
		return ErrValidation(ErrHostnameInvalid)
	}

	for label := range strings.SplitSeq(s, ".") {
		if !hostnameLabelRegex.MatchString(label) {
			return ErrValidation(ErrHostnameInvalid)
		}
	}

	return nil
}

// IsFQDN returns ErrFQDNInvalid when s is not a fully qualified domain
// name: a valid hostname with at least two labels whose TLD contains at
// least one letter (so plain IPv4 dotted-decimal does not satisfy IsFQDN).
func IsFQDN(s string) error {
	if IsHostname(s) != nil {
		return ErrValidation(ErrFQDNInvalid)
	}

	labels := strings.Split(s, ".")
	if len(labels) < 2 {
		return ErrValidation(ErrFQDNInvalid)
	}

	tld := labels[len(labels)-1]
	if !strings.ContainsFunc(tld, unicode.IsLetter) {
		return ErrValidation(ErrFQDNInvalid)
	}

	return nil
}

// IsPort returns ErrPortInvalid when n is not in the inclusive range
// [1, 65535]. The leaf is generic so callers can pass int, uint16, or any
// other numeric type without conversion noise. The comparison is performed
// in float64 so the bounds fit every type in the Numeric constraint.
func IsPort[T Numeric](n T) error {
	f := float64(n)
	if f < 1 || f > 65535 {
		return ErrValidation(ErrPortInvalid)
	}

	return nil
}

// IsRFC3339 returns ErrDateInvalid when s does not parse as a time using
// the RFC 3339 layout.
func IsRFC3339(s string) error {
	if s == "" {
		return ErrValidation(ErrDateInvalid)
	}

	_, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return ErrValidation(cerrs.Wrap(ErrDateInvalid, err))
	}

	return nil
}

// IsDate returns ErrLayoutInvalid when layout is empty, and ErrDateInvalid
// when s does not parse against the supplied time.Parse layout.
func IsDate(s, layout string) error {
	if layout == "" {
		return ErrValidation(ErrLayoutInvalid)
	}

	if s == "" {
		return ErrValidation(ErrDateInvalid)
	}

	_, err := time.Parse(layout, s)
	if err != nil {
		return ErrValidation(cerrs.Wrap(ErrDateInvalid, err))
	}

	return nil
}

// Before returns ErrTimeBefore when t is not strictly before ref.
func Before(t, ref time.Time) error {
	if !t.Before(ref) {
		return ErrValidation(ErrTimeBefore)
	}

	return nil
}

// After returns ErrTimeAfter when t is not strictly after ref.
func After(t, ref time.Time) error {
	if !t.After(ref) {
		return ErrValidation(ErrTimeAfter)
	}

	return nil
}

// BetweenTime returns ErrTimeOutOfRange when t is outside [lo, hi], and
// ErrInvalidTimeRange when the caller provided lo > hi.
func BetweenTime(t, lo, hi time.Time) error {
	if lo.After(hi) {
		return ErrValidation(ErrInvalidTimeRange)
	}

	if t.Before(lo) || t.After(hi) {
		return ErrValidation(ErrTimeOutOfRange)
	}

	return nil
}

// MinCount returns ErrCountBelowMin when len(xs) is below n. Negative
// thresholds are treated as zero.
func MinCount[T any](xs []T, n int) error {
	if n < 0 {
		n = 0
	}

	if len(xs) < n {
		return ErrValidation(ErrCountBelowMin)
	}

	return nil
}

// MaxCount returns ErrCountAboveMax when len(xs) is above n. Negative
// thresholds are treated as zero — only the empty slice passes.
func MaxCount[T any](xs []T, n int) error {
	if n < 0 {
		n = 0
	}

	if len(xs) > n {
		return ErrValidation(ErrCountAboveMax)
	}

	return nil
}

// CountInRange returns ErrCountOutOfRange when len(xs) is outside
// [lo, hi], and ErrInvalidCountRange when lo > hi.
func CountInRange[T any](xs []T, lo, hi int) error {
	if lo > hi {
		return ErrValidation(ErrInvalidCountRange)
	}

	if len(xs) < lo || len(xs) > hi {
		return ErrValidation(ErrCountOutOfRange)
	}

	return nil
}

// Unique returns ErrDuplicate when any element of xs repeats. The empty
// slice trivially passes.
func Unique[T comparable](xs []T) error {
	seen := make(map[T]struct{}, len(xs))
	for _, x := range xs {
		_, dup := seen[x]
		if dup {
			return ErrValidation(ErrDuplicate)
		}

		seen[x] = struct{}{}
	}

	return nil
}

// SortedAsc returns ErrNotSortedAsc when xs is not in non-decreasing order.
// The empty slice and single-element slice trivially pass. Equal adjacent
// elements are accepted.
func SortedAsc[T cconstraints.Ordenable](xs []T) error {
	for i := 1; i < len(xs); i++ {
		if xs[i] < xs[i-1] {
			return ErrValidation(ErrNotSortedAsc)
		}
	}

	return nil
}

// SortedDesc returns ErrNotSortedDesc when xs is not in non-increasing
// order. The empty slice and single-element slice trivially pass. Equal
// adjacent elements are accepted.
func SortedDesc[T cconstraints.Ordenable](xs []T) error {
	for i := 1; i < len(xs); i++ {
		if xs[i] > xs[i-1] {
			return ErrValidation(ErrNotSortedDesc)
		}
	}

	return nil
}

// HasKey returns ErrKeyMissing when m does not contain key. A nil map is
// treated as having no keys and returns ErrKeyMissing.
func HasKey[K comparable, V any](m map[K]V, key K) error {
	if !cutils.HasKey(key, m) {
		return ErrValidation(ErrKeyMissing)
	}

	return nil
}

// MinKeys returns ErrMinKeys when m has fewer than n keys. A nil map is
// treated as having zero keys. Negative thresholds are treated as zero.
func MinKeys[K comparable, V any](m map[K]V, n int) error {
	if n < 0 {
		n = 0
	}

	if len(m) < n {
		return ErrValidation(ErrMinKeys)
	}

	return nil
}

// MaxKeys returns ErrMaxKeys when m has more than n keys. Negative
// thresholds are treated as zero — only the empty map passes.
func MaxKeys[K comparable, V any](m map[K]V, n int) error {
	if n < 0 {
		n = 0
	}

	if len(m) > n {
		return ErrValidation(ErrMaxKeys)
	}

	return nil
}

// Optional returns a CheckFn that passes when v is the zero value of T and
// otherwise delegates to check. It encodes the canonical "validate only if
// present" idiom and stays consistent with IsRequired by using cutils.Empty
// for zero-value detection.
func Optional[T any](check CheckFn[T]) CheckFn[T] {
	return func(v T) error {
		if cutils.Empty(v) {
			return nil
		}

		return check(v)
	}
}

// AnyOf returns a CheckFn that passes as soon as any inner check passes.
// When every inner check fails, all violations are aggregated into a single
// validation error so the caller can see what was attempted. An empty
// checks list trivially passes.
func AnyOf[T any](checks ...CheckFn[T]) CheckFn[T] {
	return func(v T) error {
		var causes []error
		for _, c := range checks {
			err := c(v)
			if err == nil {
				return nil
			}

			causes = append(causes, err)
		}

		if len(causes) == 0 {
			return nil
		}

		return ErrValidation(causes...)
	}
}

// AllOf returns a CheckFn that passes only when every inner check passes.
// All violations from failing checks are aggregated into a single
// validation error. An empty checks list trivially passes.
func AllOf[T any](checks ...CheckFn[T]) CheckFn[T] {
	return func(v T) error {
		var causes []error
		for _, c := range checks {
			err := c(v)
			if err != nil {
				causes = append(causes, err)
			}
		}

		if len(causes) == 0 {
			return nil
		}

		return ErrValidation(causes...)
	}
}

// Not returns a CheckFn that passes when the inner check fails, and
// returns ErrAssertionInverted when the inner check unexpectedly passes.
func Not[T any](check CheckFn[T]) CheckFn[T] {
	return func(v T) error {
		err := check(v)
		if err == nil {
			return ErrValidation(ErrAssertionInverted)
		}

		return nil
	}
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
