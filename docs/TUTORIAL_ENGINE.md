# Tutorial: modules/compute/engine

Guía práctica para usar los 5 motores de inferencia de Yarumo.

## Arquitectura común

Todos los motores siguen el mismo patrón:

```
1. Crear engine     → NewEngine(opciones...)
2. Ejecutar         → engine.Forward() / .Query() / .Infer() / etc.
3. Leer resultado   → Result { datos + Trace }
4. (Opcional) Trace → auditoría, explicabilidad
```

---

## 1. Deductivo — Forward & Backward Chaining

Razonamiento basado en reglas: "si llueve Y no tiene paraguas → se moja".

### Forward Chaining (de hechos a conclusiones)

```go
import (
    "github.com/guidomantilla/yarumo/compute/math/logic"
    "github.com/guidomantilla/yarumo/compute/math/logic/parser"
    "github.com/guidomantilla/yarumo/compute/engine/deductive/engine"
    "github.com/guidomantilla/yarumo/compute/engine/deductive/rules"
)

// 1. Definir reglas
reglas := []rules.Rule{
    rules.NewRule("mojado",
        parser.MustParse("llueve AND NOT paraguas"),  // condición
        map[logic.Var]bool{logic.Var("mojado"): true}, // conclusión
        rules.WithPriority(1),
    ),
    rules.NewRule("resfriado",
        parser.MustParse("mojado AND frio"),
        map[logic.Var]bool{logic.Var("resfriado"): true},
        rules.WithPriority(2),
    ),
}

// 2. Hechos iniciales
hechos := logic.Fact{
    logic.Var("llueve"):   true,
    logic.Var("paraguas"): false,
    logic.Var("frio"):     true,
}

// 3. Crear engine y ejecutar
eng := engine.NewEngine(
    engine.WithMaxIterations(100),
    engine.WithStrategy(engine.PriorityOrder),
)

result := eng.Forward(hechos, reglas)

// 4. Leer resultado
snapshot := result.Facts.Snapshot()
// snapshot["mojado"] == true   (derivado por regla "mojado")
// snapshot["resfriado"] == true (derivado por regla "resfriado")

result.Steps // 2 — iteraciones hasta punto fijo
```

### Backward Chaining (de meta a prueba)

```go
// ¿Se puede probar "resfriado" desde los hechos?
proved, result := eng.Backward(hechos, reglas, logic.Var("resfriado"))
// proved: true
// result.Facts contiene la cadena de derivación
```

### Provenance — quién derivó qué

```go
provenance := result.Facts.AllProvenance()
for _, p := range provenance {
    fmt.Printf("%s = %v (origen: %s, regla: %s, paso: %d)\n",
        p.Variable, p.Value, p.Origin, p.RuleName, p.Step)
}
// llueve = true (origen: Asserted, regla: , paso: 0)
// mojado = true (origen: Derived, regla: mojado, paso: 1)
// resfriado = true (origen: Derived, regla: resfriado, paso: 2)
```

### Opciones

| Opción | Default | Descripción |
|--------|---------|-------------|
| `WithMaxIterations(n)` | 1000 | Límite de iteraciones forward |
| `WithMaxDepth(n)` | 100 | Profundidad máxima backward |
| `WithStrategy(s)` | PriorityOrder | `PriorityOrder` o `FirstMatch` |

---

## 2. Bayesiano — Redes Bayesianas

Razonamiento probabilístico: "dado que el test salió positivo, ¿cuál es la probabilidad de estar enfermo?"

```go
import (
    "github.com/guidomantilla/yarumo/compute/math/stats"
    "github.com/guidomantilla/yarumo/compute/engine/bayesian/engine"
    "github.com/guidomantilla/yarumo/compute/engine/bayesian/network"
    "github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
    bayesian "github.com/guidomantilla/yarumo/compute/engine/bayesian"
)

// 1. Definir la red
//
//    Enfermedad → Test
//    P(enfermedad) = 0.01
//    P(test+ | enfermo) = 0.95
//    P(test+ | sano) = 0.05

net := network.NewNetwork()

// Nodo raíz: enfermedad (sin padres)
enfermedadCPT := bayesian.NewCPT("enfermedad", nil)
enfermedadCPT.Set(stats.Assignment{}, stats.Distribution{"sí": 0.01, "no": 0.99})

net.AddNode(network.Node{
    Variable: "enfermedad",
    Outcomes: []stats.Outcome{"sí", "no"},
    CPT:      enfermedadCPT,
})

// Nodo hijo: test (depende de enfermedad)
testCPT := bayesian.NewCPT("test", []stats.Var{"enfermedad"})
testCPT.Set(stats.Assignment{"enfermedad": "sí"}, stats.Distribution{"positivo": 0.95, "negativo": 0.05})
testCPT.Set(stats.Assignment{"enfermedad": "no"}, stats.Distribution{"positivo": 0.05, "negativo": 0.95})

net.AddNode(network.Node{
    Variable: "test",
    Parents:  []stats.Var{"enfermedad"},
    Outcomes: []stats.Outcome{"positivo", "negativo"},
    CPT:      testCPT,
})

// 2. Observar evidencia
ev := evidence.NewEvidenceBase()
ev.Observe("test", "positivo") // el test salió positivo

// 3. Consultar
eng := engine.NewEngine(
    engine.WithAlgorithm(engine.VariableElimination),
)

result := eng.Query(net, ev, "enfermedad")

// 4. Leer posterior
result.Posterior // {"sí": ~0.16, "no": ~0.84}
// Solo 16% de probabilidad de estar enfermo con test positivo
// (paradoja del test raro — la prevalencia baja domina)
```

### Opciones

| Opción | Default | Descripción |
|--------|---------|-------------|
| `WithAlgorithm(a)` | Enumeration | `Enumeration` o `VariableElimination` |
| `WithEliminationOrder(vars)` | auto | Orden de eliminación manual |

---

## 3. Fuzzy — Inferencia Difusa (Mamdani / Sugeno)

Razonamiento con grados: "si la temperatura es ALTA y la humedad es ALTA → ventilador a velocidad ALTA".

```go
import (
    "github.com/guidomantilla/yarumo/compute/math/fuzzy"
    "github.com/guidomantilla/yarumo/compute/engine/fuzzy/engine"
    "github.com/guidomantilla/yarumo/compute/engine/fuzzy/variable"
    "github.com/guidomantilla/yarumo/compute/engine/fuzzy/rules"
)

// 1. Variables lingüísticas
tempBaja, _ := fuzzy.Triangular(0, 0, 25)
tempMedia, _ := fuzzy.Triangular(15, 25, 35)
tempAlta, _ := fuzzy.Triangular(25, 40, 40)

temperatura := variable.NewVariable("temperatura", 0, 40, []variable.Term{
    {Name: "baja", Fn: tempBaja},
    {Name: "media", Fn: tempMedia},
    {Name: "alta", Fn: tempAlta},
})

velBaja, _ := fuzzy.Triangular(0, 0, 50)
velMedia, _ := fuzzy.Triangular(25, 50, 75)
velAlta, _ := fuzzy.Triangular(50, 100, 100)

velocidad := variable.NewVariable("velocidad", 0, 100, []variable.Term{
    {Name: "baja", Fn: velBaja},
    {Name: "media", Fn: velMedia},
    {Name: "alta", Fn: velAlta},
})

// 2. Reglas difusas
reglasF := []rules.Rule{
    rules.NewRule("r1",
        []rules.Condition{{Variable: "temperatura", Term: "baja"}},
        rules.Consequent{Variable: "velocidad", Term: "baja"},
    ),
    rules.NewRule("r2",
        []rules.Condition{{Variable: "temperatura", Term: "media"}},
        rules.Consequent{Variable: "velocidad", Term: "media"},
    ),
    rules.NewRule("r3",
        []rules.Condition{{Variable: "temperatura", Term: "alta"}},
        rules.Consequent{Variable: "velocidad", Term: "alta"},
    ),
}

// 3. Crear engine y ejecutar
eng := engine.NewEngine(
    []variable.Variable{temperatura},
    []variable.Variable{velocidad},
    reglasF,
    engine.WithMethod(engine.Mamdani),
    engine.WithTNorm(fuzzy.Min),
    engine.WithDefuzzify(fuzzy.Centroid),
    engine.WithResolution(200),
)

result := eng.Infer(map[string]float64{
    "temperatura": 33, // hace calor
})

// 4. Leer salida
result.Outputs["velocidad"] // ~75.0 — ventilador rápido
```

### Sugeno (alternativa más rápida)

```go
eng := engine.NewEngine(
    []variable.Variable{temperatura},
    []variable.Variable{velocidad},
    reglasF,
    engine.WithMethod(engine.Sugeno),
    engine.WithSugenoOutputs(map[string]float64{
        "baja":  20,
        "media": 50,
        "alta":  80,
    }),
)
// Sugeno usa promedios ponderados en vez de defuzzificación
```

### Opciones

| Opción | Default | Descripción |
|--------|---------|-------------|
| `WithMethod(m)` | Mamdani | `Mamdani` o `Sugeno` |
| `WithTNorm(fn)` | Min | AND difuso: `Min`, `Product`, `Lukasiewicz` |
| `WithTConorm(fn)` | Max | OR difuso: `Max`, `ProbabilisticSum` |
| `WithDefuzzify(fn)` | Centroid | `Centroid`, `Bisector`, `MeanOfMax`, etc. |
| `WithResolution(n)` | 100 | Puntos de muestreo para defuzzificación |

---

## 4. Causal — Modelos Causales Estructurales (Pearl)

Razonamiento causal: "¿qué PASARÍA SI intervenimos X?" No solo correlación, sino causalidad.

```go
import (
    "github.com/guidomantilla/yarumo/compute/engine/causal/engine"
    "github.com/guidomantilla/yarumo/compute/engine/causal/model"
)

// 1. Definir el modelo causal (SCM)
//
//    precio → ventas → ganancia
//    publicidad → ventas
//
scm := model.NewSCM()

scm.AddVariable("precio", nil, func(p map[string]float64) float64 {
    return 100 // exógeno
})

scm.AddVariable("publicidad", nil, func(p map[string]float64) float64 {
    return 50 // exógeno
})

scm.AddVariable("ventas", []string{"precio", "publicidad"},
    func(p map[string]float64) float64 {
        return 200 - 1.5*p["precio"] + 0.8*p["publicidad"]
    },
)

scm.AddVariable("ganancia", []string{"precio", "ventas"},
    func(p map[string]float64) float64 {
        return p["ventas"] * (p["precio"] - 30) // margen
    },
)

eng := engine.NewEngine()
```

### Nivel 1: Asociación (Propagación)

```go
// ¿Qué valores resultan con estas observaciones?
result := eng.Propagate(scm, map[string]float64{
    "precio":     100,
    "publicidad": 50,
})
// result.Values: {precio:100, publicidad:50, ventas:90, ganancia:6300}
```

### Nivel 2: Intervención (do-operator)

```go
// ¿Qué pasa SI FORZAMOS el precio a 80? (do(precio=80))
// Esto corta las causas naturales de precio — es una intervención
result := eng.Intervene(scm, map[string]float64{
    "precio": 80,
})
// Las ventas suben (precio más bajo), pero el margen baja
// result.Values: {precio:80, publicidad:50, ventas:120, ganancia:6000}
```

### Nivel 2.5: Contrafactual

```go
// Dado que OBSERVAMOS precio=100 y ventas=90,
// ¿qué HUBIERA pasado si precio=80?
result := eng.Counterfactual(scm,
    map[string]float64{"precio": 100, "publicidad": 50}, // factual
    map[string]float64{"precio": 80},                     // hipotético
)
// result.Values: valores contrafactuales
```

---

## 5. MCDM — Decisión Multi-Criterio

Elegir la mejor opción cuando hay múltiples criterios en conflicto.

### AHP — Proceso Analítico Jerárquico

"¿Qué auto comprar?" — comparar criterios (precio, seguridad, consumo) por pares.

```go
import "github.com/guidomantilla/yarumo/compute/engine/mcdm/ahp"

// Matriz de comparación por pares (escala Saaty 1-9)
//           precio  seguridad  consumo
// precio      1        3         5      → precio 3x más importante que seguridad
// seguridad  1/3       1         2
// consumo    1/5      1/2        1
matrix := ahp.PairwiseMatrix{
    {1, 3, 5},
    {1.0 / 3, 1, 2},
    {1.0 / 5, 0.5, 1},
}

result, err := ahp.Analyze(matrix)
// result.Weights: [0.63, 0.26, 0.11] — pesos de cada criterio
// result.ConsistencyRatio: 0.003 (< 0.10 = consistente)
// result.Consistent: true

// Rankear alternativas con los pesos
//           precio  seguridad  consumo
// Auto A:    0.8      0.3       0.9     (barato, poco seguro, eficiente)
// Auto B:    0.4      0.9       0.5     (caro, muy seguro, regular)
// Auto C:    0.6      0.6       0.7     (intermedio)
evaluations := [][]float64{
    {0.8, 0.3, 0.9},
    {0.4, 0.9, 0.5},
    {0.6, 0.6, 0.7},
}

scores, _ := ahp.Rank(result.Weights, evaluations)
// scores: [0.69, 0.47, 0.61] → Auto A gana (precio domina)
```

### TOPSIS — Cercanía a la Solución Ideal

```go
import "github.com/guidomantilla/yarumo/compute/engine/mcdm/topsis"

// Matriz de evaluación (3 alternativas × 3 criterios)
matrix := [][]float64{
    {250, 16, 12},  // Auto A: precio, hp, consumo
    {200, 20, 15},  // Auto B
    {300, 22, 10},  // Auto C
}

criteria := []topsis.Criterion{
    {Weight: 0.5, Benefit: false}, // precio: menor es mejor
    {Weight: 0.3, Benefit: true},  // hp: mayor es mejor
    {Weight: 0.2, Benefit: false}, // consumo: menor es mejor
}

result, _ := topsis.Rank(matrix, criteria)
// result.Scores: [0.55, 0.62, 0.43] → Auto B gana
// Scores ∈ [0,1] — más cercano a 1 = más cercano al ideal
```

---

## Traces — Explicabilidad

Todos los motores producen `Trace` con pasos detallados. Útil para auditoría y debugging.

```go
// Deductivo
result := eng.Forward(facts, rules)
for _, step := range result.Trace.Steps {
    fmt.Printf("Paso %d: regla %s disparada\n", step.Number, step.RuleName)
    fmt.Printf("  Condición: %s\n", step.Condition)
    fmt.Printf("  Produjo: %v\n", step.Produced)
}

// Bayesiano
result := eng.Query(net, ev, "enfermedad")
for _, step := range result.Trace.Steps {
    fmt.Printf("Fase: %s — %s\n", step.Phase, step.Message)
}

// Fuzzy
result := eng.Infer(inputs)
for _, step := range result.Trace.Steps {
    fmt.Printf("Fase: %s\n", step.Phase) // Fuzzification, RuleEvaluation, etc.
    for _, m := range step.Memberships {
        fmt.Printf("  %s es %s: grado %.2f\n", m.Variable, m.Term, m.Degree)
    }
}
```

---

## Resumen

| Motor | Entrada | Salida | Caso de uso |
|-------|---------|--------|-------------|
| **Deductivo** | Hechos + Reglas IF-THEN | Nuevos hechos derivados | Reglas de negocio, compliance |
| **Bayesiano** | Red + Evidencia + Query | Distribución posterior | Diagnóstico, riesgo |
| **Fuzzy** | Valores numéricos | Valores numéricos difusos | Control, scoring gradual |
| **Causal** | SCM + observaciones | Valores propagados/intervenidos | Análisis what-if |
| **MCDM** | Matriz comparación/evaluación | Rankings ponderados | Selección de alternativas |
