// Package inlineassign provides a go/analysis analyzer that enforces the
// "No Inline Assignments" rule from modules/common/CODING_STANDARDS.md.
//
// The rule forbids combining assignment and condition in a single if-statement.
// The analyzer flags any *ast.IfStmt whose Init clause is non-nil, which covers
// the three forbidden forms:
//
//   - error-check:  if err := f(); err != nil { ... }
//   - map-lookup:   if v, ok := m[k]; ok { ... }
//   - type-assert:  if v, ok := x.(T); ok { ... }
//
// Any other non-nil Init clause is also flagged because the rule is uniform:
// the assignment must be split into its own statement preceding the if.
package inlineassign

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer is the go/analysis Analyzer that enforces the "No Inline Assignments"
// coding standard. Register it with singlechecker.Main (for go vet -vettool) or
// with golangci-lint's custom plugin mechanism.
var Analyzer = &analysis.Analyzer{
	Name:     "inlineassign",
	Doc:      "flags if-statements with a non-nil Init clause (No Inline Assignments rule)",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// run walks every *ast.IfStmt in the package and reports any whose Init clause
// is non-nil. The diagnostic points at the Init position so the editor highlight
// lands on the inline assignment itself rather than the if keyword.
func run(pass *analysis.Pass) (any, error) {
	insp, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if insp == nil {
		return nil, nil
	}

	insp.Preorder([]ast.Node{(*ast.IfStmt)(nil)}, func(n ast.Node) {
		stmt, _ := n.(*ast.IfStmt)
		if stmt == nil || stmt.Init == nil {
			return
		}
		pass.Reportf(
			stmt.Init.Pos(),
			"inline assignment in if-statement; split the assignment into its own statement before the if (No Inline Assignments rule)",
		)
	})

	return nil, nil
}
