// Package main demonstrates common/pointer: pointer-to-value /
// value-to-pointer adapters, nil/empty/zero introspection, and slice
// conversions between []T and []*T.
package main

import (
	"fmt"

	cpointer "github.com/guidomantilla/yarumo/common/pointer"
)

func main() {
	demoToFromPtr()
	demoZeroAndIsZero()
	demoNilAndEmpty()
	demoSlicePtr()
	demoKindIntrospection()
}

// demoToFromPtr exercises the ToPtr / FromPtr round trip.
func demoToFromPtr() {
	fmt.Println("=== ToPtr / FromPtr ===")

	p := cpointer.ToPtr(42)
	fmt.Printf("  ToPtr(42)   -> %p (deref %d)\n", p, *p)

	var nilPtr *int

	fmt.Printf("  FromPtr(nil) -> %d (zero value)\n", cpointer.FromPtr(nilPtr))
	fmt.Printf("  FromPtr(p)   -> %d\n", cpointer.FromPtr(p))
}

// demoZeroAndIsZero shows Zero[T] and IsZero[T] in action.
func demoZeroAndIsZero() {
	fmt.Println("=== Zero / IsZero ===")

	fmt.Printf("  Zero[int]()       -> %d\n", cpointer.Zero[int]())
	fmt.Printf("  Zero[string]()    -> %q\n", cpointer.Zero[string]())
	fmt.Printf("  IsZero(0)         -> %v\n", cpointer.IsZero(0))
	fmt.Printf("  IsZero(\"hello\")   -> %v\n", cpointer.IsZero("hello"))
}

// demoNilAndEmpty shows the difference between IsNil and IsEmpty.
func demoNilAndEmpty() {
	fmt.Println("=== IsNil / IsEmpty ===")

	var nilSlice []int

	fmt.Printf("  IsNil(nilSlice)        -> %v\n", cpointer.IsNil(nilSlice))
	fmt.Printf("  IsEmpty([]int{})       -> %v\n", cpointer.IsEmpty([]int{}))
	fmt.Printf("  IsEmpty([]int{1, 2})   -> %v\n", cpointer.IsEmpty([]int{1, 2}))
	fmt.Printf("  IsNotEmpty(\"yarumo\")  -> %v\n", cpointer.IsNotEmpty("yarumo"))
}

// demoSlicePtr round-trips []T ↔ []*T.
func demoSlicePtr() {
	fmt.Println("=== ToSlicePtr / FromSlicePtr ===")

	in := []string{"a", "b", "c"}
	ptrs := cpointer.ToSlicePtr(in)
	fmt.Printf("  ToSlicePtr -> %d pointers\n", len(ptrs))

	out := cpointer.FromSlicePtr(ptrs)
	fmt.Printf("  FromSlicePtr -> %v\n", out)
}

// demoKindIntrospection exercises IsStruct/IsSlice/IsMap/IsChan/IsPointer.
func demoKindIntrospection() {
	fmt.Println("=== Kind introspection ===")

	type Point struct{ X, Y int }

	fmt.Printf("  IsStruct(Point{})          -> %v\n", cpointer.IsStruct(Point{}))
	fmt.Printf("  IsSlice([]int{1})          -> %v\n", cpointer.IsSlice([]int{1}))
	fmt.Printf("  IsMap(map[string]int{})    -> %v\n", cpointer.IsMap(map[string]int{}))
	fmt.Printf("  IsChan(make(chan int))     -> %v\n", cpointer.IsChan(make(chan int)))
	fmt.Printf("  IsPointer(cpointer.ToPtr(1)) -> %v\n", cpointer.IsPointer(cpointer.ToPtr(1)))
}
