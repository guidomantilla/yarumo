package explain

import "text/template"

// deductiveTemplates maps locale to the deductive explanation template.
var deductiveTemplates = map[Locale]*template.Template{
	English: template.Must(template.New("deductive-en").Parse(
		`Decision reached in {{.Steps}} step(s).{{range .Reasons}} {{.Variable}}={{.Value}} (rule: {{.RuleName}}, step {{.Step}}).{{end}}`)),
	Spanish: template.Must(template.New("deductive-es").Parse(
		`Decision alcanzada en {{.Steps}} paso(s).{{range .Reasons}} {{.Variable}}={{.Value}} (regla: {{.RuleName}}, paso {{.Step}}).{{end}}`)),
}

// bayesianTemplates maps locale to the Bayesian explanation template.
var bayesianTemplates = map[Locale]*template.Template{
	English: template.Must(template.New("bayesian-en").Parse(
		`Posterior for {{.Query}}:{{range .Factors}} {{.Outcome}}={{printf "%.4f" .Probability}}{{end}}.`)),
	Spanish: template.Must(template.New("bayesian-es").Parse(
		`Posterior para {{.Query}}:{{range .Factors}} {{.Outcome}}={{printf "%.4f" .Probability}}{{end}}.`)),
}

// fuzzyTemplates maps locale to the fuzzy explanation template.
var fuzzyTemplates = map[Locale]*template.Template{
	English: template.Must(template.New("fuzzy-en").Parse(
		`Fuzzy outputs:{{range .Outputs}} {{.Variable}}={{printf "%.4f" .Value}}{{end}}.{{if .Memberships}} Memberships:{{range .Memberships}} {{.Variable}}/{{.Term}}={{printf "%.4f" .Degree}}{{end}}.{{end}}`)),
	Spanish: template.Must(template.New("fuzzy-es").Parse(
		`Salidas fuzzy:{{range .Outputs}} {{.Variable}}={{printf "%.4f" .Value}}{{end}}.{{if .Memberships}} Membresias:{{range .Memberships}} {{.Variable}}/{{.Term}}={{printf "%.4f" .Degree}}{{end}}.{{end}}`)),
}

// tableTemplates maps locale to the decision table explanation template.
var tableTemplates = map[Locale]*template.Template{
	English: template.Must(template.New("table-en").Parse(
		`Table ({{.HitPolicy}}): {{len .MatchedRules}} rule(s) matched.{{range .MatchedRules}} {{.RuleName}}.{{end}}`)),
	Spanish: template.Must(template.New("table-es").Parse(
		`Tabla ({{.HitPolicy}}): {{len .MatchedRules}} regla(s) coincidieron.{{range .MatchedRules}} {{.RuleName}}.{{end}}`)),
}

// scorecardTemplates maps locale to the scorecard explanation template.
var scorecardTemplates = map[Locale]*template.Template{
	English: template.Must(template.New("scorecard-en").Parse(
		`Score: {{printf "%.2f" .TotalScore}} (base: {{printf "%.2f" .BaseScore}}).{{range .Breakdown}} {{.Attribute}}: {{printf "%.2f" .Weighted}}pts.{{end}}`)),
	Spanish: template.Must(template.New("scorecard-es").Parse(
		`Puntaje: {{printf "%.2f" .TotalScore}} (base: {{printf "%.2f" .BaseScore}}).{{range .Breakdown}} {{.Attribute}}: {{printf "%.2f" .Weighted}}pts.{{end}}`)),
}

// treeTemplates maps locale to the decision tree explanation template.
var treeTemplates = map[Locale]*template.Template{
	English: template.Must(template.New("tree-en").Parse(
		`Tree decision:{{range .Path}} {{.Condition}}={{.Result}}{{end}}.`)),
	Spanish: template.Must(template.New("tree-es").Parse(
		`Decision de arbol:{{range .Path}} {{.Condition}}={{.Result}}{{end}}.`)),
}
