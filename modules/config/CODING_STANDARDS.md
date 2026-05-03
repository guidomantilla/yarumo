# Coding Standards

This package follows the conventions defined in [`modules/common/CODING_STANDARDS.md`](../common/CODING_STANDARDS.md) with the overrides documented below.

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | No | No types or interfaces defined |
| 3 | Public Interface, Private Implementation | No | No interfaces defined |
| 4 | Constructor returns interface | No | No constructors |
| 5 | Options | No | No options pattern |
| 6 | Preconfigured Default Singletons | No | One-shot bootstrap function, not a singleton |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: No Error Pattern

`Default` does not return errors. Configuration failures (e.g. invalid log level) fall back to safe defaults silently. If the package evolves to return errors, the `common/errs` pattern must be adopted.

### Override: Private Helper Functions

Private helper functions (e.g. `parseLevel`) do not require function type contracts. They are internal implementation details of the bootstrap function.

## Reviewed Packages

- [x] config
