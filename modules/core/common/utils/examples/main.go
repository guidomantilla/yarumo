// Package main demonstrates common/utils: ternary/coalesce predicates,
// string case converters (Pascal/Camel/Kebab/Snake), and a handful of
// slice / map helpers powered by the constraints package.
package main

import (
	"fmt"

	cutils "github.com/guidomantilla/yarumo/core/common/utils"
)

func main() {
	demoPredicates()
	demoStringCases()
	demoSlices()
	demoMaps()
}

// demoPredicates exercises Ternary, Coalesce, Empty, and NotEmpty.
func demoPredicates() {
	fmt.Println("=== Predicates ===")

	fmt.Printf("  Ternary(true, \"yes\", \"no\")  -> %q\n", cutils.Ternary(true, "yes", "no"))
	fmt.Printf("  Coalesce(\"\", \"\", \"pikachu\") -> %q\n", cutils.Coalesce("", "", "pikachu"))
	fmt.Printf("  Empty(\"\", nil)              -> %v\n", cutils.Empty("", nil))
	fmt.Printf("  NotEmpty(\"a\", \"b\")          -> %v\n", cutils.NotEmpty("a", "b"))
}

// demoStringCases shows the four case converters operating on the same input.
func demoStringCases() {
	fmt.Println("=== String case converters ===")

	input := "yarumo_common_utils Module"

	fmt.Printf("  Input        -> %q\n", input)
	fmt.Printf("  PascalCase   -> %q\n", cutils.PascalCase(input))
	fmt.Printf("  CamelCase    -> %q\n", cutils.CamelCase(input))
	fmt.Printf("  KebabCase    -> %q\n", cutils.KebabCase(input))
	fmt.Printf("  SnakeCase    -> %q\n", cutils.SnakeCase(input))
}

// demoSlices covers a few of the generic slice helpers.
func demoSlices() {
	fmt.Println("=== Slices ===")

	xs := []int{4, 1, 5, 1, 2, 5, 3}

	fmt.Printf("  In(3, xs)               -> %v\n", cutils.In(3, xs...))
	fmt.Printf("  Deduplicate(xs)         -> %v\n", cutils.Deduplicate(append([]int(nil), xs...)))
	fmt.Printf("  Sort(xs)                -> %v\n", cutils.Sort(append([]int(nil), xs...)))
	fmt.Printf("  Max(xs)                 -> %d\n", cutils.Max(xs))
	fmt.Printf("  Min(xs)                 -> %d\n", cutils.Min(xs))
	fmt.Printf("  Chunk(xs, 3)            -> %v\n", cutils.Chunk(xs, 3))

	even := func(x int) bool { return x%2 == 0 }
	fmt.Printf("  CountBy(xs, even)       -> %d\n", cutils.CountBy(xs, even))
}

// demoMaps covers Keys, Values, and PickByKeys.
func demoMaps() {
	fmt.Println("=== Maps ===")

	m := map[string]int{"a": 1, "b": 2, "c": 3}

	keys := cutils.Keys(m)
	cutils.Sort(keys)
	fmt.Printf("  Keys(m) (sorted)     -> %v\n", keys)

	values := cutils.Values(m)
	cutils.Sort(values)
	fmt.Printf("  Values(m) (sorted)   -> %v\n", values)

	fmt.Printf("  HasKey(\"a\", m)       -> %v\n", cutils.HasKey("a", m))

	picked := cutils.PickByKeys(m, []string{"a", "c"})
	fmt.Printf("  PickByKeys(m, a/c)   -> %v\n", picked)
}
