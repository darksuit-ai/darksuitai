package darksuitai

import (
	"fmt"
	"github.com/darksuit-ai/darksuitai/internal"
	anthropicllm "github.com/darksuit-ai/darksuitai/internal/llms/anthropic"
	"github.com/darksuit-ai/darksuitai/internal/memory"
	"github.com/darksuit-ai/darksuitai/internal/memory/embed"
	"github.com/darksuit-ai/darksuitai/internal/memory/mongodb"
	"github.com/darksuit-ai/darksuitai/internal/observability"
	"github.com/darksuit-ai/darksuitai/pkg/agent"
	"github.com/darksuit-ai/darksuitai/pkg/agent/_chat"
	"github.com/darksuit-ai/darksuitai/pkg/agent/_stream"
	ai "github.com/darksuit-ai/darksuitai/pkg/chat"
	convai "github.com/darksuit-ai/darksuitai/pkg/convchat"
	"github.com/darksuit-ai/darksuitai/pkg/tools"
	"github.com/darksuit-ai/darksuitai/types"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
	"sync"
)

// Observability re-exports (Phase 3). These let callers configure telemetry
// without importing the internal package directly.
type (
	// Observer receives run/LLM/tool telemetry events. Implement this interface
	// to send telemetry to any backend (e.g. OpenTelemetry); see
	// docs/PHASE3_OBSERVABILITY.md for a ready-to-use OpenTelemetry adapter.
	Observer = observability.Observer
	// LangSmithConfig configures the built-in LangSmith observer.
	LangSmithConfig = observability.LangSmithConfig

	// The following are re-exported so custom Observers (e.g. an OpenTelemetry
	// adapter) can be implemented without importing the internal package.
	ObserverRunInfo   = observability.RunInfo
	ObserverRunHandle = observability.RunHandle
	ObserverLLMCall   = observability.LLMCall
	ObserverToolCall  = observability.ToolCall
)

// NewStdoutObserver returns an Observer that prints one JSON event per line to
// stderr (attribute keys follow the OpenTelemetry GenAI semantic conventions).
func NewStdoutObserver() Observer { return observability.Stdout{} }

// NewLangSmithObserver returns an Observer that posts each completed agent run
// to LangSmith over its REST API.
func NewLangSmithObserver(cfg LangSmithConfig) Observer {
	return observability.NewLangSmith(cfg)
}

// Memory / context engineering re-exports (Phase 4).
type (
	// Compactor turns a long conversation into a rolling summary + recent turns.
	Compactor = memory.Compactor
	// CompactorConfig tunes compaction thresholds.
	CompactorConfig = memory.CompactorConfig
	// Summarizer condenses older turns into a running summary.
	Summarizer = memory.Summarizer
	// Embedder converts text to vector embeddings for semantic recall.
	Embedder = memory.Embedder
	// VectorStore persists and retrieves text embeddings.
	VectorStore = memory.VectorStore
	// SummaryStore persists a session's rolling summary.
	SummaryStore = memory.SummaryStore
	// MemoryTurn is a single Human/AI exchange.
	MemoryTurn = memory.Turn
	// MemoryHit is a semantic-search result.
	MemoryHit = memory.Hit
)

// NewCompactor builds a conversation compactor from a summary store and summarizer.
func NewCompactor(store SummaryStore, summarizer Summarizer, cfg CompactorConfig) *Compactor {
	return memory.NewCompactor(store, summarizer, cfg)
}

// NewInMemorySummaryStore returns a process-local summary store (tests/single-node).
func NewInMemorySummaryStore() SummaryStore { return memory.NewInMemorySummaryStore() }

// NewMongoSummaryStore returns a MongoDB-backed summary store.
func NewMongoSummaryStore(collection *mongo.Collection) SummaryStore {
	return mongodb.NewMongoSummaryStore(collection)
}

// NewInMemoryVectorStore returns a cosine-ranked in-memory vector store.
func NewInMemoryVectorStore() VectorStore { return memory.NewInMemoryVectorStore() }

// NewMongoVectorStore returns a MongoDB Atlas Vector Search-backed vector store.
func NewMongoVectorStore(collection *mongo.Collection, indexName string) VectorStore {
	return mongodb.NewMongoVectorStore(collection, indexName)
}

// NewAnthropicSummarizer returns a Summarizer backed by the Anthropic SDK.
func NewAnthropicSummarizer(apiKey, model string) Summarizer {
	return anthropicllm.NewSummarizer(apiKey, model)
}

// NewHTTPEmbedder returns an OpenAI-compatible HTTP embedder.
func NewHTTPEmbedder(apiKey, model string) Embedder { return embed.NewHTTPEmbedder(apiKey, model) }

// Create an instance of the DarkSuitAgent interface
var darkSuitAgent internal.DarkSuitAgent = internal.NewDarkSuitAgent()

/*
NewTool creates a new instance of BaseTool with the specified name, description, and tool function.
The tool function is a callback that takes a string and a slice of interfaces as input and returns a string and a slice of interfaces.

Example usage:

	myTool := darksuitAI.NewTool("exampleTool", "This is an example tool",
			func(input string,string, data []interface{}) (string, []interface{},error) {
				// Your tool logic here
				return input, data, nil
			})

	darksuitAI.ToolNodes = append(darksuitAI.ToolNodes,myTool)

	fmt.Printf("all tools created: %v",darksuitAI.ToolNodes) // to see all your tools
*/
func NewTool(name string, description string, toolFunc func(string, string, map[string]interface{}) (string, []interface{}, error)) tools.BaseTool {
	return tools.BaseTool{
		Name:        name,
		Description: description,
		ToolFunc:    toolFunc,
	}
}

/*
NewToolWithSchema creates a tool that declares a structured JSON-schema input,
for use with native (provider-side) tool calling (ToolProtocol "native").

properties is the JSON-schema "properties" object and required lists the
required property names. When the model calls the tool, the raw JSON arguments
are passed to toolFunc as a JSON string (the tool author unmarshals them).

Tools created with NewTool (no schema) continue to work in both protocols: they
are exposed natively as a single required string property named "input".

Example:

	weatherTool := darksuitai.NewToolWithSchema(
		"get_weather",
		"Get the weather at a location",
		map[string]any{
			"location": map[string]any{"type": "string", "description": "City name"},
		},
		[]string{"location"},
		func(inputJSON string, toolName string, meta map[string]interface{}) (string, []interface{}, error) {
			// unmarshal inputJSON, call your API, return the result
			return "sunny, 68F", nil, nil
		},
	)
*/
func NewToolWithSchema(name string, description string, properties map[string]any, required []string, toolFunc func(string, string, map[string]interface{}) (string, []interface{}, error)) tools.BaseTool {
	return tools.BaseTool{
		Name:        name,
		Description: description,
		ToolFunc:    toolFunc,
		InputSchema: properties,
		Required:    required,
	}
}

// ToolNodes is a slice that holds all registered tools, allowing them to be accessed by their indices.
var ToolNodes = tools.ToolNodes

/*
ToolNodesMeta is a variable that holds metadata for all registered tools.
This is useful when you need to pass extra data to the logic of a tool from other systems
*/
var ToolNodesMeta = tools.ToolNodesMeta

// GoogleSearch is a premade tool provided by the framework from the tools package.
var GoogleSearch = tools.GoogleTool

func NewMongoChatMemory(databaseURI, databaseName string) *mongo.Collection {
	return mongodb.MongoChatMemory(databaseName, databaseURI)
}

// DarkSuitAI is the main struct that users will interact with
type LLMArgs types.LLMArgs

// NewLLMArgs creates a new LLMArgs with default values
func NewLLMArgs() *LLMArgs {

	return &LLMArgs{
		ChatSystemInstruction: nil,
		ChatInstruction:       nil,
		MongoDB:               nil,
		PromptKeys:            make(map[string][]byte),
		ModelType:             make(map[string]string),
		ModelKwargs: []struct {
			MaxTokens     int      `json:"max_tokens"`
			Temperature   float64  `json:"temperature"`
			Stream        bool     `json:"stream"`
			StopSequences []string `json:"stop_sequences"`
		}{
			{
				MaxTokens:     500,
				Temperature:   0.2,
				Stream:        false,
				StopSequences: []string{},
			},
		},
		APIKey:       nil,
		ToolProtocol: "xml",
	}
}

/*
	AddAPIKey sets the API key for the LLMArgs instance.

This method allows you to securely store the API key required for authenticating requests to the chat model service.

Example:

args := darksuitAI.NewLLMArgs()

args.AddAPIKey([]byte("your-api-key"))

In this example, the byte slice containing the API key is set, enabling the chat model to authenticate and process requests.
*/
func (args *LLMArgs) AddAPIKey(apiKey []byte) {
	args.APIKey = apiKey
}

/*
	SetChatSystemInstruction sets the system-level instruction in LLMArgs.

This method allows you to define the overarching system prompt that will guide the chat model's behavior.

Example:

args := darksuitAI.NewLLMArgs()

args.SetChatSystemInstruction([]byte("Your system prompt goes here"))

In this example, the byte slice containing the system prompt is set, which will be used by the chat model to maintain context and behavior.
*/
func (args *LLMArgs) SetChatSystemInstruction(systemPrompt []byte) {
	args.ChatSystemInstruction = systemPrompt
}

/*
	SetChatInstruction sets the chat instruction in LLMArgs.

This method allows you to define the main instruction or prompt that will guide the chat model's responses.

Example:

args := darksuitAI.NewLLMArgs()

args.SetChatInstruction([]byte("Your chat instruction goes here"))

In this example, the byte slice containing the chat instruction is set, which will be used by the chat model to generate responses.
*/
func (args *LLMArgs) SetChatInstruction(prompt []byte) {
	args.ChatInstruction = prompt
}

/*
	AddPromptKey adds a key-value pair to the PromptKeys map in LLMArgs.

This method allows you to dynamically insert or update prompt-specific variables that can be used within the chat instruction template.

Example:

args := darksuitAI.NewLLMArgs()

args.AddPromptKey("year", []byte(`2024`))

args.AddPromptKey("month", []byte(`June`))

In this example, the keys "year" and "month" with their respective values "2024" and "June" are added to the PromptKeys map, which can then be referenced in the chat instruction template.
*/
func (args *LLMArgs) AddPromptKey(key string, value []byte) {
	args.PromptKeys[key] = value
}

/*
	SetModelType sets a key-value pair in the ModelType map in LLMArgs.

This method allows you to specify the type of model to be used for the chat.

Example:

args := darksuitAI.NewLLMArgs()

args.SetModelType("openai", "gpt-4o")

In this example, the key "openai" with the value "gpt-4o" is added to the ModelType map, indicating the model type to be used.
*/
func (args *LLMArgs) SetModelType(key, value string) {
	args.ModelType[key] = value
}

/*
SetToolProtocol selects how the agent invokes tools.

  - "xml"    (default): the legacy ReAct/XML protocol, supported by all providers.
  - "native": provider-side structured tool calling (currently Anthropic only).
    For non-Anthropic providers the agent automatically falls back to "xml".

Example:

	args := darksuitai.NewLLMArgs()
	args.SetModelType("anthropic", "claude-sonnet-5")
	args.SetToolProtocol("native")
*/
func (args *LLMArgs) SetToolProtocol(protocol string) {
	args.ToolProtocol = protocol
}

/*
SetObserver attaches a telemetry Observer to the agent. Run, LLM and tool events
are reported to it during Chat. Use NewStdoutObserver for local debugging or
NewLangSmithObserver to ship runs to LangSmith; implement Observer yourself to
integrate OpenTelemetry (see docs/PHASE3_OBSERVABILITY.md).

Example:

	args := darksuitai.NewLLMArgs()
	args.SetObserver(darksuitai.NewLangSmithObserver(darksuitai.LangSmithConfig{
		APIKey:  "ls-...",
		Project: "darksuitai-prod",
	}))
*/
func (args *LLMArgs) SetObserver(observer Observer) {
	args.Observer = observer
}

/*
SetCompactor enables conversation compaction: instead of injecting the raw
last-K transcript, the agent injects a rolling summary of older turns plus the
most recent turns verbatim. This keeps the context window small on long sessions
without losing important detail.

Example:

	store := darksuitai.NewMongoSummaryStore(summaryCollection)
	summarizer := darksuitai.NewAnthropicSummarizer(apiKey, "claude-haiku-4-5")
	args.SetCompactor(darksuitai.NewCompactor(store, summarizer, darksuitai.CompactorConfig{
		MaxTurns:   20,
		KeepRecent: 6,
	}))
*/
func (args *LLMArgs) SetCompactor(compactor *Compactor) {
	args.Compactor = compactor
}

/*
	SetMongoDBChatMemory sets the MongoDB collection in LLMArgs.

This method allows you to set MongoDB that will be used for storing and retrieving chat-related data.

Example:

args := darksuitAI.NewLLMArgs()

args.SetMongoDBChatMemory(mongoCollection)

In this example, the MongoDB ChatMemory is set, which will be used for chat history operations.
*/
func (args *LLMArgs) SetMongoDBChatMemory(collection *mongo.Collection) {
	args.MongoDB = collection
}

/*
	AddModelKwargs adds a new set of model arguments to the ModelKwargs slice in LLMArgs.

This method allows you to specify various parameters for the model's behavior.

Example:

args := darksuitAI.NewLLMArgs()

args.AddModelKwargs(500, 0.8, true, []string{"Human:"})

In this example, the model arguments are set with a maximum of 1500 tokens, a temperature of 0.8, streaming enabled, and a stop sequence of "Human:".
*/
func (args *LLMArgs) AddModelKwargs(maxTokens int, temperature float64, stream bool, stopSequences []string) {
	args.ModelKwargs = append(args.ModelKwargs, struct {
		MaxTokens     int      `json:"max_tokens"`
		Temperature   float64  `json:"temperature"`
		Stream        bool     `json:"stream"`
		StopSequences []string `json:"stop_sequences"`
	}{
		MaxTokens:     maxTokens,
		Temperature:   temperature,
		Stream:        stream,
		StopSequences: stopSequences,
	})
}

type LLM struct {
	ai ai.AI
}

type ConvLLM struct {
	convai convai.ConvAI
}

type AgentSynapse struct {
	synapse                agent.Synapse
	_chatAgentPreProgram   _chat.AgentPreProgram
	_streamAgentPreProgram _stream.AgentPreProgram
}

// NewLLM creates a new instance of DarkSuitAI LLM
func (cargs *LLMArgs) NewLLM() (*LLM, error) {

	return &LLM{
		ai: ai.AI{
			ChatSystemInstruction: cargs.ChatSystemInstruction,
			ChatInstruction:       cargs.ChatInstruction,
			PromptKeys:            cargs.PromptKeys,
			ModelType:             cargs.ModelType,
			ModelKwargs:           cargs.ModelKwargs,
			APIKey:                cargs.APIKey,
		},
	}, nil
}

// NewConvLLM creates a new instance of DarkSuitAI LLM
func (cargs *LLMArgs) NewConvLLM() (*ConvLLM, error) {

	return &ConvLLM{
		convai: convai.ConvAI{
			ChatSystemInstruction: cargs.ChatSystemInstruction,
			ChatInstruction:       cargs.ChatInstruction,
			PromptKeys:            cargs.PromptKeys,
			ModelType:             cargs.ModelType,
			MongoDB:               cargs.MongoDB,
			ModelKwargs:           cargs.ModelKwargs,
			APIKey:                cargs.APIKey,
		},
	}, nil
}

// NewSuitedAgent creates a new instance of DarkSuitAI Agent
func (cargs *LLMArgs) NewSuitedAgent() (*AgentSynapse, error) {

	return &AgentSynapse{
		synapse: agent.Synapse{
			SystemPrompt:          cargs.ChatSystemInstruction,
			ChatInstructionPrompt: cargs.ChatInstruction,
			PromptKeys:            cargs.PromptKeys,
			ModelType:             cargs.ModelType,
			ToolNodes:             ToolNodes,
			MongoDB:               cargs.MongoDB,
			ModelKwargs:           cargs.ModelKwargs,
			APIKey:                cargs.APIKey,
			ToolProtocol:          cargs.ToolProtocol,
			Observer:              cargs.Observer,
			Compactor:             cargs.Compactor,
		},
	}, nil
}

func (a *AgentSynapse) Program(maxIteration int, sessionId string, verbose bool) error {
	if verbose {
		// Ensure the dark suit callback runs only once
		var once sync.Once
		once.Do(func() {
			darkSuitCallback := darkSuitAgent.WakeDarkSuitAgent()
			darkSuitCallback()
		})
	}
	var promptAgent agent.PromptAgentInterface = agent.NewPromptAgent()

	// Capture native tool-calling configuration before the ReAct/XML template
	// overwrites the system prompt below. Native mode uses the raw system
	// instruction (which may be nil) rather than the XML-rendered one.
	rawSystem := a.synapse.SystemPrompt
	var provider, model string
	for p, m := range a.synapse.ModelType {
		provider, model = p, m
	}
	var maxTokens int
	var temperature float64
	if n := len(a.synapse.ModelKwargs); n > 0 {
		kw := a.synapse.ModelKwargs[n-1]
		maxTokens, temperature = kw.MaxTokens, kw.Temperature
	}

	basePrompt, sysPrompt, tools, toolNames, err := promptAgent.PreparePrompt(a.synapse.SystemPrompt,
		a.synapse.ChatInstructionPrompt, a.synapse.ToolNodes, a.synapse.PromptKeys, a.synapse.MongoDB, sessionId, a.synapse.Compactor)
	if err != nil {
		return fmt.Errorf("failed to prepare prompt: %w", err)
	}
	a.synapse.SystemPrompt = sysPrompt
	a.synapse.ChatInstructionPrompt = basePrompt
	a._chatAgentPreProgram = _chat.AgentPreProgram{
		BasePrompt:           basePrompt,
		SystemPrompt:         sysPrompt,
		Tools:                tools,
		ToolNames:            toolNames,
		AdditionalToolsMeta:  ToolNodesMeta,
		BaseRunnableCaller:   a.synapse.Basechat,
		RunnableCaller:       a.synapse.ChatIterable,
		MaxIteration:         maxIteration,
		ChatMemoryCollection: a.synapse.MongoDB,
		Verbose:              verbose,
		SessionId:            sessionId,
		ToolProtocol:         a.synapse.ToolProtocol,
		Provider:             provider,
		Model:                model,
		APIKey:               a.synapse.APIKey,
		MaxTokens:            maxTokens,
		Temperature:          temperature,
		RawSystemPrompt:      rawSystem,
		Observer:             a.synapse.Observer,
	}
	a._streamAgentPreProgram = _stream.AgentPreProgram{
		BasePrompt:           basePrompt,
		SystemPrompt:         sysPrompt,
		Tools:                tools,
		ToolNames:            toolNames,
		AdditionalToolsMeta:  ToolNodesMeta,
		BaseRunnableCaller:   a.synapse.BaseStream,
		RunnableCaller:       a.synapse.StreamIterable,
		MaxIteration:         maxIteration,
		ChatMemoryCollection: a.synapse.MongoDB,
		Verbose:              verbose,
		SessionId:            sessionId,
	}
	return nil
}

// Chat processes the input query and returns the response.
// It optionally triggers a callback if verbose mode is enabled.
//
// Parameters:
//   - input: The user's input query as a string.
//   - sessionId: A string representing the session identifier.
//
// Returns:
//   - A string containing the response from the agent.
//   - An interface containing any additional tool data.
//   - An error if the execution fails.
func (a *AgentSynapse) Chat(input string) (string, any, error) {

	query := map[string][]byte{"question": []byte(input)}

	// Route to native (provider-side) tool calling when requested and available.
	// Native tool calling is currently implemented for Anthropic only; every
	// other provider transparently falls back to the XML/ReAct executor.
	var (
		response []byte
		toolData any
		err      error
	)
	if a._chatAgentPreProgram.ToolProtocol == "native" && a._chatAgentPreProgram.Provider == "anthropic" {
		response, toolData, err = a._chatAgentPreProgram.NativeExecutor(query, a._chatAgentPreProgram.SessionId,
			a._chatAgentPreProgram.MaxIteration, a._chatAgentPreProgram.Verbose)
	} else {
		response, toolData, err = a._chatAgentPreProgram.Executor(query, a._chatAgentPreProgram.SessionId,
			a._chatAgentPreProgram.MaxIteration, a._chatAgentPreProgram.Verbose)
	}
	if err != nil {
		return "", nil, err
	}
	return string(response), toolData, nil
}

func NewStreamWriter() *_stream.StreamWriter {
	return &_stream.StreamWriter{
		Builder: &strings.Builder{},
		Ch:      make(chan string, 100),
		Done:    make(chan struct{}),
	}
}

func (a *AgentSynapse) Stream(input string) (<-chan string, error) {
	streamWriter := NewStreamWriter()
	outputChan := make(chan string)
	var builder strings.Builder
	// Add to wait group before starting goroutines
	streamWriter.Wg.Add(2)

	// Start the streaming goroutine
	go func() {
		defer streamWriter.Wg.Done() // Mark this goroutine as done when it exits
		defer streamWriter.Close()   // This will now safely close only once

		err := a._streamAgentPreProgram.StreamExecutor(
			map[string][]byte{"question": []byte(input)},
			streamWriter,
			a._streamAgentPreProgram.MaxIteration,
			a._streamAgentPreProgram.Verbose,
		)

		if err != nil {
			select {
			case outputChan <- fmt.Sprintf("Error: %v", err):
			case <-streamWriter.Done:
			}
		}
	}()

	// Start the forwarding goroutine
	go func() {
		defer streamWriter.Wg.Done()
		defer close(outputChan)

		for {
			select {
			case chunk, ok := <-streamWriter.Ch:
				if !ok {
					return
				}
				builder.WriteString(chunk)
				select {
				case outputChan <- strings.TrimSuffix(builder.String(), "</answer>"):
					builder.Reset()
				case <-streamWriter.Done:
					return

				}
			case <-streamWriter.Done:
				return
			}
		}
	}()

	a._streamAgentPreProgram.SaveChatHistory(input, builder.String(), a._streamAgentPreProgram.SessionId)
	return outputChan, nil
}

// Chat LLM exposes the LLM method for chat
func (d *LLM) Chat(prompt string) (string, error) {
	return d.ai.Chat(prompt)
}

// Stream LLM exposes the LLM method for chat stream
func (d *LLM) Stream(prompt string) <-chan string {
	outputChan := make(chan string)
	go func() {
		d.ai.Stream(prompt, outputChan)
	}()
	return outputChan
}

// Chat ConvLLM exposes the LLM method for conversational chat
func (d *ConvLLM) Chat(prompt string) (string, error) {
	return d.convai.Chat(prompt)
}

// Stream ConvLLM exposes the LLM method for conversational chat stream
func (d *ConvLLM) Stream(prompt string) <-chan string {
	outputChan := make(chan string)
	go func() {
		d.convai.Stream(prompt, outputChan)
	}()
	return outputChan
}
