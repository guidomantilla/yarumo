# Execution Plan — Crypto & Standards (YA-0001 … YA-0034)

Phased execution order for the 34 issues currently open in [`guidomantilla/yarumo`](https://github.com/guidomantilla/yarumo/issues). Tracked in GitHub via [Milestones](https://github.com/guidomantilla/yarumo/milestones); this document is the source of truth for **why** the order is what it is.

> **Out of scope**: roadmap items in [`ROADMAP_NEW_MODULES.md`](ROADMAP_NEW_MODULES.md) that don't have tickets yet (e.g. `modules/boot/`, `modules/auth/`, `modules/datasource/`). They get their own execution plan once filed.

---

## Ordering criteria

Stacked from strongest to weakest:

1. **Hard dependencies** — A blocks B (B cannot land before A merges).
2. **Hygiene first** — mechanical fixes that clean the board before introducing big features.
3. **Standards enforcement before more code** — turn on the linter *before* writing the next batch.
4. **Bugs before features** — regression risk.
5. **Foundational refactors before features that depend on them**.
6. **Quick parallel wins** — algorithm parity, mechanical adds.
7. **Polish at the end** — benchmarks, fuzz, examples.

## Status legend

| Status | Meaning |
|---|---|
| `Open` | Not yet started |
| `In progress` | Branch exists, work underway |
| `Done` | Issue closed, PR merged |

Each milestone closes when all its issues are `Done`.

---

## Phase 0 — Standards hygiene

**Theme**: mechanical cleanup, no design decisions, lowest risk. Land before any new code.
**Estimated effort**: 1–2 days.
**Parallel**: all 5 issues can ship in parallel.

| # | Title | Why here | Blocks |
|---|---|---|---|
| **YA-0001** | inline map lookups in `common/expressions` tests | Mechanical refactor; required before YA-0003 | YA-0003 |
| **YA-0002** | remove `internal/deprecated/{passwords,servers}` modules | Workspace cleanup, kills testify transitive dep | — |
| **YA-0010** | move passwords config structs from `options.go` to `types.go` | Mechanical, zero behavior change | — |
| **YA-0012** | document registry `Get()` snapshot contract | Doc-only fix, applies across 9 subpackages | — |
| **YA-0003** | enforce *No Inline Assignments* via golangci-lint | Activates the guard before more code lands | depends on YA-0001 |

---

## Phase 1 — Crypto bug fixes

**Theme**: correctness before features. Three quick fixes, two require a small design decision.
**Estimated effort**: 2–3 days.
**Parallel**: 3 quick fixes parallel; 2 design-decision items can run in parallel after the call.

| # | Title | Notes |
|---|---|---|
| **YA-0004** | tokens — wrap raw `golang-jwt` error via sentinel | Quick |
| **YA-0006** | passwords — bump ScryptN, BcryptCost to OWASP 2024 | Quick (constant change + benchmark) |
| **YA-0008** | tokens — stop pre-generating 64-byte key in `NewOptions` | Quick (lazy generation) |
| **YA-0005** | passwords — `WithXxxParams` AND validation silently rejects | **Design decision**: per-field minima vs. log warn |
| **YA-0007** | hashes — `Hash()` panics on unavailable `crypto.Hash` | **Design decision**: signature change vs. validate at `NewMethod` |

---

## Phase 2 — Algorithm parity

**Theme**: trivial registry adds. All five completely independent.
**Estimated effort**: 1–2 days.
**Parallel**: all 5 in parallel — could ship as a single PR each.

| # | Title |
|---|---|
| **YA-0013** | hashes — register SHA3-384, SHA-224, SHA-1 |
| **YA-0014** | hmacs — register HMAC-SHA384 |
| **YA-0015** | ecdsas — register ECDSA-SHA384-over-P384 |
| **YA-0016** | rsassas — register RSASSA-PSS-SHA384 |
| **YA-0017** | rsaoaep — register RSA-OAEP-SHA384 |

---

## Phase 3 — Foundation refactors

**Theme**: two refactors that unblock the bulk of Phase 4 features.
**Estimated effort**: 3–5 days.
**Parallel**: YA-0009 and YA-0021 are independent — can run in parallel.

| # | Title | Unblocks |
|---|---|---|
| **YA-0009** | tokens — introduce `Algorithm` enum, drop `jwt.SigningMethod` from public API | YA-0011, YA-0018, YA-0019 |
| **YA-0021** | signers/* and ciphers/rsaoaep — PEM marshal/parse for keys | YA-0018, YA-0033 |

---

## Phase 4 — Major features

**Theme**: the heavy lifts. Where the user-visible value lands.
**Estimated effort**: 1–2 weeks.
**Parallel**: 4 of the 6 can run in parallel; YA-0018 and YA-0019 share the YA-0009 dep.

| # | Title | Depends on | Parallel with |
|---|---|---|---|
| **YA-0011** | tokens — package doc cleanup | YA-0009 | all others |
| **YA-0019** | tokens — opaque (encrypted) tokens via AEAD | YA-0009 | YA-0018, YA-0020, YA-0022 |
| **YA-0018** | tokens — JWT RS/PS/ES/EdDSA variants | YA-0009 + YA-0021 | YA-0019, YA-0020, YA-0022 |
| **YA-0020** | passwords — `DelegatingPasswordEncoder` | — | all others |
| **YA-0030** | passwords — rename `Argon2` → `Argon2id`, add `Argon2i` | coordinates with YA-0020 | YA-0019, YA-0018 |
| **YA-0022** | certs — load/parse helpers, CSR, multi-algo SelfSigned | — | all others |

---

## Phase 5 — Capabilities

**Theme**: extend the API surface. Mostly independent.
**Estimated effort**: 1–2 weeks.
**Parallel**: most are independent; YA-0026 → YA-0027 is the only chain.

| # | Title | Notes |
|---|---|---|
| **YA-0024** | passwords — public salt API | Quick |
| **YA-0025** | tokens — `DecodeUnsafe` peek | Quick |
| **YA-0023** | streaming Hash and AEAD APIs | Medium |
| **YA-0028** | `Method` `MarshalText` / `UnmarshalText` | Enables YA-0029 conceptually |
| **YA-0029** | cross-algorithm convenience helpers driven by registry | After YA-0028 |
| **YA-0026** | new subpackage `common/crypto/kdfs/` | **Blocks YA-0027** |
| **YA-0027** | new subpackage `common/crypto/ciphers/hybrid/` (ECIES/HPKE) | After YA-0026 |
| **YA-0034** | passwords — `WithSecureDefaults` helper | Coordinates with YA-0006, YA-0030 |

---

## Phase 6 — Polish

**Theme**: tests, benchmarks, examples. The "looks like a finished library" pass.
**Estimated effort**: 1 week.
**Parallel**: all 3 can run in parallel.

| # | Title | Depends on |
|---|---|---|
| **YA-0031** | benchmarks per algorithm under `examples/` | — |
| **YA-0032** | fuzz tests for tokens and certs parsers | — |
| **YA-0033** | round-trip examples crossing process boundaries (disk persistence) | YA-0021, YA-0022 |

---

## Critical path

The longest dependency chain across all phases:

```
Phase 0:  YA-0001 ──→ YA-0003
Phase 3:  YA-0009 ──→ YA-0018 ←── YA-0021
                  │
                  ↓
Phase 4:        YA-0019
                  │
Phase 6:          ↓
                YA-0033 ←── YA-0022
                       ←── YA-0021

Phase 5:  YA-0026 ──→ YA-0027
```

**Linearized (one developer, no parallelism)**: ~6–8 weeks.
**Aggressive parallelism (4 issues in flight)**: ~3–4 weeks.

---

## How to use this doc

1. **Start of work**: pick the lowest-numbered phase that still has open issues. Within a phase, pick anything with no remaining dependencies.
2. **PR time**: reference the issue (`Closes #N`). Milestone closes automatically when its last issue closes.
3. **Replanning**: if priorities shift, edit this file *first*, then re-assign issues to milestones. The doc is canonical.
4. **New tickets**: if a new YA-NNNN is filed mid-execution, add it to the appropriate phase here and assign its milestone.

---

## Mechanism

- **Active surface**: GitHub Milestones (one per phase). [Milestones list](https://github.com/guidomantilla/yarumo/milestones).
- **Source of truth**: this document.
- **Memory pointer**: an entry in `MEMORY.md` so future Claude sessions auto-load this context.
