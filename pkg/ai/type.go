package ai

type AI struct {
	ChatInstruction []byte
	PromptKeys      map[string][]byte
	ModelType       map[string]string
	ModelKwargs     []struct {
		MaxTokens     int    `json:"max_tokens"`
		Temperature   float64  `json:"temperature"`
		Stream        bool     `json:"stream"`
		StopSequences []string `json:"stop_sequences"`
	}
}
