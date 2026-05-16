package expressions

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
