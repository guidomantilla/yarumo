package inlineassign

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

)

// TestAnalyzer is the smoke test for the inlineassign analyzer. It runs
// analysistest.Run against testdata/src/a, which contains one example of every
// forbidden form (error-check, map-lookup, type-assertion, arbitrary init) plus
// a compliant counterexample. The // want directives encode the expected
// diagnostics, so the test passes only when the analyzer reports exactly the
// forbidden lines and stays silent on the compliant function.
func TestAnalyzer(t *testing.T) {
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), Analyzer, "a")
}
