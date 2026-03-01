# Coding Standards

This package follows the conventions defined in [`modules/common/CODING_STANDARDS.md`](../../common/CODING_STANDARDS.md) with the overrides documented below.

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | Function types for exported functions |
| 3 | Public Interface, Private Implementation | No | No interfaces, concrete types only |
| 4 | Constructor returns interface | No | Standalone functions |
| 5 | Options | Yes | `Options` + `With*` with validation |
| 6 | Preconfigured Default Singletons | No | One-shot functions |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: No Parallel Tests in Functions

Function tests modify global OpenTelemetry state (providers, propagators); subtests within `TestTracer`, `TestMeter`, `TestLogger`, and `TestObserve` cannot be parallel.

### Override: No Constructor Pattern

This package exposes standalone setup functions (`Tracer`, `Meter`, `Logger`, `Observe`), not constructors that return instances.

## Reviewed Packages

- [x] otel
