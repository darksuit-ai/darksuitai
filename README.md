<div align="center">

# 🕵️ DarkSuitAI

### Production-ready AI agents in **pure Go** — no Python sidecar, no bloat.

Native tool calling · context & memory engineering · built-in observability · Anthropic · OpenAI · Gemini · Groq

[![Go Reference](https://pkg.go.dev/badge/github.com/darksuit-ai/darksuitai.svg)](https://pkg.go.dev/github.com/darksuit-ai/darksuitai)
[![Go Report Card](https://goreportcard.com/badge/github.com/darksuit-ai/darksuitai)](https://goreportcard.com/report/github.com/darksuit-ai/darksuitai)
[![Go 1.24+](https://img.shields.io/badge/go-1.24%2B-00ADD8?logo=go&logoColor=white&style=flat-square)](https://go.dev/)
[![Stars](https://img.shields.io/github/stars/darksuit-ai/darksuitai?style=flat-square)](https://star-history.com/#darksuit-ai/darksuitai)
[![License](https://img.shields.io/github/license/darksuit-ai/darksuitai?style=flat-square)](./LICENSE)

<!-- Generate this GIF with: vhs doc/demo.tape  (see doc/demo.tape) -->
<img src="./doc/demo.gif" alt="DarkSuitAI demo" width="720">

</div>

---

## Why DarkSuitAI?

Almost every AI-agent framework is Python. If your backend is Go, that means running (and babysitting) a separate Python service just to give your app agentic behavior.

**DarkSuitAI keeps agents in the same language as your backend.** It ships as a single dependency, compiles to a single binary, and gives you the modern agent stack without the abstraction tax:

- 🧠 **Native tool calling** — provider-side structured tool use (not fragile prompt-parsing), with an XML/ReAct fallback for portability.
- ♻️ **Context & memory engineering** — rolling-summary compaction and semantic (vector) recall so long sessions stay cheap and coherent.
- 🔭 **Observability built in** — run/LLM/tool telemetry via a tiny `Observer` interface, with stdout and LangSmith exporters out of the box (OpenTelemetry adapter documented).
- 🔌 **Multi-provider** — Anthropic, OpenAI, Gemini, and Groq behind one interface, all on official SDKs.
- 🛡️ **Loop guardrails** — max-iteration caps and no-progress detection so agents can't spin forever.
- ⚡ **Streaming** everywhere — chat, conversational chat, and agents.

> Deliberately lean: loop, context, and memory engineering are first-class primitives — not buried under layers of abstraction.

## Install

```bash
go get github.com/darksuit-ai/darksuitai@latest
```

Requires **Go 1.24+**. Set your provider key in a `.env` file (see [`.env.example`](./.env.example)); DarkSuitAI can load it for you.

## Quickstart (chat in ~15 lines)

```go
package main

import (
	"fmt"
	"os"

	darksuitai "github.com/darksuit-ai/darksuitai"
)

func main() {
	_ = darksuitai.LoadEnv() // loads .env from your project root

	args := darksuitai.NewLLMArgs()
	args.AddAPIKey([]byte(os.Getenv("ANTHROPIC_API_KEY")))
	args.SetModelType("anthropic", "claude-sonnet-5")
	args.AddModelKwargs(500, 0.7, false, nil)

	llm, err := args.NewLLM()
	if err != nil {
		panic(err)
	}

	resp, err := llm.Chat("In one sentence, what is DarkSuitAI good for?")
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
}
```

Streaming is one call away:

```go
for chunk := range llm.Stream("Count from 1 to 5.") {
	fmt.Print(chunk)
}
```

## Build an agent with a tool

```go
// 1. Define a tool. NewTool tools take a single string input and work in both
//    the XML and native tool-calling protocols.
weather := darksuitai.NewTool(
	"get_weather",
	"Get the current weather for a city. Input: the city name.",
	func(input, toolName string, meta map[string]interface{}) (string, []interface{}, error) {
		return fmt.Sprintf("It is 21°C and sunny in %s.", input), nil, nil
	},
)
darksuitai.ToolNodes = append(darksuitai.ToolNodes, weather)

// 2. Configure the agent.
args := darksuitai.NewLLMArgs()
args.AddAPIKey([]byte(os.Getenv("ANTHROPIC_API_KEY")))
args.SetModelType("anthropic", "claude-sonnet-5")
args.AddModelKwargs(1000, 0.5, false, nil)
args.SetToolProtocol("native")                    // provider-side tool calling (Anthropic)
args.SetObserver(darksuitai.NewStdoutObserver())  // telemetry to stderr

agent, err := args.NewSuitedAgent()
if err != nil {
	panic(err)
}
if err := agent.Program(4 /*max iterations*/, "session-1", true /*verbose*/); err != nil {
	panic(err)
}

answer, toolData, err := agent.Chat("What's the weather in Lagos?")
fmt.Println(answer, toolData, err)
```

Need structured (multi-argument) tools? Use [`NewToolWithSchema`](https://pkg.go.dev/github.com/darksuit-ai/darksuitai#NewToolWithSchema).

## How it compares

| | DarkSuitAI | Typical Python frameworks |
|---|---|---|
| Language | **Go** (single binary) | Python (separate service/runtime) |
| Tool calling | Native + XML fallback | Native |
| Providers | Anthropic, OpenAI, Gemini, Groq (official SDKs) | Varies |
| Context engineering | Compaction + vector recall built in | Add-on / DIY |
| Observability | Built-in (`Observer`, stdout/LangSmith, OTel adapter) | Add-on |
| Loop guardrails | Built-in (max-iter, no-progress) | DIY |
| Footprint | Lean, no heavy abstraction layer | Heavier |

*Different tools for different stacks — DarkSuitAI is the pragmatic choice when your services are already in Go.*

## Providers & 2026 models

| Provider | `SetModelType` key | Example model | Key env var |
|----------|--------------------|---------------|-------------|
| Anthropic | `anthropic` | `claude-sonnet-5` | `ANTHROPIC_API_KEY` |
| OpenAI | `openai` | `gpt-5.6-terra` | `OPENAI_API_KEY` |
| Gemini | `gemini` | `gemini-2.5-flash` | `GEMINI_API_KEY` |
| Groq | `groq` | `openai/gpt-oss-120b` | `GROQ_API_KEY` |

## Context & memory engineering

Long conversations blow the context window. DarkSuitAI can fold older turns into a rolling summary and keep recent turns verbatim, and it can retrieve relevant memories semantically:

```go
store := darksuitai.NewInMemorySummaryStore() // or NewMongoSummaryStore(collection)
summarizer := darksuitai.NewAnthropicSummarizer(os.Getenv("ANTHROPIC_API_KEY"), "claude-haiku-4-5")
args.SetCompactor(darksuitai.NewCompactor(store, summarizer, darksuitai.CompactorConfig{
	MaxTurns: 20, KeepRecent: 6,
}))
```

Full guide: [`docs/PHASE4_MEMORY.md`](./docs/PHASE4_MEMORY.md).

## Observability

Every run, model call, and tool execution is reported through a one-method `Observer` interface, with event fields following the OpenTelemetry GenAI semantic conventions:

```go
args.SetObserver(darksuitai.NewStdoutObserver()) // JSON events to stderr
// or ship traces to LangSmith:
args.SetObserver(darksuitai.NewLangSmithObserver(darksuitai.LangSmithConfig{
	APIKey: os.Getenv("LANGSMITH_API_KEY"), Project: "darksuitai",
}))
```

OpenTelemetry adapter + guardrails: [`docs/PHASE3_OBSERVABILITY.md`](./docs/PHASE3_OBSERVABILITY.md).

## Run the example

A full, runnable smoke test lives in [`doc/`](./doc):

```bash
cp .env.example .env   # add your key
go run ./doc
```

See [`doc/README.md`](./doc/README.md) for the testing guide (`go build`, `go vet`, `go test ./...`).

## Documentation

- [Observability & guardrails](./docs/PHASE3_OBSERVABILITY.md)
- [Context & memory engineering](./docs/PHASE4_MEMORY.md)
- [API reference (pkg.go.dev)](https://pkg.go.dev/github.com/darksuit-ai/darksuitai)

## Contributing

Contributions are very welcome — features, docs, providers, examples. See [CONTRIBUTING.md](./CONTRIBUTING.md) and the [good first issues](https://github.com/darksuit-ai/darksuitai/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).

## License

See [LICENSE](./LICENSE).
