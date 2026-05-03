package main

import (
	"fmt"
	"time"

	"github.com/guidomantilla/yarumo/compute/math/fuzzy"
	"github.com/guidomantilla/yarumo/compute/math/logic"
	"github.com/guidomantilla/yarumo/compute/math/logic/entailment"
	"github.com/guidomantilla/yarumo/compute/math/logic/parser"
	"github.com/guidomantilla/yarumo/compute/math/logic/predicate"
	"github.com/guidomantilla/yarumo/compute/math/logic/sat"
	"github.com/guidomantilla/yarumo/compute/math/logic/temporal"
	"github.com/guidomantilla/yarumo/compute/math/sets"
	"github.com/guidomantilla/yarumo/compute/math/stats"
)

func main() {
	fmt.Println("=== TUTORIAL: inference/math ===")
	fmt.Println()

	//tutorialLogic()
	//tutorialParser()
	//tutorialTransformations()
	//tutorialSAT()
	//tutorialEntailment()
	//tutorialPredicate()
	//tutorialTemporal()
	//tutorialFuzzy()
	//tutorialStats()
	//tutorialSets()
}

func tutorialLogic() {
	fmt.Println("--- 1. Lógica Proposicional ---")

	// Variables
	llueve := logic.Var("llueve")
	paraguas := logic.Var("paraguas")
	mojado := logic.Var("mojado")

	// Fórmula: llueve AND NOT paraguas => mojado
	regla := logic.ImplF{
		L: logic.AndF{L: llueve, R: logic.NotF{F: paraguas}},
		R: mojado,
	}

	// Evaluar con hechos
	hechos := logic.Fact{llueve: true, paraguas: false, mojado: true}
	fmt.Printf("  Regla: %s\n", regla)
	fmt.Printf("  Hechos: llueve=true, paraguas=false, mojado=true\n")
	fmt.Printf("  Resultado: %v\n", regla.Eval(hechos))

	// Caso donde falla
	hechos2 := logic.Fact{llueve: true, paraguas: false, mojado: false}
	fmt.Printf("  Con mojado=false: %v\n", regla.Eval(hechos2))

	// Variables de la fórmula
	fmt.Printf("  Variables: %v\n", regla.Vars())

	// Tautología y contradicción
	taut := logic.OrF{L: llueve, R: logic.NotF{F: llueve}}
	fmt.Printf("  A OR NOT A es tautología: %v\n", logic.IsTautology(taut))

	contra := logic.AndF{L: llueve, R: logic.NotF{F: llueve}}
	fmt.Printf("  A AND NOT A es contradicción: %v\n", logic.IsContradiction(contra))

	// Equivalencia
	impl := logic.ImplF{L: llueve, R: paraguas}
	equiv := logic.OrF{L: logic.NotF{F: llueve}, R: paraguas}
	fmt.Printf("  (A => B) ≡ (¬A ∨ B): %v\n", logic.Equivalent(impl, equiv))

	// Tabla de verdad
	simple := logic.AndF{L: llueve, R: paraguas}
	rows := logic.TruthTable(simple)
	fmt.Printf("  Tabla de verdad de (llueve AND paraguas): %d filas\n", len(rows))

	// FailCases
	gaps := logic.FailCases(simple)
	fmt.Printf("  Casos donde falla: %d combinaciones\n", len(gaps))

	fmt.Println()
}

func tutorialParser() {
	fmt.Println("--- 2. Parser ---")

	// Parsear texto a fórmula
	f, err := parser.Parse("llueve AND viento => paraguas")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  Parseado: %s\n", f)

	// Diferentes notaciones
	f2, _ := parser.Parse("(A ∧ B) → C")
	fmt.Printf("  Unicode: %s\n", f2)

	f3, _ := parser.Parse("NOT A OR B")
	fmt.Printf("  Palabras: %s\n", f3)

	// MustParse (panic si falla)
	f4 := parser.MustParse("A <=> B")
	fmt.Printf("  Bicondicional: %s\n", f4)

	// Evaluar fórmula parseada
	hechos := logic.Fact{
		logic.Var("llueve"):   true,
		logic.Var("viento"):   true,
		logic.Var("paraguas"): true,
	}
	fmt.Printf("  Evaluar %q con hechos: %v\n", f, f.Eval(hechos))

	fmt.Println()
}

func tutorialTransformations() {
	fmt.Println("--- 3. Transformaciones ---")

	f := parser.MustParse("A => (B => C)")
	fmt.Printf("  Original: %s\n", logic.Format(f))

	// NNF: negaciones solo en variables
	nnf := logic.ToNNF(f)
	fmt.Printf("  NNF: %s\n", logic.Format(nnf))

	// CNF: AND de ORs (útil para SAT solvers)
	cnf := logic.ToCNF(f)
	fmt.Printf("  CNF: %s\n", logic.Format(cnf))

	// DNF: OR de ANDs
	dnf := logic.ToDNF(f)
	fmt.Printf("  DNF: %s\n", logic.Format(dnf))

	// Simplificación
	complejo := parser.MustParse("A AND (A OR B)")
	simple := logic.Simplify(complejo)
	fmt.Printf("  Simplificar %q → %q\n", logic.Format(complejo), logic.Format(simple))

	fmt.Println()
}

func tutorialSAT() {
	fmt.Println("--- 4. SAT Solver (DPLL) ---")

	// Registrar el solver
	logic.RegisterSATSolver(sat.Solver())

	// Buscar asignación satisfactoria
	f := parser.MustParse("(A OR B) AND (NOT A OR C) AND (NOT B OR C)")
	asignacion, found := logic.FindSatisfyingAssignment(f)
	fmt.Printf("  Fórmula: %s\n", f)
	fmt.Printf("  Satisfacible: %v\n", found)
	if found {
		fmt.Printf("  Asignación: %v\n", asignacion)
	}

	// Fórmula insatisfacible
	f2 := parser.MustParse("A AND NOT A")
	_, found2 := logic.FindSatisfyingAssignment(f2)
	fmt.Printf("  A AND NOT A satisfacible: %v\n", found2)

	fmt.Println()
}

func tutorialEntailment() {
	fmt.Println("--- 5. Entailment ---")

	// Premisas: llueve => mojado, llueve
	// Conclusión: mojado
	premises := []logic.Formula{
		parser.MustParse("llueve => mojado"),
		parser.MustParse("llueve"),
	}
	conclusion := parser.MustParse("mojado")

	result := entailment.Entails(premises, conclusion)
	fmt.Printf("  {llueve=>mojado, llueve} ⊨ mojado: %v\n", result)

	// Con contramodelo
	premises2 := []logic.Formula{
		parser.MustParse("A => B"),
	}
	conclusion2 := parser.MustParse("B")

	result2, counter := entailment.EntailsWithCounterModel(premises2, conclusion2)
	fmt.Printf("  {A=>B} ⊨ B: %v\n", result2)
	if !result2 {
		fmt.Printf("  Contramodelo: %v\n", counter)
	}

	fmt.Println()
}

func tutorialPredicate() {
	fmt.Println("--- 6. Lógica de Predicados ---")

	// Colección de perros
	perros := predicate.Collection{
		logic.Fact{logic.Var("grande"): true, logic.Var("vacunado"): true},
		logic.Fact{logic.Var("grande"): false, logic.Var("vacunado"): true},
		logic.Fact{logic.Var("grande"): true, logic.Var("vacunado"): false},
	}

	// ForAll: ¿todos los grandes están vacunados?
	regla := parser.MustParse("grande => vacunado")
	todos, _ := predicate.ForAll(perros, regla)
	fmt.Printf("  ∀ perro: grande => vacunado: %v\n", todos)

	// Exists: ¿existe al menos un grande vacunado?
	existe, _ := predicate.Exists(perros, parser.MustParse("grande AND vacunado"))
	fmt.Printf("  ∃ perro: grande AND vacunado: %v\n", existe)

	// Count: ¿cuántos están vacunados?
	n, _ := predicate.Count(perros, parser.MustParse("vacunado"))
	fmt.Printf("  Vacunados: %d de %d\n", n, len(perros))

	// Filter: obtener solo los vacunados
	vacunados, _ := predicate.Filter(perros, parser.MustParse("vacunado"))
	fmt.Printf("  Filtrados (vacunados): %d elementos\n", len(vacunados))

	fmt.Println()
}

func tutorialTemporal() {
	fmt.Println("--- 7. Lógica Temporal ---")

	t0 := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	trace := temporal.Trace{
		{Label: "pedido_recibido", Time: t0},
		{Label: "pago_procesado", Time: t0.Add(2 * time.Minute)},
		{Label: "envio_creado", Time: t0.Add(10 * time.Minute)},
		{Label: "entregado", Time: t0.Add(48 * time.Hour)},
	}

	// ResponseWithin: ¿envío creado dentro de 30 min del pedido?
	ok, violations := temporal.ResponseWithin(trace, "pedido_recibido", "envio_creado", 30*time.Minute)
	fmt.Printf("  Envío en <30min: %v (violaciones: %d)\n", ok, len(violations))

	// Eventually: ¿se entregó?
	entregado := temporal.Eventually(trace, "entregado")
	fmt.Printf("  Eventually(entregado): %v\n", entregado)

	// Before: ¿pago antes de envío?
	antes := temporal.Before(trace, "pago_procesado", "envio_creado")
	fmt.Printf("  Before(pago, envío): %v\n", antes)

	// Elapsed: tiempo entre pedido y entrega
	dur, _ := temporal.Elapsed(trace, "pedido_recibido", "entregado")
	fmt.Printf("  Tiempo pedido→entrega: %s\n", dur)

	// Sequence: ¿siguen el orden esperado?
	seq := temporal.Sequence(trace, "pedido_recibido", "pago_procesado", "envio_creado", "entregado")
	fmt.Printf("  Secuencia correcta: %v\n", seq)

	fmt.Println()
}

func tutorialFuzzy() {
	fmt.Println("--- 8. Fuzzy ---")

	// Funciones de membresía
	baja, _ := fuzzy.Triangular(0, 0, 50)
	media, _ := fuzzy.Triangular(20, 50, 80)
	alta, _ := fuzzy.Triangular(50, 100, 100)

	// Fuzzificar un valor
	valor := 35.0
	fmt.Printf("  Temperatura: %.0f°C\n", valor)
	fmt.Printf("    Grado 'baja':  %.2f\n", fuzzy.Fuzzify(baja, valor))
	fmt.Printf("    Grado 'media': %.2f\n", fuzzy.Fuzzify(media, valor))
	fmt.Printf("    Grado 'alta':  %.2f\n", fuzzy.Fuzzify(alta, valor))

	// T-norms (AND difuso)
	a, b := fuzzy.Degree(0.7), fuzzy.Degree(0.4)
	fmt.Printf("  AND(0.7, 0.4): Min=%.2f, Product=%.2f\n",
		fuzzy.Min(a, b), fuzzy.Product(a, b))

	// T-conorms (OR difuso)
	fmt.Printf("  OR(0.7, 0.4):  Max=%.2f, ProbSum=%.2f\n",
		fuzzy.Max(a, b), fuzzy.ProbabilisticSum(a, b))

	// Complemento (NOT difuso)
	fmt.Printf("  NOT(0.7): %.2f\n", fuzzy.Complement(a))

	// Defuzzificación
	combined := fuzzy.AggregateMax(baja, media)
	xs, ys, _ := fuzzy.Sample(combined, 0, 100, 200)
	centroid := fuzzy.Centroid(xs, ys)
	fmt.Printf("  Centroid(baja ∪ media): %.2f\n", centroid)

	fmt.Println()
}

func tutorialStats() {
	fmt.Println("--- 9. Estadística ---")

	// Estadística descriptiva
	data := []float64{23, 45, 12, 67, 34, 56, 78}
	mean, _ := stats.Mean(data)
	median, _ := stats.Median(data)
	stddev, _ := stats.StdDev(data)
	fmt.Printf("  Data: %v\n", data)
	fmt.Printf("  Media: %.2f, Mediana: %.2f, StdDev: %.2f\n", mean, median, stddev)

	p75, _ := stats.Percentile(data, 75)
	minVal, _ := stats.Min(data)
	maxVal, _ := stats.Max(data)
	fmt.Printf("  P75: %.2f, Min: %.2f, Max: %.2f\n", p75, minVal, maxVal)

	// Correlación
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{2, 4, 5, 4, 5}
	corr, _ := stats.Correlation(x, y)
	slope, intercept, _ := stats.LinearRegression(x, y)
	r2, _ := stats.RSquared(x, y)
	fmt.Printf("  Correlación: %.3f, Regresión: y = %.2fx + %.2f (R²=%.3f)\n",
		corr, slope, intercept, r2)

	// Distribución normal
	n, _ := stats.NewNormal(100, 15) // IQ distribution
	fmt.Printf("  Normal(μ=100, σ=15):\n")
	fmt.Printf("    P(X ≤ 115) = %.4f\n", n.CDF(115))
	fmt.Printf("    P(X ≤ 85)  = %.4f\n", n.CDF(85))
	fmt.Printf("    Quantile(0.95) = %.2f\n", n.Quantile(0.95))

	// Bayes
	posterior, _ := stats.Bayes(
		stats.Prob(0.01), // prior: P(enfermo)
		stats.Prob(0.95), // likelihood: P(test+|enfermo)
		stats.Prob(0.05), // evidence: P(test+)
	)
	fmt.Printf("  Bayes — P(enfermo|test+) = %.4f\n", posterior)

	// Distribución discreta
	dist := stats.Distribution{"sí": 0.7, "no": 0.3}
	entropy, _ := stats.Entropy(dist)
	fmt.Printf("  Distribución {sí:0.7, no:0.3} — Entropía: %.4f bits\n", entropy)

	// Running stats (streaming)
	var rs stats.RunningStats
	for _, v := range data {
		rs.Push(v)
	}
	fmt.Printf("  RunningStats: count=%d, mean=%.2f, var=%.2f\n",
		rs.Count(), rs.Mean(), rs.Variance())

	// T-test
	t, p, _ := stats.TTest(data, 50.0)
	fmt.Printf("  T-test (μ₀=50): t=%.3f, p=%.4f\n", t, p)

	fmt.Println()
}

func tutorialSets() {
	fmt.Println("--- 10. Conjuntos ---")

	a := sets.New("perro", "gato", "pez")
	b := sets.New("gato", "pájaro")

	fmt.Printf("  A = %v\n", a.Items())
	fmt.Printf("  B = %v\n", b.Items())
	fmt.Printf("  A ∪ B = %v\n", sets.Union(a, b).Items())
	fmt.Printf("  A ∩ B = %v\n", sets.Intersection(a, b).Items())
	fmt.Printf("  A \\ B = %v\n", sets.Difference(a, b).Items())
	fmt.Printf("  A △ B = %v\n", sets.SymmetricDifference(a, b).Items())
	fmt.Printf("  B ⊆ A: %v\n", sets.IsSubset(b, a))
	fmt.Printf("  A contains 'perro': %v\n", a.Contains("perro"))

	fmt.Println()
}
