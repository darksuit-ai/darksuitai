package agent

import (
	"bytes"
	"context"
	"fmt"

	"github.com/darksuit-ai/darksuitai/internal/memory"
	"github.com/darksuit-ai/darksuitai/internal/memory/mongodb"
	"github.com/darksuit-ai/darksuitai/internal/prompts"
	"github.com/darksuit-ai/darksuitai/internal/utilities"
	"github.com/darksuit-ai/darksuitai/pkg/tools"
	"go.mongodb.org/mongo-driver/mongo"
)

// PromptAgentInterface defines the interface for preparing the prompt for the LLM.
type PromptAgentInterface interface {
	PreparePrompt(SystemPrompt []byte, ChatInstructionPrompt []byte, agentTools []tools.BaseTool, PromptKeys map[string][]byte, mongoCollection *mongo.Collection, sessionId string, compactor *memory.Compactor) ([]byte, []byte, map[string]tools.BaseTool, string, error)
}

// PromptAgent is a struct that implements the PromptAgentInterface.
type PromptAgent struct {
}

// NewPromptAgent returns a new instance of PromptAgent.
func NewPromptAgent() PromptAgentInterface {
	return &PromptAgent{}
}

// PreparePrompt is a function that implements the MultiModalAgentInterface.
// It prepares the prompt for the LLM (Language Learning Model) and returns the LLM and the prepared prompt.
func (a *PromptAgent) PreparePrompt(SystemPrompt []byte, ChatInstructionPrompt []byte, agentTools []tools.BaseTool,
	PromptKeys map[string][]byte, mongoCollection *mongo.Collection, sessionId string, compactor *memory.Compactor) ([]byte, []byte, map[string]tools.BaseTool, string, error) {

	var (
		chatHistory     bytes.Buffer
		promptMapSystem map[string][]byte
	)

	// Clean up buffers after function returns
	defer func() {
		chatHistory.Reset()
	}()

	if ChatInstructionPrompt == nil {

		internalPrompts, err := prompts.LoadPromptConfigs()
		if err != nil {
			return nil, nil, nil, "", fmt.Errorf("error loading config: %v", err)
		}

		ChatInstructionPrompt = internalPrompts.AGENTCHATINSTRUCTION
	}

	if SystemPrompt == nil {

		internalPrompts, err := prompts.LoadPromptConfigs()
		if err != nil {
			return nil, nil, nil, "", fmt.Errorf("error loading config: %v", err)
		}

		SystemPrompt = internalPrompts.AGENTSYSTEMINSTRUCTION
	}

	// Render the tool names
	tl, tn := RenderToolNames(agentTools)

	// Merge user-supplied prompt keys with the rendered tool metadata.
	// (Previously promptMap was reassigned here, silently discarding the
	// user's PromptKeys; that is now fixed.)
	promptMap := make(map[string][]byte)
	for key, value := range PromptKeys {
		promptMap[key] = value
	}
	promptMap["tool_names"] = []byte(tn)
	promptMap["tools"] = []byte(tl)

	// Inject chat history when a session and memory store are available. With a
	// compactor configured, use the compacted context (rolling summary + recent
	// turns); otherwise fall back to the last-K raw transcript.
	if sessionId != "" && mongoCollection != nil {
		var mongoMemory mongodb.ChatMemoryCollectionInterface = mongodb.NewMongoCollection(mongoCollection)
		if compactor != nil {
			if turns, retrieveErr := mongoMemory.RetrieveTurns(sessionId); retrieveErr == nil {
				if ctxStr, cErr := compactor.BuildContext(context.Background(), sessionId, turns); cErr == nil {
					chatHistory.WriteString(ctxStr)
				}
			}
		} else {
			if chatData, retrieveErr := mongoMemory.RetrieveMemoryWithK(sessionId, 6); retrieveErr == nil {
				chatHistory.WriteString(chatData)
			}
		}
		promptMap["chat_history"] = chatHistory.Bytes()
	}

	if _, exists := PromptKeys["timeZone"]; exists {
		// Get the current date and time
		currentDate, _, _, _, currentDayOfWeek, currentTime := utilities.GetCurrentDateTimeWithTimeZoneShift(string(PromptKeys["timeZone"]))

		promptMapSystem = map[string][]byte{
			"current_date":            []byte(currentDate),
			"current_day_of_the_week": []byte(currentDayOfWeek),
			"current_time":            []byte(currentTime),
		}

	}

	// Format the instruction prompt with the prepared prompt map
	llmPrompt := utilities.CustomFormat(ChatInstructionPrompt, promptMap)

	// Format the system prompt with the prepared prompt map
	llmSystemPrompt := utilities.CustomFormat(SystemPrompt, promptMapSystem)

	agentToolsMap := make(map[string]tools.BaseTool)
	for _, tool := range agentTools {
		agentToolsMap[tool.Name] = tool
	}

	// Return the LLM and the prepared prompt
	return llmPrompt, llmSystemPrompt, agentToolsMap, tn, nil
}
