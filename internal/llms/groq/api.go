package groq

import (
	"github.com/darksuit-ai/darksuitai/internal/llms/groq/types"
)

// ChatError represents a chat-related error.
type ClientChatError struct {
	error
}

type GroqChatArgs struct {
	types.ChatArgs
}


func ChatGroq(kwargs ...map[string]interface{}) GroqChatArgs {
	var args types.ChatArgs

	for _, kwarg := range kwargs {
		if val, ok := kwarg["model"]; ok {
			args.Model = val.(string)
		}
		if val, ok := kwarg["messages"]; ok {
			args.Messages = val.([]types.Message)
		}
		if val, ok := kwarg["max_tokens"]; ok {
			args.MaxTokens = val.(int)
		}
		if val, ok := kwarg["stream"]; ok {
			args.Stream = val.(bool)
		}
		if val, ok := kwarg["stop"]; ok {
			if stopVal, ok := val.([]string); ok {
				args.Stop = stopVal
			} else if val == nil {
				args.Stop = nil
			}
		}

		// ... other fields ...
	}
	return GroqChatArgs{args}
}

// Chat sends a prompt to the chat client and returns the response.
func (args GroqChatArgs) Chat(apiKey string, prompt string, assistant string) (string, error) {
	if args.ChatArgs.Messages == nil {
		args.ChatArgs.Messages = make([]types.Message, 0)
	}

	if prompt != "" {
		args.ChatArgs.Messages = append(args.ChatArgs.Messages, types.Message{Role: "user", Content: prompt})
	} else {
		args.ChatArgs.Messages = append(args.ChatArgs.Messages, types.Message{Role: "user", Content: prompt}, types.Message{Role: "assistant", Content: assistant})
	}

	args.Stream = false
	response, err := Client(apiKey, args.ChatArgs)
	if err != nil {
		return "", err
	}
	return response, err
}

// StreamCompleteChat sends a prompt to the stream complete chat client and returns the response.
func (params GroqChatArgs) StreamCompleteChat(apiKey string, prompt string, system string) (string, error) {

	if params.ChatArgs.Messages == nil {
		params.ChatArgs.Messages = make([]types.Message, 0)
	}

	if system == "" {
		params.ChatArgs.Messages = append(params.ChatArgs.Messages, types.Message{Role: "user", Content: prompt})
	} else {
		params.ChatArgs.Messages = append(params.ChatArgs.Messages, types.Message{Role: "user", Content: prompt}, types.Message{Role: "system", Content: system})
	}

	params.Stream = true

	response, err := StreamCompleteClient(apiKey, params.ChatArgs)

	if err != nil {
		return "", err
	}
	return response, err
}

// StreamChat sends a prompt to the stream chat client and returns the response.
func (params GroqChatArgs) StreamChat(apiKey string, prompt string, system string, ipcChan chan string) {
	if params.ChatArgs.Messages == nil {
		params.ChatArgs.Messages = make([]types.Message, 0)
	}

	if system == "" {
		params.ChatArgs.Messages = append(params.ChatArgs.Messages, types.Message{Role: "user", Content: prompt})
	} else {
		params.ChatArgs.Messages = append(params.ChatArgs.Messages, types.Message{Role: "user", Content: prompt}, types.Message{Role: "system", Content: system})
	}

	params.Stream = true
	go StreamClient(apiKey, params.ChatArgs, ipcChan)
}
