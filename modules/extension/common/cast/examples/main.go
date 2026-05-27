// Demo that exercises a representative slice of the cast public API:
//
//  1. ToBool, ToInt, ToString for the scalar happy path.
//  2. ToDuration / ToTime for the time-shaped converters.
//  3. ToStringSlice / ToIntSlice for the slice family.
//  4. ToStringMap for the map family.
//  5. Failure mode: ToInt rejects a non-numeric string.
//
// Every conversion delegates to spf13/cast; this package only adds the
// type-alias compliance against modules/common/cast/.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/guidomantilla/yarumo/config"
	"github.com/guidomantilla/yarumo/extension/common/cast"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/extension/common/cast/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	demos := []struct {
		title string
		fn    func(context.Context) error
	}{
		{"ToBool / ToInt / ToString (scalars)", demoScalars},
		{"ToDuration / ToTime", demoTime},
		{"ToStringSlice / ToIntSlice", demoSlices},
		{"ToStringMap", demoMap},
		{"ToInt failure path", demoFailure},
	}

	for _, d := range demos {
		fmt.Printf("=== Demo: %s ===\n", d.title)
		err := d.fn(ctx)
		if err != nil {
			return fmt.Errorf("%s: %w", d.title, err)
		}
		fmt.Println()
	}

	return nil
}

// demoScalars converts a handful of representative inputs to bool, int and string.
func demoScalars(_ context.Context) error {
	b, err := cast.ToBool("true")
	if err != nil {
		return fmt.Errorf("ToBool: %w", err)
	}
	fmt.Printf("  ToBool(\"true\") -> %v\n", b)

	n, err := cast.ToInt("42")
	if err != nil {
		return fmt.Errorf("ToInt: %w", err)
	}
	fmt.Printf("  ToInt(\"42\") -> %d\n", n)

	s, err := cast.ToString(3.14)
	if err != nil {
		return fmt.Errorf("ToString: %w", err)
	}
	fmt.Printf("  ToString(3.14) -> %q\n", s)

	return nil
}

// demoTime exercises the duration and time converters.
func demoTime(_ context.Context) error {
	d, err := cast.ToDuration("250ms")
	if err != nil {
		return fmt.Errorf("ToDuration: %w", err)
	}
	fmt.Printf("  ToDuration(\"250ms\") -> %s\n", d)

	t, err := cast.ToTime("2026-05-26T10:00:00Z")
	if err != nil {
		return fmt.Errorf("ToTime: %w", err)
	}
	fmt.Printf("  ToTime(RFC3339) -> %s\n", t.Format(time.RFC3339))

	return nil
}

// demoSlices exercises the slice converters.
func demoSlices(_ context.Context) error {
	ss, err := cast.ToStringSlice([]any{"a", "b", "c"})
	if err != nil {
		return fmt.Errorf("ToStringSlice: %w", err)
	}
	fmt.Printf("  ToStringSlice -> %v\n", ss)

	ints, err := cast.ToIntSlice([]any{"1", "2", "3"})
	if err != nil {
		return fmt.Errorf("ToIntSlice: %w", err)
	}
	fmt.Printf("  ToIntSlice -> %v\n", ints)

	return nil
}

// demoMap exercises the heterogeneous map converter.
func demoMap(_ context.Context) error {
	m, err := cast.ToStringMap(map[any]any{"k1": 1, "k2": "two"})
	if err != nil {
		return fmt.Errorf("ToStringMap: %w", err)
	}
	fmt.Printf("  ToStringMap -> %v\n", m)

	return nil
}

// demoFailure shows the error path: a non-numeric string cannot be cast to int.
func demoFailure(_ context.Context) error {
	_, err := cast.ToInt("not-a-number")
	if err == nil {
		return fmt.Errorf("expected ToInt to reject non-numeric string")
	}
	fmt.Printf("  ToInt(\"not-a-number\") -> err: %v\n", err)
	return nil
}
