# Spring REST Docs — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-restdocs
> **Analyzed**: 2026-05-16 (re-analysis after roadmap cleanup)
> **Recommendation**: PARTIAL

## 1. Project summary

Spring REST Docs (current 3.0.5, Java 17+ / Spring 6+) generates RESTful-API reference documentation **from passing integration tests**. Tests written against MockMvc / WebTestClient / REST Assured emit Asciidoctor (or Markdown) snippets via `.andDo(document("name", ...))`; an Asciidoctor build assembles them into HTML. The model is **test-driven docs**, the opposite of springdoc-openapi / swaggo / kin-openapi (which are annotation-driven or schema-first).

JVM coupling is high at the integration layer — MockMvc is a Spring testing primitive, snippet rendering uses Mustache, the HTML build relies on Asciidoctor JVM plugins. But the conceptual model (capture HTTP exchange in a test → emit snippet → include in narrative `.adoc`) is portable to Go because each of those moving parts has a stdlib or well-supported analogue (`net/http/httptest`, `text/template`, plain `.adoc` text).

The value proposition: snippets only exist when the test passes; if the wire format drifts, the test fails, so docs cannot lie about the contract.

## 2. Pareto features (top-20%)

| # | Feature | Description | Why it matters for Go microservices |
|---|---|---|---|
| 1 | Test-driven docs | Snippets emitted only on passing test; outdated docs cause test failures | swaggo (comment-driven, decoupled from runtime) and hand-edited OpenAPI YAML both rot silently. Test-driven snippets cannot. |
| 2 | Six default request/response snippets | `curl-request.adoc`, `httpie-request.adoc`, `http-request.adoc`, `http-response.adoc`, `request-body.adoc`, `response-body.adoc` | Six client-ready artifacts per endpoint with zero descriptors. The curl/httpie commands are exactly what readers want to paste. |
| 3 | Field descriptors with JSON paths | `fieldWithPath("contact.email").description(...)`; dot/bracket notation; `[]` arrays; type inference (string/number/array/object/null/varies) | Documents the JSON contract from a single source of truth (the test). Strict-by-default mode fails the test on undocumented fields. |
| 4 | Path / query / form parameter descriptors | `pathParameters(...)`, `queryParameters(...)`, `formParameters(...)`; `relaxed*` variants when partial coverage is desired | Each parameter named once. Catches drift when a handler adds a query param and docs forget it. |
| 5 | Header / cookie descriptors | `requestHeaders`, `responseHeaders`, `requestCookies`, `responseCookies` | Surfaces auth, rate-limit, CORS, and session headers with a single line each. |
| 6 | Asciidoctor `include::{snippets}/.../*.adoc[]` model | Narrative `.adoc` files include snippets; snippet folder under `target/generated-snippets` / `build/generated-snippets` | Separates stable prose from regenerated examples. Final artifact is plain HTML — embeddable in any static site. |
| 7 | Parameterized snippet directories | `document("{class-name}/{method-name}/{step}")` | Snippet layout mirrors test tree → zero naming collisions, predictable include paths. |
| 8 | Operation preprocessors | `prettyPrint()`, `maskLinks()`, `removeHeaders("Authorization", "Cookie")`, `replacePattern(regex, repl)`, `modifyUris()` | Sanitises sensitive data (tokens, PII) and normalises hosts/ports across environments before the snippet hits disk. Indispensable for public-facing docs. |
| 9 | Constraint extraction from Bean Validation | `ConstraintDescriptions(UserInput.class).descriptionsForProperty("name")` reads `@NotNull` / `@Size` and feeds the "Constraints" column | Go equivalent: read `validate:"required,min=8"` struct tags or `common/validation` rules. |
| 10 | Reusable descriptor bundles | Shared `FieldDescriptor[]` / `LinksSnippet` composed via `.and(...)` and `.andWithPrefix(...)` | Eliminates copy-paste across 30 endpoints that share a `{page, per_page, total}` envelope. |
| 11 | Mustache template overrides | Drop replacement templates into `src/test/resources/.../templates/asciidoctor/` to restyle any snippet | Customization (column order, branding) without forking. |
| 12 | Markdown alternative output | Same descriptors, Markdown templates | Useful when docs ship to GitHub READMEs or platforms that prefer Markdown over HTML. |

## 3. Long-tail features (skip)

- **Hypermedia link documentation** (`links()`, HAL/Atom format handlers, `halLinks()`, `atomLinks()`) — yarumo / DaaS / Aluna are not HATEOAS-oriented. Dead weight.
- **REST Assured integration** — third Java HTTP-client wrapper; pick one Go equivalent.
- **WebTestClient integration** — Spring WebFlux reactive specific; Go has no analogue.
- **JSON path wildcards** (`users.*.role`) and bracket-with-dot-in-key (`['e.dot']`) — corner cases; defer until a real payload needs them.
- **Subsection documentation** (`beneathPath` / `withSubsectionId`) — rare; flat descriptor tables suffice in v1.
- **Multipart request parts** (`partWithName`, `requestPartFields`, `requestPartBody`) — niche for our services; revisit when an upload endpoint shows up.
- **TestNG / manual setup variants** — `go test` is the only target.
- **Maven / Gradle plugin specifics** — yarumo is pure Go.
- **Mustache template engine** — port to `text/template` (stdlib) instead.

## 4. Mapping to Yarumo

**Existing § 1 modules with overlap**: **none**. The trimmed roadmap (§ 1) covers datasource / auth / messaging / health / boot. None of these are testing scaffolding.

**Cross-reference**: `spring-framework.md` (re-analyzed in parallel) proposes a **NEW** top-level `modules/testing/` module — fixtures, fakes, containers, contracts, LLM-test helpers. **REST Docs lands inside that proposed new module as `modules/testing/apidocs/`**. The two proposals are coupled: REST Docs apidocs presupposes `modules/testing/` exists.

**Adjacent modules that REST Docs would consume**:

- `modules/common/validation/` (shipped in Phase 2) — supplies constraint metadata that a field-descriptor pass can read to auto-populate the "Constraints" column, analogous to Spring's `ConstraintDescriptions` over Bean Validation.
- `modules/common/errs/` — typed errors (`ErrUndocumentedField`, `ErrMissingDescriptor`, `ErrSnippetEmit`).
- `modules/common/log/` — debug output on snippet emission.
- `managed/server_http` (Phase 3, milestone #9) — supplies the HTTP server under test. REST Docs scaffolding wraps an `http.Handler` or `*httptest.Server`.
- `modules/health/` (§ 1.4, planned) — `/healthz` and `/readyz` endpoints are first-class candidates for the generated example catalogue.

**Gaps this could fill**:

- DaaS will eventually need public API reference docs. Today: no plan. swaggo (annotation-driven) was the implicit answer, but annotations drift the moment a handler is refactored.
- Aluna's agent endpoints — same future need.
- A future `modules/testing/contracts/` (Pact-style) and `modules/testing/apidocs/` are sibling patterns: contracts verify external behaviour, REST-Docs-style snippets *document* it from the same integration test. The same `RequestRecorder` primitive can power both.

**Anti-patterns to avoid** (yarumo principles):

- **No DI container** — Spring's `RestDocumentationContextProvider` plugs into `WebApplicationContext`. The Go port must accept any `http.Handler` (or `*httptest.Server`) explicitly wired by the caller.
- **No annotations / no AST magic** — translate Spring's fluent `document(...)` API to plain Go function calls. No comment scanning, no `go generate` step required for the core flow.
- **No global state** — Spring's `RestDocumentationContextProvider` is per-test-method; the Go equivalent must be a per-test `t.Helper()`-friendly handle, not a package-level singleton.
- **Don't wrap Asciidoctor in Go** — emit `.adoc` text only; let the consumer run `asciidoctor` (CLI), `asciidoctor.js`, or a Go renderer of their choice.
- **No reflection-driven field discovery** — descriptors stay explicit, matching yarumo's "no magic" principle and existing options-pattern conventions.
- **Stick to `text/template`** — do not port Mustache. `text/template` covers the snippet override use case with zero external dependencies.

## 5. Recommendation

**PARTIAL** — adopt the *concept* (test-driven snippets, Asciidoctor `include::` model, field/parameter/header descriptors, operation preprocessors for sanitisation) as a focused sub-package of the proposed-new `modules/testing/`. Reject the *implementation footprint* (no Mustache port, no Bean Validation bridge proper — read `common/validation` instead, no HAL/Atom, no WebFlux equivalent, no Maven/Gradle wrapper).

The test-driven model is genuinely valuable for Go: `swaggo` requires comment annotations that decouple from runtime behaviour and silently rot; OpenAPI YAML hand-edited by humans rots faster. A snippet that comes out of a passing integration test cannot lie about the wire format because the wire format **is** what the test recorded.

This is **complementary** to swaggo / kin-openapi, not a replacement: those tools can keep producing an OpenAPI schema for tooling (Postman, codegen clients), while REST-Docs-style snippets ship the human-readable reference (curl / httpie examples, field-by-field tables, narrative `.adoc`). The two outputs converge in the same `/static/docs/` directory the application serves.

**Conditional on**: `modules/testing/` being promoted from "proposed new" to "planned" by the parallel `spring-framework.md` analysis. If the testing module never lands, this proposal stays on ice — apidocs should not be its own top-level module.

## 6. Proposed yarumo placement

**Subpackage**: `modules/testing/apidocs/` (inside proposed NEW `modules/testing/` per `spring-framework.md`).

**Internal layout**:

- `apidocs/` — top-level: `Recorder` type that wraps an `http.RoundTripper` / `http.Handler`, captures the exchange, fans out to snippet generators.
- `apidocs/snippets/` — built-in snippet generators (`curl`, `httpie`, `http-request`, `http-response`, `request-body`, `response-body`, `request-fields`, `response-fields`, `query-parameters`, `path-parameters`, `form-parameters`, `request-headers`, `response-headers`, `request-cookies`, `response-cookies`).
- `apidocs/descriptors/` — `FieldDescriptor`, `ParamDescriptor`, `HeaderDescriptor`, `CookieDescriptor` types plus JSON-path matcher (`a.b[]`, dot + bracket notation; skip wildcards in v1).
- `apidocs/preprocessors/` — `PrettyPrint`, `MaskLinks`, `RemoveHeaders("Authorization", "Cookie")`, `ReplacePattern(regex, repl)`, `ModifyURIs(scheme, host, port)`. Same composable pipeline as Spring's `OperationPreprocessor`.
- `apidocs/templates/` — Go `text/template` snippet templates with override hooks (no Mustache port).
- `apidocs/constraints/` — bridge that reads `common/validation` rules off a struct and feeds the "Constraints" column.

**Internal deps**:

- `modules/common/validation/` — constraint metadata.
- `modules/common/errs/` — typed errors.
- `modules/common/log/` — debug output.
- **No dependency on** `managed/` — apidocs is test-scope only and must not impose lifecycle on the consumer.

**Go libraries to wrap**:

- `net/http/httptest` (stdlib) — capture the exchange.
- `tidwall/gjson` (preferred) or stdlib `encoding/json` → `any` — JSON-path field traversal. gjson's native dot/bracket syntax saves writing a parser.
- `text/template` (stdlib) — snippet templates.
- *Optional* `bytesparadise/libasciidoc` for end-to-end Asciidoctor → HTML in Go; **out of scope for v1** — emit `.adoc` files and document the user-side `asciidoctor` CLI invocation.

**Out of scope for v1**:

- HAL / Atom link documentation (`linkWithRel`, `halLinks`, `atomLinks`).
- Hypermedia format auto-detection.
- WebFlux / reactive equivalents.
- Mustache-template parity (use `text/template`).
- Multipart request-parts documentation.
- HTML / final-PDF build orchestration — emit `.adoc` only.
- Auto-discovery of routes — descriptors stay manual.
- Subsection documentation (`beneathPath`).
- OpenAPI emission — defer until a consumer demands tooling integration (then lands as `apidocs/openapi/`).

## 7. Open questions

- Is there enough integration-test surface in DaaS / Aluna v1 to justify shipping `modules/testing/apidocs/` before either app has a real endpoint catalogue? Hold until `modules/testing/` itself is "Planned" (not "Brainstorm").
- Should the snippet output also emit an OpenAPI spec fragment (alongside `.adoc`) so consumers can keep swaggo / kin-openapi for tooling? **Defer** — v1 emits `.adoc` only; OpenAPI emission lands as `apidocs/openapi/` once a consumer demands it.
- Where does the `asciidoctor` invocation live? Pure documentation in the module README, or a `tools/apidocs/` CLI under § 2? Lean toward documentation-only — yarumo has avoided wrapping JVM build tools.
- How does this coexist with a future `modules/testing/contracts/` (Pact)? Both observe the same integration test; could share a `RequestRecorder` primitive. Worth scoping together when `modules/testing/` is promoted.
- Should constraint extraction read `validate:` struct tags directly (zero coupling) or go through `modules/common/validation/` (one dependency, richer semantics)? Lean toward the validation module — it's already the project's source of truth for constraint metadata.
- Should the recorder integrate with `managed/server_http` (start a real server in-test) or stay purely at the `http.Handler` level via `httptest.NewServer`? Lean toward `httptest` only — keeps apidocs test-scope and free of `managed/` lifecycle dependency.
- Markdown vs Asciidoctor as the v1 output format? Asciidoctor is Spring's default and has richer include semantics, but Markdown lands directly in GitHub READMEs. Lean toward Asciidoctor for v1 with a Markdown template pack as v1.1.

## 8. ROADMAP delta proposed (NOT applied)

Coordinated with `spring-framework.md`, which proposes adding **§ 1.6 `modules/testing/`** as a NEW top-level module.

**Add inside the `modules/testing/` section of § 1.6** (proposed by `spring-framework.md`):

> ### 1.6.x `modules/testing/apidocs/` — test-driven API reference docs
>
> **Status**: Brainstorm
> **Why a sub-package, not its own module**: scaffolding that runs inside an integration test, emitting Asciidoctor snippets. Test-scope only; no lifecycle.
> **Inspired by**: Spring REST Docs.
>
> Capture `http.Handler` exchanges via `httptest`, emit curl/httpie/http-request/http-response/request-body/response-body `.adoc` snippets, plus descriptor tables for fields, parameters, headers, and cookies. Preprocessors sanitise auth tokens and PII before snippets hit disk. Consumes `modules/common/validation/` to auto-populate the "Constraints" column.
>
> Lands once `modules/testing/` is promoted from Brainstorm to Planned. Holds until DaaS or Aluna have a real endpoint catalogue to document.

**No changes** to existing § 1.1–1.5 (datasource / auth / messaging / health / boot), § 2 (routegen), or § 4 (go-feather-lib migration tracking). REST Docs has no go-feather-lib equivalent — it is net-new scope.

**Discarded alternatives**:

- A top-level `modules/apidocs/` — rejected; would duplicate the swaggo / kin-openapi space for the schema half and would not warrant its own module for the snippet half. Stays as a sub-package.
- A `tools/apidocs/` CLI — rejected for v1; the snippet emission happens inside `go test`, not as a separate build step.
