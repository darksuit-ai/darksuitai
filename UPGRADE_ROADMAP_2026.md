# DarkSuitAI — 2026 Modernization Roadmap

*Prepared July 17, 2026. This is a plan for review — no code has been changed yet.*

## TL;DR

DarkSuitAI is a well-structured 2024-era Go agent framework, but its core design decisions predate the 2026 agent stack. The framework's own README claims it "supercedes all other agentic frameworks through its AI self-reflect action control" — that self-reflect loop is exactly the part that has been commoditized. In 2026 the differentiators are native tool calling, context/memory engineering, and observability — none of which the current codebase has.

The single highest-leverage change is **replacing text/XML ReAct parsing with providers' native tool-calling APIs**. Almost everything else (models, SDKs, memory, loop hardening) follows naturally once you're on structured tool calls.

This maps directly to the framework question from earlier: you don't need to *rewrite* onto someone else's framework, but you do need to adopt the 2026 primitives — native tool calling, context engineering, loop engineering, memory engineering — because those are now table stakes, not differentiators.

---

## 1. Current-state assessment

What the codebase does today, and where each 2024-ism lives.

| Area | Current implementation | File(s) |
|------|------------------------|---------|
| Tool calling | Text-based ReAct; model emits `<tool_call>`/`<answer>` XML that is regex/XML-parsed out of raw completion text; `\nObservation:` stop sequence | `pkg/agent/neural_gateway.go`, `pkg/agent/_chat/cortex.go` |
| LLM access | Hand-rolled HTTP clients per provider (no official SDKs) | `internal/llms/*/client.go`, `internal/llms/*/api.go` |
| Models | `gpt-4o` and Claude-3-era defaults in docs/config | `README.md`, prompt YAML |
| Memory | Raw MongoDB conversation log; retrieve last-K, string-concat into prompt | `internal/memory/mongodb/conversation_memory.go` |
| Loop control | Sequential single-tool ReAct; `for actionReady` | `pkg/agent/_chat/cortex.go` |
| Observability | `stdout` printer when `verbose=true` | `internal/utilities/printer.go` |
| Runtime | Go 1.20; `mongo-driver` v1.15.1 | `go.mod` |

### Bugs / smells found while reading (worth fixing regardless of the 2026 upgrade)

- **Global LLM is not concurrency-safe.** `runnable.go` uses a package-level `var llm llms.LLM` that every request mutates. Two concurrent agent calls will race and can clobber each other's model config. This is a correctness bug under any real load.
- **Gemini is dead code.** A full `internal/llms/gemini` provider exists but is *not wired into* the `switch` in `runnable.go` (only `openai`/`anthropic`/`groq` are). It can never be selected.
- **`RetrieveMemoryWithK` sort is broken.** `sort.Slice(chatHistory, func(i,j) { return i > j })` sorts by *index*, not timestamp — it just reverses the slice. Ordering is effectively accidental.
- **~200 lines of copy-paste.** `Basechat`, `ChatIterable`, `BaseStream`, `StreamIterable` in `runnable.go` are near-identical; only the streaming call at the end differs. The provider `switch` is duplicated four times.
- **Max-iteration limit is a TODO, not enforced.** `cortex.go` has `// TODO: allow ... determine number of iterations`; the loop has no hard cap, no no-progress detection, and no cost budget.

---

## 2. The four workstreams

### A. Native tool calling (highest impact)

**Problem.** Parsing tools out of free text is the single biggest reliability tax in the codebase. Every malformed tag, stray emoji, or "answer + tool_call in one message" is a failure the code has to detect and re-prompt around (see the error branches in `neural_gateway.go`). All four providers now expose native, typed tool calling and structured outputs; the model returns a typed tool-call object, not prose you have to scrape.

**2026 target.**
- Define tools once as JSON-schema function declarations; pass them via each SDK's native `tools` parameter.
- Consume `tool_use` / `function_call` blocks directly instead of `UnmarshalToolCall`.
- Use **structured outputs** for the final answer (typed verdict + payload) so the loop router branches on fields, not on `<answer>` tags.
- Enable **parallel tool calls** where the provider supports it (the loop currently executes one tool per turn).

**Files touched.** Retire most of `neural_gateway.go` and the XML branch of `cortex.go`; the `NewTool`/`ToolNodes` registry in `pkg/tools/basetool.go` stays but gains a JSON-schema field.

**Effort:** L · **Risk:** M (changes the core contract, but removes far more code than it adds).

---

### B. Models & SDKs

**Problem.** Custom HTTP clients mean every new API feature (structured outputs, prompt caching, context management, streaming events) has to be hand-implemented. The model IDs are two years stale.

**2026 target — adopt official SDKs and drop the hand-rolled clients:**

| Provider | Adopt SDK | Current 2026 models to default to |
|----------|-----------|-----------------------------------|
| Anthropic | `github.com/anthropics/anthropic-sdk-go` (**v1.58.0 ships in Phase 1; needs Go 1.24+**) | `claude-opus-4-8`, `claude-sonnet-5`, `claude-haiku-4-5` |
| OpenAI | `github.com/openai/openai-go` (**Responses API**) | GPT-5.x family (Sol/Terra/Luna tiers) |
| Gemini | `google.golang.org/genai` (unified SDK; old `generative-ai-go` is legacy) | `gemini-2.5-flash` and newer |
| Groq | Reuse `openai-go` with Groq base URL (OpenAI-compatible) | `openai/gpt-oss-120b`, `openai/gpt-oss-20b`, `qwen/qwen3.x` — **note `llama-3.1-8b-instant` and `llama-3.3-70b-versatile` were deprecated June 17, 2026** |

**Also:** bump `go.mod` to Go 1.22+ (required by anthropic-sdk-go), and evaluate `mongo-driver` v2.

**Effort:** L · **Risk:** M · **Payoff:** deletes `internal/llms/*/client.go` + `types/chat*.go` hand-rolled plumbing; unlocks A, C, and D for free since the SDKs expose those features natively.

---

### C. Context & memory engineering

**Problem.** Memory today is a flat conversation log concatenated into the prompt. There's no summarization, no compaction, no retrieval relevance, no separation of "working context" from "long-term memory." Long sessions will blow the context window and degrade.

**2026 target (Anthropic's published pattern; ~29% lift from context editing alone, ~39% combined with a memory tool; 84% token reduction in a 100-turn eval):**
- **Compaction** — distill the running context into a high-fidelity summary when it grows (server-side `compact_20260112` for Claude, or a summarize-and-replace pass for other providers).
- **Context editing / tool-result clearing** — drop stale tool outputs from the active window automatically.
- **Memory tool** — structured note-taking to external storage (your MongoDB layer is a natural backing store) that survives a compaction pass.
- **Retrieval** — add semantic retrieval over memory (vector index) so `RetrieveMemoryWithK` returns *relevant* history, not just the last K by (currently mis-sorted) time.

**Files touched.** Extend `internal/memory/mongodb/` into a `Memory` interface with `Working` vs `LongTerm` tiers; fix the sort bug; add a summarizer hook in the loop.

**Effort:** M–L · **Risk:** L (additive; can ship behind a flag).

---

### D. Loop & observability

**Problem.** The loop has no hard guardrails and the only observability is `stdout`. In 2026 terms this is the gap between a demo and production.

**2026 target.**
- **Loop guardrails:** enforce max iterations (finish the existing TODO), add no-progress detection (exit when iterations stop producing new info), and hard token/cost budgets.
- **Retries with backoff** on transient tool/model errors so the loop recovers instead of stalling.
- **Guardrails at three points:** input (first turn), output (final), and tool (before/after each execution).
- **Observability:** instrument with **OpenTelemetry GenAI semantic conventions** — each model call, tool execution, and reasoning step becomes a span. This is now natively ingested by Datadog/New Relic/Dynatrace with no vendor SDK.
- **Eval harness:** wire one of Promptfoo / Braintrust / LangSmith into CI so regressions are caught per-PR.

**Files touched.** `pkg/agent/_chat/cortex.go` and `_stream/cortex.go` (loop control); new `internal/telemetry/` package; `runnable.go` (retry wrapper, and fix the global-`llm` race while you're in there).

**Effort:** M · **Risk:** L.

---

## 3. Suggested sequencing

Ordered so each phase unlocks the next and risk stays contained.

1. **Phase 0 — Foundations & bug fixes (low risk, do first).** Bump to Go 1.22+, fix the global-`llm` race, fix the memory sort, wire Gemini into the switch, and collapse the four duplicated `runnable.go` methods into one. This stabilizes the base before bigger changes.
2. **Phase 1 — SDK migration (workstream B).** Swap hand-rolled clients for official SDKs, one provider at a time behind the existing `llms.LLM` interface. Update default model IDs.
3. **Phase 2 — Native tool calling (workstream A).** With SDKs in place, replace XML parsing with native tool calls + structured outputs. This is the big reliability win.
4. **Phase 3 — Loop & observability (workstream D).** Add guardrails, retries, and OTel tracing around the now-native loop.
5. **Phase 4 — Context & memory engineering (workstream C).** Layer in compaction, context editing, memory tool, and retrieval.

Each phase is independently shippable and testable; nothing after Phase 0 forces a big-bang rewrite.

---

## 4. Backward compatibility

The public API (`NewChatLLMArgs`, `NewSuitedAgent`, `NewTool`, `agent.Chat/Stream`) can be preserved throughout. The XML tool protocol can stay available behind a `SetToolProtocol("xml"|"native")` flag during Phase 2 so existing users don't break. Recommend a `v0.x → v0.(x+1)` minor bump per phase, with the SDK swap and Go 1.22 requirement flagged as the one breaking change (new major or clearly noted minimum).

---

## 5. Decisions I need from you before implementing

1. **Native vs. keep-XML-as-fallback** — clean cut to native tool calling, or dual-mode behind a flag for a release?
2. **Provider priority** — which provider to migrate first in Phase 1 (I'd suggest Anthropic, since its SDK also gives you first-party compaction/memory/context-editing for Phase 4)?
3. **Memory backing** — stay MongoDB-only, or add a vector store for semantic retrieval in Phase 4?
4. **Observability target** — OTel-only (vendor-neutral), or wire a specific platform (Datadog/LangSmith/Braintrust)?
5. **Go / driver bump** — OK to require Go 1.22+ and move to `mongo-driver` v2?

---

## Sources

- [Anthropic Go SDK (GitHub)](https://github.com/anthropics/anthropic-sdk-go) · [Claude Go SDK docs](https://platform.claude.com/docs/en/api/sdks/go) · [Claude models overview](https://platform.claude.com/docs/en/about-claude/models/overview)
- [OpenAI Go library (GitHub)](https://github.com/openai/openai-go) · [OpenAI models](https://developers.openai.com/api/docs/models) · [OpenAI Responses API in Go](https://chris.sotherden.io/openai-responses-api-using-go/)
- [Google Gen AI Go SDK (pkg.go.dev)](https://pkg.go.dev/google.golang.org/genai) · [go-genai (GitHub)](https://github.com/googleapis/go-genai)
- [Groq supported models](https://console.groq.com/docs/models) · [Groq deprecations](https://console.groq.com/docs/deprecations)
- [Effective context engineering for AI agents (Anthropic)](https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents) · [Effective harnesses for long-running agents (Anthropic)](https://www.anthropic.com/engineering/effective-harnesses-for-long-running-agents) · [Context engineering: memory, compaction, tool clearing (Claude Cookbook)](https://platform.claude.com/cookbook/tool-use-context-engineering-context-engineering-tools)
- [OpenTelemetry GenAI observability](https://opentelemetry.io/blog/2026/genai-observability/) · [Datadog GenAI semantic conventions](https://www.datadoghq.com/blog/llm-otel-semantic-convention/)
- [Loop engineering for AI agents (2026)](https://datasciencedojo.com/blog/agentic-loops-explained-from-react-to-loop-engineering-2026-guide/) · [AI agents 2026: tools, memory, evals, guardrails](https://andriifurmanets.com/blogs/ai-agents-2026-practical-architecture-tools-memory-evals-guardrails)
