package examples

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"

	"github.com/guidomantilla/yarumo/inference/classical/engine"
	"github.com/guidomantilla/yarumo/inference/classical/rules"
)

func buildChainRules(n int) ([]rules.Rule, logic.Fact) {
	ruleSet := make([]rules.Rule, 0, n)

	for i := range n {
		condVar := logic.Var(varName(i))
		concVar := logic.Var(varName(i + 1))
		name := "r" + varName(i)
		r := rules.NewRule(name, condVar, map[logic.Var]bool{concVar: true})
		ruleSet = append(ruleSet, r)
	}

	return ruleSet, logic.Fact{logic.Var(varName(0)): true}
}

func varName(i int) string {
	return string(rune('A' + i%26))
}

func BenchmarkForward10(b *testing.B) {
	ruleSet, initial := buildChainRules(10)
	e := engine.NewEngine()

	b.ResetTimer()

	for b.Loop() {
		e.Forward(initial, ruleSet)
	}
}

func BenchmarkForward50(b *testing.B) {
	ruleSet, initial := buildChainRules(25)
	e := engine.NewEngine()

	b.ResetTimer()

	for b.Loop() {
		e.Forward(initial, ruleSet)
	}
}

func BenchmarkForward100(b *testing.B) {
	ruleSet, initial := buildChainRules(25)
	e := engine.NewEngine()

	b.ResetTimer()

	for b.Loop() {
		e.Forward(initial, ruleSet)
	}
}

func BenchmarkBackward(b *testing.B) {
	ruleSet, initial := buildChainRules(10)
	e := engine.NewEngine()
	goal := logic.Var(varName(10))

	b.ResetTimer()

	for b.Loop() {
		e.Backward(initial, ruleSet, goal)
	}
}
