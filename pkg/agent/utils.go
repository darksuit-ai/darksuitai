package agent

import (
	"bytes"
	"strings"

	"github.com/darksuit-ai/darksuitai/pkg/tools"
)

// RenderToolNames is a function that returns two strings:
// 1. toolList: a concatenated string of all tool names and descriptions,
// 2. toolNames: a comma-separated string of all tool names
func RenderToolNames(agentTools []tools.BaseTool) (string, string) {
	var toolList strings.Builder
	var toolNames strings.Builder
	toolList.Grow(len(agentTools) * 20)  // pre-allocate memory
	toolNames.Grow(len(agentTools) * 20) // pre-allocate memory

	for _, tool := range agentTools {
		toolList.WriteString(tool.Name)
		toolList.WriteString(": ")
		toolList.WriteString(tool.Description)
		toolList.WriteString("\n")
		if toolNames.Len() > 0 {
			toolNames.WriteString(",")
		}
		toolNames.WriteString(tool.Name)
	}

	return toolList.String(), strings.TrimSuffix(toolNames.String(), ",")
}

func removeTextBeforeToolCall(input []byte) []byte {
	toolCallTag := []byte("<tool_call>")
	index := bytes.Index(input, toolCallTag)
	if index == -1 {
		return input
	}
	return input[index:]
}
