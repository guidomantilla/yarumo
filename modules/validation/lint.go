package validation

import (
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
	cexpressions "github.com/guidomantilla/yarumo/core/common/expressions"
	cvalidation "github.com/guidomantilla/yarumo/core/common/validation"
)

// Validate is the static linter for a Ruleset. It walks every node and
// aggregates every structural issue it finds: unknown rule names (against
// the registry from the options), malformed nodes (group + leaf mixed),
// empty groups, when: expressions that fail to parse, and unknown schema
// versions. The returned error wraps every cause and is intended for
// fail-at-boot wiring via WithLintOnLoad.
//
// Validate does not execute any rule or evaluate any expression — it only
// performs static checks.
func Validate(rs Ruleset, opts ...Option) error {
	options := NewOptions(opts...)

	var causes []error

	switch {
	case rs.Version == "" && options.strictVersion:
		causes = append(causes, ErrUnknownVersion)
	case rs.Version != "" && rs.Version != CurrentVersion:
		causes = append(causes, cerrs.Wrap(ErrUnknownVersion, errUnknownVersion(rs.Version)))
	}

	for i := range rs.Rules {
		causes = append(causes, lintNode(rs.Rules[i], options)...)
	}

	if len(causes) == 0 {
		return nil
	}

	return ErrEngine(cerrs.Wrap(ErrLintFailed, cvalidation.ErrValidation(causes...)))
}

// lintNode walks one RuleNode and returns every structural issue found.
func lintNode(n RuleNode, options *Options) []error {
	causes := lintNodeShape(n)
	causes = append(causes, lintNodeWhen(n)...)
	causes = append(causes, lintNodeName(n, options)...)

	for i := range n.Rules {
		causes = append(causes, lintNode(n.Rules[i], options)...)
	}

	return causes
}

// lintNodeShape verifies a node is either a leaf, a non-empty group, or a
// combinator with at least one nested rule.
func lintNodeShape(n RuleNode) []error {
	if isCombinator(n.Name) {
		if len(n.Rules) == 0 {
			return []error{ErrEmptyGroup}
		}

		return nil
	}

	hasGroupFields := n.Field != "" || n.When != "" || len(n.Rules) > 0
	hasLeafFields := n.Name != ""

	switch {
	case hasGroupFields && hasLeafFields:
		return []error{ErrMixedShape}
	case !hasGroupFields && !hasLeafFields:
		return []error{ErrEmptyGroup}
	case hasGroupFields && len(n.Rules) == 0:
		return []error{ErrEmptyGroup}
	}

	return nil
}

// lintNodeWhen verifies a when: expression parses.
func lintNodeWhen(n RuleNode) []error {
	if n.When == "" {
		return nil
	}

	_, err := cexpressions.Parse(n.When)
	if err == nil {
		return nil
	}

	return []error{cerrs.Wrap(ErrWhenParseFailed, err)}
}

// lintNodeName verifies the leaf name is either a registered rule or one
// of the built-in combinator pseudo-leaves.
func lintNodeName(n RuleNode, options *Options) []error {
	if n.Name == "" || isCombinator(n.Name) {
		return nil
	}

	_, ok := options.registry.Get(n.Name)
	if ok {
		return nil
	}

	return []error{cerrs.Wrap(ErrUnknownRule, errUnknownRuleName(n.Name))}
}

// toViolations converts an engine-produced error tree into a flat slice of
// Violation values. The conversion walks errs.Unwrap recursively, extracting
// the path (via pathError) and rule name (via unknownRuleError) when
// present.
func toViolations(err error) []Violation {
	if err == nil {
		return nil
	}

	var out []Violation
	walkViolations(err, "", "", &out)

	if len(out) == 0 {
		out = append(out, Violation{Message: err.Error(), Cause: err})
	}

	return out
}

// walkViolations recurses through the unwrapped causes of err, accumulating
// a path/rule prefix as it goes and emitting one Violation per leaf cause.
func walkViolations(err error, path, rule string, out *[]Violation) {
	causes := cerrs.Unwrap(err)
	if len(causes) == 0 {
		*out = append(*out, Violation{Path: path, Rule: rule, Message: err.Error(), Cause: err})

		return
	}

	for _, cause := range causes {
		info := extractMarker(cause)
		if info.ok {
			if info.path != "" {
				path = info.path
			}

			if info.rule != "" {
				rule = info.rule
			}

			continue
		}

		walkViolations(cause, path, rule, out)
	}
}

// markerInfo carries the metadata extracted from a marker error (path
// from pathError, rule name from unknownRuleError).
type markerInfo struct {
	path string
	rule string
	ok   bool
}

// extractMarker reports whether err is one of the marker errors carrying
// metadata (path or rule name) and returns the extracted values.
func extractMarker(err error) markerInfo {
	var pe *pathError

	if asMarker(err, &pe) {
		return markerInfo{path: pe.path, ok: true}
	}

	var ue *unknownRuleError

	if asMarker(err, &ue) {
		return markerInfo{rule: ue.name, ok: true}
	}

	return markerInfo{}
}

// asMarker is a tiny wrapper around errors.As that takes a generic target.
func asMarker[T any](err error, target *T) bool {
	v, ok := any(err).(T)
	if ok {
		*target = v

		return true
	}

	return false
}

// unknownVersionError carries the offending version string for AsErrorInfo.
type unknownVersionError struct {
	version string
}

// Error returns the formatted unknown-version message.
func (u *unknownVersionError) Error() string {
	return "unknown version: " + u.version
}

// errUnknownVersion creates a leaf error tagged with the offending version.
func errUnknownVersion(version string) error {
	return &unknownVersionError{version: version}
}
