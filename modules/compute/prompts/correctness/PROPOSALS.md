# Correctness Proposals

Proposals listed here require correctness analysis before implementation.
If approved, the change must propagate through the full acceptance test chain:

```
CORRECTNESS.md → specs/ACCEPTANCE_TESTS.md → specs/acceptance/*.md → tests/acceptance/*_test.go
```

---

## P-001: Allow nil equation for exogenous causal variables

**Package:** `engine/causal/model`
**Function:** `SCM.AddVariable`
**Current behavior:** Rejects `nil` equation for all variables, regardless of whether they have parents.
**Proposed behavior:** Accept `nil` equation when `len(parents) == 0` (exogenous/root variables). Exogenous variables take their values directly from observations, so an equation is not required.
**Rationale:** A root variable in a structural causal model has no upstream causes — its value is externally assigned. Requiring an equation forces callers to provide a no-op function (`func(_ map[string]float64) float64 { return 0 }`), which adds noise without adding correctness.
**Affected acceptance test:** `TestErrorContract_CausalSCM/nil_equation` in `error_contracts_test.go`
**Status:** Pending correctness review

---

## P-002: Convert timing-ratio performance tests to benchmark-only

**Section:** 5.2 (Scaling) and 5.3 (Algorithm Comparison)
**Files affected:** `07-PERFORMANCE.md`, `ACCEPTANCE_TESTS.md`, `performance_test.go`
**Current behavior:** Three test functions use `testing.Benchmark` or `time.Now()` to measure execution time ratios and fail with `t.Fatalf` when the ratio exceeds a threshold:

| Test | Mechanism | Threshold |
|------|-----------|-----------|
| `TestPerformance_Scaling_ratios/forward_chaining_ratio` | `testing.Benchmark` 10 vs 100 rules | ratio < 150 |
| `TestPerformance_Scaling_ratios/DPLL_ratio` | `testing.Benchmark` 5 vs 10 vars | ratio < 100 |
| `TestPerformance_Scaling_ratios/VE_ratio` | `testing.Benchmark` 3 vs 5 vars | ratio < 100 |
| `TestPerformance_VE_vs_Enumeration` | `time.Now()` manual, 100 iterations | ratio 0.1..10 |
| `TestPerformance_Mamdani_vs_Sugeno` | `time.Now()` manual, 100 iterations | ratio 0.1..10 |

**Problem:** These tests are inherently non-deterministic. Timing ratios depend on CPU load, OS scheduling, thermal throttling, and CI container resource contention. `TestPerformance_VE_vs_Enumeration` is already documented as flaky. The others can fail under load even when the algorithms are unchanged.

**Proposed behavior:** Split each ratio test into two parts:

1. **Correctness assertion (stays as `Test*`)** — For `VE_vs_Enumeration`, keep the posterior-match assertion (`diff > probTolerance`). This is deterministic and validates that both algorithms produce the same result. For `Mamdani_vs_Sugeno`, no correctness assertion exists — the test is purely about timing.
2. **Timing assertion (moves to `Benchmark*`)** — The ratio checks become benchmark functions that are only run explicitly with `go test -bench=...`. They report ratios via `b.ReportMetric` instead of failing.

**Rationale:** Performance tests in Section 5.1 (termination) use goroutine + `select` with generous timeouts — these are stable because they test for gross regressions (infinite loops, exponential blowup). The ratio tests in 5.2/5.3, however, assert fine-grained timing relationships that are not reproducible across environments. Converting them to benchmarks preserves the measurement capability without introducing false failures.

**What does NOT change:**
- Section 5.1 termination tests — stable, no modification needed
- Section 5.2 `Benchmark*` functions — already benchmarks, no modification needed
- `VE_vs_Enumeration` posterior-match assertion — deterministic, stays as `Test*`

**Affected acceptance tests:**
- `TestPerformance_Scaling_ratios` (3 subtests) in `performance_test.go`
- `TestPerformance_VE_vs_Enumeration` (timing ratio portion) in `performance_test.go`
- `TestPerformance_Mamdani_vs_Sugeno` in `performance_test.go`

**Status:** Pending correctness review
