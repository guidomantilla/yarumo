package utils

import (
	rand "math/rand/v2"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/text/cases"

	"github.com/guidomantilla/yarumo/pkg/common/constraints"
	"github.com/guidomantilla/yarumo/pkg/common/pointer"
)

// Ternary returns trueValue if condition is true, otherwise returns falseValue.
func Ternary[T any](condition bool, trueValue T, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}

//

// Equal is a function that checks if two values are equal.
func Equal(x any, y any) bool {
	return cmp.Equal(x, y)
}

// NotEqual is a function that checks if two values are not equal.
func NotEqual(x any, y any) bool {
	return !cmp.Equal(x, y)
}

// Nil checks if a value is nil or if it's a reference type with a nil underlying value.
func Nil(x any) bool {
	return pointer.IsNil(x)
}

// NotNil checks if a value is not nil or if it's not a reference type with a nil underlying value.
func NotNil(x any) bool {
	return pointer.IsNotNil(x)
}

// Empty checks if a value is empty.
func Empty(x any) bool {
	if Nil(x) {
		return true
	}

	val := reflect.ValueOf(x)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return true
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return val.Len() == 0
	case reflect.Interface:
		return val.IsNil()
	default:
		return pointer.IsZero(val)
	}
}

// NotEmpty checks if a value is not empty.
func NotEmpty(x any) bool {
	return !Empty(x)
}

//

// RandomString return a random string.
func RandomString(length int, opts ...Option) string {

	options := NewOptions(opts...)
	if length <= 0 || len(options.Charset) == 0 {
		return ""
	}

	b := make([]byte, length)
	for i := range b {
		b[i] = options.Charset[rand.IntN(len(options.Charset))] //nolint:gosec
	}
	return string(b)
}

// Substring return part of a string.
func Substring(str string, offset uint, length uint) string {
	rs := []rune(str)
	size := uint(len(rs))

	offset = size + offset
	if offset >= size {
		return ""
	}

	if length > size-offset {
		length = size - offset
	}

	return strings.ReplaceAll(string(rs[offset:offset+length]), "\x00", "")
}

// ChunkString returns an array of strings split into groups the length of size.
// If array can't be split evenly, the final chunk will be the remaining elements.
func ChunkString(str string, size int) []string {
	if size <= 0 {
		return []string{""}
	}

	if len(str) == 0 {
		return []string{""}
	}

	if size >= len(str) {
		return []string{str}
	}

	var chunks = make([]string, 0, ((len(str)-1)/size)+1)
	currentLen := 0
	currentStart := 0
	for i := range str {
		if currentLen == size {
			chunks = append(chunks, str[currentStart:i])
			currentLen = 0
			currentStart = i
		}
		currentLen++
	}
	chunks = append(chunks, str[currentStart:])
	return chunks
}

// Capitalize converts the first character of string to upper case and the remaining to lower case.
func Capitalize(str string, opts ...Option) string {
	options := NewOptions(opts...)
	return cases.Title(options.Lang).String(str)
}

var (
	splitWordReg         = regexp.MustCompile(`([a-z])([A-Z0-9])|([a-zA-Z])([0-9])|([0-9])([a-zA-Z])|([A-Z])([A-Z])([a-z])`)
	splitNumberLetterReg = regexp.MustCompile(`([0-9])([a-zA-Z])`)
)

// Words splits string into an array of its words.
func Words(str string) []string {
	str = splitWordReg.ReplaceAllString(str, `$1$3$5$7 $2$4$6$8$9`)
	str = splitNumberLetterReg.ReplaceAllString(str, "$1 $2")
	var result strings.Builder
	for _, r := range str {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		} else {
			result.WriteRune(' ')
		}
	}
	return strings.Fields(result.String())
}

// PascalCase converts string to pascal case.
func PascalCase(str string) string {
	items := Words(str)
	for i := range items {
		items[i] = Capitalize(items[i])
	}
	return strings.Join(items, "")
}

// CamelCase converts string to camel case.
func CamelCase(str string) string {
	items := Words(str)
	for i, item := range items {
		item = strings.ToLower(item)
		if i > 0 {
			item = Capitalize(item)
		}
		items[i] = item
	}
	return strings.Join(items, "")
}

// KebabCase converts string to kebab case.
func KebabCase(str string) string {
	items := Words(str)
	for i := range items {
		items[i] = strings.ToLower(items[i])
	}
	return strings.Join(items, "-")
}

// SnakeCase converts string to snake case.
func SnakeCase(str string) string {
	items := Words(str)
	for i := range items {
		items[i] = strings.ToLower(items[i])
	}
	return strings.Join(items, "_")
}

//

// FilterBy performs in place filtering of a rune slice based on a predicate
func FilterBy[T constraints.Comparable](slice []T, keep FilterFn[T]) []T {
	if len(slice) == 0 {
		return slice
	}

	n := 0
	for _, v := range slice {
		if keep(v) {
			slice[n] = v
			n++
		}
	}
	return slice[:n]
}

// Count returns the number of elements in a slice.
func Count[T constraints.Comparable](slice []T) int {
	return len(slice)
}

// CountBy returns the number of elements in a slice that satisfy a given predicate.
func CountBy[T constraints.Comparable](slice []T, keep FilterFn[T]) int {
	return len(FilterBy(slice, keep))
}

// ToMap converts a slice to a map where the keys are the elements of the slice and the values are their counts.
func ToMap[T constraints.Comparable](slice []T) map[T]int {
	trueFilter := func(x T) bool {
		return true
	}
	return ToMapBy(slice, trueFilter)
}

// ToMapBy converts a slice to a map where the keys are the elements of the slice that satisfy a given predicate and the values are their counts.
func ToMapBy[T constraints.Comparable](slice []T, keep FilterFn[T]) map[T]int {
	result := make(map[T]int)
	for _, v := range slice {
		if keep(v) {
			result[v]++
		}
	}
	return result
}

// In checks if a value exists in a slice.
func In[T constraints.Comparable](value T, slice []T) bool {
	if len(slice) == 0 {
		return false
	}
	for k := range slice {
		if slice[k] == value {
			return true
		}
	}
	return false
}

// NotIn checks if a value does not exist in a slice.
func NotIn[T constraints.Comparable](value T, slice []T) bool {
	return !In(value, slice)
}

// Every check if all the values exist in the slice.
func Every[T constraints.Comparable](values []T, slice []T) bool {
	if len(slice) == 0 {
		return false
	}
	for _, v := range values {
		if NotIn(v, slice) {
			return false
		}
	}
	return true
}

// Some checks if any of the values exist in the slice.
func Some[T constraints.Comparable](values []T, slice []T) bool {
	if len(slice) == 0 {
		return false
	}
	for _, v := range values {
		if In(v, slice) {
			return true
		}
	}
	return false
}

// None checks if none of the values exist in the slice.
func None[T constraints.Comparable](values []T, slice []T) bool {
	return !Some(values, slice)
}

// Union returns a slice containing unique elements from both slices.
func Union[T constraints.Comparable](a []T, b []T) []T {
	length := len(a) + len(b)
	result := make([]T, 0, length)
	seen := make(map[T]struct{}, length)

	lists := [][]T{a, b}
	for i := range lists {
		for j := range lists[i] {
			if _, ok := seen[lists[i][j]]; !ok {
				seen[lists[i][j]] = struct{}{}
				result = append(result, lists[i][j])
			}
		}
	}

	return result
}

// Intersection returns a slice containing elements that are present in both slices.
func Intersection[T constraints.Comparable](a []T, b []T) []T {
	var result []T
	seen := map[T]struct{}{}

	for i := range a {
		seen[a[i]] = struct{}{}
	}

	for i := range b {
		if _, ok := seen[b[i]]; ok {
			result = append(result, b[i])
		}
	}

	return result
}

// Difference returns two slices: the first contains elements in a that are not in b, and the second contains elements in b that are not in a.
func Difference[T constraints.Comparable](a []T, b []T) ([]T, []T) {
	var left []T
	var right []T

	seenLeft := map[T]struct{}{}
	seenRight := map[T]struct{}{}

	for i := range a {
		seenLeft[a[i]] = struct{}{}
	}

	for i := range b {
		seenRight[b[i]] = struct{}{}
	}

	for i := range a {
		if _, ok := seenRight[a[i]]; !ok {
			left = append(left, a[i])
		}
	}

	for i := range b {
		if _, ok := seenLeft[b[i]]; !ok {
			right = append(right, b[i])
		}
	}

	return left, right
}

// Map applies a function to each element of a slice and returns a new slice containing the results.
func Map[T, U constraints.Comparable](slice []T, fn ItemMapFn[T, U]) []U {
	trueFilter := func(x T) bool {
		return true
	}
	return MapBy(slice, fn, trueFilter)
}

// MapBy applies a function to each element of a slice and returns a new slice containing the results.
func MapBy[T, U constraints.Comparable](slice []T, fn ItemMapFn[T, U], keep FilterFn[T]) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		if keep(v) {
			result[i] = fn(v)
		}
	}
	return result
}

// Copy creates a copy of a slice.
func Copy[T constraints.Comparable](slice []T) []T {
	if Nil(slice) {
		return nil
	}

	s := make([]T, len(slice))
	copy(s, slice)
	return s
}

// Reverse performs in place reversal of a slice
func Reverse[T constraints.Comparable](slice []T) []T {
	slices.Reverse(slice)
	return slice
}

// Shuffle shuffles (in place) a slice
func Shuffle[T constraints.Comparable](slice []T) []T {
	if len(slice) <= 1 {
		return slice
	}

	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})

	return slice
}

// Deduplicate performs order preserving, in place deduplication of a slice
func Deduplicate[T constraints.Comparable](slice []T) []T {
	if len(slice) < 2 {
		return slice
	}

	seen := make(map[T]struct{})

	j := 0
	for k := range slice {
		if _, ok := seen[slice[k]]; ok {
			continue
		}
		seen[slice[k]] = struct{}{}
		slice[j] = slice[k]
		j++
	}

	return slice[:j]
}

// Sort sorts a slice in place using the slices package.
func Sort[T constraints.Ordenable](slice []T) []T {
	slices.Sort(slice)
	return slice
}

// Chunk returns an array of slices split into groups the length of size.
func Chunk[T constraints.Comparable](slice []T, size int) [][]T {
	var chunks [][]T
	if size <= 0 {
		return chunks
	}

	chunksNum := len(slice) / size
	if len(slice)%size != 0 {
		chunksNum += 1
	}

	result := make([][]T, 0, chunksNum)

	for i := 0; i < chunksNum; i++ {
		last := (i + 1) * size
		if last > len(slice) {
			last = len(slice)
		}
		result = append(result, slice[i*size:last:last])
	}

	return result
}

// Delete deletes the element at the specified index from a rune slice
func Delete[T constraints.Comparable](idx int, slice []T) []T {
	if len(slice) == 0 {
		return pointer.Zero[[]T]()
	}
	if idx < 0 || idx > len(slice)-1 {
		return pointer.Zero[[]T]()
	}

	return append(slice[:idx], slice[idx+1:]...)
}

// DeleteRange deletes the elements between from and to index (inclusive) from a rune slice
func DeleteRange[T constraints.Comparable](start, end int, slice []T) []T {
	if len(slice) == 0 {
		return pointer.Zero[[]T]()
	}

	if start < 0 || start > len(slice)-1 {
		return pointer.Zero[[]T]()
	}

	if end < 0 || end > len(slice)-1 {
		return pointer.Zero[[]T]()
	}

	if start > end {
		return pointer.Zero[[]T]()
	}

	return append(slice[:start], slice[end+1:]...)
}

// Push appends a value to the end of a slice and returns the modified slice.
func Push[T constraints.Comparable](slice []T, value T) []T {
	return append(slice, value)
}

// Pop removes the last element from a slice and returns it along with the modified slice.
func Pop[T constraints.Comparable](slice []T) (T, []T) {
	if len(slice) == 0 {
		return pointer.Zero[T](), pointer.Zero[[]T]()
	}

	return slice[len(slice)-1], slice[:len(slice)-1]
}

// Max returns the maximum value of a byte slice or an error in case of a nil or empty slice
func Max[T constraints.Ordenable](slice []T) T {
	if len(slice) == 0 {
		return pointer.Zero[T]()
	}

	value := slice[0]
	for k := 1; k < len(slice); k++ {
		if slice[k] > value {
			value = slice[k]
		}
	}

	return value
}

// Min returns the minimum value of a byte slice or an error in case of a nil or empty slice
func Min[T constraints.Ordenable](slice []T) T {
	if len(slice) == 0 {
		return pointer.Zero[T]()
	}

	value := slice[0]
	for k := 1; k < len(slice); k++ {
		if slice[k] < value {
			value = slice[k]
		}
	}

	return value
}

//

// HasKey checks if a key exists in a map.
func HasKey[K constraints.Comparable, V any](key K, keyValues map[K]V) bool {
	_, ok := keyValues[key]
	return ok
}

// Keys returns a slice of keys from the provided maps.
func Keys[K constraints.Comparable, V any](keyValues ...map[K]V) []K {
	var size int
	for i := range keyValues {
		size += len(keyValues[i])
	}

	result := make([]K, 0, size)

	for i := range keyValues {
		for k := range keyValues[i] {
			result = append(result, k)
		}
	}

	return result
}

// UniqueKeys returns a slice of unique keys from the provided maps.
func UniqueKeys[K constraints.Comparable, V any](keyValues ...map[K]V) []K {
	var size int
	for i := range keyValues {
		size += len(keyValues[i])
	}

	seen := make(map[K]struct{}, size)
	result := make([]K, 0)

	for i := range keyValues {
		for k := range keyValues[i] {
			if _, exists := seen[k]; exists {
				continue
			}
			seen[k] = struct{}{}
			result = append(result, k)
		}
	}

	return result
}

// Values returns a slice of values from the provided maps.
func Values[K constraints.Comparable, V any](keyValues ...map[K]V) []V {
	var size int
	for i := range keyValues {
		size += len(keyValues[i])
	}

	result := make([]V, 0, size)

	for i := range keyValues {
		for k := range keyValues[i] {
			result = append(result, keyValues[i][k])
		}
	}

	return result
}

// UniqueValues returns a slice of unique values from the provided maps.
func UniqueValues[K constraints.Comparable, V constraints.Comparable](keyValues ...map[K]V) []V {
	var size int
	for i := range keyValues {
		size += len(keyValues[i])
	}

	seen := make(map[V]struct{}, size)
	result := make([]V, 0)

	for i := range keyValues {
		for k := range keyValues[i] {
			val := keyValues[i][k]
			if _, exists := seen[val]; exists {
				continue
			}
			seen[val] = struct{}{}
			result = append(result, val)
		}
	}

	return result
}

// PickByKeys returns a map containing only the key-value pairs from the original map where the keys are in the provided keys slice.
func PickByKeys[K constraints.Comparable, V any](keyValues map[K]V, keys []K) map[K]V {
	r := make(map[K]V)
	for i := range keys {
		if v, ok := keyValues[keys[i]]; ok {
			r[keys[i]] = v
		}
	}
	return r
}

// PickByValues returns a map containing only the key-value pairs from the original map where the values are in the provided values slice.
func PickByValues[K constraints.Comparable, V constraints.Comparable](keyValues map[K]V, values []V) map[K]V {
	r := make(map[K]V)
	for k := range keyValues {
		if In(keyValues[k], values) {
			r[k] = keyValues[k]
		}
	}
	return r
}

// OmitByKeys returns a map containing only the key-value pairs from the original map where the keys are not in the provided keys slice.
func OmitByKeys[K constraints.Comparable, V any](keyValues map[K]V, keys []K) map[K]V {
	r := make(map[K]V)
	for k := range keyValues {
		r[k] = keyValues[k]
	}
	for i := range keys {
		delete(r, keys[i])
	}
	return r
}

// OmitByValues returns a map containing only the key-value pairs from the original map where the values are not in the provided values slice.
func OmitByValues[K constraints.Comparable, V constraints.Comparable](keyValues map[K]V, values []V) map[K]V {
	r := make(map[K]V)
	for k := range keyValues {
		if NotIn(keyValues[k], values) {
			r[k] = keyValues[k]
		}
	}
	return r
}
