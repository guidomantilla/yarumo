# Spring Cloud Task — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-cloud-task (v5.0.1)
> **Analyzed**: 2026-05-16 (re-analysis after roadmap cleanup; supersedes previous "absorbed by jobs/" framing)
> **Recommendation**: PARTIAL — pattern is real and recurring; placement is a small new module `modules/tasks/` plus optional sub-package for event emission. Mode is PERMISIVO so this is proposed, not auto-filed.

The previous analysis rejected Spring Cloud Task on the grounds that it was "absorbed by `modules/jobs/` (§ 3.4)". The Yarumo roadmap was trimmed since then: § 3 (Brainstorm modules — `jobs/`, `audit/`, `scheduler/`, `batch/`, etc.) is gone. The current scope is § 1 (datasource / auth / messaging / health / boot), § 2 (routegen), § 4 (migration tracking). With the brainstorm tier removed, the question becomes: do these patterns deserve a *new* top-level module, or are they covered by the modules that survived the trim?

This re-analysis says **PARTIAL**: there is a genuine recurring pattern (finite execution → DB row → lifecycle hooks → optional event emission) that the surviving modules do not cover. It is small enough — and orthogonal enough — to warrant its own thin module rather than being smeared across `messaging/` + `datasource/gorm/` + ad-hoc consumer code.

---

## 1. Project summary

Spring Cloud Task (current GA: **5.0.1**, Spring Boot 3.x / Java 17+) scopes a single concept: **a finite, short-lived JVM application whose execution lifecycle is tracked in a database**. A "task" is one `main()` invocation that runs Spring Boot's `CommandLineRunner` / `ApplicationRunner` beans and then exits — at most-once execution, no scheduling, no orchestration. The framework writes one row to `TASK_EXECUTION` at start, updates it at end (or on failure), and optionally emits lifecycle events to a broker. That is the entire core.

**Three things glued together**:

1. **A persistence model** — `TaskExecution {executionId, taskName, startTime, endTime, exitCode, exitMessage, errorMessage, arguments, externalExecutionId, parentExecutionId}` stored via `TaskRepository` (in-memory for dev, JDBC `SimpleTaskRepository` for prod). Schema is five tables: `TASK_EXECUTION`, `TASK_EXECUTION_PARAMS`, `TASK_TASK_BATCH` (Spring Batch association), `TASK_LOCK` (single-instance enforcement), `TASK_SEQ` (id sequence).
2. **A lifecycle hook surface** — `TaskExecutionListener` with `onTaskStartup` / `onTaskEnd` / `onTaskFailed`, plus the `@BeforeTask` / `@AfterTask` / `@FailedTask` annotations. Hook precedence: `onTaskEnd` > `onTaskFailed` > `onTaskStartup`.
3. **Two opt-in integrations** — Spring Batch (associate every `JobExecution` to the enclosing `TaskExecution` via `TaskBatchExecutionListener`) and Spring Cloud Stream (publish task lifecycle on a `task-events` channel; also publish batch sub-events on `job-execution-events`, `step-execution-events`, `chunk-events`, etc.).

**Explicitly not**: a scheduler, an orchestrator, a queue, or a workflow engine. No retry, no schedule, no enqueue, no distributed coordination. Spring Cloud Data Flow is the orchestrator that *launches* tasks; Spring Cloud Task is the runtime instrumentation library each task application carries to make itself observable.

**Execution model in one paragraph**: at startup, a `SmartLifecycle#start` hook fires *before* any `*Runner` bean — at that point the framework writes a `TaskExecution` row with `startTime`, the task name resolved via `TaskNameResolver` (defaults to `spring.application.name`), and the command-line arguments. Then the runners execute in the usual Spring Boot order. On normal completion (`ApplicationReadyEvent` fires), the row is updated with `endTime`, `exitCode = 0`, and any `exitMessage` set by an `@AfterTask` listener. On failure (`ApplicationFailedEvent` fires), the row is updated with `endTime`, `exitCode = 1` (or whatever an `ExitCodeExceptionMapper` returns), and the exception's stack trace in `errorMessage`. The JVM then exits.

**Key distinctions from siblings**:

| Project | Lifecycle shape | Scope |
|---|---|---|
| Spring Cloud Task | Finite (start → run → exit), once | One JVM process |
| Spring Batch | Finite, chunked (job → steps → items), restartable | One JVM process, durable per-step state |
| Spring Cloud Stream | Continuous (long-running consumer) | One JVM process, message-driven |
| Spring Cloud Data Flow | Orchestrator | Launches tasks/streams onto a platform |

Three opinionated knobs round out the surface: `spring.cloud.task.single-instance-enabled` (prevent two tasks with the same name running concurrently — implemented with a Spring Integration JDBC lock on `TASK_LOCK`), `spring.cloud.task.closecontextEnabled` (close the `ApplicationContext` after runners finish, useful when non-daemon threads would otherwise pin the JVM), and `spring.cloud.task.batch.fail-on-job-failure` (without this, a failed Spring Batch job returns exit code 0 — a bug the user has to opt out of).

Two further configuration surfaces matter for accounting completeness: `spring.cloud.task.arguments` (pre-injected command-line arguments stored in `TASK_EXECUTION_PARAMS` for audit), and the trio `spring.cloud.task.executionid` / `external-execution-id` / `parent-execution-id` (let an orchestrator stamp the row with externally-meaningful identifiers before the task starts). The latter is how Data Flow correlates an orchestrator-side launch record with the task's self-recorded execution row — a small but important integration point.

Micrometer observation (`spring.cloud.task.observation.enabled`) exposes `spring.cloud.task` (timer), `spring.cloud.task.active` (long-task timer), `spring.cloud.task.runner` (per-runner timer) with tags for execution id, name, exit code, status, and parent/external ids.

## 2. Pareto features (top-20%)

| # | Feature | Description | Why it matters for Go microservices |
|---|---|---|---|
| 1 | **Finite-execution persistence row** | One DB row per execution with `id`, `name`, `start`, `end`, `exitCode`, `errorMessage`, `arguments` | The audit primitive. Any "task", "run", "command" gets a row that looks like this. Recurring pattern in DaaS (batch decisioning runs), KnowledgeForge (PDF→rules extraction runs), migration scripts. |
| 2 | **Lifecycle hooks (`onTaskStartup` / `onTaskEnd` / `onTaskFailed`)** | Synchronous before/after/fail callbacks with mutation hook (`setExitMessage`) | Maps to Go: `type Listener interface { OnStart(ctx, *Execution) ; OnEnd(ctx, *Execution) ; OnFailed(ctx, *Execution, err) }`. Framework-agnostic, dependency-free. |
| 3 | **Exit-code propagation discipline** | Explicit rule: failed runners must affect process exit code; `fail-on-job-failure` corrects Spring Batch's quiet-fail default | Specific anti-pattern: silent exit 0 on failure. Worth codifying as a documented invariant of any module that owns finite-execution semantics. |
| 4 | **Parent/external execution id** | `parentExecutionId` + `externalExecutionId` columns let a task identify itself as a child of an orchestrator run | Useful for tying a task to an outer trigger (HTTP request id, parent workflow, Data-Flow-style launcher). In yarumo, OTel trace id covers most cross-system correlation; a single `parent_execution_id` column covers the rest. |
| 5 | **Single-instance enforcement** | DB-level lock keyed on `taskName` prevents concurrent runs of the same task | Real need (cron-style migrations, periodic reports). One-row advisory lock in Postgres or a `TASK_LOCK`-style table suffices. |
| 6 | **Lifecycle events on a broker channel** | Publish `TaskExecution` start/end to `task-events` channel via Spring Cloud Stream | Lets an orchestrator react to task transitions without polling. In yarumo this is `modules/messaging/events/` — one publish call inside the lifecycle hook. |
| 7 | **`TaskNameResolver` + `executionId` injection** | Caller (orchestrator) can pre-stamp the task with a known id and name before launch | Solves the "I launched the process but don't know its DB row until it writes one" race. The launcher writes its own row, passes the id via env/flag, the task reuses it. |

Seven features. Three are Go-relevant and not covered elsewhere in the current roadmap: **#1 (persistence row)**, **#2 (lifecycle hooks)**, **#5 (single-instance)**. **#3** is a documented invariant, **#4 and #7** are column-level concerns, **#6** is a one-line emit into the existing `messaging/events/` façade.

Note what is **not** on this list: there is no scheduling primitive, no retry primitive, no fan-out primitive, no compensation primitive. Spring Cloud Task explicitly leaves all of those to other projects. That self-restraint is what makes it small enough to summarize in seven rows — and what makes the pattern reusable outside Spring's ecosystem.

## 3. Long-tail features (skip)

- **`TaskRepository` / `TaskExplorer` / `TaskConfigurer` strategy interfaces** — Spring's three-layer abstraction (write, read, factory) exists because Spring wants to swap H2/JDBC/etc. via DI. In Go, a single Postgres-backed type with a small `Repository` interface for testability covers this. No factory layer.
- **In-memory `TaskRepository` for dev** — yarumo's testing story is testcontainers (`modules/testing/containers/` is implied by Phase-2 plans); one execution model, not two.
- **`@BeforeTask` / `@AfterTask` / `@FailedTask` annotations** — annotations exist to wire a `TaskExecutionListener` without writing the interface. Go has no annotation processor; explicit interface is the only sensible form.
- **Spring Batch integration entirely** — yarumo has no batch framework (ETL/chunked-with-restart is out of scope). The whole `TaskBatchExecutionListener` + `TASK_TASK_BATCH` table + `spring-cloud-task-batch` artifact evaporates.
- **Spring Batch sub-event channels** (`job-execution-events`, `step-execution-events`, `chunk-events`, `item-read-events`, `item-process-events`, `item-write-events`, `skip-events`) — Spring Batch's own observability surface, not Cloud Task's. If a Go consumer needs item-level observability, OTel spans inside the handler cover it.
- **`SingleStepBatchJobAutoConfiguration` / single-step starter / `AmqpItemReader` / `FlatFileItemReader` / `JdbcCursorItemReader` / `KafkaItemReader` autoconfig** — Spring-Batch-specific convenience. Not applicable.
- **Remote partitioning via `DeployerPartitionHandler`** — Spring Cloud Deployer is the launch fabric for distributed batch workers. No yarumo equivalent planned.
- **`spring.cloud.task.closecontextEnabled`** — Spring-specific workaround for non-daemon threads pinning the JVM. Go has no equivalent problem: the process exits when `main()` returns.
- **`ApplicationFailedEvent` / `ExitCodeEvent` / `ExitCodeExceptionMapper` machinery** — three layers of Spring's exit-code resolution. The output (exit code 0 or 1) is trivial in Go; the layers exist only to compose with Spring Boot's other lifecycle events. Direct `os.Exit(code)` or return-from-`Run()` is enough.
- **`spring.cloud.task.observation.enabled` (Micrometer integration)** — yarumo uses OTel directly via `modules/telemetry/otel/`. A task instrumented at source covers the same need without an opt-in flag.
- **`TaskBatchExecutionListenerBeanPostProcessor` selective-job injection** — Spring DI accident. Not a problem outside auto-wiring.
- **`@EnableTaskLauncher` task sink (Stream-based launcher)** — Spring Cloud Stream message → launch a task as a separate process. Yarumo has no "launch a binary as a subprocess" story planned; if the need arises, the consumer wires it via `modules/messaging/` + `os/exec`.
- **Spring Cloud Data Flow integration** — Data Flow is the *launcher* of Cloud Task processes (`task launch`, scheduling, audit, REST API). Out of scope.
- **`SimpleTaskAutoConfiguration` / `@EnableTask`** — Spring Boot auto-configuration. The Go equivalent is "import the package and call its constructor"; no auto-config.

## 4. Mapping to Yarumo

**Existing § 1 modules with overlap**:

| Yarumo module | What it covers from Cloud Task | What's left over |
|---|---|---|
| `modules/datasource/gorm/` (§ 1.1) | The DB connection + transaction wrapper that any `TaskRepository`-style impl needs | The row schema, insert-at-start / update-at-end semantics, single-instance lock |
| `modules/datasource/` row-level audit hooks (`CreatedBy`/`CreatedAt`) | General audit columns on any table | Task-specific lifecycle: `exitCode`, `endTime`, `errorMessage` |
| `modules/managed/` (CronWorker, BaseWorker) | Long-running components with Start/Stop/Done | Not finite execution — a worker runs forever; a task runs once |
| `modules/messaging/events/` (§ 1.3) | Pub/sub envelope for `TaskStarted` / `TaskCompleted` / `TaskFailed` | The decision of *when* to publish, and the lifecycle hook surface |
| `modules/telemetry/otel/` | Trace ids, spans around the task body, metrics | The persistent row (OTel is volatile / sampled) |
| `modules/health/` (§ 1.4) | Long-running process probes | Orthogonal — tasks don't `/healthz` |
| `modules/boot/` (§ 1.5) | App wiring for long-running services | Finite execution is a different `Run()` shape; `boot.Run` assumes "wait for signal" |

**Gap to fill**: short-lived task lifecycle tracking — DB row + start/end hooks + single-instance lock + optional event emission. None of the modules above own this. `datasource/gorm/` gives the plumbing; `messaging/events/` gives the emit; what's missing is the orchestrator: a small piece that says "here is a function; run it; persist start/end/failure to a row; publish events if a publisher is wired; enforce single-instance if a lock name is given."

**Conceptual deltas to keep clean**:

- Cloud Task tracks **JVM process executions**. Yarumo's equivalent would track **a single function invocation inside any process** — could be a CLI binary's `main`, could be a one-shot handler inside a long-running worker, could be a manual admin trigger. The unit of accounting is "one logical execution of a named task", not "one process".
- Cloud Task assumes "1 task = 1 process, then exit". Yarumo should not assume that — a long-running worker can host the registry and run named tasks on demand (e.g., HTTP-triggered batch report). The Spring assumption is JVM-flavored; dropping it makes the module more general.
- Cloud Task's Spring Batch integration is the heaviest part of the project (sub-events, partitioning, single-step starter). Yarumo doesn't ship a batch framework and doesn't plan to. The integration evaporates entirely.

**Anti-patterns to avoid**:

1. **Annotations** — Spring's `@BeforeTask` / `@AfterTask` / `@FailedTask` are sugar over an interface. Skip; expose the interface only.
2. **Five-table schema** — `TASK_EXECUTION` + `TASK_EXECUTION_PARAMS` + `TASK_TASK_BATCH` + `TASK_LOCK` + `TASK_SEQ` is over-normalized for the Go shape. Collapse to one table (`task_execution`) with `arguments JSONB`, plus an advisory lock function (Postgres `pg_advisory_lock` on a hash of the task name) — no separate lock table needed.
3. **Strategy-interface factory (`TaskConfigurer`)** — three abstractions (Repository, Explorer, factory). Collapse to one `Repository` interface for testability; nothing else needs to be swappable.
4. **`@EnableTask` autoconfig** — implicit wiring is the opposite of Yarumo's explicit BeanFn pattern. The module exposes `tasks.New(db, opts...)` and consumers call it.
5. **Two persistence backends (memory + JDBC) shipped in one library** — Yarumo Postgres-only with `Repository` interface for tests.

**Pattern recurrence in yarumo's roadmap**: the finite-execution-with-audit-row shape shows up in at least four planned contexts —
- DaaS async decisioning (submit request → row → eval → row updated)
- KnowledgeForge extraction runs (PDF → run row → extracted rules → row updated)
- Migration / one-off admin commands (CLI `migrate db up` writing its own audit row)
- Periodic reports / cron tasks (scheduled or manually triggered, need lock + audit)

The choice is whether each of these reimplements the row + hooks + lock from scratch on top of `datasource/gorm/`, or shares a thin module.

## 5. Recommendation

**PARTIAL** — adopt the persistence + hook + single-instance core as a new thin module `modules/tasks/`. Skip everything else (batch, stream-launcher, autoconfig, annotations, Data Flow integration).

Rationale for "new module" rather than "ad-hoc on top of datasource/gorm/":

- The row + lifecycle pattern recurs in multiple consumers (DaaS, KnowledgeForge, migrations, reports). Shared code prevents drift.
- The module is small (estimated ≤ 600 LoC including tests) and orthogonal to everything else in § 1.
- It has lifecycle (the task runs and finishes; the registry has Start/Stop only if hosted in a worker), so per the placement principle it does not belong in `modules/common/`.
- It has an external SDK dep (`gorm.io/gorm` via `modules/datasource/gorm/`), so again not `common/`.
- It is finite-execution semantics, not lifecycle-component semantics (`managed/` is for components that run forever).
- It does not orchestrate, schedule, queue, or distribute — those are explicitly out of scope, matching Cloud Task's own self-restraint.

Rationale for "not just merge into `modules/datasource/gorm/`":

- `gorm/` is a driver. Adding a task-execution table + Run-with-audit semantics widens its scope from "DB plumbing" to "domain primitives". Better as a downstream consumer.

Rationale for the partial-not-full scope:

- Spring Batch integration, Stream launcher, autoconfig, and Data Flow integration are absent from any yarumo roadmap and do not pay for themselves.
- Single-step batch starter is Spring-Batch-specific.
- The exit-code-mapper machinery is a workaround for Spring Boot lifecycle composition; Go's `error` return + `os.Exit` is enough.

## 6. Proposed yarumo placement

**New module**: `modules/tasks/`

**Layout sketch**:

```
modules/tasks/
  task.go               Task[T any] interface (Run(ctx, args T) error), Name() string
  execution.go          Execution struct (Id, Name, StartTime, EndTime, ExitCode,
                        ExitMessage, ErrorMessage, Arguments JSONB,
                        ParentExecutionId, ExternalExecutionId, TraceId)
  listener.go           Listener interface (OnStart, OnEnd, OnFailed)
  registry.go           Registry: Register[T](name, task), Run(ctx, name, args) error
  repository.go         Repository interface (Insert, Update, Get, List)
  repository_gorm.go    GormRepository impl (depends on modules/datasource/gorm/)
  lock.go               SingleInstance(ctx, db, name, fn) using pg_advisory_xact_lock
  options.go            WithListener, WithEventPublisher, WithSingleInstance,
                        WithParentExecutionId, WithExternalExecutionId
  events.go             Optional sub-package wiring to modules/messaging/events/
  errs/                 ErrTaskAlreadyRunning, ErrTaskNotRegistered, etc.
```

**Public surface in words**:

```go
type Task[T any] interface {
    Name() string
    Run(ctx context.Context, args T) error
}

type Listener interface {
    OnStart(ctx context.Context, exec *Execution)
    OnEnd(ctx context.Context, exec *Execution)
    OnFailed(ctx context.Context, exec *Execution, err error)
}

func New(db *gorm.DB, opts ...Option) *Registry
func Register[T any](r *Registry, task Task[T])
func Run[T any](ctx context.Context, r *Registry, name string, args T, opts ...RunOption) (*Execution, error)
```

**What it depends on**:
- `modules/datasource/gorm/` (Repository impl, advisory lock helper)
- `modules/common/uids/` (execution id generation)
- `modules/common/errs/` (typed errors)
- `modules/messaging/events/` (optional — only when `WithEventPublisher` is wired)
- `modules/telemetry/otel/` (optional — trace id capture into the row; weak dep, no import cycle)

**What it does NOT depend on**:
- `managed/` — tasks are finite; not a lifecycle component. The registry itself can optionally implement `managed.Lifecycle` if hosted in a worker, but the core has no lifecycle.
- `boot/` — `Run()` can be called from any entry point (CLI `main`, HTTP handler, scheduled trigger).
- `messaging/` proper (only the `events/` façade, opt-in).

**Sub-module decisions inside `modules/tasks/`**:

- **`tasks/cli/`** (sketch, low prio) — convenience wrapper for the "task = CLI binary that runs once and exits" pattern (the literal Cloud Task shape on K8s `Job`). Parses flags, builds args struct, calls `Run`, returns `os.Exit(code)`. 50 LoC.
- **`tasks/http/`** (sketch, low prio) — gin handler that triggers a named task via POST and returns the execution row. Useful for ops dashboards.

Both are optional and downstream of the core. Neither needs to ship in v1.

**Pattern in usage** (DaaS async decisioning example):

```go
registry := tasks.New(db, tasks.WithEventPublisher(eventsPublisher))
tasks.Register(registry, &EvaluateDecisionTask{evalSvc: svc})

// In an HTTP handler:
exec, err := tasks.Run(ctx, registry, "evaluate-decision", reqArgs,
    tasks.WithExternalExecutionId(req.RequestId))
```

The handler returns `exec.Id` immediately if running async (registry has a worker pool), or synchronously after `Run` completes.

## 7. Open questions

1. **Sync vs. async execution**: does `Run` block until the task finishes (synchronous semantics like Spring's `*Runner`), or does it enqueue into a worker pool (async)? Likely sync by default — the caller decides whether to wrap in a goroutine. Async-with-pool is a v2 feature and may overlap with the future "in-process worker" idea that the deleted `jobs/` sketch covered.
2. **Where does the single-instance lock live?** Postgres advisory locks (session or transaction scoped) are the obvious answer for Postgres-only stacks. A separate `task_lock` table is required for MySQL/SQL-Server portability. Decision: ship Postgres-only via `pg_advisory_xact_lock`; add a `task_lock` table only if a non-Postgres consumer materializes.
3. **Run hosted in a long-lived worker vs. one-shot binary**: both are valid; the registry should work in either. The CLI sub-package (`tasks/cli/`) covers one-shot; the worker case is just "import the registry, call `Run` from any goroutine". Document both in the README; no separate module split.
4. **Relationship with eventual `managed/`-style scheduler**: if periodic execution becomes a need (cron-style "run this task every Monday"), is it (a) a new `tasks/scheduler/` sub-package, (b) consumer responsibility using `managed/CronWorker` + `Run`, or (c) defer until a real use case? Likely (b) until a use case shows up.
5. **Parent/external execution id semantics**: `ParentExecutionId` is local (another row in `task_execution`); `ExternalExecutionId` is opaque caller-supplied. Should there be a third `TraceId` column auto-populated from OTel context? Probably yes — saves every consumer from doing it manually. Weak dep on `telemetry/otel/` is fine because the column is optional.
6. **Argument serialization**: store as `JSONB` (typed via `T`) or as the Spring-style `List<String>` (command-line args)? JSONB is more useful for non-CLI triggers; CLI sub-package can stringify args before storing if `List<String>` is preferred. Default to JSONB.
7. **Schema migration**: ship as a Goose / sqlc / raw SQL migration? Yarumo has no migration convention yet — this module would set a precedent. Coordinate with future `datasource/gorm/` migration story before deciding.
8. **Naming: `tasks/` vs. `jobs/`**: "tasks" matches the Cloud Task vocabulary and the "finite execution" mental model; "jobs" connotes queueing/retry (riverqueue), which this module does *not* do. Recommend `tasks/`. (The deleted brainstorm `modules/jobs/` was queue-shaped — different concept.)
9. **Exit-code discipline as test**: should the "no silent-success on failure" invariant be enforced by an acceptance test in `modules/tasks/tests/`? Likely yes — the bug class (handler returns error, row says success) is exactly what acceptance tests catch.
10. **Listener exception handling**: Spring continues other listeners if one throws. What is yarumo's policy? Likely "log and continue" for `OnStart` / `OnEnd`, but `OnFailed` errors should be aggregated into the persisted error so they aren't lost.

## 8. ROADMAP delta proposed (NOT applied)

In `docs/ROADMAP_NEW_MODULES.md`, add a new entry under § 1:

```
## 1.6. `modules/tasks/` — Finite execution lifecycle (DB row + hooks + single-instance)

**Status**: Planned (low prio — fill when first consumer materializes: DaaS async
decisioning, KnowledgeForge extraction runs, or admin one-shot tooling)
**Why a new module**: finite-execution semantics is orthogonal to managed/ (which
covers long-running components) and to datasource/gorm/ (which is a driver). The
row + lifecycle hooks + single-instance lock pattern recurs in DaaS, KnowledgeForge,
migrations, and one-off admin tools — shared code prevents drift.

**Inspired by Spring Cloud Task** (analysis: docs/spring/spring-cloud-task.md).
Imports from Cloud Task: persistence row, lifecycle hooks, single-instance lock,
exit-code/failure discipline, optional event emission.
Skipped from Cloud Task: Spring Batch integration, Stream task launcher,
auto-configuration, annotations, Data Flow integration.

**Public surface**: tasks.New(db, opts), Register[T], Run[T], Listener interface,
Execution row struct.

**Internal deps**: modules/datasource/gorm/, modules/common/uids/, modules/common/errs/.
Optional weak deps: modules/messaging/events/ (event emission), modules/telemetry/otel/
(trace id capture).

**Sub-modules sketched** (not v1): tasks/cli/ (CLI-binary wrapper),
tasks/http/ (ops-dashboard trigger).

**Out of scope**: scheduling, retry, queueing (those are not a yarumo target — see
how Cloud Task itself defers them to Data Flow / Batch / Stream).
```

Additionally, in § 4.2 (Discarded migration items), leave Spring Cloud Task **un-listed** — it was never a go-feather-lib component to migrate; this is a green-field add inspired by an external framework, captured in `docs/spring/`.

Filing action when work starts: open `YA-NNNN: modules/tasks/ — finite execution lifecycle (Spring Cloud Task port, minus Batch/Stream/Data Flow)` referencing this analysis. Likely lands after Phase 3 alongside `modules/datasource/gorm/` (which is a hard prerequisite).
