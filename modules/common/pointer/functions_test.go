package pointer

import "testing"

type sampleStruct struct{ A int }

func TestIsNil(t *testing.T) {
	t.Parallel()

	t.Run("untyped nil", func(t *testing.T) {
		t.Parallel()

		if !IsNil(nil) {
			t.Fatal("expected true")
		}
	})

	t.Run("typed nil pointer", func(t *testing.T) {
		t.Parallel()

		var p *int
		if !IsNil(p) {
			t.Fatal("expected true")
		}
	})

	t.Run("typed nil slice", func(t *testing.T) {
		t.Parallel()

		var s []int
		if !IsNil(s) {
			t.Fatal("expected true")
		}
	})

	t.Run("typed nil map", func(t *testing.T) {
		t.Parallel()

		var m map[string]int
		if !IsNil(m) {
			t.Fatal("expected true")
		}
	})

	t.Run("typed nil chan", func(t *testing.T) {
		t.Parallel()

		var ch chan int
		if !IsNil(ch) {
			t.Fatal("expected true")
		}
	})

	t.Run("typed nil func", func(t *testing.T) {
		t.Parallel()

		var fn func()
		if !IsNil(fn) {
			t.Fatal("expected true")
		}
	})

	t.Run("interface holding typed nil pointer", func(t *testing.T) {
		t.Parallel()

		var (
			p  *int
			ai any = p
		)

		if !IsNil(ai) {
			t.Fatal("expected true")
		}
	})

	t.Run("non-nil int", func(t *testing.T) {
		t.Parallel()

		if IsNil(10) {
			t.Fatal("expected false")
		}
	})

	t.Run("non-nil pointer", func(t *testing.T) {
		t.Parallel()

		x := 10
		if IsNil(&x) {
			t.Fatal("expected false")
		}
	})

	t.Run("non-nil slice", func(t *testing.T) {
		t.Parallel()

		if IsNil([]int{1}) {
			t.Fatal("expected false")
		}
	})

	t.Run("non-nil map", func(t *testing.T) {
		t.Parallel()

		if IsNil(map[string]int{"a": 1}) {
			t.Fatal("expected false")
		}
	})

	t.Run("non-nil chan", func(t *testing.T) {
		t.Parallel()

		if IsNil(make(chan int)) {
			t.Fatal("expected false")
		}
	})

	t.Run("non-nil func", func(t *testing.T) {
		t.Parallel()

		if IsNil(func() {}) {
			t.Fatal("expected false")
		}
	})
}

func TestIsNotNil(t *testing.T) {
	t.Parallel()

	t.Run("untyped nil", func(t *testing.T) {
		t.Parallel()

		if IsNotNil(nil) {
			t.Fatal("expected false")
		}
	})

	t.Run("non-nil value", func(t *testing.T) {
		t.Parallel()

		if !IsNotNil(10) {
			t.Fatal("expected true")
		}
	})
}

func TestIsEmpty(t *testing.T) {
	t.Parallel()

	t.Run("untyped nil", func(t *testing.T) {
		t.Parallel()

		if !IsEmpty(nil) {
			t.Fatal("expected true")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()

		if !IsEmpty("") {
			t.Fatal("expected true")
		}
	})

	t.Run("non-empty string", func(t *testing.T) {
		t.Parallel()

		if IsEmpty("a") {
			t.Fatal("expected false")
		}
	})

	t.Run("zero int", func(t *testing.T) {
		t.Parallel()

		if !IsEmpty(0) {
			t.Fatal("expected true")
		}
	})

	t.Run("non-zero int", func(t *testing.T) {
		t.Parallel()

		if IsEmpty(1) {
			t.Fatal("expected false")
		}
	})

	t.Run("false bool", func(t *testing.T) {
		t.Parallel()

		if !IsEmpty(false) {
			t.Fatal("expected true for zero value bool")
		}
	})

	t.Run("true bool", func(t *testing.T) {
		t.Parallel()

		if IsEmpty(true) {
			t.Fatal("expected false")
		}
	})

	t.Run("zero struct", func(t *testing.T) {
		t.Parallel()

		if !IsEmpty(sampleStruct{}) {
			t.Fatal("expected true for zero struct")
		}
	})

	t.Run("non-zero struct", func(t *testing.T) {
		t.Parallel()

		if IsEmpty(sampleStruct{A: 1}) {
			t.Fatal("expected false")
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		t.Parallel()

		var np *int
		if !IsEmpty(np) {
			t.Fatal("expected true")
		}
	})

	t.Run("pointer to empty string", func(t *testing.T) {
		t.Parallel()

		s := ""
		if !IsEmpty(&s) {
			t.Fatal("expected true")
		}
	})

	t.Run("pointer to non-empty string", func(t *testing.T) {
		t.Parallel()

		s := "a"
		if IsEmpty(&s) {
			t.Fatal("expected false")
		}
	})

	t.Run("pointer to zero int", func(t *testing.T) {
		t.Parallel()

		z := 0
		if !IsEmpty(&z) {
			t.Fatal("expected true")
		}
	})

	t.Run("pointer to non-zero int", func(t *testing.T) {
		t.Parallel()

		nz := 2
		if IsEmpty(&nz) {
			t.Fatal("expected false")
		}
	})

	t.Run("pointer to zero struct", func(t *testing.T) {
		t.Parallel()

		if !IsEmpty(&sampleStruct{}) {
			t.Fatal("expected true")
		}
	})

	t.Run("pointer to non-zero struct", func(t *testing.T) {
		t.Parallel()

		if IsEmpty(&sampleStruct{A: 1}) {
			t.Fatal("expected false")
		}
	})

	t.Run("nil slice", func(t *testing.T) {
		t.Parallel()

		var s []int
		if !IsEmpty(s) {
			t.Fatal("expected true")
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		if !IsEmpty([]int{}) {
			t.Fatal("expected true")
		}
	})

	t.Run("non-empty slice", func(t *testing.T) {
		t.Parallel()

		if IsEmpty([]int{1}) {
			t.Fatal("expected false")
		}
	})

	t.Run("empty array", func(t *testing.T) {
		t.Parallel()

		if !IsEmpty([0]int{}) {
			t.Fatal("expected true")
		}
	})

	t.Run("non-empty array", func(t *testing.T) {
		t.Parallel()

		if IsEmpty([2]int{1, 2}) {
			t.Fatal("expected false")
		}
	})

	t.Run("nil map", func(t *testing.T) {
		t.Parallel()

		var m map[string]int
		if !IsEmpty(m) {
			t.Fatal("expected true")
		}
	})

	t.Run("empty map", func(t *testing.T) {
		t.Parallel()

		if !IsEmpty(map[string]int{}) {
			t.Fatal("expected true")
		}
	})

	t.Run("non-empty map", func(t *testing.T) {
		t.Parallel()

		if IsEmpty(map[string]int{"a": 1}) {
			t.Fatal("expected false")
		}
	})

	t.Run("nil chan", func(t *testing.T) {
		t.Parallel()

		var ch chan int
		if !IsEmpty(ch) {
			t.Fatal("expected true")
		}
	})

	t.Run("non-nil empty chan", func(t *testing.T) {
		t.Parallel()

		if !IsEmpty(make(chan int)) {
			t.Fatal("expected true for len 0 chan")
		}
	})
}

func TestIsNotEmpty(t *testing.T) {
	t.Parallel()

	t.Run("untyped nil", func(t *testing.T) {
		t.Parallel()

		if IsNotEmpty(nil) {
			t.Fatal("expected false")
		}
	})

	t.Run("non-empty string", func(t *testing.T) {
		t.Parallel()

		if !IsNotEmpty("a") {
			t.Fatal("expected true")
		}
	})
}

func TestIsPointer(t *testing.T) {
	t.Parallel()

	t.Run("untyped nil", func(t *testing.T) {
		t.Parallel()

		if IsPointer(nil) {
			t.Fatal("expected false")
		}
	})

	t.Run("typed nil pointer", func(t *testing.T) {
		t.Parallel()

		var p *int
		if !IsPointer(p) {
			t.Fatal("expected true")
		}
	})

	t.Run("non-nil pointer", func(t *testing.T) {
		t.Parallel()

		x := 3
		if !IsPointer(&x) {
			t.Fatal("expected true")
		}
	})

	t.Run("non-pointer value", func(t *testing.T) {
		t.Parallel()

		if IsPointer(3) {
			t.Fatal("expected false")
		}
	})
}

func TestIsNotPointer(t *testing.T) {
	t.Parallel()

	t.Run("non-pointer value", func(t *testing.T) {
		t.Parallel()

		if !IsNotPointer(3) {
			t.Fatal("expected true")
		}
	})

	t.Run("pointer value", func(t *testing.T) {
		t.Parallel()

		x := 3
		if IsNotPointer(&x) {
			t.Fatal("expected false")
		}
	})
}

func TestZero(t *testing.T) {
	t.Parallel()

	t.Run("int", func(t *testing.T) {
		t.Parallel()

		if Zero[int]() != 0 {
			t.Fatal("expected 0")
		}
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		if Zero[string]() != "" {
			t.Fatal("expected empty string")
		}
	})

	t.Run("bool", func(t *testing.T) {
		t.Parallel()

		if Zero[bool]() != false {
			t.Fatal("expected false")
		}
	})

	t.Run("struct", func(t *testing.T) {
		t.Parallel()

		z := Zero[sampleStruct]()
		if z.A != 0 {
			t.Fatal("expected zero struct")
		}
	})

	t.Run("pointer", func(t *testing.T) {
		t.Parallel()

		if Zero[*int]() != nil {
			t.Fatal("expected nil pointer")
		}
	})
}

func TestIsZero(t *testing.T) {
	t.Parallel()

	t.Run("zero int", func(t *testing.T) {
		t.Parallel()

		if !IsZero(0) {
			t.Fatal("expected true")
		}
	})

	t.Run("non-zero int", func(t *testing.T) {
		t.Parallel()

		if IsZero(5) {
			t.Fatal("expected false")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()

		if !IsZero("") {
			t.Fatal("expected true")
		}
	})

	t.Run("non-empty string", func(t *testing.T) {
		t.Parallel()

		if IsZero("a") {
			t.Fatal("expected false")
		}
	})

	t.Run("false bool", func(t *testing.T) {
		t.Parallel()

		if !IsZero(false) {
			t.Fatal("expected true")
		}
	})

	t.Run("true bool", func(t *testing.T) {
		t.Parallel()

		if IsZero(true) {
			t.Fatal("expected false")
		}
	})
}

func TestIsNotZero(t *testing.T) {
	t.Parallel()

	t.Run("zero int", func(t *testing.T) {
		t.Parallel()

		if IsNotZero(0) {
			t.Fatal("expected false")
		}
	})

	t.Run("non-zero int", func(t *testing.T) {
		t.Parallel()

		if !IsNotZero(5) {
			t.Fatal("expected true")
		}
	})
}

func TestToPtr(t *testing.T) {
	t.Parallel()

	t.Run("int value", func(t *testing.T) {
		t.Parallel()

		p := ToPtr(7)
		if p == nil || *p != 7 {
			t.Fatal("expected pointer to 7")
		}
	})

	t.Run("string value", func(t *testing.T) {
		t.Parallel()

		p := ToPtr("hello")
		if p == nil || *p != "hello" {
			t.Fatal("expected pointer to hello")
		}
	})

	t.Run("struct value", func(t *testing.T) {
		t.Parallel()

		p := ToPtr(sampleStruct{A: 5})
		if p == nil || p.A != 5 {
			t.Fatal("expected pointer to struct")
		}
	})
}

func TestFromPtr(t *testing.T) {
	t.Parallel()

	t.Run("non-nil pointer", func(t *testing.T) {
		t.Parallel()

		v := 7
		if FromPtr(&v) != 7 {
			t.Fatal("expected 7")
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		t.Parallel()

		var p *int
		if FromPtr(p) != 0 {
			t.Fatal("expected zero value")
		}
	})

	t.Run("nil string pointer", func(t *testing.T) {
		t.Parallel()

		var p *string
		if FromPtr(p) != "" {
			t.Fatal("expected empty string")
		}
	})

	t.Run("struct pointer", func(t *testing.T) {
		t.Parallel()

		s := sampleStruct{A: 3}
		if FromPtr(&s).A != 3 {
			t.Fatal("expected struct with A=3")
		}
	})
}

func TestToSlicePtr(t *testing.T) {
	t.Parallel()

	t.Run("non-empty slice", func(t *testing.T) {
		t.Parallel()

		in := []int{1, 2, 3}

		out := ToSlicePtr(in)
		if len(out) != 3 {
			t.Fatal("length mismatch")
		}

		for i := range in {
			if out[i] == nil || *out[i] != in[i] {
				t.Fatalf("element %d mismatch", i)
			}
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		out := ToSlicePtr([]int{})
		if len(out) != 0 {
			t.Fatal("expected empty result")
		}
	})

	t.Run("nil slice", func(t *testing.T) {
		t.Parallel()

		var in []int

		out := ToSlicePtr(in)
		if len(out) != 0 {
			t.Fatal("expected empty result")
		}
	})
}

func TestFromSlicePtr(t *testing.T) {
	t.Parallel()

	t.Run("mixed nil and non-nil", func(t *testing.T) {
		t.Parallel()

		a, b := 10, 20
		ptrs := []*int{&a, nil, &b}

		vals := FromSlicePtr(ptrs)
		if len(vals) != 3 || vals[0] != 10 || vals[1] != 0 || vals[2] != 20 {
			t.Fatalf("unexpected values: %#v", vals)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		vals := FromSlicePtr([]*int{})
		if len(vals) != 0 {
			t.Fatal("expected empty result")
		}
	})

	t.Run("nil slice", func(t *testing.T) {
		t.Parallel()

		var ptrs []*int

		vals := FromSlicePtr(ptrs)
		if len(vals) != 0 {
			t.Fatal("expected empty result")
		}
	})
}

func TestIsType(t *testing.T) {
	t.Parallel()

	t.Run("int value", func(t *testing.T) {
		t.Parallel()

		if !IsType(5, "int") {
			t.Fatal("expected true")
		}
	})

	t.Run("wrong type name", func(t *testing.T) {
		t.Parallel()

		if IsType(5, "string") {
			t.Fatal("expected false")
		}
	})

	t.Run("pointer resolves to underlying type", func(t *testing.T) {
		t.Parallel()

		x := 5
		if !IsType(&x, "int") {
			t.Fatal("expected true")
		}
	})

	t.Run("untyped nil", func(t *testing.T) {
		t.Parallel()

		if IsType(nil, "int") {
			t.Fatal("expected false")
		}
	})

	t.Run("typed nil pointer does not panic", func(t *testing.T) {
		t.Parallel()

		var p *int
		if IsType(p, "int") {
			t.Fatal("expected false")
		}
	})

	t.Run("struct value", func(t *testing.T) {
		t.Parallel()

		if !IsType(sampleStruct{}, "pointer.sampleStruct") {
			t.Fatal("expected true")
		}
	})

	t.Run("pointer to struct", func(t *testing.T) {
		t.Parallel()

		if !IsType(&sampleStruct{}, "pointer.sampleStruct") {
			t.Fatal("expected true")
		}
	})

	t.Run("string value", func(t *testing.T) {
		t.Parallel()

		if !IsType("hello", "string") {
			t.Fatal("expected true")
		}
	})
}

func TestIsStruct(t *testing.T) {
	t.Parallel()

	t.Run("struct value", func(t *testing.T) {
		t.Parallel()

		if !IsStruct(sampleStruct{A: 1}) {
			t.Fatal("expected true")
		}
	})

	t.Run("pointer to struct", func(t *testing.T) {
		t.Parallel()

		if !IsStruct(&sampleStruct{A: 1}) {
			t.Fatal("expected true")
		}
	})

	t.Run("non-struct", func(t *testing.T) {
		t.Parallel()

		if IsStruct(3) {
			t.Fatal("expected false")
		}
	})

	t.Run("untyped nil", func(t *testing.T) {
		t.Parallel()

		if IsStruct(nil) {
			t.Fatal("expected false")
		}
	})

	t.Run("typed nil pointer to struct", func(t *testing.T) {
		t.Parallel()

		var p *sampleStruct
		if IsStruct(p) {
			t.Fatal("expected false for nil pointer deref to zero Value")
		}
	})
}

func TestIsChan(t *testing.T) {
	t.Parallel()

	t.Run("non-nil chan", func(t *testing.T) {
		t.Parallel()

		if !IsChan(make(chan int)) {
			t.Fatal("expected true")
		}
	})

	t.Run("pointer to chan", func(t *testing.T) {
		t.Parallel()

		c := make(chan int)
		if !IsChan(&c) {
			t.Fatal("expected true")
		}
	})

	t.Run("typed nil chan", func(t *testing.T) {
		t.Parallel()

		var ch chan int
		if !IsChan(ch) {
			t.Fatal("expected true")
		}
	})

	t.Run("pointer to typed nil chan", func(t *testing.T) {
		t.Parallel()

		var ch chan int
		if !IsChan(&ch) {
			t.Fatal("expected true")
		}
	})

	t.Run("non-chan", func(t *testing.T) {
		t.Parallel()

		if IsChan(3) {
			t.Fatal("expected false")
		}
	})

	t.Run("untyped nil", func(t *testing.T) {
		t.Parallel()

		if IsChan(nil) {
			t.Fatal("expected false")
		}
	})
}

func TestIsSlice(t *testing.T) {
	t.Parallel()

	t.Run("non-empty slice", func(t *testing.T) {
		t.Parallel()

		if !IsSlice([]int{1}) {
			t.Fatal("expected true")
		}
	})

	t.Run("array", func(t *testing.T) {
		t.Parallel()

		if !IsSlice([2]int{1, 2}) {
			t.Fatal("expected true")
		}
	})

	t.Run("pointer to slice", func(t *testing.T) {
		t.Parallel()

		s := []int{1}
		if !IsSlice(&s) {
			t.Fatal("expected true")
		}
	})

	t.Run("pointer to array", func(t *testing.T) {
		t.Parallel()

		a := [2]int{1, 2}
		if !IsSlice(&a) {
			t.Fatal("expected true")
		}
	})

	t.Run("typed nil slice", func(t *testing.T) {
		t.Parallel()

		var s []int
		if !IsSlice(s) {
			t.Fatal("expected true")
		}
	})

	t.Run("pointer to typed nil slice", func(t *testing.T) {
		t.Parallel()

		var s []int
		if !IsSlice(&s) {
			t.Fatal("expected true")
		}
	})

	t.Run("non-slice", func(t *testing.T) {
		t.Parallel()

		if IsSlice(3) {
			t.Fatal("expected false")
		}
	})

	t.Run("untyped nil", func(t *testing.T) {
		t.Parallel()

		if IsSlice(nil) {
			t.Fatal("expected false")
		}
	})
}

func TestIsMap(t *testing.T) {
	t.Parallel()

	t.Run("non-empty map", func(t *testing.T) {
		t.Parallel()

		if !IsMap(map[string]int{"a": 1}) {
			t.Fatal("expected true")
		}
	})

	t.Run("pointer to map", func(t *testing.T) {
		t.Parallel()

		m := map[string]int{"a": 1}
		if !IsMap(&m) {
			t.Fatal("expected true")
		}
	})

	t.Run("typed nil map", func(t *testing.T) {
		t.Parallel()

		var m map[string]int
		if !IsMap(m) {
			t.Fatal("expected true")
		}
	})

	t.Run("pointer to typed nil map", func(t *testing.T) {
		t.Parallel()

		var m map[string]int
		if !IsMap(&m) {
			t.Fatal("expected true")
		}
	})

	t.Run("non-map", func(t *testing.T) {
		t.Parallel()

		if IsMap(3) {
			t.Fatal("expected false")
		}
	})

	t.Run("untyped nil", func(t *testing.T) {
		t.Parallel()

		if IsMap(nil) {
			t.Fatal("expected false")
		}
	})
}
