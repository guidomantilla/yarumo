package examples

import (
	"os"
	"testing"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic2/props"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic2/sat"
)

// TestMain explicitly registers the SAT solver for examples/tests.
// This avoids package-level side effects in production code while
// allowing tests to exercise the SAT path when desired.
func TestMain(m *testing.M) {
	props.RegisterSATSolver(sat.Solver)
	// Optionally adjust threshold here if needed for tests:
	// props.SATThreshold = 12
	code := m.Run()
	os.Exit(code)
}
