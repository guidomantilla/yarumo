// Package main demonstrates the typed expression evaluator in
// common/expressions: construct an Evaluator, build a Context, evaluate
// arithmetic + boolean + property-access expressions, register a custom
// function via WithFunc, and handle ParseError vs EvalError.
package main

import (
	"errors"
	"fmt"

	cexpr "github.com/guidomantilla/yarumo/common/expressions"
)

func main() {
	demoArithmetic()
	demoLogical()
	demoBuiltins()
	demoCustomFunc()
	demoErrors()
}

// demoArithmetic evaluates plain integer and float arithmetic.
func demoArithmetic() {
	fmt.Println("=== Arithmetic ===")

	eval := cexpr.NewEvaluator()

	value, _ := eval.Evaluate("1 + 2 * 3", cexpr.Context{})
	fmt.Printf("  1 + 2 * 3 -> %v\n", value)

	value, _ = eval.Evaluate("(1 + 2) * 3", cexpr.Context{})
	fmt.Printf("  (1 + 2) * 3 -> %v\n", value)
}

// demoLogical evaluates a boolean expression that pulls values from the context.
func demoLogical() {
	fmt.Println("=== Logical ===")

	eval := cexpr.NewEvaluator()
	ctx := cexpr.Context{
		"age":    30,
		"active": true,
	}

	value, _ := eval.Evaluate("age >= 18 and active", ctx)
	fmt.Printf("  age >= 18 and active -> %v\n", value)
}

// demoBuiltins exercises the built-in len/upper helpers from DefaultFuncs.
func demoBuiltins() {
	fmt.Println("=== Builtins ===")

	eval := cexpr.NewEvaluator()
	ctx := cexpr.Context{
		"name":  "pikachu",
		"items": []any{1, 2, 3, 4, 5},
	}

	value, _ := eval.Evaluate("upper(name)", ctx)
	fmt.Printf("  upper(name) -> %v\n", value)

	value, _ = eval.Evaluate("len(items)", ctx)
	fmt.Printf("  len(items) -> %v\n", value)
}

// demoCustomFunc registers a custom function and invokes it from an expression.
func demoCustomFunc() {
	fmt.Println("=== Custom function ===")

	double := func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, errors.New("double expects 1 arg")
		}

		switch n := args[0].(type) {
		case int:
			return n * 2, nil
		case int64:
			return n * 2, nil
		case float64:
			return n * 2, nil
		default:
			return nil, fmt.Errorf("double expects a number, got %T", args[0])
		}
	}

	eval := cexpr.NewEvaluator(cexpr.WithFunc("double", double))

	value, _ := eval.Evaluate("double(21)", cexpr.Context{})
	fmt.Printf("  double(21) -> %v\n", value)
}

// demoErrors shows how Parse failures surface as ParseError and runtime
// failures as EvalError.
func demoErrors() {
	fmt.Println("=== Errors ===")

	eval := cexpr.NewEvaluator()

	_, err := eval.Evaluate("1 +", cexpr.Context{})
	fmt.Printf("  parse error -> %v\n", err)

	_, err = eval.Evaluate("missing + 1", cexpr.Context{})
	fmt.Printf("  eval error  -> %v\n", err)
}
