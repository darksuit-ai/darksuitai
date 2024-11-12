package _stream

import (
	"bytes"
	"io"
	"strings"
	"sync"

	"github.com/darksuit-ai/darksuitai/pkg/tools"
	"go.mongodb.org/mongo-driver/mongo"
)

type AgentPreProgram struct {
	BasePrompt           []byte
	SystemPrompt         []byte
	Tools                map[string]tools.BaseTool
	ToolNames            string
	AdditionalToolsMeta  map[string]interface{}
	BaseRunnableCaller   func(prompt []byte, ipcChan chan string)
	RunnableCaller       func(promptIterable []byte, ipcChan chan string)
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
	LLMResponse chan string
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

func (sw *StreamWriter) processStream(input []byte) []byte {
    // If we haven't seen the opening tag yet
    if !sw.SeenOpenTag {
        if idx := bytes.Index(input, []byte(`<answer>`)); idx != -1 {
            sw.SeenOpenTag = true
            // Return everything after the opening tag
            return append(bytes.TrimPrefix(input[idx:], []byte(`<answer>`)), ' ')
        }
        return []byte(``)
    }

    return input
}

func (sw *StreamWriter) Write(p []byte) (n int, err error) {
	select {
    case <-sw.Done:
        return 0, io.ErrClosedPipe
    default:
		cleanData := sw.processStream(p)
		if cleanData != nil {
            sw.Ch <- string(cleanData)
        }
		// sw.Ch <- string(p)
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
