package validation

import (
	"sort"
	"strings"
	"testing"
)

// expectedRules enumerates the rule names the default registry MUST expose.
// Adding a new leaf to common/validation/ should be followed by extending
// both the builtins map and this list so the test catches drift between
// the leaf catalogue and the engine's exposure. See YA-0130 (Option C).
var expectedRules = []string{
	// presence / required
	"required",
	"must_be_undefined",

	// string length / pattern
	"min_len",
	"max_len",
	"regex",

	// string content
	"contains",
	"has_prefix",
	"has_suffix",

	// string format
	"email",
	"url",
	"lowercase",
	"uppercase",
	"alpha",
	"alphanumeric",
	"numeric_string",
	"ascii",
	"hex",
	"base64",
	"trimmed",
	"jwt",
	"semver",
	"integer_string",
	"float_string",

	// unique identifier formats
	"uuid",
	"ulid",

	// network / transport
	"ip",
	"ipv4",
	"ipv6",
	"cidr",
	"mac",
	"hostname",
	"fqdn",
	"port",

	// date / time
	"rfc3339",
	"date_layout",
	"before",
	"after",
	"between_time",

	// numeric
	"min",
	"max",
	"in_range",
	"positive",
	"negative",
	"nonzero",
	"multiple_of",

	// equality / set
	"equal",
	"not_equal",
	"equal_ignore_case",
	"one_of",
	"not_in",

	// collection
	"non_empty",
	"min_count",
	"max_count",
	"count_in_range",
}

func TestBuiltins_CatalogueCoverage(t *testing.T) {
	t.Parallel()

	reg := DefaultRegistry()

	for _, name := range expectedRules {
		_, ok := reg.Get(name)
		if !ok {
			t.Errorf("expected builtin %q is missing from DefaultRegistry — wire it in builtins.go or remove from expectedRules", name)
		}
	}
}

func TestBuiltins_NoOrphanedRule(t *testing.T) {
	t.Parallel()

	reg := DefaultRegistry()

	expectedSet := make(map[string]struct{}, len(expectedRules))
	for _, name := range expectedRules {
		expectedSet[name] = struct{}{}
	}

	var extras []string
	for _, name := range reg.Names() {
		_, ok := expectedSet[name]
		if !ok {
			extras = append(extras, name)
		}
	}

	if len(extras) == 0 {
		return
	}

	sort.Strings(extras)
	t.Errorf("builtins exposes rules not declared in expectedRules: %s — add them to coverage_test.go", strings.Join(extras, ", "))
}
