# Coding Standards

This package follows the conventions defined in
[`modules/common/CODING_STANDARDS.md`](../common/CODING_STANDARDS.md)
with the overrides documented below.

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | Rule, FactBase, Engine interfaces |
| 3 | Public Interface, Private Implementation | Yes | |
| 4 | Constructor returns interface | Yes | NewRule, NewFactBase, NewEngine |
| 5 | Options | Yes | Rule (WithPriority), Engine (WithMaxIterations, WithStrategy) |
| 6 | Preconfigured Default Singletons | No | |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

- **explain/ packages use public structs**: Origin, Provenance, Phase, Step, Trace, etc. are public
  value types (like maths/logic nodes). No interface needed — they are pure data. Applies to
  classical/explain/, bayesian/explain/, and fuzzy/explain/.
- **No dto/ package**: Serialization is out of scope for the pure engine.

## Reviewed Packages

| Package | Coverage | Issues | Notes |
|---------|----------|--------|-------|
| classical/explain | 100% | 0 | Value types, no interface |
| classical/rules | 100% | 0 | |
| classical/facts | 100% | 0 | |
| classical/engine | 100% | 0 | |
| bayesian/explain | 100% | 0 | Value types, no interface |
| bayesian/network | 100% | 0 | |
| bayesian/evidence | 100% | 0 | |
| bayesian/engine | 100% | 0 | |
| fuzzy/explain | 100% | 0 | Value types, no interface |
| fuzzy/variable | 100% | 0 | |
| fuzzy/rules | 100% | 0 | |
| fuzzy/engine | 100% | 0 | |
