package types

type ChatOverloadError struct {
	Type  string `json:"error"`
	Error map[string]string
}

type StreamedResponse struct {
	Data string `json:"data"`
}

type ContentBlockStop struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type ContentBlockDelta struct {
	Type         string            `json:"type"`
	Index         int            `json:"index"`
	Delta        map[string]string `json:"delta"`
}
