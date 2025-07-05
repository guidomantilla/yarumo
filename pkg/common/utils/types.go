package utils

import (
	"github.com/guidomantilla/yarumo/pkg/common/constraints"
)

var (
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

type TernaryFn[T any] func(condition bool, trueValue T, falseValue T) T

type EqualFn func(x any, y any) bool

type NotEqualFn func(x any, y any) bool

type NilFn func(x any) bool

type NotNilFn func(x any) bool

type EmptyFn func(x any) bool

type NotEmptyFn func(x any) bool

type RandomStringFn func(length int, opts ...Option) string

type SubstringFn func(str string, offset uint, length uint) string

type ChunkStringFn func(s string, size int) []string

type CapitalizeFn func(s string, opts ...Option) string

type WordsFn func(s string) []string

type PascalCaseFn func(s string) string

type CamelCaseFn func(s string) string

type KebabCaseFn func(s string) string

type SnakeCaseFn func(s string) string

type ItemMapFn[T, U any] func(T) U

type FilterFn[T constraints.Comparable] func(x T) bool

type FilterByFn[T constraints.Comparable] func(slice []T, keep FilterFn[T]) []T

type CountFn[T constraints.Comparable] func(slice []T) int

type CountByFn[T constraints.Comparable] func(slice []T, keep FilterFn[T]) int

type ToMapFn[T constraints.Comparable] func(slice []T) map[T]int

type ToMapByFn[T constraints.Comparable] func(slice []T, keep FilterFn[T]) map[T]int

type InFn[T constraints.Comparable] func(value T, slice ...T) bool

type NotInFn[T constraints.Comparable] func(value T, slice ...T) bool

type EveryFn[T constraints.Comparable] func(values []T, slice ...T) bool

type SomeFn[T constraints.Comparable] func(values []T, slice ...T) bool

type NoneFn[T constraints.Comparable] func(values []T, slice ...T) bool

type UnionFn[T constraints.Comparable] func(slice1 []T, slice2 []T) []T

type IntersectionFn[T constraints.Comparable] func(slice1 []T, slice2 []T) []T

type DifferenceFn[T constraints.Comparable] func(slice1 []T, slice2 []T) ([]T, []T)

type MapFn[T, U constraints.Comparable] func(slice []T, fn ItemMapFn[T, U]) []U

type MapByFn[T, U constraints.Comparable] func(slice []T, fn ItemMapFn[T, U], keep FilterFn[T]) []U

type CopyFn[T constraints.Comparable] func(slice []T) []T

type ReverseFn[T constraints.Comparable] func(slice []T) []T

type ShuffleFn[T constraints.Comparable] func(slice []T) []T

type DeduplicateFn[T constraints.Comparable] func(slice []T) []T

type SortFn[T constraints.Ordenable] func(slice []T) []T

type ChunkFn[T constraints.Comparable] func(slice []T, size int) [][]T

type DeleteFn[T constraints.Comparable] func(idx int, slice []T) []T

type DeleteRangeFn[T constraints.Comparable] func(start, end int, slice []T) []T

type PushFn[T constraints.Comparable] func(slice []T, value T) []T

type PopFn[T constraints.Comparable] func(slice []T) (T, []T)

type MaxFn[T constraints.Ordenable] func(slice []T) T

type MinFn[T constraints.Ordenable] func(slice []T) T

type HasKeyFn[K constraints.Comparable, V any] func(key K, keyValues map[K]V) bool

type KeysFn[K constraints.Comparable, V any] func(keyValues ...map[K]V) []K

type UniqueKeysFn[K constraints.Comparable, V any] func(keyValues ...map[K]V) []K

type ValuesFn[K constraints.Comparable, V any] func(keyValues ...map[K]V) []V

type UniqueValuesFn[K constraints.Comparable, V constraints.Comparable] func(keyValues ...map[K]V) []V

type PickByKeysFn[K constraints.Comparable, V any] func(keyValues map[K]V, keys []K) map[K]V

type PickByValuesFn[K constraints.Comparable, V constraints.Comparable] func(keyValues map[K]V, values []V) map[K]V

type OmitByKeysFn[K constraints.Comparable, V any] func(keyValues map[K]V, keys []K) map[K]V

type OmitByValuesFn[K constraints.Comparable, V constraints.Comparable] func(keyValues map[K]V, values []V) map[K]V
