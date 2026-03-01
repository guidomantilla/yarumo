# Yarumo

[![Build](https://github.com/guidomantilla/yarumo/actions/workflows/build.yml/badge.svg)](https://github.com/guidomantilla/yarumo/actions/workflows/build.yml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](https://go.dev/)

A modular Go toolkit for building backend applications. Yarumo provides foundational packages for common operations, configuration bootstrapping, lifecycle management, observability, mathematics, and multi-paradigm inference engines — designed to be composed, not imposed.

## Modules

Yarumo is organized as a Go workspace with independent modules that can be imported separately.

| Module | Import Path | Description |
|--------|-------------|-------------|
| **common** | `github.com/guidomantilla/yarumo/common` | Core library — 16 packages covering assertions, crypto, errors, HTTP, gRPC, logging, and more |
| **config** | `github.com/guidomantilla/yarumo/config` | One-shot bootstrap for application configuration via environment variables |
| **managed** | `github.com/guidomantilla/yarumo/managed` | Lifecycle management for components (Start/Stop/Done) |
| **telemetry/otel** | `github.com/guidomantilla/yarumo/telemetry/otel` | OpenTelemetry setup for tracing, metrics, and logging |
| **maths** | `github.com/guidomantilla/yarumo/maths` | Mathematical primitives — propositional logic, discrete probability, fuzzy logic |
| **inference** | `github.com/guidomantilla/yarumo/inference` | Multi-paradigm inference engines — classical, bayesian, fuzzy |

### Dependency Graph

```
inference → maths → common
telemetry/otel    managed    config
      │              │         │
      └──────────────┴─────────┘
                     │
                   common
```

### Common Packages

| Package | Purpose |
|---------|---------|
| `assert` | Runtime assertions |
| `cast` | Type casting utilities |
| `constraints` | Generic type constraints |
| `cron` | Cron scheduling |
| `crypto` | Hashes, signers (HMAC, ECDSA, Ed25519, RSA-PSS), ciphers (AEAD, RSA-OAEP), certs, passwords, tokens |
| `diagnostics` | Flight recorder and tracing |
| `errs` | Typed error handling |
| `grpc` | gRPC utilities |
| `http` | HTTP client and helpers |
| `log` | Structured logging (slog) |
| `pointer` | Safe pointer operations |
| `random` | Random value generation |
| `rest` | REST API utilities |
| `types` | Common type definitions |
| `uids` | ID generation (UUID, ULID, XID, CUID2, NanoID) |
| `utils` | General-purpose utilities |

### Maths Packages

| Package | Purpose |
|---------|---------|
| `logic` | Propositional logic — formulas, evaluation, transforms (NNF/CNF/DNF), simplification (18 rules), analysis |
| `logic/parser` | Recursive descent parser with Unicode operators and keyword support |
| `logic/sat` | SAT solver (DPLL algorithm) |
| `logic/entailment` | Logical entailment and equivalence checking |
| `probability` | Discrete probability — CPT, Factor, Bayes theorem, distributions, entropy |
| `fuzzy` | Fuzzy logic — membership functions, t-norms/t-conorms, defuzzification methods |

### Inference Engines

| Package | Purpose |
|---------|---------|
| `classical/engine` | Forward and backward chaining with provenance tracking |
| `classical/rules` | Rule definitions with priority and conditions |
| `classical/facts` | Fact base with assert/retract/clone operations |
| `classical/explain` | Execution traces with step-by-step provenance |
| `bayesian/engine` | Exact inference — enumeration and variable elimination algorithms |
| `bayesian/network` | Bayesian network definition with DAG validation |
| `bayesian/evidence` | Evidence base for observed variables |
| `bayesian/explain` | Inference traces with posterior distributions |
| `fuzzy/engine` | Mamdani and Sugeno inference methods |
| `fuzzy/variable` | Linguistic variables with fuzzification |
| `fuzzy/rules` | Fuzzy rules with conditions, weights, and AND/OR operators |
| `fuzzy/explain` | Inference traces with membership degrees and activations |

## Getting Started

### Requirements

- Go 1.25+
- [graphviz](https://graphviz.org/) (for dependency graph generation)

### Install

```bash
go get github.com/guidomantilla/yarumo/common@latest
go get github.com/guidomantilla/yarumo/maths@latest
go get github.com/guidomantilla/yarumo/inference@latest
```

### Tooling

```bash
make install    # Install dev tools (linter, coverage, etc.)
make validate   # Run full validation: tidy, generate, format, vet, lint, coverage
make build      # Full build: validate + vulnerability check
```

## License

[Apache License 2.0](LICENSE)
