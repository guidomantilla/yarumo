# Spring HATEOAS — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-hateoas/
> **Analyzed**: 2026-05-16
> **Re-analyzed**: 2026-05-16 (editorial cleanup after roadmap trim)
> **Recommendation**: REJECT

## 1. Project summary

Spring HATEOAS (v3.0.3, 2026-03-13) is a Spring Framework library for building hypermedia-driven REST APIs. It provides `RepresentationModel`, `EntityModel<T>`, `CollectionModel<T>`, `PagedModel<T>`, server-side link builders tied to Spring MVC/WebFlux controllers, affordances, and serializers for HAL, HAL-FORMS, Collection+JSON, UBER, ALPS, and RFC-7807 Problem Details. Java/JVM coupling: **high** — `WebMvcLinkBuilder`, `methodOn(...)` proxies, JSR-303-driven HAL-FORMS templates, `@EnableHypermediaSupport`, and Jackson modules are deeply intertwined with the Spring container, Jackson, and bytecode proxies. Client side ships `Traverson` + `LinkDiscoverer`.

## 2. Pareto features (top-20%)

| # | Feature | Description | Why it matters for Go microservices |
|---|---|---|---|
| 1 | `Link` value type + IANA relations | Immutable `{href, rel, title, templated}` with `self`, `next`, `prev`, `first`, `last`, `item`, `collection` constants (RFC 8288) | Standard vocabulary if/when an API decides to expose pagination/navigation links. Trivially modeled as a Go struct. |
| 2 | RFC 7807 Problem Details | Standard error envelope (`type`, `title`, `status`, `detail`, `instance` + extensions) | The only piece with universal value: error contract standardization. Go has multiple libraries already. |
| 3 | HAL envelope (`_links`, `_embedded`) | `application/hal+json` shape grouping links/embedded resources alongside payload | If a yarumo consumer ever commits to HAL, the envelope is ~30 lines of Go. |
| 4 | URI Template (RFC 6570) expansion | `/people/{id}{?expand}` parameterization | Useful for client-side templated link expansion. `github.com/jtacoma/uritemplates` exists. |
| 5 | `CollectionModel` / `PagedModel` shape | Standard pagination envelope with `page.size`, `page.number`, `page.totalElements`, `page.totalPages` and `_links.next/prev/first/last` | Pagination contract is genuinely reusable across REST handlers. |

**Honest assessment**: of these 5, only #2 (Problem Details) and #5 (pagination envelope) deliver real value. #1, #3, #4 are 30-line implementations that don't justify a module.

## 3. Long-tail features (skip)

- **`WebMvcLinkBuilder` / `methodOn(...)`** — depends on CGLIB bytecode proxies and Spring's `HandlerMapping`. Not portable to Go. gin/chi/echo have no equivalent reflection model.
- **`RepresentationModel` base class hierarchy** — Java single-inheritance idiom. Go composition + struct embedding makes the hierarchy unnecessary.
- **`RepresentationModelAssembler` / `RepresentationModelProcessor`** — DI-container concepts. Yarumo's no-DI principle forbids them.
- **`@EnableHypermediaSupport` annotation** — annotation-driven activation. Yarumo's no-magic principle forbids it.
- **HAL-FORMS `_templates` from JSR-303** — auto-generates form metadata from `@NotNull`, `@Pattern`, `@Email`. Couples to Java validation API; Go validator/v10 tags don't map cleanly and the use case (HAL Explorer auto-rendering forms) is niche.
- **Affordances** — controller-method-to-form-metadata via bytecode introspection. Not implementable in Go without heavy code generation.
- **Collection+JSON, UBER, ALPS** — adoption ⭐⭐ or worse; nobody is asking for these.
- **`EntityLinks` / `TypedEntityLinks`** — entity-to-link lookup via DI. No use case in yarumo.
- **`CurieProvider`** — compact link relations. Real spec, near-zero real-world adoption.
- **`Traverson` (client)** — follows links by `rel`. Cute in a demo, brittle in production; real Go clients hard-code URIs or generate clients from OpenAPI.
- **`LinkDiscoverer`** — JSONPath into `_links.<rel>`. Trivial when needed; not worth a module.

## 4. Mapping to Yarumo

**Existing/planned modules with overlap**:
- `modules/common/rest/` — REST client with `RequestSpec`/`ResponseSpec`. Client-side only, no link traversal. No overlap with HATEOAS server side.
- `modules/common/http/` — `errors.go` already has HTTP error types — could naturally extend to RFC 7807.
- No REST server framework module exists, by design (per the canonical roadmap's placement principle: REST handlers live in consumer apps using gin/chi/echo directly).
- `tools/routegen/` (§ 2.1) — generates Gin routes; pure routing, not hypermedia.

**Gaps this could fill**:
- **RFC 7807 Problem Details** — DaaS and Aluna both need a standard error envelope. Currently each consumer would invent one. A small `modules/common/http/problem/` (or extension of `common/http/errors.go`) covers it.
- **Pagination envelope** — `{ items, page: { number, size, total_elements, total_pages }, links: { next, prev, first, last } }`. Reusable types worth ~50 lines.

**Anti-patterns to avoid** (from ROADMAP_NEW_MODULES.md):
- Auto-wiring (`@EnableHypermediaSupport`).
- Bytecode-proxy magic (`methodOn(controller).method()` for link building).
- Annotation-driven affordances.
- Hierarchical base classes (`extends RepresentationModel<Self>` self-bounded generics).
- Framework coupling (cannot pick gin/chi/echo for the consumer).

**Mature Go libraries to wrap**: none worth wrapping for a full HATEOAS module.
- `github.com/danielgtaylor/huma/v2` ships HAL/JSON Schema but is a whole REST framework, not a hypermedia library.
- `github.com/nvellon/hal` and `github.com/pmoule/go-jsonapi-server` exist but are unmaintained or niche.
- `github.com/moogar0880/problems` covers RFC 7807; small and decent.

## 5. Recommendation

**REJECT** the full Spring HATEOAS surface. HATEOAS adoption in the Go microservices ecosystem is near zero: production REST APIs ship URL conventions + OpenAPI + JSON Schema and call it done. The two parts that do deliver value — RFC 7807 Problem Details and the pagination envelope — are 100-line additions that belong in `modules/common/http/` (Problem Details) and as a small pagination helper consumed by handlers in DaaS/Aluna; neither warrants a `modules/hateoas/`. Spring HATEOAS's flagship machinery (`WebMvcLinkBuilder`, `methodOn`, affordances, HAL-FORMS auto-generation) is bytecode-proxy + DI-container glue that violates yarumo's explicit-wiring principles and has no idiomatic Go equivalent.

## 6. Proposed yarumo placement (if ADOPT/PARTIAL)

Not adopting; section retained for traceability of the two extracted pieces.

**Possible follow-ups** (file as low-priority issues if/when DaaS or Aluna asks):
- **`modules/common/http/problem/`** — RFC 7807 types: `Problem{Type, Title, Status, Detail, Instance, Extensions map[string]any}`, JSON marshaling, content-type constant `application/problem+json`. ~80 LOC + tests. Could also live inline in `common/http/errors.go` as a typed error.
- **`modules/common/http/pagination/`** (or co-located with handlers in DaaS) — `Page[T]` struct + `Links{Next, Prev, First, Last *Link}` + helper that computes total-pages and emits link headers (RFC 5988 / `Link:` header is more idiomatic than HAL `_links` for plain JSON APIs).

**Out of scope (v1 and forever)**: HAL serializers, HAL-FORMS, affordances, link-from-controller-method builders, Traverson, CURIEs, Collection+JSON, UBER, ALPS.

**For traceability**, if a "Considered and rejected" section is ever added back to `docs/ROADMAP_NEW_MODULES.md`, the appropriate entry would be:

| Idea | Why not |
|---|---|
| `modules/hateoas/` | Spring HATEOAS port. ~95% is bytecode-proxy + DI machinery that doesn't apply to Go. Only RFC 7807 and a pagination envelope are reusable; both fit inside `common/http/`. HATEOAS itself has near-zero adoption in Go production REST APIs. |

## 7. Open questions

- Does DaaS or Aluna want HAL-shaped responses, or are plain JSON + `Link:` headers (RFC 5988) sufficient? Default assumption: plain JSON.
- Should RFC 7807 land as a sub-package (`common/http/problem/`) or as types in `common/http/errors.go`? The errors file already exists; merging is simpler.
- Pagination: response-body envelope vs. `Link:` header vs. both? GitHub/Stripe use `Link:` header; most internal APIs use body envelope. Likely decide per-consumer.
- Is there ever a real use case for HAL-FORMS in DaaS (decision-engine UI auto-rendering rule-edit forms from server hints)? If yes, revisit — but a custom small format is probably better than HAL-FORMS.

## 8. ROADMAP delta proposed (NOT applied)

Nothing for ROADMAP_NEW_MODULES.md directly. The two extracted reusable bits (RFC 7807 Problem Details, pagination envelope) belong inside a future `modules/common/http/` (already tracked under YA-0042) — they do not warrant a new top-level entry. If/when those land, they show up as sub-packages of `common/http/`, not as new roadmap items.
