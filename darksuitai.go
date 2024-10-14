package darksuitai

import (
	_"github.com/darksuit-ai/darksuitai/internal"
	"github.com/darksuit-ai/darksuitai/internal/prompts"
	ai "github.com/darksuit-ai/darksuitai/pkg/chat"
	convai "github.com/darksuit-ai/darksuitai/pkg/convchat"
	"github.com/darksuit-ai/darksuitai/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// Create an instance of the DarkSuitAgent interface
//var darkSuitAgent internal.DarkSuitAgent = internal.NewDarkSuitAgent()

// DarkSuitAI is the main struct that users will interact with
type ChatLLMArgs types.ChatLLMArgs

// NewChatLLMArgs creates a new ChatLLMArgs with default values
func NewChatLLMArgs() *ChatLLMArgs {

	return &ChatLLMArgs{
		ChatSystemInstruction: []byte(``),
		ChatInstruction: prompts.PromptTemplate,
		PromptKeys:      make(map[string][]byte),
		ModelType:       make(map[string]string),
		ModelKwargs: []struct {
			MaxTokens     int      `json:"max_tokens"`
			Temperature   float64  `json:"temperature"`
			Stream        bool     `json:"stream"`
			StopSequences []string `json:"stop_sequences"`
		}{
			{
				MaxTokens:     500,
				Temperature:   0.2,
				Stream:        false,
				StopSequences: []string{},
			},
		},
		APIKey: []byte(``),
	}
}

/*
	AddAPIKey sets the API key for the ChatLLMArgs instance.

This method allows you to securely store the API key required for authenticating requests to the chat model service.

Example:

args := darksuitAI.NewChatLLMArgs()

args.AddAPIKey([]byte("your-api-key"))

In this example, the byte slice containing the API key is set, enabling the chat model to authenticate and process requests.
*/
func (args *ChatLLMArgs) AddAPIKey(apiKey []byte) {
	args.APIKey = apiKey
}

/*
	SetChatSystemInstruction sets the system-level instruction in ChatLLMArgs.

This method allows you to define the overarching system prompt that will guide the chat model's behavior.

Example:

args := darksuitAI.NewChatLLMArgs()

args.SetChatSystemInstruction([]byte("Your system prompt goes here"))

In this example, the byte slice containing the system prompt is set, which will be used by the chat model to maintain context and behavior.
*/
func (args *ChatLLMArgs) SetChatSystemInstruction(systemPrompt []byte) {
	args.ChatSystemInstruction = systemPrompt
}

/*
	SetChatInstruction sets the chat instruction in ChatLLMArgs.

This method allows you to define the main instruction or prompt that will guide the chat model's responses.

Example:

args := darksuitAI.NewChatLLMArgs()

args.SetChatInstruction([]byte("Your chat instruction goes here"))

In this example, the byte slice containing the chat instruction is set, which will be used by the chat model to generate responses.
*/
func (args *ChatLLMArgs) SetChatInstruction(prompt []byte) {
	args.ChatInstruction = prompt
}

/*
	AddPromptKey adds a key-value pair to the PromptKeys map in ChatLLMArgs.

This method allows you to dynamically insert or update prompt-specific variables that can be used within the chat instruction template.

Example:

args := darksuitAI.NewChatLLMArgs()

args.AddPromptKey("year", []byte(`2024`))

args.AddPromptKey("month", []byte(`June`))

In this example, the keys "year" and "month" with their respective values "2024" and "June" are added to the PromptKeys map, which can then be referenced in the chat instruction template.
*/
func (args *ChatLLMArgs) AddPromptKey(key string, value []byte) {
	args.PromptKeys[key] = value
}

/*
	SetModelType sets a key-value pair in the ModelType map in ChatLLMArgs.

This method allows you to specify the type of model to be used for the chat.

Example:

args := darksuitAI.NewChatLLMArgs()

args.SetModelType("openai", "gpt-4o")

In this example, the key "openai" with the value "gpt-4o" is added to the ModelType map, indicating the model type to be used.
*/
func (args *ChatLLMArgs) SetModelType(key, value string) {
	args.ModelType[key] = value
}

/*
	SetMongoDBCollection sets the MongoDB collection in ChatLLMArgs.

This method allows you to specify the MongoDB collection that will be used for storing and retrieving chat-related data.

Example:

args := darksuitAI.NewChatLLMArgs()

args.SetMongoDBCollection(mongoCollection)

In this example, the MongoDB collection is set, which will be used for chat data operations.
*/
func (args *ChatLLMArgs) SetMongoDBCollection(collection *mongo.Database) {
	args.MongoDB = collection
}

/*
	AddModelKwargs adds a new set of model arguments to the ModelKwargs slice in ChatLLMArgs.

This method allows you to specify various parameters for the model's behavior.

Example:

args := darksuitAI.NewChatLLMArgs()

args.AddModelKwargs(500, 0.8, true, []string{"Human:"})

In this example, the model arguments are set with a maximum of 1500 tokens, a temperature of 0.8, streaming enabled, and a stop sequence of "Human:".
*/
func (args *ChatLLMArgs) AddModelKwargs(maxTokens int, temperature float64, stream bool, stopSequences []string) {
	args.ModelKwargs = append(args.ModelKwargs, struct {
		MaxTokens     int      `json:"max_tokens"`
		Temperature   float64  `json:"temperature"`
		Stream        bool     `json:"stream"`
		StopSequences []string `json:"stop_sequences"`
	}{
		MaxTokens:     maxTokens,
		Temperature:   temperature,
		Stream:        stream,
		StopSequences: stopSequences,
	})
}

type LLM struct {
	ai ai.AI
}

type ConvLLM struct {
	convai convai.ConvAI
}

// NewLLM creates a new instance of DarkSuitAI LLM
func (cargs *ChatLLMArgs) NewLLM() (*LLM, error) {

	// Call the dark suit callback
	// darkSuitCallback := darkSuitAgent.WakeDarkSuitAgent()
	// darkSuitCallback()
	return &LLM{
		ai: ai.AI{
			ChatSystemInstruction: cargs.ChatSystemInstruction,
			ChatInstruction: cargs.ChatInstruction,
			PromptKeys:      cargs.PromptKeys,
			ModelType:       cargs.ModelType,
			ModelKwargs:     cargs.ModelKwargs,
			APIKey: cargs.APIKey,
		},
	}, nil
}

// NewConvLLM creates a new instance of DarkSuitAI LLM
func (cargs *ChatLLMArgs) NewConvLLM() (*ConvLLM, error) {
	// Call the dark suit callback
	// darkSuitCallback := darkSuitAgent.WakeDarkSuitAgent()
	// darkSuitCallback()

	return &ConvLLM{
		convai: convai.ConvAI{
			ChatSystemInstruction: cargs.ChatSystemInstruction,
			ChatInstruction: cargs.ChatInstruction,
			PromptKeys:      cargs.PromptKeys,
			ModelType:       cargs.ModelType,
			MongoDB:         cargs.MongoDB,
			ModelKwargs:     cargs.ModelKwargs,
			APIKey: cargs.APIKey,
		},
	}, nil
}

// NewSuitedAgent creates a new instance of DarkSuitAI Agent
func (cargs *ChatLLMArgs) NewSuitedAgent() (*ConvLLM, error) {
	// Call the dark suit callback
	// darkSuitCallback := darkSuitAgent.WakeDarkSuitAgent()
	// darkSuitCallback()

	return &ConvLLM{
		convai: convai.ConvAI{
			ChatSystemInstruction: cargs.ChatSystemInstruction,
			ChatInstruction: cargs.ChatInstruction,
			PromptKeys:      cargs.PromptKeys,
			ModelType:       cargs.ModelType,
			MongoDB:         cargs.MongoDB,
			ModelKwargs:     cargs.ModelKwargs,
			APIKey: cargs.APIKey,
		},
	}, nil
}

// Chat LLM exposes the LLM method for chat
func (d *LLM) Chat(prompt string) (string, error) {
	return d.ai.Chat(prompt)
}

// Stream LLM exposes the LLM method for chat stream
func (d *LLM) Stream(prompt string) <-chan string {
	return d.ai.Stream(prompt)
}

// Chat ConvLLM exposes the LLM method for conversational chat
func (d *ConvLLM) Chat(prompt string) (string, error) {
	return d.convai.Chat(prompt)
}

// Stream ConvLLM exposes the LLM method for conversational chat stream
func (d *ConvLLM) Stream(prompt string)  <-chan string  {
	return d.convai.Stream(prompt)
}
