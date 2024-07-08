package agent

type Synapse struct {
	SysPrompt         []byte
	InstructPrompt    []byte
	MetaData          map[string]interface{}
	ToolNames         []map[string]string
	ModelType         map[string]string
}

type AgentActionTypes struct {
	AgentAction map[string]string
	AgentFinish map[string][]byte
	AgentPlan   map[string][]byte
	AgentError  map[string][]byte
}
