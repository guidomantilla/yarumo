# Coding Standards

This package follows the conventions defined in
[`modules/common/CODING_STANDARDS.md`](../common/CODING_STANDARDS.md)
with the overrides documented below.

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | Engine (deductive, bayesian, fuzzy, causal), FactBase, EvidenceBase, Network, SCM, Rule (deductive, fuzzy), Variable |
| 3 | Public Interface, Private Implementation | Yes | |
| 4 | Constructor returns interface | Yes | NewEngine (×4), NewFactBase, NewEvidenceBase, NewNetwork, NewSCM, NewRule (×2), NewVariable |
| 5 | Options | Yes | engine (deductive, bayesian, fuzzy), rules (deductive, fuzzy), variable, parser |
| 6 | Preconfigured Default Singletons | No | |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

- **explain/ packages use public structs**: Origin, Provenance, Phase, Step, Trace, etc. are public
  value types (like math/logic nodes). No interface needed — they are pure data. Applies to all
  five paradigms: deductive/explain/, bayesian/explain/, fuzzy/explain/, causal/explain/, mcdm/explain/.
- **mcdm/ uses functions, not Engine interface**: AHP and TOPSIS are standalone decision methods
  exposed as functions. No unified Engine interface.
- **No dto/ package**: Serialization is out of scope for the pure engine.

## Math Packages

| Package | Coverage | Issues | Notes |
|---------|----------|--------|-------|
| logic | 100% | 0 | Formula interface, Var/Fact types |
| logic/entailment | 100% | 0 | |
| logic/parser | 100% | 0 | Options pattern |
| logic/predicate | 100% | 0 | ForAll, Exists, Count, Filter |
| logic/sat | 100% | 0 | DPLL solver |
| logic/temporal | 100% | 0 | ResponseWithin, Eventually, etc. |
| fuzzy | 100% | 0 | MembershipFn, Degree type |
| sets | 100% | 0 | Generic Set[T] |
| stats | 100% | 0 | 12 distributions, RunningStats, WindowedStats |

## Engine Packages

| Package | Coverage | Issues | Notes |
|---------|----------|--------|-------|
| deductive/explain | 100% | 0 | Value types, no interface |
| deductive/rules | 100% | 0 | Rule interface + Options |
| deductive/facts | 100% | 0 | FactBase interface |
| deductive/engine | 100% | 0 | Engine interface + Options |
| bayesian | 100% | 0 | CPT, Factor types |
| bayesian/explain | 100% | 0 | Value types, no interface |
| bayesian/network | 100% | 0 | Network interface |
| bayesian/evidence | 100% | 0 | EvidenceBase interface |
| bayesian/engine | 100% | 0 | Engine interface + Options |
| fuzzy | 100% | 0 | Error types |
| fuzzy/explain | 100% | 0 | Value types, no interface |
| fuzzy/variable | 100% | 0 | Variable interface + Options |
| fuzzy/rules | 100% | 0 | Rule interface + Options |
| fuzzy/engine | 100% | 0 | Engine interface + Options |
| causal | 100% | 0 | Error types |
| causal/explain | 100% | 0 | Value types, no interface |
| causal/model | 100% | 0 | SCM interface |
| causal/engine | 100% | 0 | Engine interface |
| mcdm | 100% | 0 | Error types |
| mcdm/explain | 100% | 0 | Value types, no interface |
| mcdm/ahp | 100% | 0 | AHP function-based |
| mcdm/topsis | 100% | 0 | TOPSIS function-based |

## Acceptance Tests

The `tests/acceptance/` module contains cross-paradigm integration tests that exercise
complete workflows spanning multiple engine packages. These tests validate that paradigms
compose correctly and produce expected results end-to-end.

- Coverage thresholds are set to 0% (tests exercise engine/math code, not local code).
- All files are `_test.go` — linter exclusions for `funlen`, `cyclop`, `gocognit`, `gocyclo` apply.
