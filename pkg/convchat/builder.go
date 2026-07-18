package convchat

import (
	"github.com/darksuit-ai/darksuitai/internal/llms"
	ant "github.com/darksuit-ai/darksuitai/internal/llms/anthropic"
	gem "github.com/darksuit-ai/darksuitai/internal/llms/gemini"
	gro "github.com/darksuit-ai/darksuitai/internal/llms/groq"
	oai "github.com/darksuit-ai/darksuitai/internal/llms/openai"
)

// buildLLM constructs a concrete LLM client for the configured provider.
//
// A new client is returned on every call so concurrent Chat/Stream
// invocations never share mutable state (this replaces the previous
// package-level `var llm`, which raced across goroutines).
//
// Returns nil if no supported provider is configured in ModelType.
func (ai ConvAI) buildLLM() llms.LLM {
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
