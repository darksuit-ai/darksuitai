package _chat

import (
	"github.com/darksuit-ai/darksuitai/internal/observability"
	"github.com/darksuit-ai/darksuitai/pkg/tools"
	"go.mongodb.org/mongo-driver/mongo"
)

type AgentPreProgram struct {
	BasePrompt           []byte
	SystemPrompt         []byte
	Tools                map[string]tools.BaseTool
	ToolNames            string
	AdditionalToolsMeta  map[string]interface{}
	BaseRunnableCaller   func(prompt []byte) (string, error)
	RunnableCaller       func(promptIterable []byte) (string, error)
	AIIdentity           []byte
	ChatMemoryCollection *mongo.Collection
	MaxIteration         int
	Verbose              bool
	SessionId            string

	// Native tool-calling configuration (used when ToolProtocol == "native").
	ToolProtocol string
	Provider     string
	Model        string
	APIKey       []byte
	MaxTokens    int
	Temperature  float64
	// RawSystemPrompt is the user's original system instruction (before the
	// ReAct/XML template is applied). Native tool calling uses this instead of
	// the XML-rendered SystemPrompt, which would otherwise instruct the model
	// to emit XML tool calls.
	RawSystemPrompt []byte

	// Observer receives run/LLM/tool telemetry events. When nil, a no-op
	// observer is used.
	Observer observability.Observer
}
