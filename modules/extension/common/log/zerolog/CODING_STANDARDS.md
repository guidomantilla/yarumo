# Coding Standards — modules/extension/common/log/zerolog/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../../../../common/CODING_STANDARDS.md).

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `var _ clog.Logger = (*logger)(nil)` in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | `Logger` interface lives in `common/log`; impl `*logger` is private |
| 4 | Constructor returns interface | Yes | `NewLogger(opts ...Option) clog.Logger` |
| 5 | Options | Yes | `Options` + `WithLevel`/`WithWriter`/`WithConsole`/`WithTimeFormat`/`WithSampling`, defaults via `NewOptions` |
| 6 | Preconfigured Default Singletons | No | No registry, no facade, no singleton. Callers construct instances directly via `NewLogger(opts...)`. |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Sibling impl to `extension/common/log/slog/`

The slog sibling predates the workspace's canonical constructor pattern and
returns its own concrete `*Logger` struct (a documented exception for
"wrappers over stdlib types"). This module starts clean with the canonical
pattern: `NewLogger` returns the `clog.Logger` interface, and the impl
`logger` struct is private.

### Override: No registry, no pluggable, no Fn aliases on the constructor

Per the same reasoning as the resilience trio
(`extension/common/resilience/{breaker,limiter,retry}`), this module
deliberately drops the registry and pluggable-struct patterns:

- **No registry.** Consumers construct one `Logger` instance via `NewLogger`
  and wire it through `clog.Use`.
- **No pluggable function fields.** The private `logger` struct holds the
  underlying `zerolog.Logger` directly; the six log methods delegate
  through the `emit` helper.

Per PACKAGES.md L68 the constructor `NewLogger(opts ...Option) clog.Logger`
does NOT declare an Fn alias or compliance var — the contract is fixed by
the `Option` type at the entry and by `clog.Logger` at the output.

### Override: Wraps `github.com/rs/zerolog`

Layout follows the canonical Shape B template: `types.go` (compliance vars
+ package doc), `levels.go` (Level enum + zerolog mapping), `logger.go`
(private impl + `NewLogger` + the six log methods), `options.go`
(`Option`/`Options`/`With*`), `internals.go` (the `emit` helper and the
typed-arg dispatch).

### Override: `osExit` test seam

`logger.go` holds a package-level `osExit = os.Exit` indirection so the
Fatal path can be exercised in tests without terminating the process.
Mirrors the seam used by the slog sibling.

## Sub-packages

None. This module exposes a single package, `zerolog`.
