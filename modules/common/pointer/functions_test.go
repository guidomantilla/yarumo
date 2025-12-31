package pointer

import "testing"

type sampleStruct struct{ A int }

func TestIsNilAndIsNotNil(t *testing.T) {
	// direct nil
	if !IsNil(nil) || IsNotNil(nil) {
		t.Fatal("nil check failed")
	}

	// nil pointer
	var p *int
	if !IsNil(p) || IsNotNil(p) {
		t.Fatal("nil pointer check failed")
	}

	// nil slice
	var s []int
	if !IsNil(s) || IsNotNil(s) {
		t.Fatal("nil slice check failed")
	}

	// nil map
	var m map[string]int
	if !IsNil(m) || IsNotNil(m) {
		t.Fatal("nil map check failed")
	}

	// nil chan
	var ch chan int
	if !IsNil(ch) || IsNotNil(ch) {
		t.Fatal("nil chan check failed")
	}

	// non-nil values
	x := 10
	if IsNil(x) || !IsNotNil(x) {
		t.Fatal("non-nil value check failed")
	}

	// nil func
	var fn func()
	if !IsNil(fn) || IsNotNil(fn) {
		t.Fatal("nil func should be nil")
	}
	// non-nil func
	fn = func() {}
	if IsNil(fn) || !IsNotNil(fn) {
		t.Fatal("non-nil func should not be nil")
	}
	// interface holding typed nil pointer
	var (
		px *int = nil
		ai any  = px
	)

	if !IsNil(ai) {
		t.Fatal("typed nil pointer inside interface should be nil")
	}
}

func TestIsEmptyAndIsNotEmpty(t *testing.T) {
	if !IsEmpty(nil) || IsNotEmpty(nil) {
		t.Fatal("nil should be empty")
	}

	// strings
	if !IsEmpty("") || IsNotEmpty("") {
		t.Fatal("empty string should be empty")
	}

	if IsEmpty("a") || !IsNotEmpty("a") {
		t.Fatal("non-empty string should not be empty")
	}

	// zero values
	if !IsEmpty(0) || IsNotEmpty(0) {
		t.Fatal("zero int should be empty")
	}

	if IsEmpty(1) || !IsNotEmpty(1) {
		t.Fatal("non-zero int should not be empty")
	}

	// pointers to zero/non-zero
	z := 0
	nz := 2

	if !IsEmpty(&z) || IsNotEmpty(&z) {
		t.Fatal("pointer to zero should be empty")
	}

	if IsEmpty(&nz) || !IsNotEmpty(&nz) {
		t.Fatal("pointer to non-zero should not be empty")
	}
	// nil pointer
	var np *int
	if !IsEmpty(np) || IsNotEmpty(np) {
		t.Fatal("nil pointer should be empty")
	}

	// slices and maps
	var s []int
	if !IsEmpty(s) || IsNotEmpty(s) {
		t.Fatal("nil slice should be empty")
	}

	s = []int{}
	if !IsEmpty(s) || IsNotEmpty(s) {
		t.Fatal("empty slice should be empty")
	}

	s = []int{1}
	if IsEmpty(s) || !IsNotEmpty(s) {
		t.Fatal("non-empty slice should not be empty")
	}

	var m map[string]int
	if !IsEmpty(m) || IsNotEmpty(m) {
		t.Fatal("nil map should be empty")
	}

	m = map[string]int{}
	if !IsEmpty(m) || IsNotEmpty(m) {
		t.Fatal("empty map should be empty")
	}

	m["a"] = 1
	if IsEmpty(m) || !IsNotEmpty(m) {
		t.Fatal("non-empty map should not be empty")
	}

	// channels
	var ch chan int
	if !IsEmpty(ch) || IsNotEmpty(ch) {
		t.Fatal("nil chan should be empty")
	}

	ch = make(chan int)
	if !IsEmpty(ch) || IsNotEmpty(ch) {
		t.Fatal("non-nil chan with len 0 is considered empty by IsEmpty")
	}
}

func TestPointerChecks(t *testing.T) {
	var p *int
	if IsPointer(p) || !IsNotPointer(p) {
		t.Fatal("nil pointer is considered not a pointer by IsPointer")
	}

	v := 3
	if !IsPointer(&v) || IsNotPointer(&v) {
		t.Fatal("&v should be a pointer")
	}

	if IsPointer(v) || !IsNotPointer(v) {
		t.Fatal("v is not a pointer")
	}
}

func TestZeroComparisons(t *testing.T) {
	if Zero[int]() != 0 {
		t.Fatal("Zero[int] should be 0")
	}

	if !IsZero(0) || IsNotZero(0) {
		t.Fatal("0 should be zero")
	}

	if IsZero(5) || !IsNotZero(5) {
		t.Fatal("5 should not be zero")
	}
}

func TestToFromPtr(t *testing.T) {
	v := 7

	pv := ToPtr(v)
	if pv == nil || *pv != v {
		t.Fatal("ToPtr failed")
	}

	if FromPtr(pv) != v {
		t.Fatal("FromPtr failed with non-nil")
	}

	var p *int
	if FromPtr(p) != 0 {
		t.Fatal("FromPtr nil should return zero value")
	}
}

func TestToFromSlicePtr(t *testing.T) {
	in := []int{1, 2, 3}

	outPtrs := ToSlicePtr(in)
	if len(outPtrs) != len(in) {
		t.Fatal("ToSlicePtr length mismatch")
	}

	for i := range in {
		if outPtrs[i] == nil || *outPtrs[i] != in[i] {
			t.Fatal("ToSlicePtr element mismatch")
		}
	}

	// FromSlicePtr with nil element should yield zero at that index
	a, b := 10, 20
	ptrs := []*int{&a, nil, &b}

	vals := FromSlicePtr(ptrs)
	if len(vals) != 3 || vals[0] != 10 || vals[1] != 0 || vals[2] != 20 {
		t.Fatalf("FromSlicePtr unexpected values: %#v", vals)
	}

	// empty slices
	none := []int{}

	nonePtrs := ToSlicePtr(none)
	if len(nonePtrs) != 0 {
		t.Fatal("ToSlicePtr of empty slice should be empty")
	}

	var nilPtrs []*int

	vals2 := FromSlicePtr(nilPtrs)
	if len(vals2) != 0 {
		t.Fatal("FromSlicePtr of nil slice should be empty")
	}
}

func TestIsType(t *testing.T) {
	var x = 5
	if !IsType(x, "int") {
		t.Fatal("expected int type")
	}

	if IsType(x, "string") {
		t.Fatal("did not expect string type")
	}

	if !IsType(&x, "int") {
		t.Fatal("pointer should resolve to underlying int type")
	}

	if IsType(nil, "int") {
		t.Fatal("nil should not be a type match")
	}
	// struct typing
	st := sampleStruct{}
	if !IsType(st, "pointer.sampleStruct") {
		t.Fatal("expected type pointer.sampleStruct")
	}

	if !IsType(&st, "pointer.sampleStruct") {
		t.Fatal("expected pointer to resolve to struct type")
	}
}

func TestKindHelpers(t *testing.T) {
	st := sampleStruct{A: 1}
	if !IsStruct(st) || !IsStruct(&st) {
		t.Fatal("struct detection failed")
	}

	if IsStruct(3) {
		t.Fatal("int is not a struct")
	}

	if IsStruct(nil) {
		t.Fatal("nil is not a struct")
	}

	c := make(chan int)
	if !IsChan(c) || !IsChan(&c) {
		t.Fatal("chan detection failed")
	}

	if IsChan(3) {
		t.Fatal("int is not a chan")
	}

	if IsChan(nil) {
		t.Fatal("nil is not a chan")
	}

	var cn chan int
	if !IsChan(cn) {
		t.Fatal("nil chan typed should still be kind chan")
	}

	if !IsChan(&cn) {
		t.Fatal("pointer to typed nil chan should be recognized as chan kind")
	}

	sl := []int{1}

	arr := [2]int{1, 2}
	if !IsSlice(sl) || !IsSlice(arr) || !IsSlice(&sl) || !IsSlice(&arr) {
		t.Fatal("slice/array detection failed")
	}

	if IsSlice(3) {
		t.Fatal("int is not a slice/array")
	}

	if IsSlice(nil) {
		t.Fatal("nil is not a slice/array")
	}

	var nilSlice []int
	if !IsSlice(nilSlice) {
		t.Fatal("typed nil slice should be recognized as slice kind")
	}

	if !IsSlice(&nilSlice) {
		t.Fatal("pointer to typed nil slice should be recognized as slice kind")
	}

	m := map[string]int{"a": 1}
	if !IsMap(m) || !IsMap(&m) {
		t.Fatal("map detection failed")
	}

	if IsMap(3) {
		t.Fatal("int is not a map")
	}

	if IsMap(nil) {
		t.Fatal("nil is not a map")
	}

	var nilMap map[string]int
	if !IsMap(nilMap) {
		t.Fatal("typed nil map should be recognized as map kind")
	}

	if !IsMap(&nilMap) {
		t.Fatal("pointer to typed nil map should be recognized as map kind")
	}
}
