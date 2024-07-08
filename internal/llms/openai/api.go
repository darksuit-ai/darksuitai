package openai

import "github.com/darksuit-ai/darksuitai/internal/llms/openai/types"

// ChatError represents a chat-related error.
type ClientChatError struct {
	error
}

type OAIChatArgs struct {
	types.ChatArgs
}

func ChatOAI(kwargs ...map[string]interface{}) OAIChatArgs {
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
	return OAIChatArgs{args}
}

// Chat sends a prompt to the chat client and returns the response.
func (args OAIChatArgs) Chat(prompt string, assistant string) (string, error) {
	if args.ChatArgs.Messages == nil {
		args.ChatArgs.Messages = make([]types.Message, 0)
	}

	if prompt != "" {
		args.ChatArgs.Messages = append(args.ChatArgs.Messages, types.Message{Role: "user", Content: prompt})
	} else {
		args.ChatArgs.Messages = append(args.ChatArgs.Messages, types.Message{Role: "user", Content: prompt}, types.Message{Role: "assistant", Content: assistant})
	}

	args.Stream = false
	response, err := Client(args.ChatArgs)
	if err != nil {
		return "", err
	}
	return response, err
}

// StreamCompleteChat sends a prompt to the stream complete chat client and returns the response.
func (params OAIChatArgs) StreamCompleteChat(prompt string, system string) (string, error) {

	if params.ChatArgs.Messages == nil {
		params.ChatArgs.Messages = make([]types.Message, 0)
	}

	if system == "" {
		params.ChatArgs.Messages = append(params.ChatArgs.Messages, types.Message{Role: "user", Content: prompt})
	} else {
		params.ChatArgs.Messages = append(params.ChatArgs.Messages, types.Message{Role: "user", Content: prompt}, types.Message{Role: "system", Content: system})
	}

	params.Stream = true

	response, err := StreamCompleteClient(params.ChatArgs)

	if err != nil {
		return "", err
	}
	return response, err
}

// StreamChat sends a prompt to the stream chat client and returns the response.
func (params OAIChatArgs) StreamChat(prompt string, system string) <-chan string {

	if params.ChatArgs.Messages == nil {
		params.ChatArgs.Messages = make([]types.Message, 0)
	}

	if system == "" {
		params.ChatArgs.Messages = append(params.ChatArgs.Messages, types.Message{Role: "user", Content: prompt})
	} else {
		params.ChatArgs.Messages = append(params.ChatArgs.Messages, types.Message{Role: "user", Content: prompt}, types.Message{Role: "system", Content: system})
	}

	params.Stream = true
	chunkchan := make(chan string)

	go func() {
		err := StreamClient(params.ChatArgs, chunkchan)
		if err != nil {
			chunkchan <- err.Error()
		}
	}()

	return chunkchan
}
