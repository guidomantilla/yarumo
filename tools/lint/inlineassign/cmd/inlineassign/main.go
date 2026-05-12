// Command inlineassign runs the inlineassign analyzer as a standalone tool.
//
// Invoke it directly:
//
//	inlineassign ./...
//
// or via go vet:
//
//	go vet -vettool=$(which inlineassign) ./...
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/guidomantilla/yarumo/tools/lint/inlineassign"
)

func main() { singlechecker.Main(inlineassign.Analyzer) }
