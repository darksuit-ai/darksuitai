package ai

// Define a common interface that both LLM types will implement.
type LLM interface {
	// Define common methods that both LLMs should have
	StreamCompleteChat(input1 string, input2 string) (string, error)
	StreamChat(input1 string, input2 string) <-chan string
	Chat(input1 string, input2 string) (string, error)
}