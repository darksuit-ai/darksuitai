package _stream

import (
	"github.com/darksuit-ai/darksuitai/pkg/tools"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
	"sync"
	"io"
)

type AgentPreProgram struct {
	BasePrompt           []byte
	SystemPrompt         []byte
	Tools                map[string]tools.BaseTool
	ToolNames            string
	AdditionalToolsMeta  map[string]interface{}
	BaseRunnableCaller   func(prompt []byte) <-chan string
	RunnableCaller       func(promptIterable []byte) <-chan string
	AIIdentity           []byte
	ChatMemoryCollection *mongo.Collection
	MaxIteration         int
	Verbose              bool
	SessionId            string
}


// LLMResult encapsulates the stream data from the channel and the complete prompt message for the cortex.
type LLMResult struct {
	// Message holds the complete prompt message sent to the Neuron
	Message []byte
	// LLMResponse is the channel that passes each stream text from the LLM's response
	LLMResponse <-chan string
}

type StreamWriter struct {
    Builder *strings.Builder
    Ch     chan string
	Done   chan struct{}  // Signal for completion
    Once   sync.Once      // Ensure single close
	Wg     sync.WaitGroup
	    // Track if we've seen the opening tag
		SeenOpenTag  bool
		// Buffer to handle tag removal across chunks
		Buffer      strings.Builder
}

func (sw *StreamWriter) Write(p []byte) (n int, err error) {
	select {
    case <-sw.Done:
        return 0, io.ErrClosedPipe
    default:
		// cleanData := sw.processStream(p)
		// if cleanData != nil {
        //     sw.Ch <- string(cleanData)
        // }
		sw.Ch <- string(p)
		return sw.Builder.Write(p)
}
}
func (sw *StreamWriter) Close() {
    sw.Once.Do(func() {
        close(sw.Done)
        close(sw.Ch)
		sw.Wg.Wait()
    })
}
