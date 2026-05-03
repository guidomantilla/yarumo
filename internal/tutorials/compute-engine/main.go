package main

import (
	"fmt"

	"github.com/guidomantilla/yarumo/compute/math/fuzzy"
	"github.com/guidomantilla/yarumo/compute/math/logic"
	"github.com/guidomantilla/yarumo/compute/math/logic/parser"
	"github.com/guidomantilla/yarumo/compute/math/stats"

	"github.com/guidomantilla/yarumo/compute/engine/bayesian"
	bengine "github.com/guidomantilla/yarumo/compute/engine/bayesian/engine"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/network"
	"github.com/guidomantilla/yarumo/compute/engine/causal/engine"
	"github.com/guidomantilla/yarumo/compute/engine/causal/model"
	dengine "github.com/guidomantilla/yarumo/compute/engine/deductive/engine"
	"github.com/guidomantilla/yarumo/compute/engine/deductive/rules"
	fengine "github.com/guidomantilla/yarumo/compute/engine/fuzzy/engine"
	frules "github.com/guidomantilla/yarumo/compute/engine/fuzzy/rules"
	"github.com/guidomantilla/yarumo/compute/engine/fuzzy/variable"
	"github.com/guidomantilla/yarumo/compute/engine/mcdm/ahp"
	"github.com/guidomantilla/yarumo/compute/engine/mcdm/topsis"
)

func main() {
	fmt.Println("=== TUTORIAL: inference/engine ===")
	fmt.Println()

	//tutorialDeductive()
	//tutorialBayesian()
	//tutorialFuzzy()
	//tutorialCausal()
	//tutorialMCDM()
}

func tutorialDeductive() {
	fmt.Println("--- 1. Motor Deductivo (Forward & Backward Chaining) ---")

	// Reglas: si llueve y no tiene paraguas → se moja
	//         si se moja y hace frío → se resfría
	reglaMojado := rules.NewRule("mojado",
		parser.MustParse("llueve AND NOT paraguas"),
		map[logic.Var]bool{logic.Var("mojado"): true},
		rules.WithPriority(1),
	)

	reglaResfriado := rules.NewRule("resfriado",
		parser.MustParse("mojado AND frio"),
		map[logic.Var]bool{logic.Var("resfriado"): true},
		rules.WithPriority(2),
	)

	reglas := []rules.Rule{reglaMojado, reglaResfriado}

	// Hechos iniciales
	hechos := logic.Fact{
		logic.Var("llueve"):   true,
		logic.Var("paraguas"): false,
		logic.Var("frio"):     true,
	}

	// Forward chaining
	eng := dengine.NewEngine(
		dengine.WithMaxIterations(100),
		dengine.WithStrategy(dengine.PriorityOrder),
	)

	result := eng.Forward(hechos, reglas)
	snapshot := result.Facts.Snapshot()

	fmt.Println("  Forward Chaining:")
	fmt.Printf("    Hechos iniciales: llueve=true, paraguas=false, frio=true\n")
	fmt.Printf("    Pasos: %d\n", result.Steps)
	fmt.Printf("    mojado = %v (derivado)\n", snapshot[logic.Var("mojado")])
	fmt.Printf("    resfriado = %v (derivado)\n", snapshot[logic.Var("resfriado")])

	// Provenance: quién derivó qué
	fmt.Println("  Provenance:")
	provenance := result.Facts.AllProvenance()
	for _, p := range provenance {
		fmt.Printf("    %s = %v (origen: %s", p.Variable, p.Value, p.Origin)
		if p.RuleName != "" {
			fmt.Printf(", regla: %s, paso: %d", p.RuleName, p.Step)
		}
		fmt.Println(")")
	}

	// Backward chaining: ¿se puede probar "resfriado"?
	proved, backResult := eng.Backward(hechos, reglas, logic.Var("resfriado"))
	fmt.Printf("  Backward: ¿resfriado? %v (pasos: %d)\n", proved, backResult.Steps)

	fmt.Println()
}

func tutorialBayesian() {
	fmt.Println("--- 2. Motor Bayesiano (Redes Bayesianas) ---")

	// Red: Enfermedad → Test
	// P(enfermedad=sí) = 1%
	// P(test=positivo | enfermo) = 95%
	// P(test=positivo | sano) = 5%

	net := network.NewNetwork()

	// Nodo raíz: enfermedad (sin padres)
	enfermedadCPT := bayesian.NewCPT("enfermedad", nil)
	enfermedadCPT.Set(stats.Assignment{}, stats.Distribution{"sí": 0.01, "no": 0.99})

	err := net.AddNode(network.Node{
		Variable: "enfermedad",
		Outcomes: []stats.Outcome{"sí", "no"},
		CPT:      enfermedadCPT,
	})
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	// Nodo hijo: test (depende de enfermedad)
	testCPT := bayesian.NewCPT("test", []stats.Var{"enfermedad"})
	testCPT.Set(stats.Assignment{"enfermedad": "sí"}, stats.Distribution{"positivo": 0.95, "negativo": 0.05})
	testCPT.Set(stats.Assignment{"enfermedad": "no"}, stats.Distribution{"positivo": 0.05, "negativo": 0.95})

	err = net.AddNode(network.Node{
		Variable: "test",
		Parents:  []stats.Var{"enfermedad"},
		Outcomes: []stats.Outcome{"positivo", "negativo"},
		CPT:      testCPT,
	})
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	// Observar: test salió positivo
	ev := evidence.NewEvidenceBase()
	ev.Observe("test", "positivo")

	// Consultar: P(enfermedad | test=positivo)
	eng := bengine.NewEngine(
		bengine.WithAlgorithm(bengine.VariableElimination),
	)

	result := eng.Query(net, ev, "enfermedad")

	fmt.Printf("  Red: Enfermedad → Test\n")
	fmt.Printf("  Evidencia: test = positivo\n")
	fmt.Printf("  P(enfermedad | test=positivo):\n")
	for outcome, prob := range result.Posterior {
		fmt.Printf("    %s: %.4f\n", outcome, prob)
	}
	fmt.Println("  (Paradoja del test raro: la prevalencia baja domina)")

	// Trace
	fmt.Printf("  Trace: %d pasos\n", len(result.Trace.Steps))
	for _, step := range result.Trace.Steps {
		fmt.Printf("    [%s] %s\n", step.Phase, step.Message)
	}

	fmt.Println()
}

func tutorialFuzzy() {
	fmt.Println("--- 3. Motor Fuzzy (Mamdani) ---")

	// Variables lingüísticas de entrada: temperatura
	tempBaja, _ := fuzzy.Triangular(0, 0, 25)
	tempMedia, _ := fuzzy.Triangular(15, 25, 35)
	tempAlta, _ := fuzzy.Triangular(25, 40, 40)

	// Term = fuzzy.Set{Name, Fn}
	temperatura := variable.NewVariable("temperatura", 0, 40, []variable.Term{
		{Name: "baja", Fn: tempBaja},
		{Name: "media", Fn: tempMedia},
		{Name: "alta", Fn: tempAlta},
	})

	// Variable de salida: velocidad del ventilador
	velBaja, _ := fuzzy.Triangular(0, 0, 50)
	velMedia, _ := fuzzy.Triangular(25, 50, 75)
	velAlta, _ := fuzzy.Triangular(50, 100, 100)

	velocidad := variable.NewVariable("velocidad", 0, 100, []variable.Term{
		{Name: "baja", Fn: velBaja},
		{Name: "media", Fn: velMedia},
		{Name: "alta", Fn: velAlta},
	})

	// Reglas difusas
	reglasF := []frules.Rule{
		frules.NewRule("r1",
			[]frules.Condition{{Variable: "temperatura", Term: "baja"}},
			frules.Consequent{Variable: "velocidad", Term: "baja"},
		),
		frules.NewRule("r2",
			[]frules.Condition{{Variable: "temperatura", Term: "media"}},
			frules.Consequent{Variable: "velocidad", Term: "media"},
		),
		frules.NewRule("r3",
			[]frules.Condition{{Variable: "temperatura", Term: "alta"}},
			frules.Consequent{Variable: "velocidad", Term: "alta"},
		),
	}

	// Crear engine
	eng := fengine.NewEngine(
		[]variable.Variable{temperatura},
		[]variable.Variable{velocidad},
		reglasF,
		fengine.WithMethod(fengine.Mamdani),
		fengine.WithTNorm(fuzzy.Min),
		fengine.WithDefuzzify(fuzzy.Centroid),
		fengine.WithResolution(200),
	)

	// Inferir con diferentes temperaturas
	for _, temp := range []float64{10, 25, 33, 38} {
		result := eng.Infer(map[string]float64{"temperatura": temp})
		fmt.Printf("  Temperatura: %.0f°C → Velocidad: %.1f%%\n",
			temp, result.Outputs["velocidad"])
	}

	// Trace detallado para una inferencia
	result := eng.Infer(map[string]float64{"temperatura": 30})
	fmt.Printf("  Trace (temp=30°C): %d pasos\n", len(result.Trace.Steps))
	for _, step := range result.Trace.Steps {
		fmt.Printf("    [%s] %s\n", step.Phase, step.Message)
		for _, m := range step.Memberships {
			fmt.Printf("      %s es %s: grado %.3f\n", m.Variable, m.Term, m.Degree)
		}
	}

	fmt.Println()
}

func tutorialCausal() {
	fmt.Println("--- 4. Motor Causal (Pearl Levels 1-2) ---")

	// Modelo: precio → ventas → ganancia, publicidad → ventas
	scm := model.NewSCM()

	_ = scm.AddVariable("precio", nil, func(p map[string]float64) float64 {
		return 100 // exógeno
	})

	_ = scm.AddVariable("publicidad", nil, func(p map[string]float64) float64 {
		return 50 // exógeno
	})

	_ = scm.AddVariable("ventas", []string{"precio", "publicidad"},
		func(p map[string]float64) float64 {
			return 200 - 1.5*p["precio"] + 0.8*p["publicidad"]
		},
	)

	_ = scm.AddVariable("ganancia", []string{"precio", "ventas"},
		func(p map[string]float64) float64 {
			return p["ventas"] * (p["precio"] - 30)
		},
	)

	eng := engine.NewEngine()

	// Nivel 1: Propagación (observación)
	result := eng.Propagate(scm, map[string]float64{
		"precio":     100,
		"publicidad": 50,
	})
	fmt.Println("  Nivel 1 — Propagación (observar):")
	for name, val := range result.Values {
		fmt.Printf("    %s = %.1f\n", name, val)
	}

	// Nivel 2: Intervención (do-operator)
	// ¿Qué pasa si FORZAMOS precio = 80?
	result2 := eng.Intervene(scm, map[string]float64{
		"precio": 80,
	})
	fmt.Println("  Nivel 2 — Intervención do(precio=80):")
	for name, val := range result2.Values {
		fmt.Printf("    %s = %.1f\n", name, val)
	}

	// Contrafactual
	result3 := eng.Counterfactual(scm,
		map[string]float64{"precio": 100, "publicidad": 50}, // factual
		map[string]float64{"precio": 80},                    // hipotético
	)
	fmt.Println("  Contrafactual — si precio hubiera sido 80:")
	for name, val := range result3.Values {
		fmt.Printf("    %s = %.1f\n", name, val)
	}

	fmt.Println()
}

func tutorialMCDM() {
	fmt.Println("--- 5. MCDM (AHP + TOPSIS) ---")

	// AHP: ¿qué criterio pesa más? (comparación por pares)
	// Criterios: precio, seguridad, consumo
	// precio 3x más importante que seguridad, 5x más que consumo
	matrix := ahp.PairwiseMatrix{
		{1, 3, 5},
		{1.0 / 3, 1, 2},
		{1.0 / 5, 0.5, 1},
	}

	ahpResult, err := ahp.Analyze(matrix)
	if err != nil {
		fmt.Printf("  AHP Error: %v\n", err)
		return
	}

	fmt.Println("  AHP — Pesos de criterios:")
	criterios := []string{"precio", "seguridad", "consumo"}
	for i, w := range ahpResult.Weights {
		fmt.Printf("    %s: %.3f\n", criterios[i], w)
	}
	fmt.Printf("  Consistencia: CR=%.4f (consistente: %v)\n",
		ahpResult.ConsistencyRatio, ahpResult.Consistent)

	// Rankear alternativas
	evaluations := [][]float64{
		{0.8, 0.3, 0.9}, // Auto A: barato, poco seguro, eficiente
		{0.4, 0.9, 0.5}, // Auto B: caro, muy seguro, regular
		{0.6, 0.6, 0.7}, // Auto C: intermedio
	}

	scores, _ := ahp.Rank(ahpResult.Weights, evaluations)
	autos := []string{"Auto A", "Auto B", "Auto C"}
	fmt.Println("  AHP — Rankings:")
	for i, s := range scores {
		fmt.Printf("    %s: %.3f\n", autos[i], s)
	}

	fmt.Println()

	// TOPSIS: cercanía a la solución ideal
	topsisMatrix := [][]float64{
		{250, 16, 12}, // Auto A: precio, potencia, consumo
		{200, 20, 15}, // Auto B
		{300, 22, 10}, // Auto C
	}

	tCriteria := []topsis.Criterion{
		{Weight: 0.5, Benefit: false}, // precio: menor es mejor
		{Weight: 0.3, Benefit: true},  // potencia: mayor es mejor
		{Weight: 0.2, Benefit: false}, // consumo: menor es mejor
	}

	topsisResult, err := topsis.Rank(topsisMatrix, tCriteria)
	if err != nil {
		fmt.Printf("  TOPSIS Error: %v\n", err)
		return
	}

	fmt.Println("  TOPSIS — Rankings (cercanía al ideal):")
	for i, s := range topsisResult.Scores {
		fmt.Printf("    %s: %.3f\n", autos[i], s)
	}

	fmt.Println()
}
