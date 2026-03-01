# Coding Standards

This package follows the conventions defined in
[`modules/common/CODING_STANDARDS.md`](../common/CODING_STANDARDS.md)
with the overrides documented below.

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | Formula interface, node types implement it |
| 3 | Public Interface, Private Implementation | Partial | Formula is public interface; all structs are public value types |
| 4 | Constructor returns interface | Partial | Parser returns Formula; NewCPT/NewFactor return concrete types |
| 5 | Options | Partial | Only logic/parser uses functional options (WithStrict) |
| 6 | Preconfigured Default Singletons | No | RegisterSATSolver is a hook, not the Use pattern |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Public Structs as Value Types

All packages use **public structs** with **public fields** as value types.
Users construct them directly — there are no constructors for data types:

    f := logic.AndF{L: logic.Var("A"), R: logic.Var("B")}
    s := fuzzy.Set{Name: "cold", Fn: fuzzy.Constant(0.5)}

NewCPT and NewFactor are convenience helpers that return concrete types,
not interface-returning constructors.

### No Receiver Asserts

Methods operate on value types (structs, not pointers). There is no nil receiver
risk. Assert calls are not needed.

### Simple Error Sentinels

All packages use simple `errors.New` sentinels instead of the `errs.TypedError`
pattern. Exception: `logic/parser` uses structured `ParseError` with fields
(Pos, Col, Msg) for position-aware error reporting.

### SAT Solver Hook

RegisterSATSolver follows a registration hook pattern instead of the Use singleton
pattern. The solver is a package-level function variable, not a named instance.

## Reviewed Packages

- [x] maths/logic
- [x] maths/logic/parser
- [x] maths/logic/sat
- [x] maths/logic/entailment
- [x] maths/probability
- [x] maths/fuzzy
