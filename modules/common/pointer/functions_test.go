package pointer

import (
	"reflect"
	"testing"
)

// helper types for struct checks
type demoStruct struct{ A int }

// helper interface and implementation used to exercise reflect.Interface path
type demoInterface interface{ Read() int }
type impl struct{}

func (impl) Read() int { return 1 }

func mustPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic, got none")
		}
	}()
	f()
}

func TestIsNilAndIsNotNil(t *testing.T) {
	// nil interface
	if !IsNil(nil) {
		t.Fatalf("IsNil(nil) = false, want true")
	}
	if IsNotNil(nil) {
		t.Fatalf("IsNotNil(nil) = true, want false")
	}

	// typed nil pointer
	var p *int
	if !IsNil(p) {
		t.Fatalf("IsNil((*int)(nil)) = false, want true")
	}
	// non-nil pointer
	x := 5
	p = &x
	if IsNil(p) {
		t.Fatalf("IsNil(&x) = true, want false")
	}

	// nil and non-nil slice
	var s []int
	if !IsNil(s) {
		t.Fatalf("IsNil(nil slice) = false, want true")
	}
	s = []int{}
	if IsNil(s) {
		t.Fatalf("IsNil(non-nil slice) = true, want false")
	}

	// nil and non-nil map
	var m map[string]int
	if !IsNil(m) {
		t.Fatalf("IsNil(nil map) = false, want true")
	}
	m = map[string]int{}
	if IsNil(m) {
		t.Fatalf("IsNil(non-nil map) = true, want false")
	}

	// nil and non-nil chan
	var ch chan int
	if !IsNil(ch) {
		t.Fatalf("IsNil(nil chan) = false, want true")
	}
	ch = make(chan int)
	if IsNil(ch) {
		t.Fatalf("IsNil(non-nil chan) = true, want false")
	}

	// function kind (nil and non-nil)
	var fn func()
	if !IsNil(fn) {
		t.Fatalf("IsNil(nil func) = false, want true")
	}
	fn = func() {}
	if IsNil(fn) {
		t.Fatalf("IsNil(non-nil func) = true, want false")
	}

	// arrays fall into Array kind in IsNil's switch and calling IsNil() on array panics
	mustPanic(t, func() {
		_ = IsNil([1]int{1})
	})

	// interface holding a typed nil pointer should be considered nil by IsNil
	var pn *int = nil
	var ai any = pn
	if !IsNil(ai) {
		t.Fatalf("IsNil(interface(nil pointer)) = false, want true")
	}
}

func TestIsEmptyAndIsNotEmpty(t *testing.T) {
	// nil cases
	if !IsEmpty(nil) || IsNotEmpty(nil) {
		t.Fatalf("nil emptiness mismatch")
	}

	// nil pointer considered empty
	var p *int
	if !IsEmpty(p) {
		t.Fatalf("IsEmpty(nil pointer) = false, want true")
	}
	// non-nil pointer should dereference and evaluate underlying emptiness
	sp := new(string)
	*sp = ""
	if !IsEmpty(sp) {
		t.Fatalf("IsEmpty(pointer to empty string) = false, want true")
	}
	*sp = "x"
	if IsEmpty(sp) {
		t.Fatalf("IsEmpty(pointer to non-empty string) = true, want false")
	}

	// empty string, slice, map, chan
	if !IsEmpty("") {
		t.Fatalf("IsEmpty(%q) = false, want true", "")
	}
	if !IsEmpty([]int{}) {
		t.Fatalf("IsEmpty(empty slice) = false, want true")
	}
	if !IsEmpty(map[string]int{}) {
		t.Fatalf("IsEmpty(empty map) = false, want true")
	}
	c := make(chan int)
	if !IsEmpty(c) { // unbuffered (len==0)
		t.Fatalf("IsEmpty(chan) = false, want true when len==0")
	}
	// buffered with element -> non-empty
	cb := make(chan int, 1)
	cb <- 1
	if IsEmpty(cb) {
		t.Fatalf("IsEmpty(buffered non-empty chan) = true, want false")
	}

	// non-empty counterparts
	if IsEmpty("x") || !IsNotEmpty("x") {
		t.Fatalf("non-empty string mismatch")
	}
	if IsEmpty([]int{1}) {
		t.Fatalf("IsEmpty(non-empty slice) = true, want false")
	}
	if IsEmpty(map[string]int{"a": 1}) {
		t.Fatalf("IsEmpty(non-empty map) = true, want false")
	}

	// arrays are also handled in IsEmpty
	var arr0 [0]int
	if !IsEmpty(arr0) {
		t.Fatalf("IsEmpty(empty array) = false, want true")
	}
	arr1 := [1]int{1}
	if IsEmpty(arr1) {
		t.Fatalf("IsEmpty(non-empty array) = true, want false")
	}

	// Numeric zero is NOT considered empty by the current implementation (default -> IsZero(reflect.Value))
	if IsEmpty(0) {
		t.Fatalf("IsEmpty(0) = true, want false per implementation")
	}

	// interface kinds: interface holding nil should be empty
	var ai any
	if !IsEmpty(ai) {
		t.Fatalf("IsEmpty(interface nil) = false, want true")
	}
	// interface holding non-nil value should not be empty (Interface branch -> IsNil false)
	ai = 123
	if IsEmpty(ai) {
		t.Fatalf("IsEmpty(interface non-nil) = true, want false")
	}

	// Cover reflect.Interface branch explicitly using a typed interface variable
	var di demoInterface // nil underlying
	if !IsEmpty(di) {    // Kind == Interface and IsNil() == true
		t.Fatalf("IsEmpty(typed interface nil) = false, want true")
	}
	// Non-nil interface value
	var dni demoInterface = impl{}
	if IsEmpty(dni) {
		t.Fatalf("IsEmpty(typed interface non-nil) = true, want false")
	}
}

func TestIsPointerAndNotPointer(t *testing.T) {
	if IsPointer(nil) {
		t.Fatalf("IsPointer(nil) = true, want false")
	}
	v := 10
	if !IsPointer(&v) {
		t.Fatalf("IsPointer(&v) = false, want true")
	}
	if IsPointer(v) {
		t.Fatalf("IsPointer(v) = true, want false")
	}
	if IsNotPointer(&v) {
		t.Fatalf("IsNotPointer(&v) = true, want false")
	}
}

func TestZeroAndComparisons(t *testing.T) {
	if Zero[int]() != 0 {
		t.Fatalf("Zero[int]() != 0")
	}
	if !IsZero(0) || IsZero(1) {
		t.Fatalf("IsZero for ints mismatch")
	}
	if !IsNotZero(1) || IsNotZero(0) {
		t.Fatalf("IsNotZero for ints mismatch")
	}
	if Zero[string]() != "" {
		t.Fatalf("Zero[string]() != \"\"")
	}
	if !IsZero("") || IsZero("x") {
		t.Fatalf("IsZero for strings mismatch")
	}
}

func TestToPtrAndFromPtr(t *testing.T) {
	p := ToPtr(42)
	if p == nil || *p != 42 {
		t.Fatalf("ToPtr failed: %v", p)
	}
	if FromPtr(p) != 42 {
		t.Fatalf("FromPtr non-nil mismatch")
	}
	var q *int
	if FromPtr(q) != Zero[int]() {
		t.Fatalf("FromPtr(nil) should return zero value")
	}
}

func TestToSlicePtrAndFromSlicePtr(t *testing.T) {
	src := []int{1, 2, 3}
	ps := ToSlicePtr(src)
	if len(ps) != len(src) {
		t.Fatalf("ToSlicePtr length mismatch: %d vs %d", len(ps), len(src))
	}
	// pointers should reference the original underlying elements
	*ps[0] = 10
	if src[0] != 10 {
		t.Fatalf("ToSlicePtr pointers do not reference original elements")
	}

	// FromSlicePtr with nil element should yield zero for that position
	a, b := 7, 9
	withNil := []*int{&a, nil, &b}
	vals := FromSlicePtr(withNil)
	if !reflect.DeepEqual(vals, []int{7, 0, 9}) {
		t.Fatalf("FromSlicePtr with nil element mismatch: %v", vals)
	}
}

func TestIsType(t *testing.T) {
	if IsType(nil, "int") {
		t.Fatalf("IsType(nil, \"int\") = true, want false")
	}
	if !IsType(5, "int") {
		t.Fatalf("IsType(5, \"int\") = false, want true")
	}
	ds := &demoStruct{}
	// For pointers, IsType should dereference and compare underlying type name
	if !IsType(ds, "pointer.demoStruct") {
		t.Fatalf("IsType(&demoStruct, \"pointer.demoStruct\") = false, want true")
	}
	// Pointer to basic type should match underlying type name
	iv := 3
	if !IsType(&iv, "int") {
		t.Fatalf("IsType(&int, \"int\") = false, want true")
	}
	// Negative case
	if IsType("x", "int") {
		t.Fatalf("IsType(\"x\", \"int\") = true, want false")
	}
}

func TestKindCheckers(t *testing.T) {
	// IsStruct
	if IsStruct(nil) {
		t.Fatalf("IsStruct(nil) = true, want false")
	}
	if !IsStruct(demoStruct{}) {
		t.Fatalf("IsStruct(struct) = false, want true")
	}
	if !IsStruct(&demoStruct{}) {
		t.Fatalf("IsStruct(&struct) = false, want true")
	}
	// pointer to non-struct should be false
	vi := 1
	if IsStruct(&vi) {
		t.Fatalf("IsStruct(&int) = true, want false")
	}
	if IsStruct(map[string]int{}) {
		t.Fatalf("IsStruct(map) = true, want false")
	}
	// nil pointer to struct should safely return false (Elem() -> zero Value -> Kind Invalid)
	var ps *demoStruct
	if IsStruct(ps) {
		t.Fatalf("IsStruct((*demoStruct)(nil)) = true, want false")
	}

	// IsChan
	if IsChan(nil) {
		t.Fatalf("IsChan(nil) = true, want false")
	}
	ch := make(chan int)
	if !IsChan(ch) {
		t.Fatalf("IsChan(chan) = false, want true")
	}
	// pointer to chan
	pch := &ch
	if !IsChan(pch) {
		t.Fatalf("IsChan(&chan) = false, want true")
	}
	// pointer to non-chan should be false
	if IsChan(&vi) {
		t.Fatalf("IsChan(&int) = true, want false")
	}
	if IsChan(123) {
		t.Fatalf("IsChan(int) = true, want false")
	}
	// nil pointer to chan should safely return false
	var pnilch *chan int
	if IsChan(pnilch) {
		t.Fatalf("IsChan((*chan int)(nil)) = true, want false")
	}

	// IsSlice (also counts arrays)
	if IsSlice(nil) {
		t.Fatalf("IsSlice(nil) = true, want false")
	}
	if !IsSlice([]int{1, 2}) {
		t.Fatalf("IsSlice(slice) = false, want true")
	}
	arr := [2]int{1, 2}
	if !IsSlice(arr) {
		t.Fatalf("IsSlice(array) = false, want true")
	}
	s := []int{1}
	if !IsSlice(&s) {
		t.Fatalf("IsSlice(&slice) = false, want true")
	}
	// pointer to array should be true
	if !IsSlice(&arr) {
		t.Fatalf("IsSlice(&array) = false, want true")
	}
	if IsSlice(123) {
		t.Fatalf("IsSlice(int) = true, want false")
	}
	// nil pointer to slice/array should safely return false
	var pnilSlice *[]int
	if IsSlice(pnilSlice) {
		t.Fatalf("IsSlice((*[]int)(nil)) = true, want false")
	}
	var pnilArray *[2]int
	if IsSlice(pnilArray) {
		t.Fatalf("IsSlice((*[2]int)(nil)) = true, want false")
	}

	// IsMap
	if IsMap(nil) {
		t.Fatalf("IsMap(nil) = true, want false")
	}
	if !IsMap(map[string]int{"a": 1}) {
		t.Fatalf("IsMap(map) = false, want true")
	}
	mm := map[string]int{}
	if !IsMap(&mm) {
		t.Fatalf("IsMap(&map) = false, want true")
	}
	// pointer to non-map should be false
	if IsMap(&vi) {
		t.Fatalf("IsMap(&int) = true, want false")
	}
	if IsMap("nope") {
		t.Fatalf("IsMap(string) = true, want false")
	}
	// nil pointer to map should safely return false
	var pnilMap *map[string]int
	if IsMap(pnilMap) {
		t.Fatalf("IsMap((*map[string]int)(nil)) = true, want false")
	}
}
