# logic2 — Plan de implementación (MVP → Evolución)

Objetivo: Construir un nuevo paquete minimalista y potente para lógica proposicional con buen DX y un motor de reglas sencillo, evitando sobre‑ingeniería. Este documento es una guía operativa con fases, tareas concretas, entregables, criterios de aceptación y ejemplos.


## Principios guía
- KISS: empezar pequeño; sólo lo necesario para casos reales.
- Una sola vía por defecto para cada operación (políticas claras y documentadas).
- Dependencias unidireccionales y módulos pequeños.
- Tests y ejemplos como documentación viva.
- Medir antes de optimizar; umbrales y límites configurables.


## Diseño de paquetes (layout inicial)
- pkg/common/maths/logic2/props
  - AST de proposiciones y funciones puras (Eval, Vars, ToNNF, ToCNF, ToDNF, Simplify, TruthTable, Is*)
- pkg/common/maths/logic2/parser
  - Lexer y parser mínimo → AST; pretty‑printer
- pkg/common/maths/logic2/sat
  - CNF aplanada y solver DPLL (unit prop, pure literal, branching simple)
- pkg/common/maths/logic2/engine
  - Hechos (FactBase), reglas y ejecución forward simple; Explain mínimo
- pkg/common/maths/logic2/examples
  - Ejemplos end‑to‑end y fragmentos reproducibles

Reglas de dependencia:
- propositions no depende de nadie
- parser → propositions
- sat → propositions
- engine → propositions (no depende de sat ni de parser)
- examples → todos


## Fase 0 — Preparación y cimientos (0.5–1 día) — COMPLETADO
Tareas:
- [x] Crear paquetes y archivos README/PLACEHOLDER por paquete.
- [x] Estándares de código y comentarios GoDoc en públicos.
- [x] Decidir política de numeración y compatibilidad (API MVP estable en propositions).

Entregables:
- [x] Estructura de directorios creada con archivos iniciales y comentarios.
- [x] README corto en logic2 con visión y ejemplo mínimo (placeholder hasta Fase 1).

Criterios de aceptación:
- [x] go vet y go test (aunque sin tests aún) ejecutan sin fallos de compilación cuando existan skeletons.


## Fase 1 — Núcleo proposicional usable (3–5 días) — COMPLETADO
Objetivo: AST, evaluación, transformaciones y utilidades con parser/pretty básico.

Estado actual (hecho en repo):
- [x] AST completo en package props: Var, TrueF, FalseF, NotF, AndF, OrF, ImplF, IffF, GroupF con String/Eval/Vars.
- [x] Transformaciones y utilidades en props: ToNNF, ToCNF, ToDNF, Simplify (constantes, doble negación, idempotencia, absorción, complementos, trivialidades Impl/Iff).
- [x] Utilidades: TruthTable, Equivalent, FailCases, IsSatisfiable/IsContradiction/IsTautology (vía truth-table/resolución provisional).
- [x] Parser mínimo en package parser con precedencias (!, &, |, =>, <=>) y MustParse. Pretty-print básico vía String() de los nodos.
- [x] Ejemplos y pruebas tipo Test*: ejemplos de parse/eval, simplificación, equivalencias, tablas de verdad, round‑trip del parser.
- [x] Tests de propiedades formales (De Morgan, doble negación, idempotencia, absorción, identidad, dominación, complemento, conmutatividad, asociatividad y distributividad) con aserciones explícitas.

Tareas detalladas:
1) AST y semántica
- [x] Tipos: Var, True, False, Not, And, Or, Impl, Iff, Group.
- [x] Métodos en cada nodo: String(), Eval(Fact) bool, Vars() []string.
- [x] Tipo Fact: map[Var]bool.

2) Transformaciones y utilidades
- [x] ToNNF, ToCNF, ToDNF (basado en reglas estándar).
- [x] Simplify con reglas: constantes, doble negación, idempotencia, absorción, complemento, trivialidades en Impl/Iff.
- [x] TruthTable, Equivalent, FailCases.
- [x] IsSatisfiable/IsContradiction/IsTautology (provisional sin SAT; SAT llegará en Fase 2).

3) Parser + pretty‑printer mínimo
- [x] Gramática soportada:
  - NOT: !
  - AND: &
  - OR: |
  - IMPL: =>
  - IFF: <=>
  - Paréntesis: ( )
  - Variables: [A-Za-z_][A-Za-z0-9_]*
- [x] Precedencia (de menor a mayor): <=>, =>, |, &, ! (unario). Asociatividad izquierda salvo unarios.
- [x] Errores con posición aproximada; mensajes claros.
- [x] Pretty printing mínimo mediante String() respetando paréntesis.

4) API pública (snapshot)
- package props
  - [x] type Var string
  - [x] type Formula interface { String() string; Eval(Fact) bool; Vars() []string }
  - [x] type Fact map[Var]bool
- package parser
  - [x] Parse(input string) (props.Formula, error)
  - [x] MustParse(input string) props.Formula // helper para tests/examples
- package props
  - [x] Simplify/ToNNF/ToCNF/ToDNF
  - [x] TruthTable, Equivalent, FailCases
  - [x] IsSatisfiable, IsContradiction, IsTautology // provisional sin SAT

5) Tests y ejemplos
- [x] Round‑trip: parse → String → parse (TestParserRoundTrip).
- [x] Propiedades: De Morgan, doble negación, idempotencia, absorción, identidad, dominación, complemento, conmutatividad, asociatividad y distributividad (asserts dedicados en examples/properties_test.go).
- [x] Tablas de verdad y casos de fallo (ejemplos).
- [x] Equivalencias clásicas: A=>B ≡ !A|B; A<=>B ≡ (A&B)|(!A&!B).

Entregables:
- [x] Paquetes props y parser listos y en uso por ejemplos.
- [x] Ejemplos: construir fórmula, simplificar, tabla de verdad, equivalencias, round‑trip.

Criterios de aceptación:
- [x] Parser round‑trip en casos canónicos.
- [x] Leyes básicas validadas con tests de propiedades (ver examples/properties_test.go).


## Fase 2 — SAT y política de decisión (2–3 días) — EN PROGRESO
Objetivo: Integrar SAT/DPLL y definir política única para Is*.

Tareas detalladas:
1) CNF aplanada
- [x] FromFormulaToCNF(f Formula) (CNF, error): Simplify → ToCNF → descomponer a cláusulas.
- [x] Manejo de True/False, literales y grupos; detección de cláusulas tautológicas.

2) DPLL
- [x] Unit propagation, pure literal, branching por cláusula más corta.
- [x] Asignación y simplificación de CNF; utilidades de copia eficiente.
- [x] Retornar (ok bool, model Assignment) donde Assignment = map[Var]bool.

3) Política Is*
- [x] nVars ≤ K (por defecto 12, configurable) → truth table.
- [x] nVars > K → SAT (CNF + DPLL) cuando hay solver registrado (registro explícito; sin init()).
- [x] IsTautology(f) = !IsSatisfiable(!f). IsContradiction(f) = !IsSatisfiable(f).

4) Entailment
- [x] Entails(KB, phi) bool = ¬IsSatisfiable((∧KB) ∧ ¬phi).
- [x] EntailsWithCounterModel(KB, phi) (bool, Assignment) opcional (si SAT disponible).

5) Validación y benchmarks ligeros
- [x] Tests cruzados: para nVars ≤ 8, comparar truth table vs SAT (resultados idénticos) — ver examples/sat_policy_test.go.
- [x] Benchmarks: familias sintéticas para medir umbral K.

Benchmarks y escalamiento:
- Ubicación: pkg/common/maths/logic2/examples/benchmarks_test.go
- Cómo ejecutar: `go test -bench=. -benchmem ./pkg/common/maths/logic2/examples`
- Qué observar (criterio de “escalamiento razonable”):
  - Para nVars > K, el camino SAT debe superar claramente a truth-table en tiempo (la tabla crece 2^n).
  - En familias 3-CNF con m≈4n, DPLL debe escalar sub-exponencialmente en la práctica para n moderados (p. ej., n≤200 en máquina local) sin agotar memoria.
  - Documentar números de referencia locales (tiempo/op y bytes/op) y, si es necesario, ajustar SATThreshold.
- Próximo paso: capturar resultados en README o un doc de benchmarks y marcar el criterio como cumplido.

Entregables:
- [x] Paquete sat con CNF+DPLL.
- [x] props.IsSatisfiable delega según política con umbral K y hook de registro para SAT.
- [x] Paquete entailment con Entails/EntailsWithCounterModel y pruebas.

Criterios de aceptación:
- [x] Suite cruzada pasa al 100% en nVars pequeños (tests agregados).
- [ ] Benchmarks muestran escalamiento razonable (pendiente; ver sección "Benchmarks y escalamiento").


## Fase 3 — Motor de reglas (MVP simple) (3–4 días) — COMPLETADO
Objetivo: Hechos booleanos, reglas proposicionales con consecuencia literal, ejecución a punto fijo y explicación mínima.

Tareas detalladas:
1) Datos y API
- [x] type FactBase map[propositions.Var]bool
  - [x] Get(v) (bool, bool); Set(v, val); Retract(v); Merge(other).
- [x] type Rule struct { ID string; When propositions.Formula; Then propositions.Var }
- [x] type Engine struct { Facts FactBase; Rules []Rule }
- [x] Métodos:
  - [x] Assert(v propositions.Var), Retract(v propositions.Var)
  - [x] FireOnce() (fired []string) // evalúa reglas y dispara una pasada
  - [x] RunToFixpoint(maxIters int) (fired []string)
  - [x] Query(goal propositions.Formula) (ok bool, why *Explain)

2) Evaluación
- [x] Al evaluar When se consulta FactBase primero (Eval usa únicamente Facts; variables no presentes se asumen false en el MVP).
- [x] Disparo: si When evalúa true, Then se fija a true en Facts (sólo si no estaba ya true) y se registra el ID.

3) Explain mínimo
- [x] type Explain struct { ID string; Expr string; Value bool; Why string; Kids []*Explain }
- [x] Construir árbol durante Query con mensajes simples (p. ej., "left false", "both true", etc.).
- [x] PrettyExplain(*Explain) para impresión legible (mínimo; mejoras en Fase 4).

4) Ciclos y límites
- [x] maxIters en RunToFixpoint evita loops; determinismo por orden de Rules.

5) Ejemplos
- [x] 3–5 reglas sobre vars A..F; Assert hechos; RunToFixpoint; Query y PrettyExplain (examples/engine_test.go).

Entregables:
- [x] Paquete engine con API estable.
- [x] Ejemplos end‑to‑end en examples.

Criterios de aceptación:
- [x] RunToFixpoint converge dentro de maxIters; Explain es legible y consistente.


## Fase 4 — DX y serialización (2–3 días, opcional)
Objetivo: Intercambio de reglas y explicaciones, mejora de usabilidad.

Tareas (detalladas):
- DTOs explícitos y versionados (v1):
  - RuleDTO {version,id,when,then} donde when es string (parseable) y then es nombre de variable.  
  - RuleSetDTO {version,rules}.  
  - ExplainDTO {expr,value,why,kids[]} sin referencias cíclicas.  
  - Nota: No serializar structs internos del AST; usar strings de fórmulas como contrato estable.
- Serialización JSON/YAML (opcional YAML):
  - LoadRulesJSON/SaveRulesJSON (y variantes YAML si se incluye): io.Reader/io.Writer para testabilidad.  
  - Round‑trip estable (encode→decode→encode ≈ estable modulo espacios).
- Pretty‑printing y helpers:
  - PrettyExplainTo(w io.Writer, *Explain) (determinista).  
  - Mantener String() canónica; añadir Format(f, opts) opcional (Unicode/espaciado) si aplica.
- Parser DX y errores mejores:
  - ParseError con posición (byte/columna) y mensaje claro.  
  - (Opcional) ParseWith(input, ParseOptions{Strict}) para activar/desactivar sinónimos.  
  - Soporte de sinónimos/Unicode en lexer (keywords AND/OR/NOT/THEN/IFF; operadores &&,||,->,<->; símbolos ¬∧∨→⇒↔⇔; TRUE/FALSE) sin usar reemplazos globales.
- Observabilidad mínima del motor:
  - EngineStats {RulesEvaluated, RulesFired, Iterations}.  
  - PrettyExplain mantiene orden estable de hijos; Kids en orden determinista.
- Utilidades de intercambio:
  - LoadFactsJSON/SaveFactsJSON: map plano {"A":true,...}.  
  - Serializar contra‑modelos SAT: AssignmentDTO map[string]bool.  
  - (Opcional) CNFDTO para inspección/depuración.
- CLI opcional (playground):
  - logic2cli: parse/sat/entails/run con archivos JSON de reglas y hechos.
- Golden tests y determinismo:
  - Golden files de PrettyExplain.  
  - Round‑trip de RuleSetDTO JSON/YAML.  
  - Orden estable de Vars() y salidas.
- Documentación:
  - Errores comunes del parser, límites (SATThreshold, maxIters) y guías de uso.

Entregables:
- engine/dto.go con RuleDTO, RuleSetDTO, ExplainDTO y helpers ToDTO/FromDTO.  
- engine/serialize.go con Load/Save JSON (y YAML opcional).  
- engine/pretty.go con PrettyExplainTo.  
- parser/errors.go con ParseError y (opcional) parser/options.go con ParseWith.  
- lexer actualizado con sinónimos/Unicode (si se activa esta parte en Fase 4).  
- Tests de golden y round‑trip.

Criterios de aceptación:
- Round‑trip JSON/YAML para conjuntos de reglas pequeñas; PrettyExplainTo produce salida estable (golden).  
- Parser reporta posición coherente en errores típicos; si se activa sinónimos, identifiers como "ANDY" no se corrompen.  
- EngineStats visible y consistente tras RunToFixpoint.


## Fase 5 — Mejoras del motor (opcional)
- Agenda de activaciones con estrategia FIFO como baseline.
- Índices por variable para reevaluación incremental eficiente.
- Métricas básicas (conteo de disparos, iteraciones).
- Backward chaining básico con ProofTree (si se requiere).


## Fase 6 — Primer orden (futuro/no prioritario)
- Términos, funciones, predicados n‑arios.
- Cuantificadores (∀, ∃), Skolemización y clausificación.
- Unificación (Robinson) y resolución.
- Integración parcial con motor de reglas.


## Hitos y checklist
- Hito A (Fase 1): AST + Parser + Simplify + TruthTable + Is* (sin SAT).
- Hito B (Fase 2): SAT operativo + política Is*; Entails.
- Hito C (Fase 3): Engine MVP con Explain + ejemplo completo.
- Hito D (Fase 4): Serialización y DX.

Checklist rápido por fase está al inicio de cada sección.


## Ejemplo objetivo (tras Fase 3)
```text
// Proposiciones
f := parser.MustParse("(A & B) => C")
if !props.Entails([]props.Formula{parser.MustParse("A"), parser.MustParse("B")}, parser.MustParse("C")) {
    panic("should entail")
}

// Motor de reglas
rules := []engine.Rule{
  {ID: "r1", When: parser.MustParse("A & B"), Then: props.Var("C")},
  {ID: "r2", When: parser.MustParse("C => D"), Then: props.Var("D")},
}
eng := engine.Engine{Facts: engine.FactBase{}, Rules: rules}
eng.Assert(props.Var("A"))
eng.Assert(props.Var("B"))
eng.RunToFixpoint(5)
ok, why := eng.Query(parser.MustParse("C"))
fmt.Println("C?", ok)
fmt.Println(engine.PrettyExplain(why))
```


## Fuera de alcance del MVP
- CDCL avanzado, heurísticas VSIDS, learned clauses.
- Agenda con prioridades complejas, recency/specificity.
- Lógica de primer orden completa.


## Riesgos y mitigaciones
- Complejidad del parser → gramática mínima, tests de round‑trip.
- Rendimiento de SAT → umbral K configurable y documentación; benchmarks para ajustar K.
- Bucles en reglas → maxIters estricto y documentado; pruebas con ciclos simples.


## Convenciones de código y testing
- GoDoc en públicos; nombres claros y consistentes (Var, Formula, Fact).
- Tests con tablas de casos y property‑based cuando aporte valor.
- Ejemplos en package examples ejecutables con `go test` (Example*).



## Anexo A — Política Is* (detallada)

Esta sección documenta explícitamente el punto “3) Política Is*” de la Fase 2, con detalles de configuración, registro del solver SAT y referencias a código.

### ¿Qué resuelve?
Provee una política única y clara para decidir cómo evaluar satisfacibilidad/tautología/contradicción:
- Para fórmulas con pocas variables se usa tabla de verdad (brute force).
- Para fórmulas con muchas variables se usa SAT (CNF aplanada + DPLL), si hay un solver SAT registrado.

### Comportamiento por defecto
- Umbral K por defecto: `12` (configurable en tiempo de ejecución).
- Si `nVars ≤ K` → se usa tabla de verdad.
- Si `nVars > K` y hay solver SAT registrado → se usa SAT.
- Si no hay solver registrado o este falla → se hace fallback a tabla de verdad.
- Definiciones derivadas:
  - `IsContradiction(f) = !IsSatisfiable(f)`
  - `IsTautology(f) = !IsSatisfiable(!f)`

### API y configuración
- Umbral configurable:
  - `props.SATThreshold int` (por defecto `12`).
- Registro del solver SAT (sin efectos colaterales de init):
  - `props.RegisterSATSolver(fn func(Formula) (ok bool, result bool))`

Ejemplo de registro (p. ej. en `main.go` o en `TestMain`):
```text
import (
  p "github.com/guidomantilla/yarumo/pkg/common/maths/logic2/props"
  s "github.com/guidomantilla/yarumo/pkg/common/maths/logic2/sat"
)

func init() {
  p.RegisterSATSolver(s.Solver)
  // p.SATThreshold = 12 // opcional: ajustar K
}
```

### Implementación (referencias de código)
- `pkg/common/maths/logic2/props/functions.go`
  - `SATThreshold` (líneas ~42–45).
  - `RegisterSATSolver` y variable `satSolver` (líneas ~46–53).
  - `IsSatisfiable` con la lógica de política (líneas ~55–75).
  - `IsContradiction` y `IsTautology` derivadas (líneas ~77–83).
- `pkg/common/maths/logic2/sat/solver.go`
  - `func Solver(f Formula) (bool, bool)` convierte a CNF y llama `DPLL`.
- `pkg/common/maths/logic2/sat/cnf.go`, `dpll.go`
  - Conversión a CNF aplanada y solver DPLL.
- Registro en pruebas (para usar SAT en tests):
  - `pkg/common/maths/logic2/examples/setup_test.go` usa `TestMain` para registrar el solver.

### Validación
- Pruebas cruzadas (tabla vs SAT) para `nVars ≤ 8`:
  - `pkg/common/maths/logic2/examples/sat_policy_test.go`.
- Caso de umbral (fórmula con >K variables) también en `sat_policy_test.go`.

### Notas
- El paquete legacy `pkg/common/maths/logic` mantiene su propia implementación; la política Is* aquí descrita aplica al paquete nuevo `logic2`.
- No se utiliza `init()` para registrar el solver en producción; el registro es explícito para evitar efectos ocultos.
