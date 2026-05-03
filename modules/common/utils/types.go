package utils

import (
	cconstraints "github.com/guidomantilla/yarumo/common/constraints"
)

var (
	_ CoalesceFn[any]          = Coalesce
	_ TernaryFn[any]           = Ternary
	_ EqualFn                  = Equal
	_ NotEqualFn               = NotEqual
	_ NilFn                    = Nil
	_ NotNilFn                 = NotNil
	_ EmptyFn                  = Empty
	_ NotEmptyFn               = NotEmpty
	_ RandomStringFn           = RandomString
	_ SubstringFn              = Substring
	_ ChunkStringFn            = ChunkString
	_ CapitalizeFn             = Capitalize
	_ WordsFn                  = Words
	_ PascalCaseFn             = PascalCase
	_ CamelCaseFn              = CamelCase
	_ KebabCaseFn              = KebabCase
	_ SnakeCaseFn              = SnakeCase
	_ FilterByFn[any]          = FilterBy
	_ CountFn[any]             = Count
	_ CountByFn[any]           = CountBy
	_ ToMapFn[any]             = ToMap
	_ ToMapByFn[any]           = ToMapBy
	_ InFn[any]                = In
	_ NotInFn[any]             = NotIn
	_ EveryFn[any]             = Every
	_ SomeFn[any]              = Some
	_ NoneFn[any]              = None
	_ UnionFn[any]             = Union
	_ IntersectionFn[any]      = Intersection
	_ DifferenceFn[any]        = Difference
	_ MapFn[any, any]          = Map
	_ MapByFn[any, any]        = MapBy
	_ CopyFn[any]              = Copy
	_ ReverseFn[any]           = Reverse
	_ ShuffleFn[any]           = Shuffle
	_ DeduplicateFn[any]       = Deduplicate
	_ SortFn[int]              = Sort
	_ ChunkFn[any]             = Chunk
	_ DeleteFn[any]            = Delete
	_ DeleteRangeFn[any]       = DeleteRange
	_ PushFn[any]              = Push
	_ PopFn[any]               = Pop
	_ MaxFn[int]               = Max
	_ MinFn[int]               = Min
	_ HasKeyFn[any, any]       = HasKey
	_ KeysFn[any, any]         = Keys
	_ UniqueKeysFn[any, any]   = UniqueKeys
	_ ValuesFn[any, any]       = Values
	_ UniqueValuesFn[any, any] = UniqueValues
	_ PickByKeysFn[any, any]   = PickByKeys
	_ PickByValuesFn[any, any] = PickByValues
	_ OmitByKeysFn[any, any]   = OmitByKeys
	_ OmitByValuesFn[any, any] = OmitByValues
)

// --- Predicate function types ---

// CoalesceFn is the function type for Coalesce.
type CoalesceFn[T any] func(values ...T) T

// TernaryFn is the function type for Ternary.
type TernaryFn[T any] func(condition bool, trueValue T, falseValue T) T

// EqualFn is the function type for Equal.
type EqualFn func(x any, y any) bool

// NotEqualFn is the function type for NotEqual.
type NotEqualFn func(x any, y any) bool

// NilFn is the function type for Nil.
type NilFn func(x any) bool

// NotNilFn is the function type for NotNil.
type NotNilFn func(x any) bool

// EmptyFn is the function type for Empty.
type EmptyFn func(x ...any) bool

// NotEmptyFn is the function type for NotEmpty.
type NotEmptyFn func(x ...any) bool

// --- String function types ---

// RandomStringFn is the function type for RandomString.
type RandomStringFn func(length int, options ...Option) string

// SubstringFn is the function type for Substring.
type SubstringFn func(str string, offset uint, length uint) string

// ChunkStringFn is the function type for ChunkString.
type ChunkStringFn func(s string, size int) []string

// CapitalizeFn is the function type for Capitalize.
type CapitalizeFn func(s string, options ...Option) string

// WordsFn is the function type for Words.
type WordsFn func(s string) []string

// PascalCaseFn is the function type for PascalCase.
type PascalCaseFn func(s string) string

// CamelCaseFn is the function type for CamelCase.
type CamelCaseFn func(s string) string

// KebabCaseFn is the function type for KebabCase.
type KebabCaseFn func(s string) string

// SnakeCaseFn is the function type for SnakeCase.
type SnakeCaseFn func(s string) string

// --- Slice callback types ---

// ItemMapFn is the function type for element transformation in Map and MapBy.
type ItemMapFn[T, U any] func(T) U

// FilterFn is the function type for element filtering predicates.
type FilterFn[T cconstraints.Comparable] func(x T) bool

// --- Slice function types ---

// FilterByFn is the function type for FilterBy.
type FilterByFn[T cconstraints.Comparable] func(slice []T, keep FilterFn[T]) []T

// CountFn is the function type for Count.
type CountFn[T cconstraints.Comparable] func(slice []T) int

// CountByFn is the function type for CountBy.
type CountByFn[T cconstraints.Comparable] func(slice []T, keep FilterFn[T]) int

// ToMapFn is the function type for ToMap.
type ToMapFn[T cconstraints.Comparable] func(slice []T) map[T]int

// ToMapByFn is the function type for ToMapBy.
type ToMapByFn[T cconstraints.Comparable] func(slice []T, keep FilterFn[T]) map[T]int

// InFn is the function type for In.
type InFn[T cconstraints.Comparable] func(value T, slice ...T) bool

// NotInFn is the function type for NotIn.
type NotInFn[T cconstraints.Comparable] func(value T, slice ...T) bool

// EveryFn is the function type for Every.
type EveryFn[T cconstraints.Comparable] func(values []T, slice ...T) bool

// SomeFn is the function type for Some.
type SomeFn[T cconstraints.Comparable] func(values []T, slice ...T) bool

// NoneFn is the function type for None.
type NoneFn[T cconstraints.Comparable] func(values []T, slice ...T) bool

// UnionFn is the function type for Union.
type UnionFn[T cconstraints.Comparable] func(slice1 []T, slice2 []T) []T

// IntersectionFn is the function type for Intersection.
type IntersectionFn[T cconstraints.Comparable] func(slice1 []T, slice2 []T) []T

// DifferenceFn is the function type for Difference.
type DifferenceFn[T cconstraints.Comparable] func(slice1 []T, slice2 []T) ([]T, []T)

// MapFn is the function type for Map.
type MapFn[T, U cconstraints.Comparable] func(slice []T, fn ItemMapFn[T, U]) []U

// MapByFn is the function type for MapBy.
type MapByFn[T, U cconstraints.Comparable] func(slice []T, fn ItemMapFn[T, U], keep FilterFn[T]) []U

// CopyFn is the function type for Copy.
type CopyFn[T cconstraints.Comparable] func(slice []T) []T

// ReverseFn is the function type for Reverse.
type ReverseFn[T cconstraints.Comparable] func(slice []T) []T

// ShuffleFn is the function type for Shuffle.
type ShuffleFn[T cconstraints.Comparable] func(slice []T) []T

// DeduplicateFn is the function type for Deduplicate.
type DeduplicateFn[T cconstraints.Comparable] func(slice []T) []T

// SortFn is the function type for Sort.
type SortFn[T cconstraints.Ordenable] func(slice []T) []T

// ChunkFn is the function type for Chunk.
type ChunkFn[T cconstraints.Comparable] func(slice []T, size int) [][]T

// DeleteFn is the function type for Delete.
type DeleteFn[T cconstraints.Comparable] func(idx int, slice []T) []T

// DeleteRangeFn is the function type for DeleteRange.
type DeleteRangeFn[T cconstraints.Comparable] func(start, end int, slice []T) []T

// PushFn is the function type for Push.
type PushFn[T cconstraints.Comparable] func(slice []T, value T) []T

// PopFn is the function type for Pop.
type PopFn[T cconstraints.Comparable] func(slice []T) (T, []T)

// MaxFn is the function type for Max.
type MaxFn[T cconstraints.Ordenable] func(slice []T) T

// MinFn is the function type for Min.
type MinFn[T cconstraints.Ordenable] func(slice []T) T

// --- Map function types ---

// HasKeyFn is the function type for HasKey.
type HasKeyFn[K cconstraints.Comparable, V any] func(key K, keyValues map[K]V) bool

// KeysFn is the function type for Keys.
type KeysFn[K cconstraints.Comparable, V any] func(keyValues ...map[K]V) []K

// UniqueKeysFn is the function type for UniqueKeys.
type UniqueKeysFn[K cconstraints.Comparable, V any] func(keyValues ...map[K]V) []K

// ValuesFn is the function type for Values.
type ValuesFn[K cconstraints.Comparable, V any] func(keyValues ...map[K]V) []V

// UniqueValuesFn is the function type for UniqueValues.
type UniqueValuesFn[K cconstraints.Comparable, V cconstraints.Comparable] func(keyValues ...map[K]V) []V

// PickByKeysFn is the function type for PickByKeys.
type PickByKeysFn[K cconstraints.Comparable, V any] func(keyValues map[K]V, keys []K) map[K]V

// PickByValuesFn is the function type for PickByValues.
type PickByValuesFn[K cconstraints.Comparable, V cconstraints.Comparable] func(keyValues map[K]V, values []V) map[K]V

// OmitByKeysFn is the function type for OmitByKeys.
type OmitByKeysFn[K cconstraints.Comparable, V any] func(keyValues map[K]V, keys []K) map[K]V

// OmitByValuesFn is the function type for OmitByValues.
type OmitByValuesFn[K cconstraints.Comparable, V cconstraints.Comparable] func(keyValues map[K]V, values []V) map[K]V
