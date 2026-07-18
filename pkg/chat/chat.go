package chat

import (
	"fmt"

	"github.com/darksuit-ai/darksuitai/internal/prompts"
	"github.com/darksuit-ai/darksuitai/internal/utilities"
)

func (ai AI) Chat(prompt string) (string, error) {
	llm := ai.buildLLM()
	promptMap := ai.PromptKeys
	promptMap["query"] = []byte(prompt)
	if ai.ChatInstruction == nil {
		internalPrompts, err := prompts.LoadPromptConfigs()
		if err != nil {
			return "", fmt.Errorf("error loading config: %v", err)
		}

		ai.ChatInstruction = internalPrompts.CHATINSTRUCTION
	}
	promptTemplate := utilities.CustomFormat(ai.ChatInstruction, promptMap)
	resp, err := llm.StreamCompleteChat(string(ai.APIKey), string(promptTemplate), string(ai.ChatSystemInstruction))
	return resp, err
}
