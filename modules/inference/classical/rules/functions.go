package rules

import (
	"slices"
	"sort"

	"github.com/guidomantilla/yarumo/maths/logic"
)

// SortByPriority returns a copy of the rules sorted by priority ascending (stable).
func SortByPriority(ruleSet []Rule) []Rule {
	sorted := make([]Rule, len(ruleSet))
	copy(sorted, ruleSet)

	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Priority() < sorted[j].Priority()
	})

	return sorted
}

// Variables returns all variables referenced by the rule (condition + conclusion), sorted and deduplicated.
func Variables(r Rule) []logic.Var {
	vars := r.Condition().Vars()
	conclusion := r.Conclusion()

	for v := range conclusion {
		vars = append(vars, v)
	}

	slices.Sort(vars)

	return slices.Compact(vars)
}
