# Testing DarkSuitAI

This folder contains a runnable smoke test (`main.go`) and the instructions
below for verifying the library.

## 1. Resolve dependencies (once)

The 2026 modernization pulls in official SDKs (Anthropic, Google Gen AI,
OpenAI). Resolve them and write `go.sum`:

```bash
go mod tidy
```

> Requires **Go 1.24+** (the Anthropic and Gen AI SDKs mandate it) and network
> access to the Go module proxy.

## 2. Static checks (no API keys, no network calls)

```bash
go build ./...   # everything compiles
go vet ./...     # static analysis
```

## 3. Unit tests (no API keys)

The memory/context-engineering core ships with real unit tests (cosine
similarity, in-memory vector store, and the compaction state machine):

```bash
go test ./...            # run all tests
go test -v ./internal/memory/   # just the memory tests, verbose
```

These are hermetic — they use fakes/in-memory stores and never call a provider.

## 4. Live smoke test (needs one provider API key)

`main.go` exercises the real entry points end-to-end.

### Option A — a `.env` file in the project root (recommended)

Go does **not** read `.env` automatically. DarkSuitAI provides
`darksuitai.LoadEnv()`, which the example calls at startup; it walks up from the
working directory to find `.env`, so it works from the root or any subdirectory.

```bash
cp .env.example .env      # then edit .env and fill in your key(s)
go run ./doc
```

Your app does the same — call it once at startup:

```go
func main() {
    _ = darksuitai.LoadEnv()            // loads project-root .env into the environment
    key := os.Getenv("ANTHROPIC_API_KEY")
    args := darksuitai.NewLLMArgs()
    args.AddAPIKey([]byte(key))
    // ...
}
```

### Option B — export the variables yourself

```bash
export DARKSUIT_PROVIDER=anthropic         # anthropic | openai | gemini | groq
export ANTHROPIC_API_KEY=sk-ant-...        # or OPENAI_API_KEY / GEMINI_API_KEY / GROQ_API_KEY
go run ./doc
```

It runs, in order:

1. **Chat** — a single non-streaming completion.
2. **Stream** — token streaming.
3. **Agent + tool (XML)** — a ReAct agent that calls a `get_weather` tool, with
   stdout telemetry (JSON events on stderr).
4. **Agent + tool (native)** — the same, using provider-side structured tool
   calling (Anthropic only; skipped otherwise).
5. **Compaction setup** — wires a rolling-summary compactor (needs
   `ANTHROPIC_API_KEY` for the summarizer).

Model defaults per provider (override in `main.go` if needed):

| `DARKSUIT_PROVIDER` | Default model | Key env var |
|---------------------|---------------|-------------|
| `anthropic`         | `claude-sonnet-5` | `ANTHROPIC_API_KEY` |
| `openai`            | `gpt-5.6-terra`   | `OPENAI_API_KEY` |
| `gemini`            | `gemini-2.5-flash`| `GEMINI_API_KEY` |
| `groq`              | `openai/gpt-oss-120b` | `GROQ_API_KEY` |

## 5. Optional: persistent memory (MongoDB)

Chat memory, rolling summaries, and vector recall persist to MongoDB. To try
them, provide a connection and swap the in-memory stores for the Mongo ones:

```go
db := darksuitai.NewMongoChatMemory(mongoURI, "mydb")     // *mongo.Collection
args.SetMongoDBChatMemory(db)

// compaction persisted to Mongo instead of memory:
store := darksuitai.NewMongoSummaryStore(summaryCollection)

// semantic recall (requires an Atlas Vector Search index — see docs/PHASE4_MEMORY.md):
vs := darksuitai.NewMongoVectorStore(memCollection, "vector_index")
```

## Verifying without network (maintainers)

If the module proxy is unavailable, the packages can be type-checked against
local copies of the SDKs using `replace` directives pointing at cloned repos or
minimal stubs. This is how the modernization was validated; see the notes in
`docs/` for the per-phase verification approach.

## Where to look

- `docs/PHASE3_OBSERVABILITY.md` — telemetry, guardrails, OpenTelemetry adapter.
- `docs/PHASE4_MEMORY.md` — compaction and vector memory.
- `UPGRADE_ROADMAP_2026.md` — the full modernization plan and rationale.
