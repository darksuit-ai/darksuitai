package _stream

import (
	"context"
	"strings"
)

type streamState struct {
	buffer      []string
	accumulated strings.Builder
	isToolCall  bool
	isStreaming bool
}

const (
	wordBufferSize = 5
	toolCallMarker = "<tool_call>"
)

func _streamDifferentiator(ctx context.Context, writer *StreamWriter, llmStreamData LLMResult) ([]byte, bool) {
	state := streamState{
		buffer: make([]string, 0, wordBufferSize),
	}

	// Process incoming stream data
	for syllable := range llmStreamData.LLMResponse {
		// If already streaming, write directly
		if state.isStreaming {
			_, _ = writer.Write([]byte(toRawStringLiteral(syllable)))
			continue
		}

		processStreamChunk(syllable, &state, writer)
	}

	// If we found a tool call, return the accumulated content
	if state.isToolCall {
		return []byte(state.accumulated.String()), true
	}

	// Flush any remaining buffered content
	if len(state.buffer) > 0 {
		content := strings.Join(state.buffer, "")
		if _, err := writer.Write([]byte(toRawStringLiteral(content))); err != nil {
			return nil, false
		}
	}

	// No tool call found
	return nil, false
}

func processStreamChunk(syllable string, state *streamState, writer *StreamWriter) {

	// Accumulate content
	state.accumulated.WriteString(syllable)

	// print(state.accumulated.String())
	if !state.isToolCall {

		// Skip empty chunks
		trimmed := strings.TrimSpace(syllable)

		// Buffer words for tool call detection
		state.buffer = append(state.buffer, trimmed)

		// Process buffer when it reaches the desired size
		if len(state.buffer) >= wordBufferSize {
			content := strings.Join(state.buffer, "")

			if strings.Contains(content, toolCallMarker) {
				state.isToolCall = true
			} else {

				// Start streaming if no tool call found
				state.isStreaming = true
				_, _ = writer.Write([]byte(toRawStringLiteral(content + " ")))
			}

		}
	}

}

func toRawStringLiteral(s string) string {
	replacer := strings.NewReplacer(
		// `\`, `\\`,
		"\n", `\n`,
		// "\r", `\r`,
		// "\t", `\t`,
		// `"`, `\"`,
	)
	return replacer.Replace(s)
}
