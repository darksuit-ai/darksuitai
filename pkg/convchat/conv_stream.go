package convchat

import (
	ant "github.com/darksuit-ai/darksuitai/internal/llms/anthropic"
	oai "github.com/darksuit-ai/darksuitai/internal/llms/openai"
	gro "github.com/darksuit-ai/darksuitai/internal/llms/groq"
	"github.com/darksuit-ai/darksuitai/internal/memory/mongodb"
	"github.com/darksuit-ai/darksuitai/internal/utilities"
)


func (ai ConvAI) Stream(prompt string) <-chan string  {
	kwargs := make([]map[string]interface{}, 5)
	for key := range ai.ModelType {
		switch key {
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
			llm = oai.ChatOAI(kwargs...)
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

			llm = ant.ChatAnth(kwargs...)
		case "groq":
			for k, v := range ai.ModelKwargs {
				kwargs[k] = map[string]interface{}{
					"model":          ai.ModelType["groq"],
					"max_tokens":     v.MaxTokens,
					"temperature":    v.Temperature,
					"stream":         v.Stream,
					"stop": v.StopSequences,
				}

			}

			llm = gro.ChatGroq(kwargs...)
		default:
			llm = nil
		}
	}
	promptMap := ai.PromptKeys
	promptMap["query"] = []byte(prompt)
	promptMap["chat_history"] = []byte(mongodb.RetrieveMemoryWithK(ai.MongoDB, "", 6))
	promptTemplate := utilities.CustomFormat(ai.ChatInstruction, promptMap)
	respChan := llm.StreamChat(string(ai.APIKey),string(promptTemplate), string(ai.ChatSystemInstruction))
	return respChan
}
