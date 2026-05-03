package expressions

import (
	"math"
	"slices"
	"strconv"
	"strings"
)

// DefaultFuncs returns the built-in function registry.
func DefaultFuncs() map[string]Func {
	return map[string]Func{
		"len":      builtinLen,
		"sum":      builtinSum,
		"min":      builtinMin,
		"max":      builtinMax,
		"avg":      builtinAvg,
		"abs":      builtinAbs,
		"contains": builtinContains,
		"lower":    builtinLower,
		"upper":    builtinUpper,
	}
}

func builtinLen(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, ErrEval("len: expected 1 argument, got "+strconv.Itoa(len(args)), ErrArgCount)
	}

	switch v := args[0].(type) {
	case string:
		return float64(len(v)), nil
	case []any:
		return float64(len(v)), nil
	default:
		return nil, ErrEval("len: unsupported type "+formatValue(args[0]), ErrTypeMismatch)
	}
}

// requireSliceArg validates that a function received exactly one slice argument
// and converts all elements to float64.
func requireSliceArg(name string, args []any) ([]float64, error) {
	if len(args) != 1 {
		return nil, ErrEval(name+": expected 1 argument, got "+strconv.Itoa(len(args)), ErrArgCount)
	}

	s, ok := toSlice(args[0])
	if !ok {
		return nil, ErrEval(name+": expected slice, got "+formatValue(args[0]), ErrTypeMismatch)
	}

	nums := make([]float64, 0, len(s))

	for i, v := range s {
		n, nok := toFloat64(v)
		if !nok {
			return nil, ErrEval(name+": element "+strconv.Itoa(i)+" is not numeric", ErrTypeMismatch)
		}

		nums = append(nums, n)
	}

	return nums, nil
}

func builtinSum(args ...any) (any, error) {
	nums, err := requireSliceArg("sum", args)
	if err != nil {
		return nil, err
	}

	total := 0.0
	for _, n := range nums {
		total += n
	}

	return total, nil
}

func builtinMin(args ...any) (any, error) {
	nums, err := requireSliceArg("min", args)
	if err != nil {
		return nil, err
	}

	if len(nums) == 0 {
		return nil, ErrEval("min: empty slice", ErrArgCount)
	}

	result := nums[0]

	for _, n := range nums[1:] {
		if n < result {
			result = n
		}
	}

	return result, nil
}

func builtinMax(args ...any) (any, error) {
	nums, err := requireSliceArg("max", args)
	if err != nil {
		return nil, err
	}

	if len(nums) == 0 {
		return nil, ErrEval("max: empty slice", ErrArgCount)
	}

	result := nums[0]

	for _, n := range nums[1:] {
		if n > result {
			result = n
		}
	}

	return result, nil
}

func builtinAvg(args ...any) (any, error) {
	nums, err := requireSliceArg("avg", args)
	if err != nil {
		return nil, err
	}

	if len(nums) == 0 {
		return nil, ErrEval("avg: empty slice", ErrArgCount)
	}

	total := 0.0
	for _, n := range nums {
		total += n
	}

	return total / float64(len(nums)), nil
}

func builtinAbs(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, ErrEval("abs: expected 1 argument, got "+strconv.Itoa(len(args)), ErrArgCount)
	}

	n, ok := toFloat64(args[0])
	if !ok {
		return nil, ErrEval("abs: expected numeric, got "+formatValue(args[0]), ErrTypeMismatch)
	}

	return math.Abs(n), nil
}

func builtinContains(args ...any) (any, error) {
	if len(args) != 2 {
		return nil, ErrEval("contains: expected 2 arguments, got "+strconv.Itoa(len(args)), ErrArgCount)
	}

	switch collection := args[0].(type) {
	case string:
		substr, ok := toString(args[1])
		if !ok {
			return nil, ErrEval("contains: expected string argument, got "+formatValue(args[1]), ErrTypeMismatch)
		}
		return strings.Contains(collection, substr), nil
	case []any:
		return slices.Contains(collection, args[1]), nil
	default:
		return nil, ErrEval("contains: unsupported collection type "+formatValue(args[0]), ErrTypeMismatch)
	}
}

func builtinLower(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, ErrEval("lower: expected 1 argument, got "+strconv.Itoa(len(args)), ErrArgCount)
	}

	s, ok := toString(args[0])
	if !ok {
		return nil, ErrEval("lower: expected string, got "+formatValue(args[0]), ErrTypeMismatch)
	}

	return strings.ToLower(s), nil
}

func builtinUpper(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, ErrEval("upper: expected 1 argument, got "+strconv.Itoa(len(args)), ErrArgCount)
	}

	s, ok := toString(args[0])
	if !ok {
		return nil, ErrEval("upper: expected string, got "+formatValue(args[0]), ErrTypeMismatch)
	}

	return strings.ToUpper(s), nil
}
