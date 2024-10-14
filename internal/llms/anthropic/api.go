package anthropic

import "github.com/darksuit-ai/darksuitai/internal/llms/anthropic/types"

// ChatError represents a chat-related error.
type ClientChatError struct {
	error
}

type AnthChatArgs struct {
	types.ChatArgs
}

func ChatAnth(kwargs ...map[string]interface{}) AnthChatArgs {
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
		if val, ok := kwarg["stop_sequences"]; ok {
			if stopVal, ok := val.([]string); ok {
				args.StopSequences = stopVal
			} else if val == nil {
				args.StopSequences = nil
			}
		}

		// ... other fields ...
	}
	return AnthChatArgs{args}
}

// ChatClient sends a prompt to the chat client and returns the response.
func (params AnthChatArgs) Chat(apiKey string, prompt string, system string) (string, error) {
	if params.ChatArgs.Messages == nil {
		params.ChatArgs.Messages = make([]types.Message, 0)
	}

	params.ChatArgs.Messages = append(params.ChatArgs.Messages, types.Message{Role: "user", Content: prompt})
	if system != "" {
	params.ChatArgs.System = system
	}

	params.Stream = false
	response, err := Client(apiKey, params.ChatArgs)
	if err != nil {
		return "", err
	}
	return response, err
}

func (params AnthChatArgs) StreamCompleteChat(apiKey string, prompt string, system string) (string, error) {
	if params.ChatArgs.Messages == nil {
		params.ChatArgs.Messages = make([]types.Message, 0)
	}

	params.ChatArgs.Messages = append(params.ChatArgs.Messages, types.Message{Role: "user", Content: prompt})
	if system != "" {
	params.ChatArgs.System = system
	}

	params.Stream = true

	response, err := StreamCompleteClient(apiKey, params.ChatArgs)

	if err != nil {
		return "", err
	}
	return response, err
}

func (params AnthChatArgs) StreamChat(apiKey string, prompt string, system string) <-chan string {
	if params.ChatArgs.Messages == nil {
		params.ChatArgs.Messages = make([]types.Message, 0)
	}

	params.ChatArgs.Messages = append(params.ChatArgs.Messages, types.Message{Role: "user", Content: prompt})
	if system != "" {
	params.ChatArgs.System = system
	}

	params.Stream = true
	chunkchan := make(chan string)

	go func() {
		defer close(chunkchan)
		err := StreamClient(apiKey,params.ChatArgs, chunkchan)
		if err != nil {
			chunkchan <- err.Error()
		}
	}()

	return chunkchan
}
