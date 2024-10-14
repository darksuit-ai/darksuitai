package agent

// Synapse struct represents the core structure for managing the state and data of the agent.
type Synapse struct {
	SysPrompt         []byte                 // SysPrompt holds the system prompt for the agent.
	InstructPrompt    []byte                 // InstructPrompt holds the instructional prompt for the agent.
	MetaData          map[string]interface{} // MetaData contains various metadata related to the agent's state.
	ToolNames         string                 // ToolNames is a string of tool names available to the agent.
	ModelType         map[string]string      // ModelType specifies the type of model used by the agent.
	QueryForTimeRange string                 // QueryForTimeRange holds the query string for a specific time range.
	RagToolResponse   []map[string]string
	RagData []byte
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
