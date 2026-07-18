# Changelog

All notable changes to DarkSuitAI are documented here. Format based on
[Keep a Changelog](https://keepachangelog.com/); this project aims to follow
[Semantic Versioning](https://semver.org/).

## [0.0.9] — 2026 modernization

The big one: DarkSuitAI moves from hand-rolled HTTP clients and prompt-parsing to
the official provider SDKs, native tool calling, first-class context/memory
engineering, and built-in observability — while keeping the public API stable.

### ✨ Highlights

- **Official SDKs for every provider.** Anthropic, OpenAI, Gemini, and Groq now
  run on their official Go SDKs. All the bespoke `net/http` clients, rate
  limiters, retry loops, and SSE scanners are gone.
- **Native tool calling.** Provider-side structured tool use (Anthropic), with
  the XML/ReAct protocol kept as a portable fallback — selectable per agent.
- **Context & memory engineering.** Rolling-summary compaction and semantic
  (vector) recall keep long sessions cheap and coherent.
- **Observability built in.** A tiny `Observer` interface with stdout and
  LangSmith exporters; event fields follow the OpenTelemetry GenAI semantic
  conventions (OTel adapter documented).
- **Loop guardrails.** Max-iteration caps and no-progress detection.

### Added

- `SetToolProtocol("native"|"xml")` — dual-mode tool calling; native tool use
  for Anthropic, XML/ReAct for everything else.
- `NewToolWithSchema(...)` — tools with structured (multi-argument) JSON-schema
  input for native tool calling. `NewTool` still works in both protocols.
- Observability: `SetObserver`, `NewStdoutObserver`, `NewLangSmithObserver`, and
  the `Observer` interface (see `docs/PHASE3_OBSERVABILITY.md`).
- Memory: `SetCompactor`, `NewCompactor`, in-memory and MongoDB summary stores,
  in-memory and MongoDB Atlas vector stores, `NewAnthropicSummarizer`,
  `NewHTTPEmbedder` (see `docs/PHASE4_MEMORY.md`).
- `LoadEnv()` — loads a `.env` file from the project root (walks up parents) so
  `os.Getenv` and the provider SDKs can see your keys.
- Runnable example and testing guide under `doc/`; `.env.example`.
- Contributor scaffolding: `CONTRIBUTING.md`, issue/PR templates.

### Changed

- **Requires Go 1.24+** (the Anthropic and Google Gen AI SDKs mandate it).
- Providers reimplemented on official SDKs; OpenAI and Groq now share one
  Chat Completions path (`internal/llms/openaicompat`), Groq via base-URL swap.
- Default model IDs in docs updated to 2026 models (`claude-sonnet-5`, etc.).
- README rewritten and corrected (the previous examples referenced APIs that no
  longer matched the code).

### Fixed

- **Data race:** removed a package-level `llm` variable shared across requests
  in the agent, chat, and convchat packages — concurrent calls could clobber
  each other's model config.
- **Gemini was dead code:** the provider existed but was never wired into the
  provider switch; it's now selectable.
- **Memory ordering bug:** `RetrieveMemoryWithK` sorted by slice index instead of
  timestamp; history now renders chronologically (Human → AI per turn).
- **Temperature was ignored:** the `temperature` kwarg was never read for any
  provider (agents silently ran at 0); it's now plumbed through.
- **Dropped prompt keys:** user-supplied `PromptKeys` were discarded during agent
  prompt preparation; fixed.
- Enforced the `maxIteration` limit in the XML agent loop (previously a no-op).
- Removed ~200 lines of duplicated provider-switch code.

### ⚠️ Breaking changes

- **Go 1.24+ is now required.**
- **New dependencies** are introduced (official SDKs). After upgrading, run:
  ```bash
  go mod tidy
  ```
- **Anthropic requests no longer send `temperature`.** Anthropic's 2026 models
  (e.g. `claude-sonnet-5`, `claude-opus-4.x`) deprecated the parameter and reject
  requests that include it. Temperature set via `AddModelKwargs` is ignored for
  Anthropic; other providers still honor it.

### Known limitations

- Native tool calling is Anthropic-only so far (OpenAI-compatible providers are
  next — the shared `openaicompat` layer makes this straightforward).
- Native streaming *with* tool calls isn't wired yet (`Chat` uses the native tool
  loop; `Stream` uses the XML path).
- Per-call token usage isn't yet surfaced from every SDK into the `Observer`.

### Upgrade notes

1. Upgrade to Go 1.24+.
2. `go get github.com/darksuit-ai/darksuitai@v0.0.9 && go mod tidy`
3. Put keys in `.env` and call `darksuitai.LoadEnv()` at startup (or export them).
4. If you relied on Anthropic temperature, note it's no longer sent.


[0.0.9]: https://github.com/darksuit-ai/darksuitai/releases/tag/v0.0.9
