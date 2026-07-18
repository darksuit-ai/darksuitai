package types

import (
	"github.com/darksuit-ai/darksuitai/internal/memory"
	"github.com/darksuit-ai/darksuitai/internal/observability"
	"go.mongodb.org/mongo-driver/mongo"
)

type LLMArgs struct {
	ChatSystemInstruction []byte
	ChatInstruction       []byte
	PromptKeys            map[string][]byte
	ModelType             map[string]string
	MongoDB               *mongo.Collection
	ModelKwargs           []struct {
		MaxTokens     int      `json:"max_tokens"`
		Temperature   float64  `json:"temperature"`
		Stream        bool     `json:"stream"`
		StopSequences []string `json:"stop_sequences"`
	}
	APIKey []byte
	// ToolProtocol selects how the agent invokes tools: "xml" (default, the
	// legacy ReAct/XML protocol supported by all providers) or "native"
	// (provider-side structured tool calling; currently Anthropic only).
	ToolProtocol string
	// Observer receives run/LLM/tool telemetry events. When nil, telemetry is
	// disabled (a no-op observer is used).
	Observer observability.Observer
	// Compactor, when set, enables conversation compaction (rolling summary +
	// recent turns) in place of raw last-K memory retrieval.
	Compactor *memory.Compactor
}
