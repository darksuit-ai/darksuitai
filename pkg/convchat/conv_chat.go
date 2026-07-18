package convchat

import (
	"github.com/darksuit-ai/darksuitai/internal/memory/mongodb"
	"github.com/darksuit-ai/darksuitai/internal/utilities"
)

func (ai ConvAI) Chat(prompt string) (string, error) {
	llm := ai.buildLLM()
	promptMap := ai.PromptKeys
	promptMap["query"] = []byte(prompt)
	var mongoMemory mongodb.ChatMemoryCollectionInterface = mongodb.NewMongoCollection(ai.MongoDB)
	chatData, _ := mongoMemory.RetrieveMemoryWithK("", 6)
	promptMap["chat_history"] = []byte(chatData)
	promptTemplate := utilities.CustomFormat(ai.ChatInstruction, promptMap)
	resp, err := llm.StreamCompleteChat(string(ai.APIKey), string(promptTemplate), string(ai.ChatSystemInstruction))
	return resp, err
}
