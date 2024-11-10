package agent

import (
	"github.com/darksuit-ai/darksuitai/pkg/tools"
	"go.mongodb.org/mongo-driver/mongo"
)

// Synapse struct represents the core structure for managing the state and data of the agent.
type Synapse struct {
	SystemPrompt          []byte // SysPrompt holds the system prompt for the agent.
	ChatInstructionPrompt []byte // InstructPrompt holds the instructional prompt for the agent.
	ToolNodes             []tools.BaseTool
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
}

// AIResponse represents the structured response of an AI assistant
type ToolCall struct {
	// Thought represents the AI's reasoning or interpretation of a user's request
	Thought string `json:"Thought"`

	// Action indicates the type of action the AI decides to take
	Action string `json:"Action"`

	// Input represents any input parameters for the action
	Input interface{} `json:"Input"`

	// Feedback provides information about the action being taken
	Feedback string `json:"Feedback"`
}

// AIResponse represents the structured response of an AI assistant
type ChatCall struct {
	SnapshotAIResponse string `json:"snapshotairesponse"`
}

// AgentActionTypes represents the various stages and types of actions an agent can take.
type AgentActionTypes struct {
	// AgentAction holds the current action the agent is performing.
	AgentAction map[string][]byte

	// AgentFinish holds the final actions or results after the agent completes its task.
	AgentFinish map[string][]byte

	// AgentPlan holds the planned actions or steps the agent intends to take.
	AgentPlan map[string][]byte

	// AgentError holds any errors encountered by the agent during its process.
	AgentError map[string][]byte
}

// LLMResult encapsulates the stream data from the channel and the complete prompt message for the cortex.
type LLMResult struct {
	// Message holds the complete prompt message sent to the Neuron
	Message []byte
	// LLMResponse is the channel that passes each stream text from the LLM's response
	LLMResponse <-chan string
}
