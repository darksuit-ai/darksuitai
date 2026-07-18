package _chat

import (
	"fmt"
	"time"

	ant "github.com/darksuit-ai/darksuitai/internal/llms/anthropic"
	"github.com/darksuit-ai/darksuitai/internal/memory/mongodb"
	"github.com/darksuit-ai/darksuitai/internal/observability"
	"github.com/darksuit-ai/darksuitai/internal/utilities"
)

// NativeExecutor runs the agent using provider-side (native) structured tool
// calling instead of the XML/ReAct protocol handled by Executor.
//
// It mirrors Executor's contract — same inputs, same ([]byte answer, tool
// metadata, error) return — so pkg callers can route to either based on the
// configured ToolProtocol without any other changes. It currently supports the
// Anthropic provider (migrated to the official SDK in Phase 1); callers are
// expected to fall back to Executor for other providers.
func (prePrompt *AgentPreProgram) NativeExecutor(queryPrompt map[string][]byte, sessionId string, maxIterations int, verbose bool) (result []byte, toolData any, execErr error) {
	question := string(queryPrompt["question"])

	// Observability: start a run span and close it on every return path.
	obs := prePrompt.Observer
	if obs == nil {
		obs = observability.Noop{}
	}
	runHandle := obs.StartRun(observability.RunInfo{
		SessionID: sessionId,
		Provider:  prePrompt.Provider,
		Model:     prePrompt.Model,
		Protocol:  "native",
		Input:     question,
	})
	defer func() {
		runHandle.End(string(result), execErr)
	}()

	// Translate the registered tools into provider-agnostic specs.
	specs := make([]ant.ToolSpec, 0, len(prePrompt.Tools))
	for _, tool := range prePrompt.Tools {
		specs = append(specs, ant.ToolSpec{
			Name:        tool.Name,
			Description: tool.Description,
			Properties:  tool.InputSchema,
			Required:    tool.Required,
		})
	}

	// Collect raw tool metadata (matching Executor's toolData return shape).
	var toolResponseList []interface{}

	exec := func(name, input string) (string, bool) {
		tool, found := prePrompt.Tools[name]
		if !found {
			return fmt.Sprintf("You tried to use the tool %q, but it doesn't exist. You must use any of these available tools: [%s].", name, prePrompt.ToolNames), true
		}
		start := time.Now()
		result, rawToolResponse, toolErr := tool.ToolFunc(input, tool.Name, prePrompt.AdditionalToolsMeta)
		if toolErr != nil {
			runHandle.ToolEnd(observability.ToolCall{Name: tool.Name, Input: input, Output: toolErr.Error(), IsError: true, Duration: time.Since(start)})
			return toolErr.Error(), true
		}
		runHandle.ToolEnd(observability.ToolCall{Name: tool.Name, Input: input, Output: result, Duration: time.Since(start)})
		toolResponseList = append(toolResponseList, map[string]interface{}{tool.Name: rawToolResponse})
		return result, false
	}

	cfg := ant.ToolLoopConfig{
		APIKey:        string(prePrompt.APIKey),
		Model:         prePrompt.Model,
		MaxTokens:     prePrompt.MaxTokens,
		Temperature:   prePrompt.Temperature,
		System:        string(prePrompt.RawSystemPrompt),
		MaxIterations: maxIterations,
		Verbose:       verbose,
	}

	finalText, err := ant.RunToolLoop(cfg, question, specs, exec)
	if err != nil {
		runHandle.Error("tool_loop", err)
		return nil, nil, err
	}

	if verbose {
		utilities.Printer("", finalText, "green")
	}

	// Persist the exchange, mirroring Executor's memory behaviour.
	if prePrompt.ChatMemoryCollection != nil {
		var mongoMemory mongodb.ChatMemoryCollectionInterface = mongodb.NewMongoCollection(prePrompt.ChatMemoryCollection)
		_ = mongoMemory.AddConversationToMemory(sessionId, question, finalText)
	}

	if toolResponseList != nil {
		return []byte(finalText), toolResponseList, nil
	}
	return []byte(finalText), []string{}, nil
}
