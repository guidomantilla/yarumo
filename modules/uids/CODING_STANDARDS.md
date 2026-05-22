# Coding Standards — modules/uids/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../common/CODING_STANDARDS.md).

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `UID` interface + private `uid` impl; `UIDFn`, `IsUIDFn`, `NewUIDFn`, `RegisterFn`, `LookupFn`, `SupportedFn`, `ErrAlgorithmNotSupportedFn`, `ErrGenerationFn` in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | `UID` is public, `uid` is private |
| 4 | Constructor returns interface | Yes | `NewUID` returns `UID` |
| 5 | Options | No | Trivial 2-arg constructor; no Options struct |
| 6 | Preconfigured Default Singletons | Yes | Each provider sub-module exposes a preconfigured singleton (`cuid2.Cuid2`, `nanoid.NanoID`, `uuid.UuidV4`/`uuid.UuidV7`, `ulid.Ulid`, `xid.XId`). No `init()`-based auto-registration: consumers either use the singleton directly or call `uids.Register(...)` explicitly at startup. |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Top-level module (not under common/)

Five third-party UID providers (`github.com/akshayvadher/cuid2`,
`github.com/devmiek/nanoid-go`, `github.com/google/uuid`,
`github.com/oklog/ulid/v2`, `github.com/rs/xid`) collectively account for
a significant share of `modules/common/`'s transitive footprint. Most
consumers of `common` only need one or two of them. For that reason the
UID registry lives at the top-level module layer (`modules/uids/`) and
each provider lives in its own sub-module (`modules/uids/<provider>/`).
Consumers import the providers they need directly (no blank imports, no
init-based side effects):

```go
import "github.com/guidomantilla/yarumo/uids/uuid"

// Use the preconfigured singleton:
id, err := uuid.UuidV7.Generate()

// Or call the free function:
id, err := uuid.UUIDv7()

// Or register for centralized lookup at startup, if needed:
uids.Register(uuid.UuidV4)
uids.Register(uuid.UuidV7)
```

### Override: Shape A (registry + interface, no third-party deps)

`modules/uids/` itself is a Shape A package — its public API is the
registry functions (`Register`, `Lookup`, `Supported`) plus the
`NewUID` constructor (trivial — returns an immutable `UID` under interface,
no Options pattern). State (the `methods` map) is mutated through
`Register` only, mirroring the precedent set by `common/uids/extensions.go`.
Provider sub-modules are Shape A leaves that only expose free
`<ALGO>` and `Is<ALGO>` functions.
