# Coding Standards — modules/validation/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../common/CODING_STANDARDS.md).

## Module-specific overrides

- `builtins.go` and `yaml.go` carry package-level lookup tables
  (`builtins`, `reservedKeys`). They are intentionally global for O(1)
  dispatch and are exempt from `gochecknoglobals` via `.golangci.yml`.
- Engine errors use `EngineType = "validation-engine"`. Leaf violations
  retain the `"validation"` type inherited from `common/validation/` so
  callers can group user-facing violations under a single category.
- The `Engine.Validate(obj, ctx)` contract treats missing context fields in a
  `when:` expression as `false` (the conditional block is skipped). Other
  evaluation failures surface as `ErrWhenEvalFailed`.

## Examples

`examples/main.go` is a runnable demonstration: it loads a YAML ruleset,
runs the engine against two sample objects, and prints the aggregated
errors via `errs.AsErrorInfo`.
