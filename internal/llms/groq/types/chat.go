package types

type ChatError struct {
	Error map[string]string
}

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

type ChatArgs struct {
	Model       string      `json:"model"`
	Messages    []Message   `json:"messages"`
	MaxTokens   int         `json:"max_tokens"`
	Temperature float64     `json:"temperature"`
	TopP        float64     `json:"top_p"`
	Seed        int         `json:"seed,omitempty"`
	Stream      bool        `json:"stream,omitempty"`
	Stop        interface{} `json:"stop"`
}

// Message is a message in a chat request
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Logprobs     interface{} `json:"logprobs"` // Assuming logprobs can be null
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int     `json:"prompt_tokens"`
		PromptTime       float64 `json:"prompt_time"`
		CompletionTokens int     `json:"completion_tokens"`
		CompletionTime   float64 `json:"completion_time"`
		TotalTokens      int     `json:"total_tokens"`
		TotalTime        float64 `json:"total_time"`
	} `json:"usage"`
	SystemFingerprint string `json:"system_fingerprint"`
	XGroq             struct {
		ID string `json:"id"`
	} `json:"x_groq"`
}


// type ChatCompletionChunkResponse struct {
// 	ID                string                              `json:"id"`
// 	Object            string                              `json:"object"`
// 	Created           int64                               `json:"created"`
// 	Model             string                              `json:"model"`
// 	SystemFingerprint string                              `json:"system_fingerprint"`
// 	Choices           []ChatCompletionChunkResponseChoice `json:"choices"`
// 	XGroq             struct {
// 		ID string `json:"id"`

// 		Usage ChatCompletionChunkResponseUsage `json:"usage"`
// 	} `json:"x_groq"`
// }
