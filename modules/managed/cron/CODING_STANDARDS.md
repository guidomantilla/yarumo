# Coding Standards — modules/managed/cron/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../../common/CODING_STANDARDS.md).

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `Scheduler` interface in `types.go`; `var _ Scheduler = (*cron.Cron)(nil)` |
| 3 | Public Interface, Private Implementation | Yes | `Scheduler` is public; the impl is `*cron.Cron` from `github.com/robfig/cron/v3` |
| 4 | Constructor returns interface | Yes | `NewScheduler(options ...cron.Option) Scheduler` |
| 5 | Options | No | Options are forwarded as the underlying `cron.Option` variadic; no module-owned `Options` struct |
| 6 | Preconfigured Default Singletons | No | No singleton; each `NewScheduler` call owns its inner `*cron.Cron` |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Top-level module (not under common/)

`modules/common/` is a pure library with no lifecycle opinions. The
`Scheduler` interface returned by `NewScheduler` is a lifecycle-owning
abstraction — `Start()` launches a goroutine, `Stop()` halts it and
returns a context that completes when running jobs finish. For that
reason the cron wrapper lives at the top-level module layer alongside
`modules/managed/`, `modules/managed/grpc/`, `modules/managed/cache/`, `modules/managed/telemetry/`
and `modules/config/`, never inside `modules/common/`.

### Override: Exception shape (thin wrapper over external library)

This is an Exception package per `modules/PACKAGES.md` — a thin wrapper
over `github.com/robfig/cron/v3` whose only purpose is to expose the
`Scheduler` interface so consumers can program against the abstraction
instead of the concrete `*cron.Cron`. Layout is a minimal `types.go`
(package doc + interface + compliance var) + `scheduler.go` (the
`NewScheduler` constructor). No `functions.go`, no `errors.go`, no
`options.go` — the underlying library owns options, errors and helpers.
