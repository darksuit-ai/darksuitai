package chat

import (
	"fmt"

	ant "github.com/darksuit-ai/darksuitai/internal/llms/anthropic"
	gro "github.com/darksuit-ai/darksuitai/internal/llms/groq"
	oai "github.com/darksuit-ai/darksuitai/internal/llms/openai"
	"github.com/darksuit-ai/darksuitai/internal/prompts"
	"github.com/darksuit-ai/darksuitai/internal/utilities"
)

var llm LLM

func (ai AI) Chat(prompt string) (string, error) {
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
	if ai.ChatInstruction == nil{
		internalPrompts, err := prompts.LoadPromptConfigs()
		if err != nil {
			return "", fmt.Errorf("error loading config: %v", err)
		}

		ai.ChatInstruction = internalPrompts.CHATINSTRUCTION
	}
	promptTemplate := utilities.CustomFormat(ai.ChatInstruction, promptMap)
	resp, err := llm.StreamCompleteChat(string(ai.APIKey),string(promptTemplate), string(ai.ChatSystemInstruction))
	return resp, err
}
