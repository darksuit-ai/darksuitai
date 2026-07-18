package agent

import (
	"github.com/darksuit-ai/darksuitai/internal/llms"
	ant "github.com/darksuit-ai/darksuitai/internal/llms/anthropic"
	gem "github.com/darksuit-ai/darksuitai/internal/llms/gemini"
	gro "github.com/darksuit-ai/darksuitai/internal/llms/groq"
	oai "github.com/darksuit-ai/darksuitai/internal/llms/openai"
	"github.com/darksuit-ai/darksuitai/internal/utilities"
)

// buildLLM constructs the concrete LLM client for the configured provider.
//
// It returns a freshly built client on every call so that concurrent agent
// invocations never share mutable state. Previously a package-level `var llm`
// was assigned here, which raced across goroutines; keeping the client local
// removes that data race entirely.
//
// Returns nil if no supported provider is configured in ModelType.
func (ai Synapse) buildLLM() llms.LLM {
	// One kwargs map per configured ModelKwargs entry. The provider
	// constructors apply each map in order, so the last (user-supplied)
	// entry wins — preserving the original behaviour.
	kwargs := make([]map[string]interface{}, len(ai.ModelKwargs))

	for provider := range ai.ModelType {
		switch provider {
		case "openai":
			for k, v := range ai.ModelKwargs {
				kwargs[k] = map[string]interface{}{
					"model":          ai.ModelType["openai"],
					"max_tokens":     v.MaxTokens,
					"temperature":    v.Temperature,
					"stream":         v.Stream,
					"stop_sequences": v.StopSequences,
				}
			}
			return oai.ChatOAI(kwargs...)
		case "anthropic":
			for k, v := range ai.ModelKwargs {
				kwargs[k] = map[string]interface{}{
					"model":          ai.ModelType["anthropic"],
					"max_tokens":     v.MaxTokens,
					"temperature":    v.Temperature,
					"stream":         v.Stream,
					"stop_sequences": v.StopSequences,
				}
			}
			return ant.ChatAnth(kwargs...)
		case "gemini":
			for k, v := range ai.ModelKwargs {
				kwargs[k] = map[string]interface{}{
					"model":       ai.ModelType["gemini"],
					"max_tokens":  v.MaxTokens,
					"temperature": v.Temperature,
					"stream":      v.Stream,
					"stop":        v.StopSequences,
				}
			}
			return gem.ChatGEM(kwargs...)
		case "groq":
			for k, v := range ai.ModelKwargs {
				kwargs[k] = map[string]interface{}{
					"model":       ai.ModelType["groq"],
					"max_tokens":  v.MaxTokens,
					"temperature": v.Temperature,
					"stream":      v.Stream,
					"stop":        v.StopSequences,
				}
			}
			return gro.ChatGroq(kwargs...)
		}
	}
	return nil
}

// Basechat runs the first (non-iterative) LLM call for the agent.
func (ai Synapse) Basechat(prompt []byte) (string, error) {
	llm := ai.buildLLM()
	promptTemplate := utilities.CustomFormat(prompt, ai.PromptKeys)
	return llm.StreamCompleteChat(string(ai.APIKey), string(promptTemplate), string(ai.SystemPrompt))
}

// ChatIterable runs a follow-up LLM call carrying the agent's thought plus a tool response.
func (ai Synapse) ChatIterable(thoughtWithToolResponse []byte) (string, error) {
	llm := ai.buildLLM()
	promptTemplate := utilities.CustomFormat(thoughtWithToolResponse, ai.PromptKeys)
	return llm.StreamCompleteChat(string(ai.APIKey), string(promptTemplate), string(ai.SystemPrompt))
}

// BaseStream runs the first LLM call in streaming mode, writing chunks to ipcChan.
func (ai Synapse) BaseStream(prompt []byte, ipcChan chan string) {
	llm := ai.buildLLM()
	promptTemplate := utilities.CustomFormat(prompt, ai.PromptKeys)
	llm.StreamChat(string(ai.APIKey), string(promptTemplate), string(ai.SystemPrompt), ipcChan)
}

// StreamIterable runs a follow-up streaming LLM call carrying the agent's thought plus a tool response.
func (ai Synapse) StreamIterable(thoughtWithToolResponse []byte, ipcChan chan string) {
	llm := ai.buildLLM()
	promptTemplate := utilities.CustomFormat(thoughtWithToolResponse, ai.PromptKeys)
	llm.StreamChat(string(ai.APIKey), string(promptTemplate), string(ai.SystemPrompt), ipcChan)
}
