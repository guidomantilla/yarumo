# logic2 — Plan de implementación (MVP → Evolución)

Objetivo: Construir un nuevo paquete minimalista y potente para lógica proposicional con buen DX y un motor de reglas sencillo, evitando sobre‑ingeniería. Este documento es una guía operativa con fases, tareas concretas, entregables, criterios de aceptación y ejemplos.


## Principios guía
- KISS: empezar pequeño; sólo lo necesario para casos reales.
- Una sola vía por defecto para cada operación (políticas claras y documentadas).
- Dependencias unidireccionales y módulos pequeños.
- Tests y ejemplos como documentación viva.
- Medir antes de optimizar; umbrales y límites configurables.


## Diseño de paquetes (layout inicial)
- pkg/common/maths/logic2/propositions
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


## Fase 1 — Núcleo proposicional usable (3–5 días)
Objetivo: AST, evaluación, transformaciones y utilidades con parser/pretty básico.

Tareas detalladas:
1) AST y semántica
- Tipos: Var, True, False, Not, And, Or, Impl, Iff, Group.
- Métodos en cada nodo: String(), Eval(Fact) bool, Vars() []string.
- Tipo Fact: map[Var]bool; helpers de set/merge (internos del paquete).

2) Transformaciones y utilidades
- ToNNF, ToCNF, ToDNF (basado en reglas estándar).
- Simplify con reglas: constantes, doble negación, idempotencia, absorción, complemento, trivialidades en Impl/Iff.
- TruthTable, Equivalent, FailCases.
- IsSatisfiable/IsContradiction/IsTautology (provisional: truth table o resolución simple si se reusa; SAT se integra en Fase 2).

3) Parser + pretty‑printer mínimo
- Gramática soportada:
  - NOT: !
  - AND: &
  - OR: |
  - IMPL: =>
  - IFF: <=>
  - Paréntesis: ( )
  - Variables: [A-Za-z_][A-Za-z0-9_]*
- Precedencia (de menor a mayor): <=>, =>, |, &, ! (unario, mayor precedencia). Asociatividad izquierda salvo unarios.
- Errores con posición aproximada; mensajes claros.
- Pretty printing con paréntesis mínimos según precedencia.

4) API pública (snapshot)
- package propositions
  - type Var string
  - type Formula interface { String() string; Eval(Fact) bool; Vars() []string }
  - type Fact map[Var]bool
  - Parse(input string) (Formula, error)
  - MustParse(input string) Formula // helper para tests/examples
  - Simplify/ToNNF/ToCNF/ToDNF
  - TruthTable, Equivalent, FailCases
  - IsSatisfiable, IsContradiction, IsTautology // provisional sin SAT

5) Tests
- Round‑trip: f == Parse(f.String()) para el subconjunto soportado.
- Propiedades: De Morgan, doble negación, idempotencia, absorción.
- Tablas de verdad de ejemplos pequeños.

Entregables:
- Paquetes propositions y parser con cobertura mínima (>70% en propositions y parser básicos).
- Ejemplos: construir fórmula, simplificar, tabla de verdad.

Criterios de aceptación:
- Todas las leyes básicas pasan; parser round‑trip en casos canónicos.


## Fase 2 — SAT y política de decisión (2–3 días)
Objetivo: Integrar SAT/DPLL y definir política única para Is*.

Tareas detalladas:
1) CNF aplanada
- FromFormulaToCNF(f Formula) (CNF, error): Simplify → ToCNF → descomponer a cláusulas.
- Manejo de True/False, literales y grupos; detección de cláusulas tautológicas.

2) DPLL
- Unit propagation, pure literal, branching por cláusula más corta.
- Asignación y simplificación de CNF; utilidades de copia eficiente.
- Retornar (ok bool, model Assignment) donde Assignment = map[Var]bool.

3) Política Is*
- nVars ≤ K (por defecto 12, configurable) → truth table.
- nVars > K → SAT (CNF + DPLL).
- IsTautology(f) = !IsSatisfiable(!f). IsContradiction(f) = !IsSatisfiable(f).

4) Entailment
- Entails(KB, phi) bool = ¬IsSatisfiable((∧KB) ∧ ¬phi).
- EntailsWithCounterModel(KB, phi) (bool, Assignment) opcional (si SAT disponible).

5) Validación y benchmarks ligeros
- Tests cruzados: para nVars ≤ 8, comparar truth table vs SAT (resultados idénticos).
- Benchmarks: familias sintéticas para medir umbral K.

Entregables:
- Paquete sat con CNF+DPLL.
- propositions.IsSatisfiable delega según política (exponer parámetro opcional o variable de paquete para K).

Criterios de aceptación:
- Suite cruzada pasa al 100%; benchmarks muestran escalamiento razonable.


## Fase 3 — Motor de reglas (MVP simple) (3–4 días)
Objetivo: Hechos booleanos, reglas proposicionales con consecuencia literal, ejecución a punto fijo y explicación mínima.

Tareas detalladas:
1) Datos y API
- type FactBase map[propositions.Var]bool
  - Get(v) (bool, bool); Set(v, val); Retract(v); Merge(other).
- type Rule struct { ID string; When propositions.Formula; Then propositions.Var }
- type Engine struct { Facts FactBase; Rules []Rule }
- Métodos:
  - Assert(v propositions.Var), Retract(v propositions.Var)
  - FireOnce() (fired []string) // evalúa reglas y dispara una pasada
  - RunToFixpoint(maxIters int) (fired []string)
  - Query(goal propositions.Formula) (ok bool, why *Explain)

2) Evaluación
- Al evaluar When se consulta FactBase primero: si aparece Var en hechos, usar ese valor; si no, utilizar valor “libre” (se interpreta como variable no asignada → debe venir desde Facts; si no está, Eval usa false por defecto o se exige que todas las variables del When estén en Facts; decidir: MVP usará false por defecto y se documenta). Simplificar: para Query/When, Eval usa únicamente Facts.
- Disparo: si When evalúa true, Then se fija a true en Facts.

3) Explain mínimo
- type Explain struct { ID string; Expr string; Value bool; Why string; Kids []*Explain }
- Construir árbol durante Query: para hojas Var, Why="fact: A=true" o "fact: A=false"; para operadores, mensajes simples p.ej. "A & B es falso porque B=false".

4) Ciclos y límites
- maxIters en RunToFixpoint evita loops; documentar determinismo (orden de Rules).

5) Ejemplos
- 3–5 reglas sobre vars A..F; Assert algunos hechos; RunToFixpoint; Query y PrettyExplain.

Entregables:
- Paquete engine con API estable.
- Ejemplos end‑to‑end en examples.

Criterios de aceptación:
- RunToFixpoint converge dentro de maxIters; Explain es legible y consistente.


## Fase 4 — DX y serialización (2–3 días, opcional)
Objetivo: Intercambio de reglas y explicaciones, mejora de usabilidad.

Tareas:
- JSON/YAML para Rule y Explain (DTOs simples sin referencias cíclicas).
- ParseRules([]byte) y PrettyExplain(*Explain) helpers.
- Documentación de errores comunes y guías de uso.

Entregables:
- Subpaquete o archivo dto.go en engine y funciones en parser para serializar fórmulas si aplica.

Criterios de aceptación:
- Round‑trip JSON/YAML para conjuntos de reglas pequeñas; PrettyExplain produce salida estable.


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
```go
// Proposiciones
f := propositions.MustParse("(A & B) => C")
if !propositions.Entails([]propositions.Formula{propositions.MustParse("A"), propositions.MustParse("B")}, propositions.MustParse("C")) {
    panic("should entail")
}

// Motor de reglas
rules := []engine.Rule{
  {ID: "r1", When: propositions.MustParse("A & B"), Then: propositions.Var("C")},
  {ID: "r2", When: propositions.MustParse("C => D"), Then: propositions.Var("D")},
}
eng := engine.Engine{Facts: engine.FactBase{}, Rules: rules}
eng.Assert(propositions.Var("A"))
eng.Assert(propositions.Var("B"))
eng.RunToFixpoint(5)
ok, why := eng.Query(propositions.MustParse("C"))
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
