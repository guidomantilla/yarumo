// Package main demonstrates the contract-only common/cast package. The
// package declares function-type aliases (ToBoolFn, ToIntFn, ToDurationFn,
// ...) so consumers can type-check casting implementations against the
// contract without pulling external dependencies. Concrete implementations
// live in modules/extension/common/cast/ (the spf13/cast wrapper); this
// demo writes its own toy implementations and asserts they satisfy the
// contract via the declared Fn types.
package main

import (
	"fmt"
	"strconv"
	"time"

	ccast "github.com/guidomantilla/yarumo/common/cast"
)

func main() {
	demoToBool()
	demoToInt()
	demoToString()
	demoToDuration()
}

// demoToBool wires a tiny strconv.ParseBool wrapper into ccast.ToBoolFn.
func demoToBool() {
	fmt.Println("=== ToBoolFn ===")

	var toBool ccast.ToBoolFn = func(v any) (bool, error) {
		s, ok := v.(string)
		if !ok {
			return false, fmt.Errorf("not a string: %T", v)
		}

		return strconv.ParseBool(s)
	}

	got, _ := toBool("true")
	fmt.Printf("  toBool(\"true\") -> %v\n", got)

	_, err := toBool(42)
	fmt.Printf("  toBool(42) -> %v\n", err)
}

// demoToInt wires a tiny implementation into ccast.ToIntFn.
func demoToInt() {
	fmt.Println("=== ToIntFn ===")

	var toInt ccast.ToIntFn = func(v any) (int, error) {
		switch x := v.(type) {
		case int:
			return x, nil
		case string:
			return strconv.Atoi(x)
		default:
			return 0, fmt.Errorf("unsupported: %T", v)
		}
	}

	n, _ := toInt("123")
	fmt.Printf("  toInt(\"123\") -> %d\n", n)
}

// demoToString wires a fmt.Sprint wrapper into ccast.ToStringFn.
func demoToString() {
	fmt.Println("=== ToStringFn ===")

	var toString ccast.ToStringFn = func(v any) (string, error) {
		return fmt.Sprint(v), nil
	}

	s, _ := toString(3.14)
	fmt.Printf("  toString(3.14) -> %q\n", s)
}

// demoToDuration wires a time.ParseDuration wrapper into ccast.ToDurationFn.
func demoToDuration() {
	fmt.Println("=== ToDurationFn ===")

	var toDuration ccast.ToDurationFn = func(v any) (time.Duration, error) {
		s, ok := v.(string)
		if !ok {
			return 0, fmt.Errorf("not a string: %T", v)
		}

		return time.ParseDuration(s)
	}

	d, _ := toDuration("250ms")
	fmt.Printf("  toDuration(\"250ms\") -> %s\n", d)
}
