# Tutorial: modules/compute/math

Guía práctica para usar los paquetes matemáticos de Yarumo.

## 1. Lógica Proposicional (`logic/`)

La base de todo. Fórmulas lógicas evaluables, transformables, simplificables.

### Tipos fundamentales

```go
type Var string              // Variable proposicional: "llueve", "hay_tráfico"
type Fact map[Var]bool       // Asignación parcial: {"llueve": true, "hay_tráfico": false}
type Formula interface {     // Cualquier fórmula: And, Or, Not, Impl, Iff, True, False
    Eval(facts Fact) bool
    Vars() []Var
    String() string
}
```

### Crear y evaluar fórmulas

```go
import "github.com/guidomantilla/yarumo/compute/math/logic"

// Variables
llueve := logic.Var("llueve")
paraguas := logic.Var("paraguas")

// Fórmula: llueve => paraguas
regla := logic.ImplF{L: llueve, R: paraguas}

// Evaluar con hechos
hechos := logic.Fact{llueve: true, paraguas: true}
regla.Eval(hechos) // true

hechos2 := logic.Fact{llueve: true, paraguas: false}
regla.Eval(hechos2) // false (llueve pero no hay paraguas)
```

### Combinaciones

```go
// AND: llueve AND hay_viento
logic.AndF{L: llueve, R: logic.Var("hay_viento")}

// OR: llueve OR nieva
logic.OrF{L: llueve, R: logic.Var("nieva")}

// NOT: !llueve
logic.NotF{F: llueve}

// Bicondicional: llueve <=> mojado
logic.IffF{L: llueve, R: logic.Var("mojado")}
```

### Parser — de texto a fórmula

```go
import "github.com/guidomantilla/yarumo/compute/math/logic/parser"

// Texto natural → fórmula
f, err := parser.Parse("llueve AND viento => paraguas")

// Operadores soportados:
//   NOT: !, ~, ¬, not
//   AND: &, &&, ∧, and, ^
//   OR:  |, ||, ∨, or, v
//   IMPL: =>, ->, →
//   IFF:  <=>, <->, ↔
//   Constantes: true, T, ⊤, false, F, ⊥

f2, _ := parser.Parse("(A ∧ B) → C")
f3 := parser.MustParse("¬A ∨ B")  // panic si falla
```

### Transformaciones

```go
import "github.com/guidomantilla/yarumo/compute/math/logic"

f, _ := parser.Parse("A => (B => C)")

// Forma Normal Negativa (negaciones solo en variables)
nnf := logic.ToNNF(f)

// Forma Normal Conjuntiva (AND de ORs) — útil para SAT
cnf := logic.ToCNF(f)

// Forma Normal Disyuntiva (OR de ANDs)
dnf := logic.ToDNF(f)

// Simplificación algebraica (18 reglas hasta punto fijo)
simple := logic.Simplify(f)
logic.Format(simple) // representación legible
```

### Análisis

```go
f, _ := parser.Parse("A AND (NOT A)")

// ¿Es satisfacible?
logic.IsSatisfiable(f)   // false — contradicción

// ¿Es tautología?
logic.IsTautology(parser.MustParse("A OR NOT A"))  // true

// ¿Son equivalentes?
logic.Equivalent(
    parser.MustParse("A => B"),
    parser.MustParse("NOT A OR B"),
) // true

// Tabla de verdad completa
rows := logic.TruthTable(parser.MustParse("A AND B"))
// [{A:true,B:true} => true, {A:true,B:false} => false, ...]

// Casos donde la fórmula falla
gaps := logic.FailCases(parser.MustParse("A OR B"))
// [{A:false, B:false}]
```

### SAT Solver (DPLL)

```go
import "github.com/guidomantilla/yarumo/compute/math/logic/sat"

// Registrar el solver (una vez, al inicio)
logic.RegisterSATSolver(sat.Solver())

// Ahora FindSatisfyingAssignment usa DPLL
asignacion, found := logic.FindSatisfyingAssignment(
    parser.MustParse("(A OR B) AND (NOT A OR C)"),
)
// found: true, asignacion: {A: true, B: true, C: true} (o similar)
```

### Entailment — ¿las premisas implican la conclusión?

```go
import "github.com/guidomantilla/yarumo/compute/math/logic/entailment"

premises := []logic.Formula{
    parser.MustParse("llueve => mojado"),
    parser.MustParse("llueve"),
}
conclusion := parser.MustParse("mojado")

entailment.Entails(premises, conclusion) // true: {llueve=>mojado, llueve} ⊨ mojado
```

## 2. Lógica de Predicados (`logic/predicate/`)

Cuantificadores sobre colecciones finitas de hechos.

```go
import "github.com/guidomantilla/yarumo/compute/math/logic/predicate"

// Una colección = slice de Facts (cada Fact es un "objeto")
perros := predicate.Collection{
    logic.Fact{logic.Var("grande"): true, logic.Var("vacunado"): true},
    logic.Fact{logic.Var("grande"): false, logic.Var("vacunado"): true},
    logic.Fact{logic.Var("grande"): true, logic.Var("vacunado"): false},
}

regla := parser.MustParse("grande => vacunado")

// ∀ perro: grande => vacunado?
ok, _ := predicate.ForAll(perros, regla) // false (el 3ro es grande y no vacunado)

// ∃ perro: grande AND vacunado?
ok, _ = predicate.Exists(perros, parser.MustParse("grande AND vacunado")) // true

// ¿Cuántos cumplen?
n, _ := predicate.Count(perros, parser.MustParse("vacunado")) // 2

// Filtrar los que cumplen
vacunados, _ := predicate.Filter(perros, parser.MustParse("vacunado"))
// 2 elementos
```

## 3. Lógica Temporal (`logic/temporal/`)

Aserciones sobre secuencias de eventos en el tiempo.

```go
import "github.com/guidomantilla/yarumo/compute/math/logic/temporal"

trace := temporal.Trace{
    {Label: "pedido_recibido", Time: t0},
    {Label: "pago_procesado", Time: t0.Add(2 * time.Minute)},
    {Label: "envio_creado", Time: t0.Add(10 * time.Minute)},
    {Label: "entregado", Time: t0.Add(48 * time.Hour)},
}

// ¿Cada pedido tiene envío en máximo 30 min?
ok, violations := temporal.ResponseWithin(trace, "pedido_recibido", "envio_creado", 30*time.Minute)
// ok: true (10 min < 30 min)

// ¿El evento ocurre al menos una vez?
temporal.Eventually(trace, "entregado") // true

// ¿"pago" siempre ocurre antes de "envio"?
temporal.Before(trace, "pago_procesado", "envio_creado") // true

// ¿Cuánto tiempo entre dos eventos?
dur, _ := temporal.Elapsed(trace, "pedido_recibido", "entregado") // 48h

// ¿Los eventos siguen esta secuencia?
temporal.Sequence(trace, "pedido_recibido", "pago_procesado", "envio_creado") // true
```

## 4. Fuzzy (`fuzzy/`)

Lógica difusa: grados de pertenencia, no true/false binario.

```go
import "github.com/guidomantilla/yarumo/compute/math/fuzzy"

// Funciones de membresía
baja, _ := fuzzy.Triangular(0, 0, 50)       // "temperatura baja"
media, _ := fuzzy.Triangular(20, 50, 80)    // "temperatura media"
alta, _ := fuzzy.Triangular(50, 100, 100)   // "temperatura alta"

// Fuzzificar un valor crisp
fuzzy.Fuzzify(baja, 30)   // 0.4 (parcialmente baja)
fuzzy.Fuzzify(media, 30)  // 0.33 (parcialmente media)
fuzzy.Fuzzify(alta, 30)   // 0.0 (nada alta)

// T-norms (AND difuso)
fuzzy.Min(0.4, 0.7)       // 0.4
fuzzy.Product(0.4, 0.7)   // 0.28

// T-conorms (OR difuso)
fuzzy.Max(0.4, 0.7)       // 0.7

// Complemento (NOT difuso)
fuzzy.Complement(0.4)     // 0.6

// Operaciones sobre funciones de membresía
clipped := fuzzy.Clip(media, 0.5)    // cortar a grado máximo 0.5
scaled := fuzzy.Scale(media, 0.5)    // escalar proporcionalmente
combined := fuzzy.AggregateMax(baja, media)  // unión (max)

// Muestreo y defuzzificación
xs, ys, _ := fuzzy.Sample(combined, 0, 100, 200) // 200 puntos en [0,100]
crisp := fuzzy.Centroid(xs, ys)    // centro de gravedad
crisp2 := fuzzy.MeanOfMax(xs, ys)  // promedio de máximos
```

## 5. Estadística (`stats/`)

Distribuciones, tests de hipótesis, estadística descriptiva.

### Tipos base

```go
import "github.com/guidomantilla/yarumo/compute/math/stats"

type Var string                      // Variable aleatoria
type Outcome string                  // Valor posible: "sí", "no"
type Prob float64                    // Probabilidad [0,1]
type Distribution map[Outcome]Prob   // {"sí": 0.7, "no": 0.3}
```

### Distribuciones de probabilidad

```go
// Discreta
d := stats.Distribution{"llueve": 0.3, "no_llueve": 0.7}
stats.IsValid(d)                         // true (suma ~1.0)
stats.Entropy(d)                         // Shannon entropy
stats.Complement(stats.Prob(0.3))        // 0.7

// Bayes
posterior, _ := stats.Bayes(
    stats.Prob(0.01),  // prior: P(enfermo)
    stats.Prob(0.95),  // likelihood: P(test+|enfermo)
    stats.Prob(0.05),  // evidence: P(test+)
) // 0.19 — P(enfermo|test+)
```

### Distribuciones continuas

```go
// Normal
n, _ := stats.NewNormal(100, 15) // μ=100, σ=15 (IQ)
n.PDF(100)        // densidad en 100
n.CDF(115)        // P(X ≤ 115) ≈ 0.84
n.Mean()          // 100
n.Variance()      // 225
n.Quantile(0.95)  // valor donde P(X ≤ x) = 0.95

// Otras: Exponential, Uniform, Beta, Gamma, ChiSquared, StudentT,
//        Lognormal, Weibull, FDist

// Discretas: Poisson, Binomial
p := stats.Poisson{Lambda: 5}
p.PMF(3)           // P(X = 3)
p.CDFDiscrete(3)   // P(X ≤ 3)
```

### Estadística descriptiva

```go
data := []float64{23, 45, 12, 67, 34, 56, 78}

stats.Mean(data)           // 45.0
stats.Median(data)         // 45.0
stats.StdDev(data)         // desviación estándar
stats.Variance(data)       // varianza poblacional
stats.SampleVariance(data) // varianza muestral (Bessel)
stats.Percentile(data, 75) // percentil 75
stats.Min(data)            // 12
stats.Max(data)            // 78
stats.Mode(data)           // moda

// Bivariado
x := []float64{1, 2, 3, 4, 5}
y := []float64{2, 4, 5, 4, 5}
stats.Correlation(x, y)        // Pearson r
stats.LinearRegression(x, y)   // slope, intercept
stats.RSquared(x, y)           // R²
```

### Tests de hipótesis

```go
// t-test: ¿la media es significativamente diferente de μ?
t, p, _ := stats.TTest(data, 50.0)

// Welch t-test: ¿dos muestras tienen medias diferentes?
t, p, _ = stats.TTestTwoSample(muestra1, muestra2)

// Chi-cuadrado: ¿los datos observados coinciden con los esperados?
chi2, p, _ = stats.ChiSquaredTest(observed, expected)
```

### Running Stats (Welford — streaming)

```go
var rs stats.RunningStats
rs.Push(10)
rs.Push(20)
rs.Push(30)
rs.Count()          // 3
rs.Mean()           // 20.0
rs.Variance()       // varianza poblacional
rs.SampleVariance() // varianza muestral
```

## 6. Conjuntos (`sets/`)

Operaciones de conjuntos genéricos.

```go
import "github.com/guidomantilla/yarumo/compute/math/sets"

a := sets.New("perro", "gato", "pez")
b := sets.New("gato", "pájaro")

sets.Union(a, b)              // {perro, gato, pez, pájaro}
sets.Intersection(a, b)       // {gato}
sets.Difference(a, b)         // {perro, pez}
sets.SymmetricDifference(a,b) // {perro, pez, pájaro}
sets.IsSubset(b, a)           // false
a.Contains("perro")           // true
a.Len()                       // 3
```

## Mapa de dependencias

```
logic/
  ├── parser/      → Parse("A AND B") → Formula
  ├── sat/         → DPLL solver, se registra en logic/
  ├── entailment/  → premises ⊨ conclusion
  ├── predicate/   → ForAll, Exists, Count, Filter sobre Collection
  └── temporal/    → ResponseWithin, Before, Sequence sobre Trace

fuzzy/             → MembershipFn, TNorm, Defuzzify (independiente de logic/)

stats/             → Distribuciones, Bayes, tests, descriptiva (independiente)

sets/              → Operaciones de conjuntos genéricos (independiente)
```
