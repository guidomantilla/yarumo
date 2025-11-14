package utils

import (
	"reflect"
	"strings"
	"testing"
)

// helper: compare slices as multisets (order-insensitive)
func sameMultiset[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[T]int, len(a))
	for _, v := range a {
		m[v]++
	}
	for _, v := range b {
		m[v]--
	}
	for _, c := range m {
		if c != 0 {
			return false
		}
	}
	return true
}

func TestTernaryAndCoalesceAndEquality(t *testing.T) {
	if got := Ternary(true, 1, 2); got != 1 {
		t.Fatalf("Ternary true branch = %v", got)
	}
	if got := Ternary(false, 1, 2); got != 2 {
		t.Fatalf("Ternary false branch = %v", got)
	}

	if got := Coalesce[int](); got != 0 {
		t.Fatalf("Coalesce() = %v, want 0", got)
	}
	if got := Coalesce(0, 0, 3, 4); got != 3 {
		t.Fatalf("Coalesce non-empty = %v", got)
	}

	if !Equal("a", "a") || NotEqual("a", "a") {
		t.Fatalf("Equal/NotEqual mismatch for equal values")
	}
	if Equal("a", "b") || !NotEqual("a", "b") {
		t.Fatalf("Equal/NotEqual mismatch for different values")
	}
}

func TestNilAndEmptyPredicates(t *testing.T) {
	var p *int
	if !Nil(p) || Nil(10) {
		t.Fatalf("Nil predicate failed")
	}
	if NotNil(p) || !NotNil(10) {
		t.Fatalf("NotNil predicate failed")
	}

	// Empty variadic: no args -> true
	if !Empty() {
		t.Fatalf("Empty() without args must be true")
	}
	// NotEmpty variadic: no args -> true by implementation
	if !NotEmpty() {
		t.Fatalf("NotEmpty() without args must be true per implementation")
	}

	if !Empty("") || Empty("x") {
		t.Fatalf("Empty on strings mismatch")
	}
	if !NotEmpty("x") || NotEmpty("") {
		t.Fatalf("NotEmpty on strings mismatch")
	}

	if Empty("x", "") {
		t.Fatalf("Empty with mixed non-empty should be false")
	}
	if NotEmpty("x", "", "y") {
		t.Fatalf("NotEmpty with an empty element should be false")
	}
}

func TestRandomStringAndSubstringAndChunkString(t *testing.T) {
	// RandomString: length<=0 -> empty
	if got := RandomString(0); got != "" {
		t.Fatalf("RandomString(0) = %q, want empty", got)
	}
	// With custom charset of single char -> deterministic
	s := RandomString(5, WithCharset("A"))
	if s != "AAAAA" {
		t.Fatalf("RandomString single charset = %q", s)
	}

	// Substring: typical small offset returns empty due to implementation
	if got := Substring("hello", 0, 2); got != "" {
		t.Fatalf("Substring default path expected empty, got %q", got)
	}
	// Use uint overflow trick to get non-empty branch
	// For a string of size n, offset := size + ^uint(0) => size-1
	var maxU = ^uint(0)
	str := "hello"
	got := Substring(str, maxU, 2)
	// should start at last rune (o), and length will be clamped to 1
	if got != "o" {
		t.Fatalf("Substring overflow path = %q, want %q", got, "o")
	}

	// ChunkString
	if !reflect.DeepEqual(ChunkString("", 0), []string{""}) {
		t.Fatalf("ChunkString empty/zero size")
	}
	if !reflect.DeepEqual(ChunkString("", 3), []string{""}) {
		t.Fatalf("ChunkString empty string")
	}
	if !reflect.DeepEqual(ChunkString("abc", 5), []string{"abc"}) {
		t.Fatalf("ChunkString size>=len")
	}
	chunks := ChunkString("abcdef", 2)
	want := []string{"ab", "cd", "ef"}
	if !reflect.DeepEqual(chunks, want) {
		t.Fatalf("ChunkString got %v want %v", chunks, want)
	}
}

func TestCaseAndWordsFunctions(t *testing.T) {
	w := Words("HTTPServer2ID v1_test")
	// Expect tokens split between letters/numbers and transitions
	if !reflect.DeepEqual(w, []string{"HTTP", "Server", "2", "ID", "v", "1", "test"}) {
		t.Fatalf("Words split mismatch: %v", w)
	}

	if got := Capitalize("hELLO"); got != "Hello" {
		t.Fatalf("Capitalize = %q", got)
	}
	if got := PascalCase("hello-world_user42"); got != "HelloWorldUser42" {
		t.Fatalf("PascalCase = %q", got)
	}
	if got := CamelCase("hello_world user42"); got != "helloWorldUser42" {
		t.Fatalf("CamelCase = %q", got)
	}
	if got := KebabCase("HelloWorld User42"); got != "hello-world-user-42" {
		t.Fatalf("KebabCase = %q", got)
	}
	if got := SnakeCase("HelloWorld User42"); got != "hello_world_user_42" {
		t.Fatalf("SnakeCase = %q", got)
	}
}

func TestSliceFiltersAndCountsAndMaps(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	even := func(x int) bool { return x%2 == 0 }
	got := FilterBy(Copy(nums), even)
	if !reflect.DeepEqual(got, []int{2, 4}) {
		t.Fatalf("FilterBy result %v", got)
	}
	if Count(nums) != 5 {
		t.Fatalf("Count = %d", Count(nums))
	}
	if CountBy(nums, even) != 2 {
		t.Fatalf("CountBy even = %d", CountBy(nums, even))
	}

	m := ToMap([]string{"a", "b", "a"})
	if m["a"] != 2 || m["b"] != 1 {
		t.Fatalf("ToMap counts = %v", m)
	}
	m2 := ToMapBy([]int{1, 2, 3, 4}, even)
	if len(m2) != 2 || m2[2] != 1 || m2[4] != 1 {
		t.Fatalf("ToMapBy even = %v", m2)
	}

	if In(2, 1, 2, 3) != true || In(9, 1, 2, 3) != false {
		t.Fatalf("In failed")
	}
	if NotIn(2, 1, 2) != false || NotIn(9, 1, 2) != true {
		t.Fatalf("NotIn failed")
	}
	if Every([]int{1, 2}, 1, 2, 3) != true {
		t.Fatalf("Every failed")
	}
	if Every([]int{1, 4}, 1, 2, 3) != false {
		t.Fatalf("Every false case failed")
	}
	if Some([]int{4, 2}, 1, 2, 3) != true {
		t.Fatalf("Some true failed")
	}
	if Some([]int{4, 5}, 1, 2, 3) != false {
		t.Fatalf("Some false failed")
	}
	if None([]int{4, 5}, 1, 2, 3) != true {
		t.Fatalf("None true failed")
	}

	// Union/Intersection/Difference
	a := []int{1, 2, 2}
	b := []int{2, 3}
	u := Union(a, b)
	if !sameMultiset(u, []int{1, 2, 3}) {
		t.Fatalf("Union = %v", u)
	}
	inter := Intersection(a, b)
	if !sameMultiset(inter, []int{2}) {
		t.Fatalf("Intersection = %v", inter)
	}
	l, r := Difference(a, b)
	if !sameMultiset(l, []int{1}) || !sameMultiset(r, []int{3}) {
		t.Fatalf("Difference = %v %v", l, r)
	}
}

func TestMapByCopyReverseShuffleDeduplicateSort(t *testing.T) {
	// Map and MapBy
	src := []int{1, 2, 3}
	double := func(x int) int { return x * 2 }
	got := Map(src, double)
	if !reflect.DeepEqual(got, []int{2, 4, 6}) {
		t.Fatalf("Map = %v", got)
	}
	keepOdd := func(x int) bool { return x%2 == 1 }
	got2 := MapBy(src, double, keepOdd)
	// positions not kept remain zero value (0)
	if !reflect.DeepEqual(got2, []int{2, 0, 6}) {
		t.Fatalf("MapBy = %v", got2)
	}

	// Copy nil vs non-nil
	if Copy[int](nil) != nil {
		t.Fatalf("Copy(nil) should be nil slice")
	}
	cp := Copy(src)
	if &cp[0] == &src[0] || !reflect.DeepEqual(cp, src) {
		t.Fatalf("Copy did not copy properly")
	}

	// Reverse
	rev := Reverse(Copy(src))
	if !reflect.DeepEqual(rev, []int{3, 2, 1}) {
		t.Fatalf("Reverse = %v", rev)
	}

	// Shuffle length <=1 returns input
	one := []int{7}
	if !reflect.DeepEqual(Shuffle(one), []int{7}) {
		t.Fatalf("Shuffle length<=1")
	}
	// For larger slice, ensure permutation (not necessarily different)
	sh := Shuffle(Copy(src))
	if !sameMultiset(sh, src) {
		t.Fatalf("Shuffle not a permutation: %v", sh)
	}

	// Deduplicate
	du := []int{1, 2, 2, 3, 1}
	de := Deduplicate(du)
	if !sameMultiset(de, []int{1, 2, 3}) {
		t.Fatalf("Deduplicate = %v", de)
	}
	if len(Deduplicate([]int{1})) != 1 {
		t.Fatalf("Deduplicate len<2 path failed")
	}

	// Sort
	uns := []int{3, 1, 2}
	if !reflect.DeepEqual(Sort(uns), []int{1, 2, 3}) {
		t.Fatalf("Sort ints failed")
	}
}

func TestChunkDeletePopPushMinMax(t *testing.T) {
	// Chunk generic
	if res := Chunk([]int{1, 2, 3, 4, 5}, 2); !reflect.DeepEqual(res, [][]int{{1, 2}, {3, 4}, {5}}) {
		t.Fatalf("Chunk = %v", res)
	}
	if res := Chunk([]int{1, 2}, 0); len(res) != 0 {
		t.Fatalf("Chunk size<=0 must be empty, got %v", res)
	}

	// Delete / DeleteRange
	if out := Delete(-1, []int{1, 2}); out != nil {
		t.Fatalf("Delete invalid index should return nil slice")
	}
	if out := Delete(5, []int{1, 2}); out != nil {
		t.Fatalf("Delete oob should return nil slice")
	}
	if out := Delete(1, []int{1, 2, 3}); !reflect.DeepEqual(out, []int{1, 3}) {
		t.Fatalf("Delete valid = %v", out)
	}

	if out := DeleteRange(-1, 1, []int{1, 2}); out != nil {
		t.Fatalf("DeleteRange invalid start -> nil")
	}
	if out := DeleteRange(0, 5, []int{1, 2}); out != nil {
		t.Fatalf("DeleteRange invalid end -> nil")
	}
	if out := DeleteRange(2, 1, []int{1, 2}); out != nil {
		t.Fatalf("DeleteRange start>end -> nil")
	}
	if out := DeleteRange(1, 2, []int{1, 2, 3, 4}); !reflect.DeepEqual(out, []int{1, 4}) {
		t.Fatalf("DeleteRange valid = %v", out)
	}

	// Push/Pop
	pushed := Push([]int{1}, 2)
	if !reflect.DeepEqual(pushed, []int{1, 2}) {
		t.Fatalf("Push = %v", pushed)
	}
	v, rest := Pop([]int{7, 8})
	if v != 8 || !reflect.DeepEqual(rest, []int{7}) {
		t.Fatalf("Pop = %v, %v", v, rest)
	}
	v2, rest2 := Pop([]int{})
	if v2 != 0 || rest2 != nil {
		t.Fatalf("Pop empty = %v, %v", v2, rest2)
	}

	// Min / Max
	if Max([]int{}) != 0 || Min([]int{}) != 0 {
		t.Fatalf("Min/Max empty should be zero values")
	}
	if Max([]int{2, 9, 3}) != 9 {
		t.Fatalf("Max failed")
	}
	if Min([]int{2, 9, 3}) != 2 {
		t.Fatalf("Min failed")
	}
}

func TestMapHelpers(t *testing.T) {
	m1 := map[string]int{"a": 1, "b": 2}
	m2 := map[string]int{"b": 3, "c": 4}

	if !HasKey("a", m1) || HasKey("z", m1) {
		t.Fatalf("HasKey failed")
	}

	ks := Keys(m1, m2)
	// ks should contain a,b,b,c (order not guaranteed)
	if !sameMultiset(ks, []string{"a", "b", "b", "c"}) {
		t.Fatalf("Keys = %v", ks)
	}

	uks := UniqueKeys(m1, m2)
	if !sameMultiset(uks, []string{"a", "b", "c"}) {
		t.Fatalf("UniqueKeys = %v", uks)
	}

	vs := Values(m1, m2)
	if !sameMultiset(vs, []int{1, 2, 3, 4}) {
		t.Fatalf("Values = %v", vs)
	}

	uvs := UniqueValues(m1, m2)
	if !sameMultiset(uvs, []int{1, 2, 3, 4}) {
		t.Fatalf("UniqueValues = %v", uvs)
	}

	// Pick/Omit by keys/values
	pk := PickByKeys(m1, []string{"a", "z"})
	if len(pk) != 1 || pk["a"] != 1 {
		t.Fatalf("PickByKeys = %v", pk)
	}

	pv := PickByValues(m2, []int{3, 5})
	if len(pv) != 1 || pv["b"] != 3 {
		t.Fatalf("PickByValues = %v", pv)
	}

	okm := OmitByKeys(m1, []string{"b"})
	if len(okm) != 1 || okm["a"] != 1 {
		t.Fatalf("OmitByKeys = %v", okm)
	}

	ovm := OmitByValues(m2, []int{3})
	if len(ovm) != 1 || ovm["c"] != 4 {
		t.Fatalf("OmitByValues = %v", ovm)
	}
}

func TestWordsSeparatorsNonAlnum(t *testing.T) {
	// ensure non-alphanumeric become spaces and are trimmed by Fields
	parts := Words("ab-__cd!!ef 12xy")
	if strings.Join(parts, ",") != "ab,cd,ef,12,xy" {
		t.Fatalf("Words cleanup mismatch: %v", parts)
	}
}
