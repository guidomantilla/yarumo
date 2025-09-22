# Logic Package Roadmap (yarumo)

## Objetivo
Construir un paquete de lógica en Go que soporte lógica proposicional completa, SAT para escalabilidad, motor de reglas con trazabilidad y una ruta a lógica de primer orden.

## Estado actual
- Conectivos: Var, True, False, Not, And, Or, Impl, Iff, Group
- Evaluación/compilación: PredicatesRegistry.Compile/Evaluate, EvalNode
- Reglas: RuleSet con forward chaining y consecuencias en registry
- Utilidades: tablas de verdad, tautología, contradicción, satisfacibilidad

## Roadmap por fases
### Fase 1 — Proposicional + DX
- [x] Completar ToNNF/ToCNF/ToDNF para todos los conectivos
- [x] SAT (DPLL) e integración en IsSatisfiable/IsTautology/IsContradiction
- [x] Entails(KB, φ) por refutación
- [ ] Separar KnowledgeBase vs FactBase; adaptar Evaluate
- [ ] Mejorar EvalNode (explicaciones, ids, referencia a regla)
- [ ] Parser de fórmulas y pretty-printer
- [ ] Serialización JSON/YAML para reglas y EvalNode

### Fase 2 — Motor de reglas
- [ ] Agenda de activaciones con prioridades/estrategia de conflicto
- [ ] Reevaluación incremental basada en cambios de hechos
- [ ] Detección de ciclos/recursión con límites
- [ ] Backward chaining básico con ProofTree
- [ ] Métricas y observabilidad

### Fase 3 — Primer Orden (opcional)
- [ ] Términos, funciones, predicados n-arios
- [ ] Cuantificadores (∀, ∃), Skolemización y clausificación
- [ ] Unificación (Robinson) y resolución
- [ ] Dominios e interpretaciones

### Fase 4 — Calidad y docs
- [ ] Tests de propiedades y benchmarks
- [ ] Ejemplos end-to-end
- [ ] Documentación de diseño y uso

## Tareas técnicas detalladas
### Propositions
- [x] Implementar ToNNF/ToCNF/ToDNF para NotF, OrF, ImplF, IffF, GroupF
- [ ] Tests de tablas de verdad y propiedades (De Morgan, doble negación)

### SAT y entailment
- [ ] sat/DPLL con unit-prop, pure-literal, branching heuristics simples
- [ ] Integrar SAT en Is* cuando Vars > k (p. ej., 12)
- [ ] Entails(KB, φ) = ¬Satisfiable(KB ∧ ¬φ)
- [ ] Benchmarks comparando contra truth table

### Engine (logic)
- [ ] FactBase: map[Var]bool con Get/Set/Retract/Merge
- [ ] KnowledgeBase: PredicatesRegistry[T]
- [ ] Evaluate: consultar FactBase antes que registry
- [ ] API Engine: Assert, Retract, Query, Explain

### RuleSet
- [ ] Agenda: tipo Activation {RuleID, Priority, Evidence}
- [ ] Estrategia de conflicto: first, specificity, recency
- [ ] Índice por variables; reevaluación incremental
- [ ] Detección de ciclos y max-depth
- [ ] BackwardChaining(goal) con ProofTree
- [ ] Explicaciones por regla (por qué/por qué no)

### Parser/Serialización
- [ ] Gramática: !/~ para NOT, & para AND, | para OR, => para IMPL, <=> para IFF
- [ ] Lexer+parser a AST existente; pretty-printer configurable
- [ ] JSON/YAML: Rule {label, formula, consequence, priority}
- [ ] Serializar EvalNode (label, expr, value, nodes, ruleId)

### Primer Orden (si aplica)
- [ ] Estructuras de términos y predicados n-arios
- [ ] Unificación y resolución
- [ ] Skolemización/clausificación
- [ ] Integración parcial con motor de reglas

## Criterios de aceptación
- [ ] IsSatisfiable usa SAT para nVars grandes y valida contra truth table para nVars pequeños
- [ ] Entails pasa suite de reglas clásicas (modus ponens, etc.)
- [ ] RuleSet con agenda evita loops y documenta activaciones
- [ ] BackwardChaining retorna ProofTree legible
- [ ] Parser round-trip f == Parse(f.String())

## Riesgos y mitigaciones
- SAT costoso → DPLL simple, límites y documentación
- Explosión de reglas → incrementalidad + índices por variable
- Parser ambiguo → BNF y precedencias con tests

## Referencias de código actuales
- pkg/common/maths/logic/forward_chaining.go
- pkg/common/maths/logic/registry.go
- pkg/common/maths/logic/eval_node.go
- pkg/common/maths/logic/propositions/*.go
- pkg/common/maths/logic/predicates/*.go
- pkg/common/maths/logic/entailment/*.go
- pkg/common/maths/logic/sat/*.go
