package validation

import (
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
	cvalidation "github.com/guidomantilla/yarumo/core/common/validation"
)

// Expand walks the ruleset tree and replaces every node with Use="name"
// by the rules registered under rs.Defines[name]. Cycles (a define that
// references itself, directly or transitively) and undefined references
// produce errors aggregated into a single domain error.
//
// Expand returns a new Ruleset; rs is not mutated. After expansion the
// returned ruleset is safe to pass to NewEngine — the engine itself never
// sees Use or Defines.
func Expand(rs Ruleset) (Ruleset, error) {
	if len(rs.Defines) == 0 {
		// Still walk the tree to catch dangling Use references that have
		// no matching define.
		out, err := expandNodes(rs.Rules, rs.Defines, nil)
		if err != nil {
			return Ruleset{}, err
		}

		return Ruleset{Version: rs.Version, Rules: out}, nil
	}

	out, err := expandNodes(rs.Rules, rs.Defines, nil)
	if err != nil {
		return Ruleset{}, err
	}

	return Ruleset{Version: rs.Version, Rules: out}, nil
}

// expandNodes walks a slice of RuleNode and expands every Use reference,
// honouring stack to detect cycles.
func expandNodes(nodes []RuleNode, defines map[string][]RuleNode, stack []string) ([]RuleNode, error) {
	out := make([]RuleNode, 0, len(nodes))

	for i := range nodes {
		expanded, err := expandNode(nodes[i], defines, stack)
		if err != nil {
			return nil, err
		}

		out = append(out, expanded...)
	}

	return out, nil
}

// expandNode expands one node. A Use reference produces a slice of
// substituted rules (potentially many). A non-Use node returns itself with
// its nested Rules expanded.
func expandNode(n RuleNode, defines map[string][]RuleNode, stack []string) ([]RuleNode, error) {
	if n.Use != "" {
		return expandUse(n, defines, stack)
	}

	if len(n.Rules) == 0 {
		return []RuleNode{n}, nil
	}

	children, err := expandNodes(n.Rules, defines, stack)
	if err != nil {
		return nil, err
	}

	clone := n
	clone.Rules = children

	return []RuleNode{clone}, nil
}

// expandUse substitutes a Use reference with the matching define.
func expandUse(n RuleNode, defines map[string][]RuleNode, stack []string) ([]RuleNode, error) {
	for _, ref := range stack {
		if ref == n.Use {
			return nil, ErrEngine(cerrs.Wrap(ErrCycleDetected, errUseRef(n.Use)))
		}
	}

	body, ok := defines[n.Use]
	if !ok {
		return nil, ErrEngine(cerrs.Wrap(ErrUndefinedUse, errUseRef(n.Use)))
	}

	expanded, err := expandNodes(body, defines, append(stack, n.Use))
	if err != nil {
		return nil, err
	}

	// If the calling node has Field/When/Optional/Message wrap the expanded
	// rules into a group so those modifiers apply uniformly.
	if n.Field != "" || n.When != "" || n.Optional || n.Message != "" {
		wrapped := n
		wrapped.Use = ""
		wrapped.Rules = expanded

		return []RuleNode{wrapped}, nil
	}

	return expanded, nil
}

// RulesetFor[T] verifies that every field: path in rs resolves against T
// at compile time. Returns rs unchanged on success, or an error listing
// every bad path.
//
// Validation walks the ruleset tree; nested groups apply paths relative to
// their parent field, matching the runtime engine semantics. Slice indices
// and map accesses pass through unchanged (their structure is verified at
// runtime when GetField actually walks the value).
func RulesetFor[T any](rs Ruleset) (Ruleset, error) {
	var zero T

	err := bindNodes(rs.Rules, "", zero)
	if err != nil {
		return Ruleset{}, err
	}

	return rs, nil
}

// Bind is the dynamic counterpart of RulesetFor: callers pass a sample
// value when the type is not known at compile time.
func Bind(rs Ruleset, sample any) (Ruleset, error) {
	err := bindNodes(rs.Rules, "", sample)
	if err != nil {
		return Ruleset{}, err
	}

	return rs, nil
}

// bindNodes walks a slice of nodes, joining nested field paths and
// verifying each via GetField against the sample value.
func bindNodes(nodes []RuleNode, parent string, sample any) error {
	var causes []error

	for i := range nodes {
		errs := bindNode(nodes[i], parent, sample)
		causes = append(causes, errs...)
	}

	if len(causes) == 0 {
		return nil
	}

	return ErrEngine(cvalidation.ErrValidation(causes...))
}

// bindNode verifies a single node's field path and recurses into nested
// rules with the joined path.
func bindNode(n RuleNode, parent string, sample any) []error {
	var causes []error

	path := parent
	if n.Field != "" {
		path = joinPath(parent, n.Field)

		_, err := cvalidation.GetField(sample, path)
		if err != nil {
			causes = append(causes, cerrs.Wrap(ErrUnknownField, errFieldPath(path)))
		}
	}

	for i := range n.Rules {
		errs := bindNode(n.Rules[i], path, sample)
		causes = append(causes, errs...)
	}

	return causes
}

// useRefError carries the offending Use reference name for AsErrorInfo.
type useRefError struct {
	name string
}

// Error returns the use-reference message.
func (u *useRefError) Error() string {
	return "use reference: " + u.name
}

// errUseRef creates a leaf error tagged with the offending use name.
func errUseRef(name string) error {
	return &useRefError{name: name}
}

// fieldPathError carries the offending field path for AsErrorInfo.
type fieldPathError struct {
	path string
}

// Error returns the field-path message.
func (f *fieldPathError) Error() string {
	return "field path: " + f.path
}

// errFieldPath creates a leaf error tagged with the offending field path.
func errFieldPath(path string) error {
	return &fieldPathError{path: path}
}
