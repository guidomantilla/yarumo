// Package validation provides function-typed leaf validators and reflection
// helpers used by callers (handlers, services) and by higher-level
// config-driven engines such as modules/validation/.
//
// Leaves are plain functions: they take a value (and any extra parameters)
// and return an error. They never depend on struct tags. The package error
// type embeds errs.TypedError with Type = "validation" so violations group
// cleanly under errs.AsErrorInfo.
package validation

import (
	"time"

	cconstraints "github.com/guidomantilla/yarumo/core/common/constraints"
	cuids "github.com/guidomantilla/yarumo/core/common/uids"
)

var (
	_ CheckFn[string]           = IsEmail
	_ CheckFn[string]           = IsURL
	_ CheckUIDFn                = IsUID
	_ CheckFn[string]           = IsJWT
	_ CheckFn[string]           = IsSemver
	_ CheckFn[string]           = IsRequired[string]
	_ CheckFn[string]           = MustBeUndefined[string]
	_ CheckFn[string]           = IsLowercase
	_ CheckFn[string]           = IsUppercase
	_ CheckFn[string]           = IsAlpha
	_ CheckFn[string]           = IsAlphanumeric
	_ CheckFn[string]           = IsNumeric
	_ CheckFn[string]           = IsASCII
	_ CheckFn[string]           = IsHex
	_ CheckFn[string]           = IsBase64
	_ CheckFn[string]           = IsTrimmed
	_ MinLenFn                  = MinLen
	_ MaxLenFn                  = MaxLen
	_ MatchesRegexFn            = MatchesRegex
	_ ContainsFn                = Contains
	_ HasPrefixFn               = HasPrefix
	_ HasSuffixFn               = HasSuffix
	_ MinFn[int]                = Min[int]
	_ MaxFn[int]                = Max[int]
	_ InRangeFn[int]            = InRange[int]
	_ PositiveFn[int]           = Positive[int]
	_ NegativeFn[int]           = Negative[int]
	_ NonZeroFn[int]            = NonZero[int]
	_ MultipleOfFn[int]         = MultipleOf[int]
	_ CheckFn[string]           = IsIntegerString
	_ CheckFn[string]           = IsFloatString
	_ EqualFn[string]           = Equal[string]
	_ NotEqualFn[string]        = NotEqual[string]
	_ EqualIgnoreCaseFn         = EqualIgnoreCase
	_ OneOfFn[string]           = OneOf[string]
	_ NotInFn[string]           = NotIn[string]
	_ CheckFn[string]           = IsIP
	_ CheckFn[string]           = IsIPv4
	_ CheckFn[string]           = IsIPv6
	_ CheckFn[string]           = IsCIDR
	_ CheckFn[string]           = IsMAC
	_ CheckFn[string]           = IsHostname
	_ CheckFn[string]           = IsFQDN
	_ IsPortFn[int]             = IsPort[int]
	_ CheckFn[string]           = IsRFC3339
	_ IsDateFn                  = IsDate
	_ BeforeFn                  = Before
	_ AfterFn                   = After
	_ BetweenTimeFn             = BetweenTime
	_ MinCountFn[int]           = MinCount[int]
	_ MaxCountFn[int]           = MaxCount[int]
	_ CountInRangeFn[int]       = CountInRange[int]
	_ UniqueFn[int]             = Unique[int]
	_ SortedAscFn[int]          = SortedAsc[int]
	_ SortedDescFn[int]         = SortedDesc[int]
	_ HasKeyFn[string, int]     = HasKey[string, int]
	_ MinKeysFn[string, int]    = MinKeys[string, int]
	_ MaxKeysFn[string, int]    = MaxKeys[string, int]
	_ CollectionCheckFn[string] = NonEmpty[string]
	_ EachFn[string]            = Each[string]
	_ FieldFn                   = GetField
)

// Numeric is the constraint accepted by the numeric leaf validators. It
// covers every Go integer and floating-point type.
type Numeric interface {
	cconstraints.Number
}

// CheckFn is the function type for a generic leaf validator over a single value.
type CheckFn[T any] func(value T) error

// CollectionCheckFn is the function type for a leaf validator over a slice.
type CollectionCheckFn[T any] func(xs []T) error

// MinLenFn is the function type for MinLen.
type MinLenFn func(s string, n int) error

// MaxLenFn is the function type for MaxLen.
type MaxLenFn func(s string, n int) error

// MatchesRegexFn is the function type for MatchesRegex.
type MatchesRegexFn func(s string, pattern string) error

// CheckUIDFn is the function type for IsUID. The leaf takes the input
// string and a cuids.IsUIDFn predicate that encodes the chosen algorithm
// (UUID, ULID, XID, NanoID, CUID2, …). Callers supply the predicate to
// keep this package independent of any concrete UID library.
type CheckUIDFn func(s string, f cuids.IsUIDFn) error

// ContainsFn is the function type for Contains.
type ContainsFn func(s, substr string) error

// HasPrefixFn is the function type for HasPrefix.
type HasPrefixFn func(s, prefix string) error

// HasSuffixFn is the function type for HasSuffix.
type HasSuffixFn func(s, suffix string) error

// MinFn is the function type for Min.
type MinFn[T Numeric] func(v, lo T) error

// MaxFn is the function type for Max.
type MaxFn[T Numeric] func(v, hi T) error

// InRangeFn is the function type for InRange.
type InRangeFn[T Numeric] func(v, lo, hi T) error

// PositiveFn is the function type for Positive.
type PositiveFn[T Numeric] func(v T) error

// NegativeFn is the function type for Negative.
type NegativeFn[T Numeric] func(v T) error

// NonZeroFn is the function type for NonZero.
type NonZeroFn[T Numeric] func(v T) error

// MultipleOfFn is the function type for MultipleOf.
type MultipleOfFn[T cconstraints.Integer] func(v, factor T) error

// EqualFn is the function type for Equal.
type EqualFn[T comparable] func(v, expected T) error

// NotEqualFn is the function type for NotEqual.
type NotEqualFn[T comparable] func(v, forbidden T) error

// EqualIgnoreCaseFn is the function type for EqualIgnoreCase.
type EqualIgnoreCaseFn func(s, expected string) error

// OneOfFn is the function type for OneOf.
type OneOfFn[T comparable] func(v T, allowed []T) error

// NotInFn is the function type for NotIn.
type NotInFn[T comparable] func(v T, forbidden []T) error

// IsPortFn is the function type for IsPort.
type IsPortFn[T Numeric] func(n T) error

// IsDateFn is the function type for IsDate.
type IsDateFn func(s, layout string) error

// BeforeFn is the function type for Before.
type BeforeFn func(t, ref time.Time) error

// AfterFn is the function type for After.
type AfterFn func(t, ref time.Time) error

// BetweenTimeFn is the function type for BetweenTime.
type BetweenTimeFn func(t, lo, hi time.Time) error

// MinCountFn is the function type for MinCount.
type MinCountFn[T any] func(xs []T, n int) error

// MaxCountFn is the function type for MaxCount.
type MaxCountFn[T any] func(xs []T, n int) error

// CountInRangeFn is the function type for CountInRange.
type CountInRangeFn[T any] func(xs []T, lo, hi int) error

// UniqueFn is the function type for Unique.
type UniqueFn[T comparable] func(xs []T) error

// SortedAscFn is the function type for SortedAsc.
type SortedAscFn[T cconstraints.Ordenable] func(xs []T) error

// SortedDescFn is the function type for SortedDesc.
type SortedDescFn[T cconstraints.Ordenable] func(xs []T) error

// HasKeyFn is the function type for HasKey.
type HasKeyFn[K comparable, V any] func(m map[K]V, key K) error

// MinKeysFn is the function type for MinKeys.
type MinKeysFn[K comparable, V any] func(m map[K]V, n int) error

// MaxKeysFn is the function type for MaxKeys.
type MaxKeysFn[K comparable, V any] func(m map[K]V, n int) error

// EachFn is the function type for Each.
type EachFn[T any] func(xs []T, check func(T) error) error

// FieldFn is the function type for GetField.
type FieldFn func(obj any, path string) (any, error)
