package types

type StreamedResponse struct {
	Data string `json:"data"`
}


type ChatCompletionChunkResponseUsage struct {
	QueueTime        float64 `json:"queue_time"`
	PromptTokens     int     `json:"prompt_tokens"`
	PromptTime       float64 `json:"prompt_time"`
	CompletionTokens int     `json:"completion_tokens"`
	CompletionTime   float64 `json:"completion_time"`
	TotalTokens      int     `json:"total_tokens"`
	TotalTime        float64 `json:"total_time"`
}

type ChatCompletionChunkResponseDelta struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type ChatCompletionChunkResponseChoice struct {
	Index        int `json:"index"`
	Delta        ChatCompletionChunkResponseDelta
	Logprobs     interface{} `json:"logprobs"` // Assuming logprobs can be null
	FinishReason interface{} `json:"finish_reason"`
}

type ChatCompletionChunkResponse struct {
	ID                string                              `json:"id"`
	Object            string                              `json:"object"`
	Created           int64                               `json:"created"`
	Model             string                              `json:"model"`
	SystemFingerprint string                              `json:"system_fingerprint"`
	Choices           []ChatCompletionChunkResponseChoice `json:"choices"`
	XGroq             struct {
		ID string `json:"id"`
		Usage ChatCompletionChunkResponseUsage `json:"usage,omitempty"`
	} `json:"x_groq"`
}

type ChatCompletionDone struct{
	Done []string
}