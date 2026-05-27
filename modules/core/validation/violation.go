package validation

// Violation is the structured result of a single rule failure. It carries
// the offending field path, the rule that fired, the parameters that were
// applied, a human-readable message, and the underlying cause.
//
// Engine.Run returns a slice of Violation so callers building UIs or
// telemetry pipelines can iterate, filter, or aggregate without parsing
// error strings.
type Violation struct {
	// Path is the dotted field path that produced the violation, or empty
	// when the violation is at the root level.
	Path string

	// Rule is the registered rule name (leaf) that failed, or empty when
	// the violation comes from the engine itself (unknown rule, bad when:,
	// bad params).
	Rule string

	// Params is the positional parameter list passed to the failing rule,
	// preserved so callers can render the constraint that was checked
	// (e.g. "min_len(3)").
	Params []any

	// Message is a short human-readable description of the failure,
	// suitable for UI display.
	Message string

	// Cause is the underlying error (typically a sentinel from
	// common/validation or a domain *Error). Callers can errors.Is /
	// errors.As against it.
	Cause error
}
