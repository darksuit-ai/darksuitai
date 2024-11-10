package types

type ChatError struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// ChatError represents a chat-related error.
type ClientChatError struct {
	error
}

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

type ChatArgs struct {
	Model         string      `json:"model"`
	MaxTokens     int         `json:"max_tokens"`
	Messages      []Message   `json:"messages"`
	Temperature   float64     `json:"temperature"`
	Stream        bool        `json:"stream,omitempty"`
	Stop interface{} `json:"stop"`
}

// Message is a message in a chat request
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	Content      []map[string]string `json:"content"`
	ID           string              `json:"id"`
	Model        string              `json:"model"`
	Role         string              `json:"role"`
	StopReason   string              `json:"stop_reason"`
	StopSequence string              `json:"stop_sequence"`
	Type         string              `json:"type"`
	Usage        map[string]int
}
