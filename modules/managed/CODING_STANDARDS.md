# Coding Standards

This package follows the conventions defined in [`modules/common/CODING_STANDARDS.md`](../common/CODING_STANDARDS.md) with the overrides documented below.

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `BuildFn[I, C]` + adapter → interface checks |
| 3 | Public Interface, Private Implementation | Yes | Interfaces public, adapters private |
| 4 | Constructor returns interface | Yes | `NewXxx` returns the adapter interface |
| 5 | Options | No | No options pattern |
| 6 | Preconfigured Default Singletons | No | Factory/builder pattern, not singletons |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Builders

The `Build*` functions are not classic constructors. They launch async goroutines and return a `Component[C]` plus a `StopFn`. They do not require individual function types — the generic `BuildFn[I, C]` compliance check covers all of them.

### Override: No Parallel Tests in Builders

Builder tests modify global state (logger) and launch goroutines; subtests within `TestBuild*` cannot be parallel.

## Reviewed Packages

- [x] managed
