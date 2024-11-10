package _chat

import (
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
}
