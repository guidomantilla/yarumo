# Acceptance Tests — Functional Invariants

> These tests express **eternal domain truths**. They don't say "function X returns Y".
> They say "the system has this property". If a refactor breaks one of these, the refactor
> has a bug — not the test.
>
> Date: 2026-03-02 | Version: 1

---

## Guiding principle

These are **behavioral invariants** — the contract that survives any refactor. It doesn't matter
if tomorrow you replace the parser, change engine internals, move packages between modules, or
rewrite the SDK completely. These scenarios **remain true**.

---

## Abstraction layers

| Layer | What it protects | Analogy |
|---|---|---|
| **maths** | Mathematical axioms and algebraic properties | "The laws of physics don't change" |
| **inference** | Reasoning correctness | "If premises are true and logic is valid, the conclusion is true" |
| **sdk** | Service orchestration and contracts | "A judge applies the laws correctly to the case" |
| **daas** | User experience and business flows | "A citizen can present a case and receive an explained verdict" |

---

## modules/maths — Axioms and properties

These protect **mathematical truths**. It doesn't matter how you implement the parser or
evaluator — these properties are immutable.

### logic/

**Evaluation:**
- A tautology always evaluates to true regardless of variable values
- A contradiction always evaluates to false regardless of variable values
- The 22 formal laws of propositional logic hold (De Morgan, absorption, idempotence, double negation, distribution, etc.)

**Equivalence and transformations:**
- A formula transformed to NNF/CNF/DNF is logically equivalent to the original
- A simplified formula is logically equivalent to the original
- If two formulas are equivalent, they have the same truth table

**Satisfiability:**
- Every tautology is satisfiable
- No contradiction is satisfiable
- If a formula is satisfiable, there exists at least one assignment that makes it true
- If A entails B, then every model of A is a model of B

**Parsing (round-trip):**
- Any well-formed formula parsed and then serialized to string produces a formula logically equivalent to the original

### probability/

**Kolmogorov axioms:**
- Probabilities in a distribution always sum to 1.0
- No probability is negative
- No probability exceeds 1.0

**Bayes:**
- P(A|B) computed via Bayes is consistent with the definition: P(A|B) = P(B|A) * P(A) / P(B)
- Marginalizing a variable from a factor produces a factor with one fewer variable whose probabilities still sum correctly

**CPT:**
- Each row of a valid CPT sums to 1.0 per parent configuration

### fuzzy/

**Membership:**
- Membership value is always in [0, 1]
- A value outside the universe of discourse has membership 0
- The peak of a triangular function has membership 1

**T-norms / S-norms:**
- min(a, b) <= a and min(a, b) <= b (t-norm property)
- max(a, b) >= a and max(a, b) >= b (s-norm property)

**Defuzzification:**
- The centroid result always falls within the [min, max] range of the universe of discourse

---

## modules/inference — Reasoning correctness

These protect that **conclusions are correct**. Concrete business scenarios with known
inputs and outputs.

### classical/ — "The rule engine never reaches a wrong conclusion"

**Soundness (no false derivations):**
- Given a set of rules and initial facts, the engine never produces a fact that is not logically derivable from the rules

**Completeness (no missed truths):**
- Given a set of rules and facts, the engine finds ALL derivable facts, not just some

**Concrete scenarios:**
- Loan eligibility: 4 rules, 6 distinct client profiles -> each profile gets the correct decision
- Rules with priority: if two rules compete, the higher-priority one wins
- First-match: the engine stops after the first applicable rule
- Derivation chain: rule A derives fact X, rule B uses fact X to derive fact Y -> Y is derived
- No applicable rules: if no rule applies, the result is only the initial facts (nothing invented)

**Provenance:**
- Every derived fact has a trace: which rule derived it, at which step, from which facts

**Backward chaining:**
- Given a goal and rules, the engine determines if the goal is reachable
- If reachable, it identifies which initial facts support it

### bayesian/ — "Computed probabilities are mathematically correct"

**Algorithm equivalence:**
- Enumeration and Variable Elimination produce the same posterior distribution (within numerical tolerance)

**Concrete scenarios (textbook):**
- Medical diagnosis: symptoms -> disease probability (verified against manual calculation)
- 3-node network with evidence -> exact known posterior
- No evidence -> posterior = prior (absence of evidence changes nothing)
- Complete evidence -> degenerate distribution (certainty)

**Network validation:**
- A network with cycles is rejected
- A network with incomplete CPT is rejected
- A network with probabilities that don't sum to 1 is rejected

### fuzzy/ — "Fuzzy inference produces reasonable and reproducible results"

**Concrete scenarios:**
- Tipping system: service=excellent, food=good -> high tip (verified against manual calculation)
- Extremes: inputs at minimum -> output near minimum; inputs at maximum -> output near maximum
- Interpolation: intermediate inputs -> intermediate output (monotonicity)

**Mamdani vs Sugeno equivalence:**
- For the same scenario, both methods produce results in the same reasonable range (not identical, but coherent)

**Stability:**
- Small changes in input produce small changes in output (continuity)

---

## sdks/decisions — Service contracts

These protect **correct orchestration**. It doesn't matter how the engine works internally —
the service fulfills its contract.

### Execution (Service.Execute)

**Basic contract:**
- A service configured with a ruleset and binder, given a valid request, returns a result with outcome and explanation — never panic, never empty result
- The requested paradigm (classical/bayesian/fuzzy) determines which engine is used — no accidental mixing

**Multi-paradigm scenarios with the SAME domain:**
- Loan approval (classical): applicant with good credit -> approved, with bad credit -> rejected
- Loan risk (bayesian): same applicant -> risk probability distribution
- Loan priority (fuzzy): same applicant -> numeric priority score
- All three return coherent results for the same applicant

### Explanation (Explainer)

**Contract:**
- Every classical decision explains which rules fired and in what order
- Every bayesian decision explains the posterior distribution
- Every fuzzy decision explains the outputs and memberships
- The Spanish explanation says the same thing as the English one (locale coherence)
- If a custom explainer is injected, it is used instead of the default

### Cascade (CascadePipeline)

**Contract:**
- A chain of N stages produces N partial results + 1 final result
- The result of stage 1 feeds stage 2 (the converter works)
- The combined explanation reflects all stages
- If a stage fails, the chain fails with a clear error indicating which stage failed

**Concrete scenarios:**
- Eligibility (classical) -> Risk assessment (bayesian): an eligible client can have high or low risk
- Eligibility (classical) -> Priority (fuzzy): an eligible client receives a priority score

### Validation (Validator)

**Contract — problem detection:**
- A ruleset with two contradictory rules -> validation reports the contradiction
- A ruleset with a redundant rule -> validation identifies it
- A ruleset with uncovered input combinations -> validation reports the gaps
- A ruleset with simplifiable conditions -> validation suggests the simplification
- A clean ruleset -> validation returns Valid=true, empty lists

**Contract — no false positives:**
- A ruleset without contradictions is never reported as contradictory
- A ruleset without redundancies is never reported as redundant

### Audit (AuditLog)

**Contract:**
- When audit is configured, each successful execution generates an entry with: timestamp, ruleset name/version, paradigm, request, result, explanation, duration
- When audit is NOT configured, executions work normally (no error, no side effect)
- If audit fails to record, the execution returns an error (not silenced)

### Repository

**CRUD contract:**
- Save -> Get returns what was saved (round-trip)
- Save two versions of the same ruleset -> both coexist
- List returns all saved rulesets
- Delete -> Get returns error (not found)
- Get of a non-existent ruleset returns a clear error

---

## apps/daas — User flows (aspirational)

These protect the **user experience**. They are end-to-end user stories.

### Complete lifecycle

> "An analyst creates a ruleset, validates it, deploys it, executes decisions against it,
> and queries the history of decisions made"

- Create ruleset via API -> 201 Created
- Validate ruleset -> report with no errors
- Execute decision with valid input -> result with explanation
- Query audit trail -> the decision appears with all its fields

### Versioning

> "An analyst updates a ruleset without breaking in-flight decisions"

- Create ruleset v1 -> execute decision -> result A
- Create ruleset v2 (different rules) -> execute decision -> result B (different)
- Execute decision explicitly requesting v1 -> still returns result A

### Validation as guardian

> "A ruleset with errors cannot be deployed"

- Create ruleset with contradictions -> validate -> report with errors
- Attempt to execute decision against unvalidated ruleset -> error or clear warning

### Errors and resilience

> "Invalid inputs produce clear errors, not crashes"

- Request without ruleset name -> 400 error with descriptive message
- Request with paradigm the ruleset doesn't support -> clear error
- Request with incomplete facts -> error indicating what's missing
- Two concurrent requests for different rulesets -> both correct, no interference

### Multi-paradigm from the API

> "The same endpoint supports all three paradigms"

- POST /decisions with paradigm=classical -> result with boolean facts
- POST /decisions with paradigm=bayesian -> result with distribution
- POST /decisions with paradigm=fuzzy -> result with numeric outputs
- All three for the same domain are coherent

---

## Cross-cutting observations

1. **Golden scenario**: Loan Approval. It's the domain that crosses all 4 modules and all 3
   paradigms. It should be THE main test case — the same loan evaluated by classical logic,
   bayesian probability, and fuzzy logic.

2. **Round-trip is a recurring pattern**: parse->serialize, save->get, create->read. Protects
   data integrity across the entire chain.

3. **"Don't invent" is as important as "don't omit"**: soundness (no false positives) matters
   as much as completeness (no false negatives).

4. **daas scenarios are compositions of sdk scenarios**, which in turn are compositions of
   inference scenarios, which are based on maths scenarios. The pyramid is natural.

5. **Performance scenarios** (e.g., "1000 concurrent executions don't degrade correctness")
   are not yet included — to be added in a future iteration.
