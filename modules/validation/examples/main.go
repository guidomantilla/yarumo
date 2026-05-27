// Package main demonstrates the config-driven validation engine: a YAML
// ruleset is loaded once and applied to multiple target objects with
// different per-request contexts.
package main

import (
	"fmt"
	"os"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"

	"github.com/guidomantilla/yarumo/validation"
)

// Pokemon is a sample domain type validated by this example.
type Pokemon struct {
	ID    string
	Name  string
	Email string
	Phone string
	Level int
	Owner Owner
}

// Owner mirrors a nested object so the nested fixture has a target.
type Owner struct {
	Email string
	Tags  []string
}

func main() {
	rulesetPath := "fixtures/conditional.yaml"
	if len(os.Args) > 1 {
		rulesetPath = os.Args[1]
	}

	data, err := os.ReadFile(rulesetPath)
	if err != nil {
		fmt.Printf("read ruleset: %v\n", err)

		return
	}

	rs, err := validation.LoadYAML(data)
	if err != nil {
		fmt.Printf("load ruleset: %v\n", err)

		return
	}

	eng := validation.NewEngine(rs)

	demoCase(eng, "POST without ID", Pokemon{Name: "Pikachu"}, map[string]any{"method": "POST"})
	demoCase(eng, "POST with ID", Pokemon{ID: "abc", Name: "Pikachu"}, map[string]any{"method": "POST"})
	demoCase(eng, "GET with valid UUID",
		Pokemon{ID: "550e8400-e29b-41d4-a716-446655440000"},
		map[string]any{"method": "GET"})
	demoCase(eng, "GET with invalid UUID", Pokemon{ID: "nope"}, map[string]any{"method": "GET"})
	demoCase(eng, "CO phone OK", Pokemon{Phone: "+571234567890"}, map[string]any{"country": "CO"})
	demoCase(eng, "CO phone bad", Pokemon{Phone: "abc"}, map[string]any{"country": "CO"})
}

// demoCase runs the engine once and prints a labelled outcome.
func demoCase(eng validation.Engine, label string, p Pokemon, ctx map[string]any) {
	fmt.Println("===", label, "===")

	err := eng.Validate(p, ctx)
	if err == nil {
		fmt.Println("  OK")

		return
	}

	for _, info := range cerrs.AsErrorInfo(err) {
		fmt.Printf("  type=%s\n", info.Type)

		for _, msg := range info.Messages {
			fmt.Printf("    - %s\n", msg)
		}
	}
}
