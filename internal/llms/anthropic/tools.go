package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// This file implements Phase 2 "native tool calling" for Anthropic using the
// official SDK's structured tool-use API (tool_use / tool_result content
// blocks) instead of parsing <tool_call>/<answer> XML out of free text.
//
// It is deliberately self-contained: all SDK types stay inside this package so
// the agent layer can drive a native tool loop without importing the SDK. The
// legacy XML/ReAct executor remains the default and is still fully supported;
// this path is selected only when the caller opts into the "native" tool
// protocol AND the configured provider is Anthropic.

// ToolSpec is a provider-agnostic description of a callable tool.
type ToolSpec struct {
	Name        string
	Description string
	// Properties is the JSON-schema "properties" object for the tool input.
	// When nil, the tool is exposed with a single required string property
	// named "input" (the single-string contract used by the XML executor).
	Properties map[string]any
	Required   []string
}

// ToolExecutor runs a single tool call and returns the textual result.
//
//   - name  is the tool name the model chose.
//   - input is the decoded string argument for default (single-string) tools,
//     or the raw JSON arguments for tools declared with a custom schema.
//
// Returning isError=true marks the tool_result as an error so the model can
// recover on the next turn.
type ToolExecutor func(name, input string) (result string, isError bool)

// ToolLoopConfig carries the model/runtime configuration for RunToolLoop.
type ToolLoopConfig struct {
	APIKey        string
	Model         string
	MaxTokens     int
	Temperature   float64
	System        string
	MaxIterations int
	Verbose       bool
}

// RunToolLoop drives a native tool-use conversation to completion.
//
// It sends the user's message together with the tool definitions, executes any
// tool_use blocks the model returns via exec, feeds the tool_result blocks back,
// and repeats until the model responds without requesting a tool or until
// MaxIterations is reached. It returns the final assistant text.
func RunToolLoop(cfg ToolLoopConfig, userInput string, specs []ToolSpec, exec ToolExecutor) (string, error) {
	client, err := newSDKClient(cfg.APIKey)
	if err != nil {
		return "", err
	}

	model := cfg.Model
	if model == "" {
		model = "claude-sonnet-5"
	}
	maxTokens := int64(cfg.MaxTokens)
	if maxTokens <= 0 {
		maxTokens = 1024
	}
	maxIterations := cfg.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 5
	}

	// Build SDK tool params and remember which tools use the default schema.
	isDefaultSchema := make(map[string]bool, len(specs))
	tools := make([]anthropic.ToolUnionParam, 0, len(specs))
	for _, spec := range specs {
		toolParam := anthropic.ToolParam{Name: spec.Name}
		if spec.Description != "" {
			toolParam.Description = anthropic.String(spec.Description)
		}
		if spec.Properties == nil {
			isDefaultSchema[spec.Name] = true
			toolParam.InputSchema = anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"input": map[string]any{
						"type":        "string",
						"description": spec.Description,
					},
				},
				Required: []string{"input"},
			}
		} else {
			toolParam.InputSchema = anthropic.ToolInputSchemaParam{
				Properties: spec.Properties,
				Required:   spec.Required,
			}
		}
		tp := toolParam
		tools = append(tools, anthropic.ToolUnionParam{OfTool: &tp})
	}

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(userInput)),
	}

	var lastText string
	for iter := 0; iter < maxIterations; iter++ {
		params := anthropic.MessageNewParams{
			Model:     anthropic.Model(model),
			MaxTokens: maxTokens,
			Messages:  messages,
			Tools:     tools,
		}
		if cfg.System != "" {
			params.System = []anthropic.TextBlockParam{{Text: cfg.System}}
		}
		// temperature intentionally omitted: 2026 Claude models deprecate it.

		message, err := client.Messages.New(context.Background(), params)
		if err != nil {
			return lastText, fmt.Errorf("anthropic: tool loop request failed: %w", err)
		}

		// Collect this turn's text and any tool calls.
		var turnText strings.Builder
		var toolUses []anthropic.ToolUseBlock
		for _, block := range message.Content {
			switch b := block.AsAny().(type) {
			case anthropic.TextBlock:
				turnText.WriteString(b.Text)
			case anthropic.ToolUseBlock:
				toolUses = append(toolUses, b)
			}
		}
		if turnText.Len() > 0 {
			lastText = turnText.String()
			if cfg.Verbose {
				fmt.Printf("[assistant] %s\n", lastText)
			}
		}

		// No tools requested: the model has produced its final answer.
		if len(toolUses) == 0 {
			return lastText, nil
		}

		// Record the assistant turn (including tool_use blocks) verbatim.
		messages = append(messages, message.ToParam())

		// Execute each requested tool and gather tool_result blocks.
		toolResults := make([]anthropic.ContentBlockParamUnion, 0, len(toolUses))
		for _, tu := range toolUses {
			input := deriveToolInput(tu.Input, isDefaultSchema[tu.Name])
			if cfg.Verbose {
				fmt.Printf("[tool_use] %s(%s)\n", tu.Name, input)
			}
			result, isErr := exec(tu.Name, input)
			if cfg.Verbose {
				fmt.Printf("[tool_result] %s\n", result)
			}
			toolResults = append(toolResults, anthropic.NewToolResultBlock(tu.ID, result, isErr))
		}
		messages = append(messages, anthropic.NewUserMessage(toolResults...))
	}

	// Iteration budget exhausted; return whatever text we have.
	if lastText == "" {
		return "", fmt.Errorf("anthropic: tool loop reached max iterations (%d) without a final answer", maxIterations)
	}
	return lastText, nil
}

// deriveToolInput turns the model's raw JSON tool arguments into the string
// passed to a tool's ToolFunc. For default single-string tools it extracts the
// "input" property; for custom-schema tools it passes the raw JSON through.
func deriveToolInput(raw json.RawMessage, defaultSchema bool) string {
	if !defaultSchema {
		return string(raw)
	}
	var args map[string]any
	if err := json.Unmarshal(raw, &args); err != nil {
		return string(raw)
	}
	switch v := args["input"].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		if b, err := json.Marshal(v); err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", v)
	}
}
