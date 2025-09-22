# logic2 — Paquete de lógica proposicional, SAT y motor de reglas (MVP)

Este paquete provee utilidades de lógica proposicional en Go con foco en simplicidad y buen DX: AST + parser, transformaciones (NNF/CNF/DNF), simplificación, SAT (DPLL) con política de decisión, entailment por refutación, un motor de reglas (forward) sencillo con explicaciones, y utilidades de serialización (DTOs JSON v1).

La documentación a continuación refleja el estado actual del paquete basado en las tareas marcadas como [x] en PLAN.md.


## Índice
- Visión y estado
- Instalación e importación
- Layout de paquetes
- Características implementadas
  - Parser y DX
  - Proposiciones (AST, Eval, Simplify, NNF/CNF/DNF, TruthTable, Equivalent, Is*)
  - Política Is* (umbral SAT)
  - SAT (CNF + DPLL)
  - Entailment (KB ⊨ φ)
  - Motor de reglas (Engine MVP) + PrettyExplain
  - Serialización (DTOs v1, JSON) y Facts JSON
- Ejemplos rápidos
- Benchmarks y escalamiento
- Límites conocidos y fuera de alcance del MVP


## Visión y estado
- Fase 1 — Núcleo proposicional usable: COMPLETADO
- Fase 2 — SAT y política Is*: EN PROGRESO (SAT operativo, política Is* y entailment listos; pendiente documentar métricas de escalamiento)
- Fase 3 — Motor de reglas (MVP): COMPLETADO
- Fase 4 — DX y serialización: EN PROGRESO (DTOs/JSON/Pretty y parser DX listos)

Consulte pkg/common/maths/logic2/PLAN.md para el roadmap detallado.


## Instalación e importación
```text
import (
  p "github.com/guidomantilla/yarumo/pkg/common/maths/logic2/props"
  "github.com/guidomantilla/yarumo/pkg/common/maths/logic2/parser"
  "github.com/guidomantilla/yarumo/pkg/common/maths/logic2/engine"
  logic2sat "github.com/guidomantilla/yarumo/pkg/common/maths/logic2/sat"
)
```


## Layout de paquetes
- logic2/props: AST y funciones puras (Eval, Vars, Simplify, ToNNF/CNF/DNF, TruthTable, Is*)
- logic2/parser: lexer+parser con sinónimos/Unicode opcionales; errores con posición
- logic2/sat: CNF aplanada + DPLL (unit-prop, pure-literal, branching por cláusula más corta)
- logic2/entailment: Entails y EntailsWithCounterModel
- logic2/engine: FactBase, Rule, Engine (forward simple) y Explain + PrettyExplain
- logic2/examples: pruebas y ejemplos reproducibles

Dependencias internas: props ← parser | sat | engine; entailment → props,sat; examples → todos.


## Características implementadas

### Parser y DX
- Gramática canónica: `!` `&` `|` `=>` `<=>` y paréntesis. Identificadores `[A-Za-z_][A-Za-z0-9_]*`.
- Sinónimos/Unicode (por defecto ON): palabras clave `AND/OR/NOT/THEN/IFF`, operadores `&&, ||, ->, <->`, símbolos `¬ ∧ ∨ → ⇒ ↔ ⇔`, literales `TRUE/FALSE`.
- Modo estricto opcional: `ParseWith(input, ParseOptions{Strict: true})` sólo acepta la gramática canónica; los sinónimos se tratan como identificadores.
- Errores estructurados: `ParseError{Pos, Col, Msg}` con posición (byte/columna) y mensajes claros.

API principal:
```text
f, err := parser.Parse("(A & B) => C")
f2, err := parser.ParseWith("A ∧ B ⇒ C", parser.ParseOptions{Strict: false})
```


### Proposiciones (AST y utilidades)
- AST: `Var, TrueF, FalseF, NotF, AndF, OrF, ImplF, IffF, GroupF` con `String()`, `Eval(Fact)`, `Vars()`.
- Transformaciones: `ToNNF`, `ToCNF`, `ToDNF`.
- Simplifcación: constantes, doble negación, idempotencia, absorción, complemento, trivialidades Impl/Iff.
- Utilidades: `TruthTable`, `Equivalent`, `FailCases`.
- Is*: `IsSatisfiable`, `IsContradiction`, `IsTautology` (con política Is*; ver abajo).


### Política Is* (umbral SAT)
- Política única para `IsSatisfiable(f)`:
  - Si `#vars ≤ K` (por defecto `12`): usa tabla de verdad (exhaustiva).
  - Si `#vars > K` y hay solver SAT registrado: usa SAT (CNF+DPLL).
  - Fallback automático a tabla de verdad si no hay solver registrado.
- Configuración:
```text
p.SATThreshold = 12 // opcional
p.RegisterSATSolver(logic2sat.Solver) // registro explícito del backend SAT
```
- Derivadas: `IsContradiction(f) = !IsSatisfiable(f)`, `IsTautology(f) = !IsSatisfiable(!f)`.


### SAT (CNF + DPLL)
- Conversión a CNF aplanada: `sat.FromFormulaToCNF(f) (CNF, error)`.
- Solver DPLL con propagación unitaria, literales puros y branching por cláusula más corta: `sat.DPLL(cnf, nil)` → `(ok, model)`.


### Entailment (consecución lógica)
- `entailment.Entails(KB, phi) bool` implementa refutación: `¬IsSatisfiable((∧KB) ∧ ¬phi)`.
- `entailment.EntailsWithCounterModel(KB, phi) (bool, sat.Assignment)` devuelve un contramodelo cuando no hay entailment (vía SAT si disponible).


### Motor de reglas (MVP simple)
- Datos: `engine.FactBase` (mapa de hechos), `engine.Rule{ID, When, Then}`.
- Ejecución: `Engine.FireOnce()`, `RunToFixpoint(maxIters)`, `Assert/ Retract`.
- Consulta: `Query(goal) (ok bool, why *Explain)`.
- Explicaciones: árbol `Explain{Expr, Value, Why, Kids}` y `engine.PrettyExplain`/`PrettyExplainTo(io.Writer, *Explain)` deterministas.
- Semántica especial de firing: si `When` es `A => B` y `Then == B`, la regla dispara cuando `A` es verdadero.


### Serialización (DTOs v1) y Facts JSON
- DTOs (v1) en `engine/dto.go`:
  - `RuleDTO{version,id,when,then}`; `RuleSetDTO{version,rules}`; `ExplainDTO{expr,value,why,kids[]}`.
- JSON helpers en `engine/serialize.go`:
  - `LoadRulesJSON(r) ([]Rule, error)`, `SaveRulesJSON(w, rules)`.
  - `LoadFactsJSON(r) (FactBase, error)`, `SaveFactsJSON(w, facts)`.
- Las fórmulas se serializan como strings (contrato estable). No se serializa el AST interno.


## Ejemplos rápidos

### Parsear, evaluar y simplificar
```text
f := parser.MustParse("(A & B) => C")
facts := p.Fact{p.Var("A"): true, p.Var("B"): true, p.Var("C"): false}
_ = f.Eval(facts) // false
s := p.Simplify(parser.MustParse("A | (A & B)")) // → A
```

### Satisfacibilidad con política Is*
```text
p.RegisterSATSolver(logic2sat.Solver)
if p.IsSatisfiable(parser.MustParse("A | !A")) { /* tautología */ }
```

### Entailment
```text
kb := []p.Formula{parser.MustParse("A => B"), parser.MustParse("A")}
phi := parser.MustParse("B")
ok := entailment.Entails(kb, phi) // true
```

### Motor de reglas y explicación
```text
rules := []engine.Rule{
  {ID: "r1", When: parser.MustParse("A & B"), Then: p.Var("C")},
  {ID: "r2", When: parser.MustParse("C => D"), Then: p.Var("D")},
}
eng := engine.Engine{Facts: engine.FactBase{}, Rules: rules}
eng.Assert(p.Var("A")); eng.Assert(p.Var("B"))
eng.RunToFixpoint(5)
ok, why := eng.Query(parser.MustParse("D"))
fmt.Print(engine.PrettyExplain(why))
```

### Serialización JSON de reglas
```text
var buf bytes.Buffer
_ = engine.SaveRulesJSON(&buf, rules)
restored, _ := engine.LoadRulesJSON(&buf)
_ = restored
```


## Benchmarks y escalamiento
- Ubicación: `pkg/common/maths/logic2/examples/benchmarks_test.go`.
- Ejecutar: `go test -bench=. -benchmem ./pkg/common/maths/logic2/examples`.
- Criterio esperado (documentación pendiente de números): para `#vars > K` el camino SAT debe superar a truth-table; en familias 3-CNF con m≈4n, DPLL debe escalar razonablemente para n moderados.


## Límites conocidos y fuera de alcance del MVP
- No incluye CDCL avanzado ni heurísticas VSIDS.
- El motor de reglas es forward simple (sin agenda ni prioridades).
- Lógica de primer orden está fuera de alcance por ahora.


## Licencia
Ver LICENSE en la raíz del repositorio.
