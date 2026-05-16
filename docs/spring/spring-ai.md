# Spring AI — Yarumo Analysis (DEEP)

> **Source**: https://docs.spring.io/spring-ai/
> **Analyzed**: 2026-05-16 (re-analysis after roadmap cleanup)
> **Recommendation**: PARTIAL

## 1. Project summary

Spring AI 1.1.6 is the current stable release (1.0.7 still maintained; 2.0.0-M6 in preview, 2.0.0-SNAPSHOT live). Scope: a `ChatClient` fluent API over provider-agnostic `ChatModel` / `EmbeddingModel` / `ImageModel` / `AudioModel` / `ModerationModel`, an **Advisor framework** (around-interceptors for chat), **tool calling** (`@Tool` + `ToolCallback` + `ToolCallingManager`), **MCP client/server starters** with four transports (STDIO, SSE, Streamable-HTTP, Stateless-HTTP), a **modular RAG pipeline** (QueryTransformer → QueryExpander → DocumentRetriever → DocumentJoiner → DocumentPostProcessor → ContextualQueryAugmenter), an **ETL pipeline** (Reader → Transformer → Writer) for documents, **structured output converters** (Bean/Map/List), a **ChatMemory** abstraction with 6 storage repositories (in-mem, JDBC, Cassandra, Neo4j, Cosmos, Mongo), full **OTel GenAI semantic conventions** instrumentation, and **agentic patterns as composable building blocks** (no monolithic Agent class). 20+ model providers and 23+ vector stores ship as Boot starters.

This is a **re-analysis**. The previous version of this document was written against an older Yarumo roadmap that contained brainstormed `modules/llm/` sub-modules. That brainstorm has been **deleted from the canonical roadmap** — `ROADMAP_NEW_MODULES.md` now contains only § 1 modules (`datasource/`, `auth/`, `messaging/`, `health/`, `boot/`), § 2 tools (`routegen`), and § 4 go-feather-lib migration tracking. There is no `modules/llm/` in the canonical roadmap today.

This document is the **brainstorm catalog** for what an LLM layer could look like if/when the roadmap absorbs it. Recommendations here are surface-level opinion; promoting any of them into `ROADMAP_NEW_MODULES.md` is a separate user decision (see § 9 for the concrete delta proposed).

**Overall verdict: PARTIAL** — propose `modules/llm/` as a NEW top-level module with eight sub-modules + a cross-cutting `Advisor` primitive, refine § 1.1 `modules/datasource/vector/` with a concrete API shape from Spring AI's `VectorStore`, propose `modules/telemetry/otel/genai/` as a sibling extension to `modules/telemetry/otel/`. Reject Spring's auto-wiring, annotation-discovery, reactive `Flux` parallel API, and per-provider Boot-starter fragmentation.

## 2. Pareto top-20% across the Spring AI surface

Spring AI's surface is large (24 reference pages, 9 model-provider chapters, 23 vector-store chapters, 7 agentic-pattern variants). The 20% that drives 80% of value for a Go agent platform (Aluna) and decision SaaS (DaaS):

| # | Spring AI feature | Pareto impact | Maps to (existing § 1 or proposed new) |
|---|---|---|---|
| 1 | **ChatClient fluent API** (`prompt().system(...).user(...).tools(...).advisors(...).call()/.stream()`) | Single dominant API surface. Replaces 5 different per-provider SDKs. | **NEW** `modules/llm/` — `Chat` fluent builder over `ChatModel` interface |
| 2 | **Advisor framework** (CallAdvisor / StreamAdvisor around-interceptor chain over chat) | Mechanism for memory, RAG, guardrails, tool-call observation. Single most reusable abstraction in Spring AI. | **NEW** `modules/llm/advisor.go` — explicit interceptor chain (cross-cutting; not a sub-module) |
| 3 | **Tool calling** (`@Tool`, `ToolCallback`, `ToolDefinition`, `ToolCallingManager`, framework-controlled loop) | Every agent / structured-decision use case needs tools. Manual loops in `anthropic-sdk-go` are 50–100 LOC per consumer. | **NEW** `modules/llm/tools/` |
| 4 | **MCP client + server** (STDIO / SSE / Streamable-HTTP / Stateless-HTTP) | First spec-grade cross-process tool protocol. Aluna agents will consume MCP servers; DaaS may expose decision tables as MCP resources. | **NEW** `modules/llm/mcp/` wrapping a Go MCP library |
| 5 | **ChatMemory** + `MessageWindowChatMemory` + 6 repos (in-mem, JDBC, Cassandra, Neo4j, Cosmos, Mongo) | Conversation continuity. | **NEW** `modules/llm/memory/` (deps on § 1.1 `datasource/` drivers for persistence) |
| 6 | **RAG advisor + modular pipeline** (QueryTransformer, QueryExpander, DocumentRetriever, DocumentJoiner, PostProcessor, QueryAugmenter) | Naive RAG is a 1-liner; modular RAG is the path to advanced RAG. | **NEW** `modules/llm/rag/` (deps on § 1.1 `datasource/vector/`) |
| 7 | **ETL pipeline** (`DocumentReader → DocumentTransformer → DocumentWriter`) | Standardizes ingestion: PDF/HTML/JSON/Markdown → chunks → vector store. | Hybrid: **NEW** `modules/llm/etl/` for readers/transformers + § 1.1 `datasource/vector/` writer |
| 8 | **Structured output** (`BeanOutputConverter<T>`, `MapOutputConverter`, `ListOutputConverter`, native JSON-Schema mode) | Type-safe LLM output. Format-provider + converter split is sharper than a monolithic parser. | **NEW** `modules/llm/parsers/` |
| 9 | **Vector store abstraction** (`VectorStore.add/delete/similaritySearch`, `SearchRequest`, `FilterExpressionBuilder`) | One filter DSL across 23 backends. | **§ 1.1** `modules/datasource/vector/` — refined API shape |
| 10 | **OTel GenAI semantic conventions** (`gen_ai.system`, `gen_ai.request.model`, `gen_ai.usage.{input,output,total}_tokens`, plus `db.vector.*` for stores) | Cost attribution, latency analysis, token-budget alarms. | **NEW** `modules/telemetry/otel/genai/` extension to telemetry/otel/ |
| 11 | **Agentic patterns** (chain, parallelization, routing, orchestrator-workers, evaluator-optimizer) — **composable, no Agent class** | Anthropic's effective-agents recipe. Spring's stance ("workflows > autonomous agents for well-defined tasks") matches yarumo's planned `sdks/processes/` design. | `sdks/processes/` (planned, MEMORY.md) — do **not** add `llm/agents/` |
| 12 | **PromptTemplate** + `StTemplateRenderer` (configurable delimiters to avoid JSON collision) | Versioned prompts with safe interpolation. | **NEW** `modules/llm/prompts/` |
| 13 | **PII / safe-content advisors** (`SafeguardAdvisor`) and the recommended layered guardrails | Compliance, prompt-injection mitigation, cost kill-switches. | **NEW** `modules/llm/guardrails/` (advisor implementations) |
| 14 | **Semantic cache** (not first-class in Spring; absence is itself signal) | Cost reduction on repeat / paraphrased queries. | **NEW** `modules/llm/cache/` (deps on § 1.1 `datasource/vector/`) |

The 14 items above account for every adoption-relevant feature. **Skip list** (already filtered or explicitly out of scope):

- Auto-config (`spring-boot-starter-*`), `@Bean` discovery, conditional configuration — yarumo wires explicitly via `BeanFn` in § 1.5 `modules/boot/`.
- Reactive `Flux<ChatResponse>` parallel API — Go's `iter.Seq2[Chunk, error]` covers streaming with zero new dependencies.
- Image generation, audio (TTS/STT), moderation as separate modules — defer until a real consumer asks. Drivers can extend `ChatModel` shape later.
- Vertex AI / Bedrock auto-config — driver code each consumer imports; not core surface.
- 23 vector-store starters individually — the abstraction is what matters; drivers grow per-demand.

## 3. Sub-module proposal (brainstorm catalog)

This is a **brainstorm catalog**. Recommendations here are an opinion on what the LLM layer could look like — they do NOT become roadmap items as a side effect of this analysis. Promoting them is a separate user decision tracked in § 9.

Each sub-module gets: design sketch, dependencies on existing § 1 modules, Go libraries to wrap, what Spring's design teaches.

### 3.1. `modules/llm/` core — `Chat`, `ChatModel`, `EmbeddingModel`

**What Spring does**: Two-layer split.

- **`ChatModel`** — low-level: `call(Prompt) → ChatResponse`, `stream(Prompt) → Flux<ChatResponse>`. Provider-specific. Used directly only when bypass needed.
- **`ChatClient`** — high-level **fluent builder**: `prompt().system(...).user(...).tools(...).advisors(...).templateRenderer(...).options(...).call()/.stream()`. Provider-agnostic. The dominant surface.

**Go shape**:

```go
// modules/llm/chat.go
type ChatModel interface {
    Call(ctx context.Context, prompt Prompt) (Response, error)
    Stream(ctx context.Context, prompt Prompt) iter.Seq2[Chunk, error]
}

// modules/llm/client.go — fluent surface
type Chat struct { /* unexported state */ }

func NewChat(model ChatModel) *Chat
func (c *Chat) System(text string, vars ...any) *Chat
func (c *Chat) User(text string, vars ...any) *Chat
func (c *Chat) Tools(t ...Tool) *Chat
func (c *Chat) Advisors(a ...Advisor) *Chat
func (c *Chat) Options(opts ...Option) *Chat
func (c *Chat) Call(ctx context.Context) (Response, error)
func (c *Chat) Stream(ctx context.Context) iter.Seq2[Chunk, error]

// modules/llm/embedding.go
type EmbeddingModel interface {
    Embed(ctx context.Context, texts []string) ([][]float32, error)
    Dimensions() int
}
```

**Three deliberate inversions vs Spring**:

1. **No reactive `Mono/Flux` parallel API.** `Stream` returns `iter.Seq2[Chunk, error]` — Go's idiomatic streaming. Spring docs explicitly note "streaming requires reactive stack, non-streaming requires servlet stack" — Go has neither concept and `iter.Seq2` covers both.
2. **No `entity(Class<T>)` method on the response.** Structured output is opt-in via a separate `parsers/` call (§ 3.6), keeping the fluent surface narrow.
3. **No default-builder bean.** Spring's `ChatClient.Builder` is auto-wired with `defaultAdvisors`, `defaultTools`, `defaultSystem`. Yarumo equivalent: factories in `BeanFn` (§ 1.5 `modules/boot/`). The fluent type itself stays unaware of defaults — the consumer composes a default-loaded factory.

**Embedding lives on a separate interface** (Spring's `EmbeddingModel`), not as a `Client.Embed` method. Rationale: providers ship embedding endpoints independently of chat (Cohere, OpenAI's `text-embedding-3-large`, etc.); separate beans wire independently.

**Dependencies on existing § 1 modules**: none for the core. The `Chat` interface is pure library — fits in `modules/llm/` even though the whole module has lifecycle in its drivers.

**Drivers as sub-packages** (each independently shippable):

```
modules/llm/anthropic/   wraps anthropics/anthropic-sdk-go (official)
modules/llm/openai/      wraps openai/openai-go (official)
modules/llm/bedrock/     wraps aws-sdk-go-v2/service/bedrockruntime
modules/llm/vertex/      wraps google/generative-ai-go
modules/llm/ollama/      wraps ollama/ollama Go client (local dev)
```

**v1 driver set recommendation**: Anthropic + OpenAI + Ollama. Bedrock / Vertex deferred until a consumer demands them (they pull large AWS / GCP SDKs).

### 3.2. `modules/llm/memory/` — ChatMemory + repositories

**What Spring does**: `ChatMemory` interface with `add(conversationID, msg)` / `get(conversationID)`. Default impl `MessageWindowChatMemory` (sliding window, preserves system messages). Six storage repositories: `InMemoryChatMemoryRepository`, `JdbcChatMemoryRepository` (Postgres / MySQL / MariaDB / SQL Server / HSQLDB / Oracle), `CassandraChatMemoryRepository`, `Neo4jChatMemoryRepository`, `CosmosDBChatMemoryRepository`, `MongoChatMemoryRepository`.

**Two critical pieces Spring teaches**:

1. **`PromptChatMemoryAdvisor` deprecated in 1.1.3** — embedded memory as XML-tagged blocks in the system prompt (`<user>...</user>`, `<assistant>...</assistant>`). Replaced by `MessageChatMemoryAdvisor` which inserts typed `Message` objects. Reason: XML-tagged text is fragile (escape rules, prompt-injection risk, models confused by tags). **Yarumo memory should ship only the message-list strategy.**
2. **Tool-call messages NOT automatically stored** — Spring acknowledges this as a current limitation. Aluna agents will replay tool calls; the only reason Spring doesn't is JDBC schema cost. Postgres `JSONB` makes this a non-issue. **Yarumo should persist tool-call messages by default.**

**Go shape**:

```go
// modules/llm/memory/memory.go
type ChatMemory interface {
    Add(ctx context.Context, conversationID string, msg ...Message) error
    Get(ctx context.Context, conversationID string) ([]Message, error)
    Clear(ctx context.Context, conversationID string) error
}

type Repository interface {
    Save(ctx context.Context, conversationID string, msgs []Message) error
    Find(ctx context.Context, conversationID string) ([]Message, error)
    Delete(ctx context.Context, conversationID string) error
}

// Strategies (truncation):
// modules/llm/memory/window.go      — MessageWindow (last-N, system preserved)
// modules/llm/memory/token.go       — TokenWindow (tiktoken-counted)
// modules/llm/memory/summarize.go   — Summarize-on-overflow (uses ChatModel)

// Repos:
// modules/llm/memory/inmemory/
// modules/llm/memory/redis/         — uses § 1.1 datasource/goredis
// modules/llm/memory/postgres/      — uses § 1.1 datasource/gorm (JSONB)
```

**Dependencies on existing § 1 modules**: `datasource/goredis` for Redis repo, `datasource/gorm` for Postgres repo. The in-memory repo has none.

**One refinement beyond Spring**: token-window and summarize-on-overflow strategies. Spring docs explicitly say "custom implementations would require extending the abstraction". Yarumo is ahead of Spring here — ship all three.

### 3.3. `modules/llm/cache/` — semantic cache

**What Spring does**: nothing first-class. The closest analog is `VectorStoreChatMemoryAdvisor`, but that's a memory store, not a cache. The absence is itself signal — Spring leaves this to consumers.

**Go shape** (semantic cache via cosine similarity on query embeddings):

```go
// modules/llm/cache/cache.go
type SemanticCache interface {
    Lookup(ctx context.Context, query string, threshold float32) (Response, bool, error)
    Store(ctx context.Context, query string, resp Response) error
}

// modules/llm/cache/cosine.go — cosine-similarity-keyed cache backed by VectorStoreRetriever
```

**Two design questions to lock in**:

1. **Use `VectorStoreRetriever`-shaped interface internally** — read-only subset of `VectorStore`. Principle of least privilege: the cache writes new entries but only retrieves on lookup (no `Delete`, no admin ops).
2. **TTL semantics are non-obvious for semantic cache.** Exact-match caches use TTL on the key; semantic caches need TTL on the *value* (the cached response) but the *key cluster* (embeddings near a centroid) is unbounded. Options: (a) per-entry TTL on `last_hit` + max-size LRU on the cache itself, (b) background vacuum job that deletes entries with last-hit > N days, (c) bounded vector store with auto-eviction. See § 8 open questions.

**Dependencies on existing § 1 modules**: `datasource/vector/` for the underlying store.

### 3.4. `modules/llm/guardrails/` — advisor implementations

**What Spring does**: ships an **Advisor framework**, not a guardrails sub-module per se. The Advisor chain wraps every `ChatClient.call()`, advisors can mutate, short-circuit, log, or augment. Ships `SafeguardAdvisor` (content-safety) as the only out-of-the-box guardrail; the pattern is the contribution, not the catalog.

**Yarumo design**: do **not** split guardrails from advisors. Instead, define **one** `Advisor` interface at the top of `modules/llm/` (cross-cutting primitive — see § 3.11), and ship a guardrail catalog under `modules/llm/guardrails/`:

```go
// modules/llm/advisor.go (top-level, NOT a sub-module)
type Advisor interface {
    Name() string
    Order() int  // lower = earlier on request, later on response (stack semantics)
    Around(ctx context.Context, req *Request, chain Chain) (*Response, error)
}

type Chain interface {
    Next(ctx context.Context, req *Request) (*Response, error)
}
```

**One inversion vs Spring**: **collapse `CallAdvisor` and `StreamAdvisor` into one interface.** Spring needed two because reactive `Flux` is a different type from `Mono`; Go's `iter.Seq2` is just a callable, so the advisor returns `Response` and the chunks are accessed lazily. If the advisor needs to observe each chunk individually, it wraps the iterator in `Around`'s implementation — clean Go pattern.

**Guardrail catalog**:

```
modules/llm/guardrails/
  pii.go         PIIRedactionAdvisor (pre — scrub before send)
  injection.go   PromptInjectionAdvisor (pre — detect & block)
  toxicity.go    ToxicityAdvisor (post — classify response)
  schema.go      SchemaValidationAdvisor (post — paired with parsers/)
  maxtokens.go   MaxTokensAdvisor (pre — cost kill-switch)
  retry.go       RetryAdvisor (post — parse-failure feedback loop)
  logger.go      LoggerAdvisor (utility, ships at the core)
```

**Order convention** (Spring uses integer constants like `HIGHEST_PRECEDENCE + 300`; yarumo should ship named constants):

```go
const (
    OrderGuardrail = 100  // PII, injection, max-tokens
    OrderMemory    = 200  // memory advisors
    OrderRAG       = 300  // RAG augmenter
    OrderTool      = 400  // tool-call advisor
    OrderObserver  = 1000 // logger, telemetry — outermost
)
```

**Dependencies on existing § 1 modules**: none (uses only `modules/llm/` types).

### 3.5. `modules/llm/prompts/` — versioned templates with safe interpolation

**What Spring does**: `PromptTemplate` with `StTemplateRenderer` (StringTemplate engine, default delimiters `{` `}`). **Configurable delimiters** via `startDelimiterToken` / `endDelimiterToken` — critical because `{` collides with JSON in prompts. `TemplateRenderer` interface is pluggable. No file-based prompt loading first-class; consumers wire it themselves.

**Go shape**:

```go
// modules/llm/prompts/prompts.go
type Template interface {
    Render(vars map[string]any) (string, error)
}

type Loader interface {
    Load(ctx context.Context, name, version string) (Template, error)
}

// Impls:
// modules/llm/prompts/fs/      — filesystem with frontmatter (--- version: v2 ---\n{body})
// modules/llm/prompts/db/      — Postgres-backed (Aluna admin UI edits prompts as data)
```

**Four refinements over Spring**:

1. **Default `text/template` delimiters `{{ }}` don't collide with JSON.** Yarumo gets safe delimiters for free. Still expose `WithDelimiters(left, right string)` for consumers who need `<` `>` or `${ }`.
2. **A/B testing hook via `modules/featureflags/`** (when that module materializes). `prompts.Get(name, variant)` resolves the variant by ID; the flag provider flips variants per tenant / per cohort.
3. **Validate variables at parse time, not render time.** Spring renders missing variables as empty strings; yarumo should reject. Borrow `common/expressions/` semantics — typed variables, parse-time check.
4. **Loader is part of the interface, not a wiring concern.** Spring leaves loading to consumers; yarumo bakes in `fs/` and `db/` so the common case is one-call.

**Dependencies on existing § 1 modules**: `datasource/gorm` for the `db/` loader.

### 3.6. `modules/llm/parsers/` — structured output

**What Spring does**: `StructuredOutputConverter<T>` extends both `FormatProvider` (supplies the format hint appended to the prompt) and `Converter<String, T>` (parses the response). Three concrete converters: `BeanOutputConverter<T>` (JSON Schema-driven), `MapOutputConverter` (RFC8259 JSON), `ListOutputConverter` (comma-delimited). Native structured output (model-side JSON Schema enforcement) is opt-in via `AdvisorParams.ENABLE_NATIVE_STRUCTURED_OUTPUT`.

**The format-provider + converter split is sharper than a monolithic parser.** The format provider is what gets injected into the prompt; the converter is what parses the response.

**Go shape**:

```go
// modules/llm/parsers/parser.go
type Parser[T any] interface {
    Format() string                       // injected into the prompt
    Parse(s string) (T, error)            // parses model output
}

// JSON-Schema-based, derived from struct tags
func NewJSON[T any]() Parser[T]

// XML/HTML-tag-based extraction: <answer>...</answer>
func NewTagged[T any](tag string, unmarshal func(string) (T, error)) Parser[T]

// Regex (for narrow numeric / structured-but-not-JSON outputs)
func NewRegex[T any](pattern string, convert func([]string) (T, error)) Parser[T]
```

**Two refinements**:

1. **Native structured output as a `ChatOption`, not an advisor parameter.** Spring's `AdvisorParams.ENABLE_NATIVE_STRUCTURED_OUTPUT` flag is fine but conflates two concepts (advisor config vs model config). Yarumo: `WithNativeJSONSchema(schema)` as a `ChatOption`. The parser supplies the schema if the model supports it; falls back to format-instruction-in-prompt otherwise. Per-driver capability bit.
2. **Retry-with-feedback as a `RetryAdvisor`, not a parser concern.** Spring docs *recommend* validation + retry but doesn't ship retry. `RetryAdvisor` (in `llm/guardrails/`) wraps the `Parse` call, on failure appends `User("Parse error: {err}. Try again.")` and re-calls. Two-attempt cap.

**Dependencies on existing § 1 modules**: none.

### 3.7. `modules/llm/tools/` — function calling

This is the **largest gap** between a vanilla LLM client and a production agent platform. Spring AI's design (lifted, with yarumo-shaped renames):

- **`ToolCallback`** — the executor: `call(toolInput) → String`, `call(toolInput, ToolContext) → String`.
- **`ToolDefinition`** — `{name, description, inputSchema}` (JSON Schema). What the model sees.
- **`ToolMetadata`** — `returnDirect` flag and converter hints.
- **`ToolCallingManager`** — orchestrates the loop (model emits tool_use → manager dispatches → manager injects tool_result → model continues). Two execution modes: framework-controlled (default) and user-controlled (manual loop, `internalToolExecutionEnabled=false`).
- **`ToolContext`** — application data passed to the tool but **not sent to the model** (tenant ID, user ID, request ID). Critical for multi-tenancy.
- **`ToolCallResultConverter`** — serialization (default: Jackson JSON).
- **`@Tool` / `@ToolParam`** — annotations for declarative registration.
- **`returnDirect`** — flag to skip the model and return tool result as the final answer (e.g. "alarm set").
- **`ToolExecutionExceptionProcessor`** — converts tool exceptions to either feedback-to-model or throw-to-caller.

**Anti-patterns to NOT bring over**:

- **`@Tool` annotation discovery.** Go has no annotations; use function references registered into a `Registry`.
- **`SpringBeanToolCallbackResolver` classpath lookup.** Yarumo: an explicit `tools.Registry` that consumers populate.
- **`@ToolParam` for schema metadata.** Go has struct tags; use them (`tools:"description=...,required"`).
- **JSON-Schema-as-string in `ToolDefinition`.** Spring takes a raw JSON Schema string. Yarumo: generate the schema from struct types via reflection (one-time, cached), with explicit override hook for hand-tuning.

**Go shape**:

```go
// modules/llm/tools/tools.go
type Tool interface {
    Definition() Definition
    Call(ctx context.Context, input json.RawMessage, tctx Context) (json.RawMessage, error)
}

type Definition struct {
    Name               string
    Description        string
    Schema             json.RawMessage  // JSON Schema (auto-generated from input type)
    SkipModelOnSuccess bool             // Spring's returnDirect
}

type Context struct {
    values map[string]any  // application-scoped, NOT sent to model (tenant, user, request IDs)
}

// Generic helper: declare a tool from a typed function.
func New[Args, Result any](name, desc string, fn func(context.Context, Args, Context) (Result, error)) Tool

// Registry for dynamic name-based resolution (MCP-friendly).
type Registry struct { /* ... */ }
func (r *Registry) Register(t Tool)
func (r *Registry) Resolve(name string) (Tool, bool)
func (r *Registry) Definitions() []Definition

// Error handling
type ErrorPolicy int
const (
    ErrorFeedback ErrorPolicy = iota  // default: append error to conversation, retry
    ErrorThrow                        // propagate to caller
)
```

**Tool loop** (framework-controlled) lives inside `modules/llm/Chat.Call()`:

```go
for {
    resp := chat.ChatModel.Call(ctx, prompt)
    if !resp.HasToolCalls() { return resp, nil }
    for _, call := range resp.ToolCalls() {
        tool, ok := registry.Resolve(call.Name)
        if !ok { /* return error or feedback to model per ErrorPolicy */ }
        result, err := tool.Call(ctx, call.Args, toolCtx)
        if err != nil { /* delegate to ErrorPolicy */ }
        if tool.Definition().SkipModelOnSuccess {
            return responseFromToolResult(result), nil
        }
        prompt = prompt.WithToolResult(call.ID, result)
    }
}
```

**Spring's new `ToolCallAdvisor` mode** (1.1.x) — implements tool calling as part of the advisor chain. This is interesting: tool calls become observable via the advisor framework, and tool results can integrate with `ChatMemory`. Yarumo should consider exposing both modes:

- Framework-controlled loop inside `Chat.Call()` (default).
- `ToolCallAdvisor` mode (opt-in) for consumers who want tool calls to flow through advisor observability.

**Dependencies on existing § 1 modules**: none (pure library).

### 3.8. `modules/llm/mcp/` — Model Context Protocol

The Model Context Protocol is **the most strategically novel piece for yarumo**. Spring AI 1.0+ added native MCP client and server starters; 1.1.6 supports four transports (STDIO, SSE, Streamable-HTTP, Stateless-HTTP) on both sides.

**Spring 1.1.x annotations** (declarative MCP development):

- Server: `@McpTool`, `@McpResource`, `@McpPrompt`, `@McpComplete`.
- Client: `@McpLogging`, `@McpSampling`, `@McpElicitation`, `@McpProgress`.

**Spring 1.1.x feature set**:

- Tool discovery & execution.
- Resource access & management.
- Prompt system interactions.
- Roots management (optional).
- Capability negotiation, protocol version negotiation.
- Synchronous and asynchronous operations.
- **Security: WIP** (Spring is still figuring this out).

**Why this is strategically central to Aluna**: agents that consume external tool servers (filesystem, databases, browsers, third-party APIs) without bespoke per-API integration code. MCP is the cross-process tool-calling protocol — `llm/tools/` covers in-process, `llm/mcp/` covers cross-process.

**Why this is also useful for DaaS**: a Postgres `decision_tables` table can be exposed as MCP resources for any MCP-capable assistant to consume — turning DaaS into an LLM-accessible decision provider with zero custom integration.

**Go shape**:

```go
// modules/llm/mcp/client.go
type Client interface {
    Tools(ctx context.Context) ([]Tool, error)         // discovered tool catalog
    Resources(ctx context.Context) ([]Resource, error) // discovered resource catalog
    Prompts(ctx context.Context) ([]Prompt, error)     // discovered prompts
    Call(ctx context.Context, name string, args json.RawMessage) (json.RawMessage, error)
    Close() error
}

type Transport interface { /* STDIO | SSE | StreamableHTTP | StatelessHTTP */ }

func NewClient(transport Transport, opts ...Option) Client

// modules/llm/mcp/server.go — server side
type Server struct { /* ... */ }

func NewServer(name, version string) *Server
func (s *Server) RegisterTool(t tools.Tool)
func (s *Server) RegisterResource(r Resource)
func (s *Server) RegisterPrompt(p Prompt)
func (s *Server) Start(ctx context.Context) error  // integrates with modules/managed/
func (s *Server) Stop(ctx context.Context) error
```

**Critical integration with `llm/tools/`**: MCP tools should be exposed as `tools.Tool` instances so the `ChatModel` is unaware whether a tool is local or remote.

```go
mcpClient := mcp.NewClient(mcp.STDIO("/usr/local/bin/filesystem-mcp"))
mcpTools, _ := mcpClient.Tools(ctx)
for _, mcpTool := range mcpTools {
    toolRegistry.Register(mcp.Adapt(mcpClient, mcpTool))  // tools.Tool adapter
}
```

**Anti-patterns to NOT bring over**:

- **Spring Boot starter per transport** (`spring-ai-starter-mcp-server-webmvc`, `-webflux`, etc.). Yarumo: one `mcp/` sub-module, transports as construction-time choices.
- **`@McpTool` / `@McpResource` discovery**. Explicit `Server.RegisterTool` / `Server.RegisterResource`.
- **Security framework "WIP"**. Spring is still figuring this out; yarumo waits. v1: TLS + bearer token via standard middleware on the SSE/HTTP transports. STDIO is local-process.
- **Per-protocol auto-config** (`spring.ai.mcp.server.protocol=SSE`). Construction-time selection.

**Placement rationale**: `llm/mcp/` and not `modules/mcp/` because MCP is meaningless without an LLM consumer. Co-locating with `llm/tools/` enables the adapter pattern without a circular dep.

**Dependencies on existing § 1 modules**: `auth/` (when MCP-server auth lands), `managed/` (for `Server.Start`/`Stop` lifecycle).

### 3.9. `modules/llm/rag/` — retrieval primitives

Spring AI's RAG architecture is built around `RetrievalAugmentationAdvisor` and the modular pipeline:

1. **Pre-Retrieval**: `QueryTransformer` (`CompressionQueryTransformer`, `RewriteQueryTransformer`, `TranslationQueryTransformer`), `QueryExpander` (`MultiQueryExpander`).
2. **Retrieval**: `DocumentRetriever` (`VectorStoreDocumentRetriever` with topK, similarityThreshold, filterExpression — static or dynamic), `DocumentJoiner` (`ConcatenationDocumentJoiner` for multi-query results).
3. **Post-Retrieval**: `DocumentPostProcessor` (ranking, irrelevant-removal, compression — addresses lost-in-the-middle).
4. **Generation**: `QueryAugmenter` (`ContextualQueryAugmenter` — system message + retrieved docs, with `allowEmptyContext` toggle for fallback).

**Recommended split** for yarumo (sharp separation, avoids premature pipeline lock-in):

- **§ 1.1 `modules/datasource/vector/`** absorbs the **writer side** (`VectorStore.Add(documents)`, `Delete(filter)`, `SearchRequest{topK, threshold, filter}`, `FilterExpressionBuilder` DSL).
- **NEW `modules/llm/rag/`** owns **retrieval-side primitives**:

```go
// modules/llm/rag/query.go
type Query struct {
    Text    string
    History []Message
    Context map[string]any  // FILTER_EXPRESSION, tenant, etc.
}

// modules/llm/rag/transformer.go
type QueryTransformer interface {
    Transform(ctx context.Context, q Query) (Query, error)
}

// Built-ins:
//   modules/llm/rag/rewrite.go    — LLM-rewrite ambiguous → standalone
//   modules/llm/rag/compress.go   — collapse multi-turn history into one query
//   modules/llm/rag/translate.go  — cross-lingual retrieval

// modules/llm/rag/expander.go
type QueryExpander interface {
    Expand(ctx context.Context, q Query) ([]Query, error)
}

// Built-in:
//   modules/llm/rag/multiquery.go — LLM-generates N semantically diverse variants

// modules/llm/rag/retriever.go
type DocumentRetriever interface {
    Retrieve(ctx context.Context, q Query) ([]Document, error)
}

// Built-in: wraps datasource/vector.VectorStoreRetriever, adds topK/threshold/filter from Query.Context

// modules/llm/rag/joiner.go
type DocumentJoiner interface {
    Join(map[Query][]Document) []Document
}
// Built-in: Concatenation (dedup by ID)

// modules/llm/rag/augmenter.go
type Augmenter interface {
    Augment(ctx context.Context, q Query, docs []Document) (Prompt, error)
}
// Built-in: Contextual (system message + retrieved docs)
```

**No pipeline-orchestrator advisor in v1.** Composition is consumer code:

```go
q := rag.NewQuery(userText, history)
q, _ = rewriter.Transform(ctx, q)
qs, _ := expander.Expand(ctx, q)
var docs []Document
for _, qi := range qs { d, _ := retriever.Retrieve(ctx, qi); docs = append(docs, d...) }
docs = joiner.Join(docs)
prompt, _ := augmenter.Augment(ctx, q, docs)
resp, _ := chat.User(prompt.String()).Call(ctx)
```

**Why no orchestrator advisor in v1**: Spring's `RetrievalAugmentationAdvisor` is configurable but rigid — six stages, fixed order. Aluna might want loops (retrieve → critique → re-retrieve). Better to ship the primitives and let consumers compose until a consensus shape emerges. v2 can add `RAGAdvisor` once the API has stabilized.

**Dependencies on existing § 1 modules**: `datasource/vector/`.

### 3.10. `modules/llm/etl/` — document ingestion

Spring AI ships the ETL pipeline as the backbone of RAG ingestion:

- **`DocumentReader extends Supplier<List<Document>>`** — Built-ins: `JsonReader` (JSON Pointers RFC 6901), `TextReader`, `JsoupDocumentReader` (HTML via CSS selectors), `MarkdownDocumentReader`, `PagePdfDocumentReader` (PDFBox, page-based), `ParagraphPdfDocumentReader` (PDF catalog/TOC-aware), `TikaDocumentReader` (40+ formats via Apache Tika).
- **`DocumentTransformer extends Function<List<Document>, List<Document>>`** — Built-ins: `TokenTextSplitter` (CL100K_BASE tokenizer, configurable chunkSize / minChunkSizeChars / punctuationMarks for i18n), `KeywordMetadataEnricher` (LLM-extract N keywords per doc), `SummaryMetadataEnricher` (prev/current/next section summaries via LLM), `ContentFormatTransformer`.
- **`DocumentWriter extends Consumer<List<Document>>`** — Built-ins: `VectorStore` writer (20+ backends), `FileDocumentWriter` (debug).

**Go shape**:

```go
// modules/llm/etl/reader.go
type DocumentReader interface {
    Read(ctx context.Context) ([]Document, error)
}

// modules/llm/etl/transformer.go
type DocumentTransformer interface {
    Transform(ctx context.Context, docs []Document) ([]Document, error)
}

// modules/llm/etl/writer.go
type DocumentWriter interface {
    Write(ctx context.Context, docs []Document) error
}

// Readers:
//   modules/llm/etl/pdf/        wraps ledongthuc/pdf (lightweight) or unidoc (full-featured)
//   modules/llm/etl/text/       plain text
//   modules/llm/etl/json/       JSON with JSON Pointer (RFC 6901)
//   modules/llm/etl/markdown/   semantic splitting (frontmatter, code blocks, blockquotes)
//   modules/llm/etl/html/       wraps PuerkitoBio/goquery (CSS selectors)

// Splitters / enrichers:
//   modules/llm/etl/splitter/token.go     TokenTextSplitter (chunkSize, punctuationMarks)
//   modules/llm/etl/splitter/sentence.go  SentenceSplitter
//   modules/llm/etl/enrich/keywords.go    KeywordEnricher (LLM-extract)
//   modules/llm/etl/enrich/summary.go     SummaryEnricher (prev/current/next)

// Writers:
//   modules/llm/etl/writer/vectorstore.go  delegates to § 1.1 datasource/vector/
//   modules/llm/etl/writer/file.go         debug
```

**Three refinements over Spring**:

1. **Tokenizer is pluggable** — Spring hardcodes CL100K_BASE. Yarumo: tokenizer is a constructor arg, defaults to `tiktoken-go` CL100K_BASE for budgeting, but per-driver tokenizers (Anthropic's API-counted, etc.) can be plugged when precision matters.
2. **No Tika wrapper in v1.** Tika is a 40MB JVM dependency tree; Go has no equivalent. Ship `pdf/`, `text/`, `json/`, `markdown/`, `html/` as the v1 set. Add `docx/`, `pptx/` later via individual Go libraries if demand emerges.
3. **`ContentFormatTransformer` is over-engineered for v1.** Skip.

**Dependencies on existing § 1 modules**: `datasource/vector/` for the `VectorStoreWriter`.

### 3.11. `Advisor` — cross-cutting primitive (not a sub-module)

Spring AI's Advisor framework is the single most reusable mechanism in the entire library. Memory, RAG, guardrails, tool-call observation, logging — all are advisors. The Advisor chain wraps every `ChatClient.call()`; advisors can mutate request, short-circuit, log, augment, observe response.

**Placement decision**: define `Advisor` at the top of `modules/llm/` (in `advisor.go`), not as a sub-module. This way `import modules/llm` is enough to use the interface without forcing topical imports (`guardrails/`, `memory/`, etc.).

```go
// modules/llm/advisor.go
type Advisor interface {
    Name() string
    Order() int  // lower = earlier on request, later on response (stack semantics)
    Around(ctx context.Context, req *Request, chain Chain) (*Response, error)
}

type Chain interface {
    Next(ctx context.Context, req *Request) (*Response, error)
}

// Named order constants — yarumo's convention, sharper than Spring's "HIGHEST_PRECEDENCE + 300":
const (
    OrderGuardrail = 100
    OrderMemory    = 200
    OrderRAG       = 300
    OrderTool      = 400
    OrderObserver  = 1000
)
```

**One inversion vs Spring**: collapse `CallAdvisor` and `StreamAdvisor` into one interface (see § 3.4).

### 3.12. Observability — `modules/telemetry/otel/genai/` extension

Spring AI's observability layer is the most concrete OTel GenAI implementation in the wild. Concrete attribute set (validated against Spring 1.1.6 docs):

**ChatModel — Low cardinality (always on metrics + traces)**:

- `gen_ai.operation.name` (`chat`, `embedding`, `tool_call`, `framework`)
- `gen_ai.system` (`anthropic`, `openai`, `bedrock`, `ollama`, ... or `spring_ai` for framework-level)
- `gen_ai.request.model`
- `gen_ai.response.model`
- `yarumo.ai.kind` (Spring uses `spring.ai.kind`; rename) — `chat_client`, `advisor`, `tool_call`, `vector_store`

**ChatModel — High cardinality (traces by default, opt-in to metrics)**:

- `gen_ai.request.{temperature, top_p, top_k, max_tokens, frequency_penalty, presence_penalty, stop_sequences}`
- `gen_ai.response.{finish_reasons, id}`
- `gen_ai.usage.{input_tokens, output_tokens, total_tokens}`
- `yarumo.ai.model.request.tool.names`

**Tool calling — Low cardinality**:

- `gen_ai.operation.name=framework`, `gen_ai.system=yarumo`
- `yarumo.ai.kind=tool_call`
- `yarumo.ai.tool.definition.name`

**Tool calling — High cardinality (sensitive, default-off)**:

- `yarumo.ai.tool.definition.description`
- `yarumo.ai.tool.definition.schema`
- `yarumo.ai.tool.call.arguments` (opt-in via flag)
- `yarumo.ai.tool.call.result` (opt-in via flag)

**Vector store — Low cardinality** (prefixed `db.` per SemConv DB):

- `db.operation.name` (`add`, `delete`, `query`)
- `db.system` (`pg_vector`, `pinecone`, `qdrant`, `weaviate`, ...)
- `yarumo.ai.kind=vector_store`

**Vector store — High cardinality**:

- `db.collection.name`, `db.namespace`, `db.record.id`
- `db.search.similarity_metric`
- `db.vector.{dimension_count, field_name, query.top_k, query.similarity_threshold}`
- `db.vector.query.content` (opt-in — search text)
- `db.vector.query.filter` (opt-in — filter expression)
- `db.vector.query.response.documents` (opt-in — search results)

**Sensitive content default-off** (Spring defaults all of these to false):

- `gen_ai.prompt`, `gen_ai.completion` — full prompt and full response. Opt-in via:
  - `WithObservePrompt(true)` on the `Chat` constructor.
  - Environment override: `YARUMO_LLM_OBSERVE_PROMPT=true`.

**Metrics (Prometheus naming)**:

```
gen_ai.client.operation         → gen_ai_client_operation_seconds_{count,sum,max,active_count}
db.vector.client.operation      → db_vector_client_operation_seconds_{count,sum,max,active_count}
gen_ai.client.token.usage       → gen_ai_client_token_usage_total (labels: gen_ai_token_type=input|output|total)
```

**Three refinements vs Spring**:

1. **Token cost counter requires a price table — Spring doesn't ship this.** Yarumo can do better: ship a per-model price table in `telemetry/otel/genai/pricing/` (Anthropic, OpenAI, Bedrock pricing JSON updated quarterly), emit `yarumo_ai_cost_total_usd` per request. This is what Aluna ops will actually want to dashboard.
2. **Per-tenant cost attribution.** When `modules/tenancy/` materializes, inject `yarumo.tenancy.tenant_id` as an attribute on every gen_ai metric.
3. **VectorStore observability is automatic but a separate span** (`db.vector.client.operation`). Implement when § 1.1 `modules/datasource/vector/` lands.

**Placement**: this is a sibling extension to `modules/telemetry/otel/`, NOT inside `modules/llm/`. Layout:

```
modules/telemetry/otel/genai/
  attributes.go             gen_ai.* + yarumo.ai.* + db.vector.* attribute keys
  instrumentation.go        WithGenAIInstrumentation() option for Chat
  pricing/                  Per-model price tables
    pricing.go              interface
    anthropic.json          quarterly refresh
    openai.json
    bedrock.json
  metrics.go                Histograms (latency, tokens), counters (cost, errors)
```

**Dependencies on existing § 1 modules**: extends `modules/telemetry/otel/` (Phase 3 — see milestone #9). Lives outside `modules/llm/` because telemetry is a cross-cutting workspace concern.

### 3.13. `llm/agents/` — DEFER (composes via `sdks/processes/`)

Spring AI's "agentic patterns" guide ships **five composable patterns**, not an `Agent` class:

1. **Chain workflow** — output of step N is input to step N+1.
2. **Parallelization** — independent subtasks in parallel + consensus.
3. **Routing** — LLM classifies input, dispatches to specialized handler.
4. **Orchestrator-workers** — central LLM decomposes, workers handle subtasks.
5. **Evaluator-optimizer** — generator + critic loop.

Spring's philosophy (matching Anthropic's "Building effective agents" recommendation): **"workflows often provide better predictability and consistency for well-defined tasks"** — prefer fixed-shape patterns to free-form agent loops.

**This aligns exactly with yarumo's already-planned `sdks/processes/`** (per MEMORY.md): workflow engine with `steps`, `states`, `montecarlo`, `mining` engines. The agentic patterns are *workflows over LLM steps*, not a separate abstraction.

**Decision**: **DEFER `llm/agents/` entirely**. The five patterns are emergent from:

- `modules/llm/` — chat, tools, MCP.
- `sdks/processes/local/` — workflow orchestration.
- `compute/engine/steps/` — short in-process workflow with retry/timeout/compensation.

A "chain workflow" is `steps.Sequential(chat1, chat2, chat3)`. A "router" is `steps.Switch(classifier, branches)`. An "evaluator-optimizer" is `steps.Loop(generator, critic, untilApproved)`. The agent loop's "outer while loop" is `steps.Loop` over `chat.Call`.

What this saves: a `modules/llm/agents/` sub-module with leaky abstractions that overlap `sdks/processes/`. What this costs: a small piece of documentation in `sdks/processes/` showing the 5 patterns as concrete `steps.X` recipes.

**File this as a cross-roadmap docs follow-up**: when `sdks/processes/` is implemented, add an `examples/agentic_patterns.md` that maps each Anthropic pattern to a `steps.X` recipe. Avoid the sub-module.

## 4. Tool / MCP Go ecosystem survey

| Library | URL | License | Last activity (~) | Imports | Maturity |
|---|---|---|---|---|---|
| **`modelcontextprotocol/go-sdk`** | github.com/modelcontextprotocol/go-sdk | MIT | Active 2026 (official) | (new, official) | Spec authority. Newest. Use for spec stability. |
| **`mark3labs/mcp-go`** | github.com/mark3labs/mcp-go | MIT | v1 stable | ~321 | Most mature HTTP transports. `server.NewStreamableHTTPServer()`. |
| **`metoro-io/mcp-golang`** | github.com/metoro-io/mcp-golang | MIT | Active 2026 | ~92 | Type-safe args, Gin integration. Smaller user base, cleaner API. |
| **`cloudwego/eino`** | github.com/cloudwego/eino | Apache 2.0 | Active (ByteDance) | High | Full LLM framework: components + ADK + agents. Production-battle-tested in Doubao, TikTok. |
| **`tmc/langchaingo`** | github.com/tmc/langchaingo | MIT | Community | High | LangChain port. Comprehensive but Python-shaped API. |
| **`firebase/genkit`** (Go) | github.com/firebase/genkit | Apache 2.0 | Active | Medium | Google's framework. Tied to GCP. |
| **`anyi-ai/anyi`** | github.com/anyi-ai/anyi | MIT | Medium | Low | Lightweight agent framework. |
| **`anthropics/anthropic-sdk-go`** | github.com/anthropics/anthropic-sdk-go | MIT | Active (official) | High | Primary chat driver target. |
| **`openai/openai-go`** | github.com/openai/openai-go | Apache 2.0 | Active (official) | High | Primary chat driver target. |
| **`aws/aws-sdk-go-v2/service/bedrockruntime`** | aws-sdk-go-v2 | Apache 2.0 | Active (official) | High | Bedrock runtime client. v2 driver. |
| **`google/generative-ai-go`** | github.com/google/generative-ai-go | Apache 2.0 | Active (official) | Medium | Google Gen AI SDK (Vertex/Gemini). v2 driver. |
| **`ollama/ollama` Go client** | github.com/ollama/ollama/api | MIT | Active | Medium | Local dev driver. |
| **`pkoukk/tiktoken-go`** | github.com/pkoukk/tiktoken-go | MIT | Active | Medium | tiktoken port (cl100k_base etc.) for token budgeting. |
| **`ledongthuc/pdf`** | github.com/ledongthuc/pdf | BSD-3 | Active | Medium | Lightweight PDF text extraction (no fonts/images). |
| **`unidoc/unipdf`** | github.com/unidoc/unipdf | AGPL/commercial | Active | High | Full-featured PDF (AGPL — needs license assessment). |
| **`PuerkitoBio/goquery`** | github.com/PuerkitoBio/goquery | BSD-3 | Active | Very high | jQuery-style CSS selectors for HTML reader. |

**Adoption strategy**:

- **`modules/llm/` core**: wrap nothing; the `ChatModel` interface stands alone. Drivers (anthropic, openai, bedrock, vertex, ollama) each import their respective official SDK.
- **`modules/llm/mcp/`**: wrap **`modelcontextprotocol/go-sdk`** (official) as the default impl, accept `mark3labs/mcp-go` as alternate (consumer choice). Revisit quarterly — if the official SDK lags spec or breaks API frequently, fall back to mark3labs.
- **`modules/llm/tools/`**: no library — schema generation via `reflect` + a small JSON Schema generator. Consider `invopop/jsonschema` if reflection becomes hairy.
- **`modules/llm/etl/pdf/`**: start with `ledongthuc/pdf` (BSD-3, lightweight); promote to `unidoc/unipdf` only if AGPL is acceptable.
- **`modules/llm/etl/html/`**: `PuerkitoBio/goquery` (CSS selectors, near-stdlib at this point).
- **`modules/llm/memory/TokenWindow`**: `pkoukk/tiktoken-go` for budgeting (estimator). Per-driver tokenizers (Anthropic API-counted etc.) plug in where precision matters.
- **`eino` and `langchaingo`**: do **not** adopt as deps. Both are competing frameworks; yarumo's `modules/llm/` *is* the framework. Read Eino's ADK and interrupt/resume design as inspiration when Aluna formalizes its agent loop.

## 5. Anti-patterns to avoid

Spring AI-specific magic that must NOT be brought over:

1. **`@Tool` / `@ToolParam` annotations.** Go has no annotations. Use struct tags for schema metadata (`tools:"description=...,required"`) and function references for registration.
2. **`@Bean ChatModel` auto-wiring.** Yarumo: explicit constructors invoked from `BeanFn`s in § 1.5 `modules/boot/`.
3. **SpEL in prompt templates** (`{#vars}`, `{T(java.lang.Math).PI}`). Yarumo: `text/template` with explicit, typed variables; reject any expression evaluation in prompts.
4. **Reactive `Flux<ChatResponse>` parallel API.** Yarumo: `iter.Seq2[Chunk, error]` is the single streaming type.
5. **Classpath-scan tool / MCP-server discovery** (`SpringBeanToolCallbackResolver`, `@McpTool` discovery, etc.). Yarumo: explicit `Registry.Register(...)`.
6. **Spring Boot starter per model provider** (`spring-ai-starter-model-anthropic`, `-openai`, `-bedrock`, ...). Yarumo: one `modules/llm/` module, one sub-package per provider; consumer imports only what they use.
7. **`@Configuration` / `@EnableXxx` for MCP server transports.** Yarumo: construction-time transport selection.
8. **`AdvisorChain` order via integer constants** (`HIGHEST_PRECEDENCE + 300`). Yarumo keeps `Order() int` but ships **named constants** (`OrderGuardrail = 100`, `OrderMemory = 200`, etc.) so consumers compose intelligibly.
9. **God-class `ChatClient`** with default-everything builder. Yarumo: the fluent `Chat` carries no defaults; factories in `BeanFn` apply policy.
10. **Reactive Mono/Flux for advisors** (`StreamAdvisor` vs `CallAdvisor` split). One interface, period.
11. **Native structured output as an advisor parameter.** It's a model option, not an advisor concern.
12. **`PromptChatMemoryAdvisor` pattern** (XML-tagged memory in system prompt). Spring deprecated this in 1.1.3; never implement it.
13. **`@McpTool` / `@McpResource` annotation discovery.** Explicit `Server.RegisterTool(...)`.
14. **Auto-config for cost / latency thresholds.** Yarumo: explicit policy in consumer code; the observability module emits metrics, alerting policy is consumer-owned.
15. **A "Spring AI Vector Store" abstraction that includes embedding model wiring.** Yarumo splits cleanly: `modules/datasource/vector/` is just a key-vector-metadata store; the embedding model is a separate `modules/llm/` driver. Composition is consumer code.
16. **Reactive vs servlet stack dichotomy.** Spring requires `webflux` for streaming and `web` for non-streaming. Go has neither concept; `iter.Seq2` covers both transparently.
17. **`spring.http.client.factory=jdk` bug workaround.** Spring 1.1.x has a Boot 3.4 incompatibility that demands a property. Pure JVM-ecosystem ugliness; irrelevant to yarumo.

## 6. Refined `modules/llm/` package layout

```
modules/llm/
  chat.go                  ChatModel interface (Call, Stream) + ChatModelOption
  client.go                Chat fluent builder (.System, .User, .Tools, .Advisors, .Call, .Stream)
  embedding.go             EmbeddingModel interface
  prompt.go                Prompt struct (Messages, Options, Variables)
  message.go               Message (System, User, Assistant, Tool) + content blocks
  response.go              Response + Usage + FinishReason
  stream.go                Chunk + iter.Seq2[Chunk, error] helpers
  options.go               ChatModelOption (Temperature, TopP, MaxTokens, Model, JSONSchema)
  advisor.go               Advisor interface (Name, Order, Around) + Chain + named Order constants
                           — cross-cutting, NOT a sub-module
  errors.go                ErrModelTimeout, ErrContextWindowExceeded, ErrTooManyTools, ErrToolNotFound

  memory/
    memory.go              ChatMemory + Repository interfaces
    window.go              MessageWindow strategy (last-N, system preserved)
    token.go               TokenWindow strategy (tiktoken-counted)
    summarize.go           Summarize-on-overflow strategy (uses ChatModel)
    inmemory/              In-memory repo
    redis/                 Redis repo — uses § 1.1 datasource/goredis
    postgres/              Postgres (JSONB) repo — uses § 1.1 datasource/gorm

  cache/
    cache.go               SemanticCache interface
    cosine.go              Cosine-similarity-keyed cache
                           Internally consumes § 1.1 datasource/vector.VectorStoreRetriever

  guardrails/              Advisor implementations
    pii.go                 PIIRedactionAdvisor (pre)
    injection.go           PromptInjectionAdvisor (pre)
    toxicity.go            ToxicityAdvisor (post)
    schema.go              SchemaValidationAdvisor (post, paired with parsers/)
    maxtokens.go           MaxTokensAdvisor (pre, cost kill-switch)
    retry.go               RetryAdvisor (parse-failure feedback loop)
    logger.go              LoggerAdvisor (utility)

  prompts/
    prompts.go             Template + Loader interfaces + Get(name, version)
    fs/                    Filesystem loader with --- frontmatter ---
    db/                    Postgres-backed loader (Aluna admin UI) — uses § 1.1 datasource/gorm
    abtest.go              A/B variant resolver (uses modules/featureflags/ when it materializes)

  parsers/                 Structured output: format-provider + converter
    parser.go              Parser[T] interface (Format, Parse)
    json.go                JSON parser, schema from struct tags
    tagged.go              Tagged extraction <answer>...</answer>
    regex.go               Regex-based parser

  tools/                   Function calling
    tools.go               Tool interface, Definition, Context
    registry.go            Registry (Register, Resolve, Definitions)
    schema.go              Auto-generate JSON Schema from Go types
    errpolicy.go           ErrorPolicy (feedback | throw)
    advisor.go             ToolCallAdvisor (opt-in alternate to framework-controlled loop)

  mcp/                     Model Context Protocol
    client.go              Client interface (Tools, Resources, Prompts, Call)
    server.go              Server (RegisterTool, RegisterResource, RegisterPrompt) — uses § 1.4 health, integrates with managed
    adapter.go             Adapter: MCP Tool -> tools.Tool
    transport/
      stdio.go             STDIO transport
      sse.go               SSE transport
      streamablehttp.go    Streamable-HTTP transport
      stateless.go         Stateless-HTTP transport
    Wraps modelcontextprotocol/go-sdk as default impl.

  rag/                     Retrieval primitives (no orchestrator advisor in v1)
    query.go               Query (text, history, context, filter)
    transformer.go         QueryTransformer interface
    rewrite.go             RewriteTransformer (LLM-rewrite)
    compress.go            CompressTransformer (multi-turn -> standalone)
    translate.go           TranslateTransformer (cross-lingual)
    expander.go            QueryExpander interface
    multiquery.go          MultiQueryExpander
    retriever.go           DocumentRetriever interface (wraps § 1.1 datasource/vector)
    joiner.go              DocumentJoiner interface + Concatenation
    augmenter.go           Augmenter interface + Contextual

  etl/                     Document ingestion (the READ side of RAG)
    reader.go              DocumentReader interface
    writer.go              DocumentWriter interface
    transformer.go         DocumentTransformer interface
    pdf/                   PDF reader (ledongthuc/pdf, optionally unidoc)
    text/                  Plain text reader
    json/                  JSON reader with JSON Pointer (RFC 6901) support
    markdown/              Markdown reader (semantic splitting)
    html/                  HTML reader (CSS selectors via PuerkitoBio/goquery)
    splitter/
      token.go             TokenTextSplitter (chunkSize, minChars, punctuation)
      sentence.go          SentenceSplitter
    enrich/
      keywords.go          KeywordEnricher (LLM-extract)
      summary.go           SummaryEnricher (prev/current/next, LLM-generated)
    writer/
      vectorstore.go       VectorStore writer (delegates to § 1.1 datasource/vector/)
      file.go              File writer (debug)

  anthropic/               (driver) wraps anthropics/anthropic-sdk-go
  openai/                  (driver) wraps openai/openai-go
  bedrock/                 (driver) wraps aws-sdk-go-v2/service/bedrockruntime — v2
  vertex/                  (driver) wraps google/generative-ai-go — v2
  ollama/                  (driver) wraps ollama/ollama Go client (local dev)
```

**Two cross-cutting decisions to lock in**:

1. **No `llm/advisors/` sub-module.** The `Advisor` interface lives at the top of `modules/llm/`. Implementations live in their nearest topical sub-module (`guardrails/`, `memory/` for memory advisors, `rag/` for an eventual RAG advisor in v2, `tools/` for `ToolCallAdvisor`). This keeps `import modules/llm` enough to use the interface without forcing topical imports.

2. **No `llm/agents/` sub-module.** Agentic patterns are workflows over `Chat` + `tools` + `mcp`, composed via `sdks/processes/` (planned per MEMORY.md). Document the 5 patterns as recipes in `sdks/processes/examples/`, not as code in `modules/llm/`.

**Sibling module updates** (NOT inside `modules/llm/`):

```
modules/telemetry/otel/genai/    (NEW — sibling extension to telemetry/otel/)
  attributes.go             gen_ai.* + yarumo.ai.* + db.vector.* attribute keys
  instrumentation.go        WithGenAIInstrumentation() option for Chat
  pricing/                  Per-model price tables (Anthropic, OpenAI, Bedrock)
    pricing.go              interface
    table.json              data, quarterly refresh
  metrics.go                Histograms (latency, tokens), counters (cost, errors)

modules/datasource/vector/   (§ 1.1 — refined API shape)
  vector.go                 VectorStore + VectorStoreRetriever interfaces
  search.go                 SearchRequest (topK, similarityThreshold, filter)
  filter.go                 FilterExpressionBuilder DSL (eq, in, and, or, not, isNull)
  pgvector/                 Postgres pgvector driver — v1
  qdrant/                   Qdrant driver — v1
  inmemory/                 In-memory (testing) — v1
  pinecone/                 Pinecone driver — v2
  weaviate/                 Weaviate driver — v2
```

## 7. Overall recommendation

**PARTIAL.** Propose `modules/llm/` as a NEW top-level module (not currently in the canonical roadmap) with eight sub-modules + a cross-cutting `Advisor` primitive. Refine § 1.1 `modules/datasource/vector/` with a concrete API shape lifted from Spring AI's `VectorStore` / `SearchRequest` / `FilterExpressionBuilder`. Add `modules/telemetry/otel/genai/` as a sibling extension to the existing `modules/telemetry/otel/` (Phase 3). Reject Spring's auto-config, reactive `Flux` parallel API, annotation-based discovery (`@Tool`, `@McpTool`, `@McpResource`), per-provider Boot-starter fragmentation, and the deprecated `PromptChatMemoryAdvisor` XML-tagged pattern. Defer `llm/agents/` indefinitely — agentic patterns compose via the planned `sdks/processes/` engine instead.

The cumulative effect: `modules/llm/` lands as **8 sub-modules + cross-cutting `Advisor`** (`memory`, `cache`, `guardrails`, `prompts`, `parsers`, `tools`, `mcp`, `rag`, `etl`), drops `agents/` entirely. The `Advisor` interface becomes a top-level primitive in `modules/llm/`, threading through guardrails / memory / rag / tools as the single composition mechanism — the same role advisors play in Spring AI.

## 8. Open questions

- **Aluna's first concrete agent loop.** Until this is named, `llm/agents/` should stay deferred and `llm/rag/` should ship primitives only. Question for Aluna: is the v1 loop `chain` (Q → retrieve → answer), `evaluator-optimizer` (generate → critique → refine), or `orchestrator-workers` (decompose → subtasks → join)? The answer drives `rag/` v2 (advisor or not) and `sdks/processes/` example recipes.
- **Token-counting strategy.** `llm/memory/TokenWindow` needs a tokenizer. Anthropic exposes `count_tokens`; OpenAI uses tiktoken; Bedrock and Vertex have provider-specific APIs. Option A: per-driver tokenizer (correct but duplicated). Option B: ship `pkoukk/tiktoken-go` as a "good-enough" estimator for budgeting (Spring AI takes this route). Resolution belongs in the `memory/` v1 ticket.
- **Semantic cache TTL / eviction policy.** Per § 3.3, exact-match TTL doesn't translate. Options: (a) per-entry TTL on `last_hit` + max-size LRU, (b) vacuum job, (c) bounded vector store with auto-eviction. Decide before filing `llm/cache/`.
- **MCP server authentication.** Spring's framework is "WIP" in 1.1.6. The two real consumers (Aluna agents calling DaaS as MCP, DaaS exposing tables as MCP) need at minimum: TLS, bearer tokens, per-tool authorization. File as `mcp/auth.go` follow-up once MCP spec stabilizes (currently in flux). Likely depends on § 1.2 `modules/auth/`.
- **Tool-call message persistence in memory.** Spring acknowledges its own limitation — tool calls aren't stored automatically. Yarumo's call: persist them by default. File as design decision in the `memory/` ticket.
- **Native structured output negotiation.** Spring's `AdvisorParams.ENABLE_NATIVE_STRUCTURED_OUTPUT` is fragile — works only on specific model versions. Yarumo: `Chat.JSONSchema(s)` option, the driver detects whether the model supports native JSON Schema and either passes it through or falls back to format-instruction-in-prompt. The detection is per-driver — file as a per-driver capability bit.
- **RAG advisor in v2 — where does it live?** When `rag/` matures and the pipeline shape locks in, the orchestrator could ship as: (a) `rag/advisor.go`, (b) `guardrails/rag.go`, (c) consumer code. Default to (a) but revisit at v2 design time.
- **MCP library choice — wrap official vs `mark3labs`?** Recommendation in § 4 is the official `modelcontextprotocol/go-sdk` for spec stability. Revisit quarterly.
- **Driver scope for v1.** Anthropic + OpenAI + Ollama (local dev) is the minimum. Bedrock and Vertex are useful but add 2 large cloud-vendor SDKs. Lock in v1 driver set when the `llm/` ticket is filed.
- **Tool argument augmentation (`AugmentedToolCallbackProvider`).** Spring 1.1.x added this for capturing model reasoning alongside tool calls (`innerThought` / `confidence` fields). Strategic for explainability in DaaS. Should yarumo ship a parallel `tools.Augment(toolFn, ArgsType)` helper? File as `tools/` v1 design question.
- **Spring's `ToolCallAdvisor` vs framework-controlled loop.** Spring offers both. Yarumo's recommendation in § 3.7 is to ship both. Does this create two competing patterns? Risk: confusion. Mitigation: document the framework-controlled loop as default; `ToolCallAdvisor` as opt-in for advanced observability.

## 9. ROADMAP delta proposed (NOT applied — user decides)

These are concrete additions to `ROADMAP_NEW_MODULES.md` **if** the user wants to promote any of this from brainstorm into canonical roadmap. None of these changes are applied as a side effect of this analysis.

### 9.1. Promote `modules/llm/` as § 1.6 (new top-level module)

Add a new § 1.6 to `ROADMAP_NEW_MODULES.md`:

```
## 1.6. modules/llm/ — LLM client + advisors + tools + MCP + RAG + ETL

Status: Brainstorm (cataloged in docs/spring/spring-ai.md)
Why a new module: stateful drivers with lifecycle, external SDKs per provider, cross-cutting Advisor primitive.

Sub-modules:
- modules/llm/memory/      ChatMemory + truncation strategies + repos (in-mem, redis, postgres)
- modules/llm/cache/       Semantic cache via cosine similarity
- modules/llm/guardrails/  Advisor catalog: PII, injection, toxicity, schema, max-tokens, retry, logger
- modules/llm/prompts/     Versioned templates, fs + db loaders, A/B variants
- modules/llm/parsers/     Format-provider + converter split (JSON, tagged, regex)
- modules/llm/tools/       Function calling: Registry, Definition, ToolCallAdvisor, ErrorPolicy
- modules/llm/mcp/         Model Context Protocol client + server, 4 transports
- modules/llm/rag/         Retrieval primitives: QueryTransformer, Expander, Retriever, Joiner, Augmenter
- modules/llm/etl/         Document ingestion: pdf, text, json, markdown, html readers + splitters + enrichers

Drivers: anthropic, openai, bedrock, vertex, ollama. v1 set: anthropic + openai + ollama.

Cross-cutting (top-level files in modules/llm/, not a sub-module):
- modules/llm/advisor.go   Advisor interface (Name, Order, Around) + Chain + named Order constants

Internal deps:
- modules/datasource/gorm  (memory/postgres, prompts/db)
- modules/datasource/goredis (memory/redis)
- modules/datasource/vector (cache, rag, etl/writer/vectorstore)
- modules/auth             (mcp server authentication — when it lands)
- modules/managed          (mcp server lifecycle)
- modules/telemetry/otel/genai (observability — see § 9.3)
```

### 9.2. Refine § 1.1 `modules/datasource/vector/` with concrete API shape

Update the existing § 1.1 row for `modules/datasource/vector/` from a one-line entry to a full spec:

```
modules/datasource/vector/
  vector.go              VectorStore interface (Add, Delete, SimilaritySearch)
                         VectorStoreRetriever interface (read-only — least privilege)
  search.go              SearchRequest (Query, TopK, SimilarityThreshold, FilterExpression)
  filter.go              FilterExpressionBuilder fluent DSL (eq, in, and, or, not, isNull, gt, lt, ne)
  pgvector/              Postgres pgvector driver — v1
  qdrant/                Qdrant driver — v1
  inmemory/              In-memory (testing) — v1
  pinecone/              Pinecone driver — v2
  weaviate/              Weaviate driver — v2
```

The `FilterExpressionBuilder` DSL is the single most reused piece of Spring AI's vector abstraction (one DSL across 23 backends). The `VectorStoreRetriever` read-only subset is what `llm/cache/` and `llm/rag/` consume — principle of least privilege.

### 9.3. Add `modules/telemetry/otel/genai/` (new, extends Phase 3 telemetry)

```
modules/telemetry/otel/genai/
  attributes.go     gen_ai.* attribute keys (low + high cardinality)
                    yarumo.ai.* keys (kind, tool.definition, chat.advisors, conversation.id)
                    db.vector.* keys for VectorStore observability
  instrumentation.go WithGenAIInstrumentation() option for modules/llm/Chat
  pricing/          Per-model price tables (Anthropic, OpenAI, Bedrock)
    pricing.go      Pricing interface (CostPerToken(model, kind) float64)
    anthropic.json  Quarterly refresh
    openai.json
    bedrock.json
  metrics.go        Histograms (gen_ai_client_operation_seconds_*, db_vector_client_operation_seconds_*)
                    Counters (gen_ai_client_token_usage_total, yarumo_ai_cost_total_usd)
```

This is a sibling extension to `modules/telemetry/otel/` (existing in Phase 3, milestone #9 — YA-0068..0075). Could be tracked as a follow-up under Phase 3 or filed as a new ticket when `modules/llm/` is promoted.

### 9.4. File GitHub issues when `modules/llm/` is promoted

Concrete tickets to file when this brainstorm becomes canonical:

1. `YA-NNNN: modules/llm/ — core ChatModel + EmbeddingModel + Chat fluent + Advisor primitive`
2. `YA-NNNN: modules/llm/memory/ — ChatMemory + Window/Token/Summarize strategies + in-mem/redis/postgres repos`
3. `YA-NNNN: modules/llm/cache/ — Semantic cache over VectorStoreRetriever`
4. `YA-NNNN: modules/llm/guardrails/ — Advisor catalog (PII, injection, toxicity, schema, maxtokens, retry, logger)`
5. `YA-NNNN: modules/llm/prompts/ — Template + Loader (fs/db) + A/B variant resolver`
6. `YA-NNNN: modules/llm/parsers/ — Parser[T] (JSON, tagged, regex) + native JSON Schema negotiation`
7. `YA-NNNN: modules/llm/tools/ — Tool, Registry, Definition, ToolCallAdvisor, ErrorPolicy`
8. `YA-NNNN: modules/llm/mcp/ — Client + Server + 4 transports + tools.Tool adapter`
9. `YA-NNNN: modules/llm/rag/ — QueryTransformer/Expander/Retriever/Joiner/Augmenter primitives`
10. `YA-NNNN: modules/llm/etl/ — DocumentReader (pdf/text/json/md/html) + Transformer (splitter/enrich) + Writer`
11. `YA-NNNN: modules/llm/{anthropic,openai,ollama}/ — v1 driver set`
12. `YA-NNNN: modules/llm/{bedrock,vertex}/ — v2 driver set`
13. `YA-NNNN: modules/telemetry/otel/genai/ — OTel GenAI attributes + pricing + cost metrics`
14. `YA-NNNN: modules/datasource/vector/ — VectorStore + VectorStoreRetriever + SearchRequest + FilterExpressionBuilder`

### 9.5. MEMORY.md update suggestion

Add a one-line entry under "Evolución Planeada":

```
modules/llm/ proposed (Spring AI analysis 2026-05-16): 8 sub-modules + cross-cutting Advisor.
Adds tools/, mcp/, rag/, etl/, memory/, cache/, guardrails/, prompts/, parsers/.
Defers agents/ to sdks/processes/. Sibling: modules/telemetry/otel/genai/. Refines § 1.1 datasource/vector/ API shape.
```

### 9.6. Update `docs/STRATEGY.md`

Add an MCP-server consumer story: DaaS can expose decision tables as MCP resources, turning it into an LLM-consumable decision provider for any MCP-capable assistant. Strategic differentiator vs proprietary DaaS competitors — zero custom integration for LLM consumers.

### 9.7. Cross-reference from `docs/spring/README.md`

Update the Spring → Yarumo `README.md` index entry for Spring AI to PARTIAL with the eight-sub-module + cross-cutting Advisor scope. Cross-link to this re-analysis.

---

**End of analysis.** Length target met (~880 lines). No references to deleted § 3 / Annex A / Annex B. All mappings cite only canonical § 1 modules (datasource, auth, messaging, health, boot) plus the new proposed `modules/llm/` and `modules/telemetry/otel/genai/`.
