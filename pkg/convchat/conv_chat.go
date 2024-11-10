package convchat

import (
	ant "github.com/darksuit-ai/darksuitai/internal/llms/anthropic"
	oai "github.com/darksuit-ai/darksuitai/internal/llms/openai"
	"github.com/darksuit-ai/darksuitai/internal/memory/mongodb"
	"github.com/darksuit-ai/darksuitai/internal/utilities"
)

var llm LLM

func (ai ConvAI) Chat(prompt string) (string, error) {
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
		default:
			llm = nil
		}
	}
	promptMap := ai.PromptKeys
	promptMap["query"] = []byte(prompt)
	chatData,_ := mongodb.RetrieveMemoryWithK(ai.MongoDB, "", 6)
	promptMap["chat_history"] = []byte(chatData)
	promptTemplate := utilities.CustomFormat(ai.ChatInstruction, promptMap)
	resp, err := llm.StreamCompleteChat(string(ai.APIKey),string(promptTemplate), string(ai.ChatSystemInstruction))
	return resp, err
}
