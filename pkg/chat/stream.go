package chat

import (
	"fmt"

	"github.com/darksuit-ai/darksuitai/internal/prompts"
	"github.com/darksuit-ai/darksuitai/internal/utilities"
)

func (ai AI) Stream(prompt string, ipcChan chan string) {
	llm := ai.buildLLM()
	promptMap := ai.PromptKeys
	if prompt != "" {
		promptMap["query"] = []byte(prompt)
	}
	if ai.ChatInstruction == nil {
		internalPrompts, err := prompts.LoadPromptConfigs()
		if err != nil {

			ipcChan <- fmt.Sprintf("error loading config: %v", err)

		}

		ai.ChatInstruction = internalPrompts.CHATINSTRUCTION
	}
	promptTemplate := utilities.CustomFormat(ai.ChatInstruction, promptMap)
	llm.StreamChat(string(ai.APIKey), string(promptTemplate), string(ai.ChatSystemInstruction), ipcChan)

}
