package types

import "go.mongodb.org/mongo-driver/mongo"

type ChatLLMArgs struct {
	ChatSystemInstruction []byte
	ChatInstruction       []byte
	PromptKeys            map[string][]byte
	ModelType             map[string]string
	MongoDB               *mongo.Database
	ModelKwargs           []struct {
		MaxTokens     int      `json:"max_tokens"`
		Temperature   float64  `json:"temperature"`
		Stream        bool     `json:"stream"`
		StopSequences []string `json:"stop_sequences"`
	}
}
