# Strategy

Product positioning, application ideas, and licensing decisions for the yarumo / DaaS / aluna stack. Engineering work lives in [`ROADMAP_NEW_MODULES.md`](ROADMAP_NEW_MODULES.md), [`ROADMAP_COMPUTE.md`](ROADMAP_COMPUTE.md), [`ROADMAP_DECISIONS.md`](ROADMAP_DECISIONS.md), and [`ROADMAP_ONTOLOGIES.md`](ROADMAP_ONTOLOGIES.md). `modules/common/` and `modules/telemetry/otel/` work is tracked directly via tickets in milestones [Phase 7](https://github.com/guidomantilla/yarumo/milestones/8) and [Phase 8](https://github.com/guidomantilla/yarumo/milestones/9). Concrete product blueprints live in [`products/`](products/).

---

## Product layers

```
Yarumo  = SDKs + modules (Go libraries). Not a product.
Aluna   = agentic AI platform (product, runtime). External: /Users/raven/Workspace/guidomau/aluna.
DaaS    = decision-as-a-service product (frontend + backend + DB). First SDK consumer.
Socratic= text-to-text tool in sdks/decisions/socratic/ (proprietary license, not Apache 2.0).
```

Relationships are intentionally loose for now:

- Aluna may consume DaaS as an API, or DaaS may run on Aluna.
- Aluna agents use the Decision Engine and Socratic as tools.
- KnowledgeForge (PDF→rules extractor) feeds the socratic.

---

## DaaS — vertical defines the product

A generic decision engine ("bring your rules and run them") competes with GoRules, Drools, IBM ODM — engines with years of ecosystem advantage. A vertical-less DaaS is a commodity.

BeClever proves it: nobody buys "a decision engine". They buy "credit approval in 500 ms" or "25% reduction in default". The engine is invisible; what is sold is the business outcome.

### What the vertical defines

| Aspect | Generic | With vertical |
|---|---|---|
| Binder | User builds it | Pre-built for the domain |
| Rulesets | Empty, start from scratch | Templates and examples |
| Integrations | Generic (REST, DB) | Specific (bureaus, DIAN, SIC, …) |
| UI | Generic rule editor | Domain analyst flows |
| Explanations | Technical (rule X fired in step N) | Regulatory ("rejected per policy 4.2.3") |
| Customer | "Anyone" (i.e. nobody) | Concrete persona / company with budget |
| Pricing | Hard to justify | Tied to business value |

### Candidate verticals

High affinity with the multi-paradigm engine:

| Vertical | Why it fits | Paradigms | Reference |
|---|---|---|---|
| **Credit / lending** | Large LATAM market, BeClever validates demand | Classical (rules) + Bayesian (risk) + Fuzzy (gradual scoring) | [`LENDING_BRAIN.md`](products/LENDING_BRAIN.md) / [`UNDERWRITE_IQ.md`](products/UNDERWRITE_IQ.md) |
| **Regulatory compliance** | Real pain in Colombian SMBs (DIAN, SIC) | Classical (rules) + Bayesian (audit risk) | [`COMPLIANCE_ENGINE.md`](products/COMPLIANCE_ENGINE.md) / [`GOV_RULES.md`](products/GOV_RULES.md) |
| **Collections** | Segmentation + prioritization = naturally multi-paradigm | Classical (stages) + Fuzzy (priority) + Bayesian (P[paid]) | BeClever-adjacent |
| **Fraud detection** | The 3 paradigms cooperate in cascade | Bayesian (suspicion) + Fuzzy (severity) + Classical (escalation) | FraudNet (catalog 1.8 below) |
| **HR screening** | Explainable fairness as differentiator | Fuzzy (scoring) + Classical (fairness rules) | HRScreen (catalog 4.5 below) |

### Open questions

- Which vertical first? Criteria: market size, customer access, integration complexity, time-to-value.
- Is the vertical built inside `apps/daas` or as a separate app consuming DaaS?
- Are the generic SDK and the vertical the same product or two products on the same engine?
- How do we validate demand before building? (interviews, landing page, MVP?)

### Competitive reference

| Competitor | Vertical | Engine | Market |
|---|---|---|---|
| BeClever | Lending + collections + cards | Closed, black box | LATAM finance, 40+ customers |
| FICO | Scoring + fraud + lending | FICO Blaze Advisor | Global enterprise |
| Alloy | Credit underwriting + KYC | Proprietary | US fintech |
| Algonew | Credit + fraud with AI agents | Proprietary | LATAM |

See `docs/BENCHMARK_RULE_ENGINES.md` for the full competitive benchmark.

---

## Application ideas — yarumo + DaaS + aluna synergy

Brainstorm combining the capabilities of `compute/math/` (logic, probability, fuzzy), `compute/engine/` (deductive, bayesian, fuzzy, causal, mcdm) and the aluna platform (LATAM SMB agentic SaaS).

> **Status: ideation only.** Anything here that should ship as a real product moves to `docs/products/<name>.md` and gets a roadmap entry.
>
> All ideas are realizable with the current state of `compute/*` (100% coverage, all paradigms operative). Future extensions (`predicate/`, `temporal/`) open additional dimensions but are not prerequisites.

### Category 1 — Apps independent of DaaS and aluna

Products built with yarumo in domains unrelated to DaaS or aluna.

| # | Name | Description | yarumo packages | Complexity |
|---|---|---|---|---|
| 1.1 | **SmartTune** | Go library for fuzzy PID controllers in IoT/embedded | `compute/engine/fuzzy` (Mamdani) | Low |
| 1.2 | **StructHealth** | Structural monitoring: Bayes for failure probability + classical rules for actions | `compute/engine/bayesian` + `compute/engine/deductive` | Medium |
| 1.3 | **LogicLab** | Interactive propositional-logic tutor (truth tables, SAT, step-by-step transforms) | `compute/math/logic` (all sub-packages) | Low |
| 1.4 | **InferencePlayground** | Multi-paradigm sandbox: define a problem in 3 paradigms and compare results with traces | the 3 engines + the 3 maths | Medium |
| 1.5 | **AlertGuard** | Infra alert correlation: Bayes for root cause + backward chaining for remediation | `compute/engine/bayesian` + `compute/engine/deductive` | Medium |
| 1.6 | **ConfigValidator** | CLI that validates YAML/JSON configs against declarative rules using SAT and forward chaining | `compute/math/logic`, `compute/math/logic/sat`, `compute/engine/deductive` | Low |
| 1.7 | **RiskGrade** | Credit risk scorer: fuzzy (gradual scoring) + Bayes (default probability) | `compute/engine/fuzzy` + `compute/engine/bayesian` | Medium |
| 1.8 | **FraudNet** | Transaction anomaly detection: Bayes + fuzzy + classical rules cascade | the 3 engines | High |

### Category 2 — Apps with DaaS and/or aluna synergy

| # | Name | Description | Overlap | Complexity |
|---|---|---|---|---|
| 2.1 | **DecisionSkill** | Aluna skill that wraps a DaaS ruleset — deterministic auditable decisions as a tool | DaaS engine inside aluna | Medium |
| 2.2 | **SmartValidator** | Aluna vertical: document validation (invoices, contracts) with LLM extraction + Bayes risk + compliance rules | DaaS rules + aluna agents | High |
| 2.3 | **EvalBot** | Aluna vertical: vendor evaluation with multi-criteria fuzzy + Bayesian reliability | aluna Procurement | Medium |
| 2.4 | **AuditTrail+** | Aluna runtime middleware: validate every agent action against a ruleset before execution | DaaS enforcement in aluna | Low-Medium |
| 2.5 | **WorkflowOptimizer** | Aluna vertical: process efficiency analysis with Sugeno fuzzy inference | aluna PMO | Low |
| 2.6 | **ClassifyRoute** | Smart router: Bayes + fuzzy to dispatch jobs to the optimal Skill automatically | improves aluna platform | Medium |

### Category 3 — How DaaS gets supercharged by yarumo and aluna

Concrete ways `compute/math/`, `compute/engine/` and aluna patterns make DaaS more powerful than a basic rule system.

| # | Name | Source | Description |
|---|---|---|---|
| 3.1 | **SAT Validation** | `compute/math/logic/sat` | Validate ruleset consistency with DPLL before deploy — detect contradictions |
| 3.2 | **Coverage Analysis** | `compute/math/logic/sat` | Enumerate input combinations and find gaps (scenarios with no rule) |
| 3.3 | **Rule Simplification** | `compute/math/logic` | Apply 18 simplification rules to complex conditions — readability + comparison |
| 3.4 | **Dual-Mode Execution** | `compute/engine/deductive` | Expose forward (data-driven) + backward (goal-driven) chaining as 2 query modes |
| 3.5 | **Bayesian Overlay** | `compute/engine/bayesian` | Attach probabilistic confidence to empirical rules — `P(rule_applies | data_quality)` |
| 3.6 | **Fuzzy Thresholds** | `compute/engine/fuzzy` | Replace hard thresholds (`income > 5000`) with fuzzy conditions (`income IS high`) |
| 3.7 | **AI Rule Authoring** | aluna pattern | Use THINK-ACT-OBSERVE for NL→rules with SAT validation in the loop |
| 3.8 | **Trace-to-NL** | aluna pattern | Convert technical traces to natural-language explanations via Claude API — for auditors and regulators |

### Category 4 — Extending and amplifying aluna

| # | Name | Type | Description | yarumo packages | Complexity |
|---|---|---|---|---|---|
| 4.1 | **Reconciliation Engine** | Improves CierreExpress | Classical rules for deterministic transaction matching, LLM only for ambiguous cases — reduces API calls 60-80% | `compute/engine/deductive` | Medium |
| 4.2 | **Variance Analysis** | Improves CierreExpress | Fuzzy significance scoring for budget variances — prioritize what really matters | `compute/engine/fuzzy` (Sugeno) | Low |
| 4.3 | **Win-Rate Estimation** | Improves PropuestaYa | Bayesian `P(win | client, pricing, competition)` — calibrate effort per opportunity | `compute/engine/bayesian` | Medium |
| 4.4 | **ComplianceGuard** | New vertical | Colombian regulatory compliance (DIAN, SIC) with classical rules + Bayes audit risk | `compute/engine/deductive` + `compute/engine/bayesian` | Medium |
| 4.5 | **HRScreen** | New vertical | CV screening with fuzzy scoring + classical fairness rules — explainable ranking | `compute/engine/fuzzy` + `compute/engine/deductive` | Low-Medium |
| 4.6 | **ProcureScore** | New vertical | Quote comparison: LLM extracts data + multi-criteria fuzzy evaluates | `compute/engine/fuzzy` | Low |
| 4.7 | **CostGuard** | Improves platform | Fuzzy controller for token budget — gradual degradation instead of hard cutoff | `compute/engine/fuzzy` (Mamdani) | Low |
| 4.8 | **ReliabilityNet** | Improves platform | Bayesian `P(success | conditions)` to predict Skill reliability and auto-scale | `compute/engine/bayesian` | Medium |

### AI's role in DaaS

AI is convenience, not architecture. DaaS works without it — the user just writes rules by hand and reads raw traces. AI is the human interface, not the engine. The engine's value is that it is **not probabilistic**.

| Task | What it does | Local model OK? |
|---|---|---|
| NL → facts (known fields) | Extract structured data from free text | Yes — 7-8B is enough |
| Trace → NL explanation | Reformat technical trace to human language | Yes — 7-8B is enough |
| NL → rules (authoring) | Generate `A & B => C` from natural language | Depends on complexity |
| Ruleset validation | Detect contradictions, uncovered cases | No AI needed — `compute/math/logic` does it (SAT, entailment) |

What AI does **NOT** do: it does not execute rules, decide, or replace auditability. "Why was this credit rejected?" must be a traceable rule chain, not "the model said so."

**Recommended strategy**:

1. Start with Sonnet via API for everything.
2. Define 20 input/output examples per task.
3. Measure error rate.
4. Downgrade to Haiku or local for what works.
5. Do not optimize before measuring.

---

## Licensing — channels and open questions

Pending detailed analysis. Initial summary:

### Channels

| Channel | Model | Target |
|---|---|---|
| **Community** | Free, self-hosted, limited | Devs, adoption, community |
| **Cloud** | SaaS, paid, managed | Enterprises that don't want to operate |
| **Aluna embedded** | Bundled with aluna, not standalone | Agentic platform users |

### Current license per layer

| Layer | License | Rationale |
|---|---|---|
| `modules/compute/math` | Apache 2.0 | Foundation; attracts academics and developers |
| `modules/compute/engine` | Apache 2.0 | Visible technical differentiation |
| `sdks/decisions/core` | Apache 2.0 | Adoption driver, builds ecosystem |
| `apps/daas` | TBD | Product — where monetization happens |

### Open questions

- What features differentiate Community (free) from Cloud (paid)?
- Community limits: number of rulesets? executions/day? users? paradigms?
- Is the aluna version a separate channel or a feature tier of Cloud?
- Are companion modules (storage/postgres, llm/, endpoints/) open or part of paid?
- Dual licensing (open core + commercial) or single license with feature gates?

### Market reference

| Engine | Model |
|---|---|
| Drools | Full open (Apache 2.0), revenue via Red Hat support |
| GoRules | ZEN engine open (MIT), BRMS commercial |
| OpenRules | Engine open, Decision Manager commercial |
| IBM ODM | Full commercial |
| NRules, MS RulesEngine, json-rules-engine | Full open, no direct revenue |

See `docs/BENCHMARK_RULE_ENGINES.md` for the full benchmark.
