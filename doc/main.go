// Command doc is a runnable smoke-test / usage example for DarkSuitAI.
//
// It exercises the main entry points end-to-end: a plain LLM chat, streaming,
// an agent with a tool (both the XML and native tool-calling protocols),
// telemetry via the stdout observer, and conversation compaction.
//
// Run it:
//
//	# pick a provider and set its key
//	export DARKSUIT_PROVIDER=anthropic          # anthropic | openai | gemini | groq
//	export ANTHROPIC_API_KEY=sk-ant-...         # or OPENAI_API_KEY / GEMINI_API_KEY / GROQ_API_KEY
//	go run ./doc
//
// Sections that need a capability you haven't configured (e.g. native tool
// calling requires the anthropic provider) are skipped with a note, so the
// program is safe to run with any single provider key.
package main

import (
	"fmt"
	"os"
	"strings"

	darksuitai "github.com/darksuit-ai/darksuitai"
)

// providerKey returns the configured provider and its API key from the
// environment, defaulting to anthropic.
func providerKey() (provider, key, model string) {
	provider = strings.ToLower(os.Getenv("DARKSUIT_PROVIDER"))
	if provider == "" {
		provider = "anthropic"
	}
	switch provider {
	case "openai":
		return provider, os.Getenv("OPENAI_API_KEY"), "gpt-5.6-terra"
	case "gemini":
		return provider, os.Getenv("GEMINI_API_KEY"), "gemini-2.5-flash"
	case "groq":
		return provider, os.Getenv("GROQ_API_KEY"), "openai/gpt-oss-120b"
	default:
		return "anthropic", os.Getenv("ANTHROPIC_API_KEY"), "claude-sonnet-5"
	}
}

func section(title string) {
	fmt.Printf("\n========== %s ==========\n", title)
}

func main() {
	// Load .env from the project root (walking up from the working directory)
	// so os.Getenv below can see keys defined there. Go does not read .env
	// automatically. A missing .env is fine.
	if err := darksuitai.LoadEnv(); err != nil {
		fmt.Println("warning: could not load .env:", err)
	}

	provider, apiKey, model := providerKey()
	fmt.Printf("Provider=%s  Model=%s\n", provider, model)
	if apiKey == "" {
		fmt.Printf("No API key set for %s. Set the appropriate *_API_KEY env var and re-run.\n", provider)
		os.Exit(1)
	}

	// demoChat(provider, apiKey, model)
	demoStream(provider, apiKey, model)
	demoAgentWithTool(provider, apiKey, model, "xml")
	if provider == "anthropic" {
		demoAgentWithTool(provider, apiKey, model, "native")
	} else {
		section("NATIVE TOOL CALLING")
		fmt.Println("Skipped: native tool calling currently supports the anthropic provider only.")
	}
	demoCompaction(provider, apiKey, model)
}

// demoChat: a single non-streaming completion.
func demoChat(provider, apiKey, model string) {
	section("CHAT")
	args := darksuitai.NewLLMArgs()
	args.AddAPIKey([]byte(apiKey))
	args.SetModelType(provider, model)
	args.AddModelKwargs(500, 0.7, false, []string{})

	llm, err := args.NewLLM()
	if err != nil {
		fmt.Println("NewLLM error:", err)
		return
	}
	resp, err := llm.Chat("In one sentence, what is DarkSuitAI good for?")
	if err != nil {
		fmt.Println("Chat error:", err)
		return
	}
	fmt.Println(resp)
}

// demoStream: token streaming.
func demoStream(provider, apiKey, model string) {
	section("STREAM")
	args := darksuitai.NewLLMArgs()
	args.AddAPIKey([]byte(apiKey))
	args.SetModelType(provider, model)
	args.AddModelKwargs(200, 0.7, true, []string{})

	llm, err := args.NewLLM()
	if err != nil {
		fmt.Println("NewLLM error:", err)
		return
	}
	for chunk := range llm.Stream("Count from 1 to 5, one number per line.") {
		fmt.Print(chunk)
	}
	fmt.Println()
}

// demoAgentWithTool: an agent that can call a weather tool, using either the
// "xml" (ReAct) or "native" tool-calling protocol, with stdout telemetry.
func demoAgentWithTool(provider, apiKey, model, protocol string) {
	section("AGENT + TOOL (" + strings.ToUpper(protocol) + ")")

	// Register a simple tool. NewTool tools take a single string input and work
	// in both protocols.
	weatherTool := darksuitai.NewTool(
		"getweather",
		"Get the current weather for a city. Input: the city name.",
		func(input, toolName string, meta map[string]interface{}) (string, []interface{}, error) {
			raw := fmt.Sprintf(`{"city":%q,"tempC":21,"conditions":"sunny"}`, input)
			return fmt.Sprintf("It is 21°C and sunny in %s.", input), []interface{}{raw}, nil
		},
	)
	darksuitai.ToolNodes = append(darksuitai.ToolNodes, weatherTool)

	args := darksuitai.NewLLMArgs()
	args.AddAPIKey([]byte(apiKey))
	args.SetModelType(provider, model)
	args.AddModelKwargs(1000, 0.5, false, []string{"\nObservation:"})
	args.SetToolProtocol(protocol)
	// Telemetry: prints one JSON event per line to stderr.
	args.SetObserver(darksuitai.NewStdoutObserver())

	agent, err := args.NewSuitedAgent()
	if err != nil {
		fmt.Println("NewSuitedAgent error:", err)
		return
	}
	// maxIterations=4, a session id, verbose=true.
	if err := agent.Program(4, "demo-session", true); err != nil {
		fmt.Println("Program error:", err)
		return
	}
	resp, toolData, err := agent.Chat("What's the weather in Lagos?")
	if err != nil {
		fmt.Println("Chat error:", err)
		return
	}
	fmt.Println("\nAnswer:", resp)
	fmt.Printf("Tool data: %v\n", toolData)
}

// demoCompaction: shows wiring conversation compaction (rolling summary +
// recent turns). Uses an in-memory summary store and an Anthropic summarizer.
func demoCompaction(provider, apiKey, model string) {
	section("COMPACTION SETUP")

	// The summarizer uses Anthropic; only wire it when an Anthropic key exists.
	anthKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthKey == "" {
		fmt.Println("Skipped: set ANTHROPIC_API_KEY to enable the summarizer.")
		return
	}

	store := darksuitai.NewInMemorySummaryStore() // or NewMongoSummaryStore(collection)
	summarizer := darksuitai.NewAnthropicSummarizer(anthKey, "claude-haiku-4-5")
	compactor := darksuitai.NewCompactor(store, summarizer, darksuitai.CompactorConfig{
		MaxTurns:   20,
		KeepRecent: 6,
	})

	args := darksuitai.NewLLMArgs()
	args.AddAPIKey([]byte(apiKey))
	args.SetModelType(provider, model)
	args.AddModelKwargs(1000, 0.5, false, []string{})
	args.SetCompactor(compactor) // agent now injects compacted history when it has a session + Mongo memory

	if _, err := args.NewSuitedAgent(); err != nil {
		fmt.Println("NewSuitedAgent error:", err)
		return
	}
	fmt.Println("Compactor configured. On long sessions the agent will inject a rolling")
	fmt.Println("summary plus recent turns instead of the full raw transcript.")
}
