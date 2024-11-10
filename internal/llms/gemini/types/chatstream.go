package types

// Define the data model
type ChatCompletionChunk struct {
	ID               string    `json:"id"`
	Object           string    `json:"object"`
	Created          int64     `json:"created"`
	Model            string    `json:"model"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Choices          []Choice  `json:"choices"`
}

type Choice struct {
	Index        int    `json:"index"`
	Delta        Delta  `json:"delta"`
	LogProbs     *string `json:"logprobs"`
	FinishReason *string `json:"finish_reason"`
}

type Delta struct {
	Role    string  `json:"role"`
	Content string  `json:"content"`
}
