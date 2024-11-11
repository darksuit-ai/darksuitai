package _stream

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/darksuit-ai/darksuitai/internal/memory/mongodb"
	"github.com/darksuit-ai/darksuitai/internal/utilities"
	"github.com/darksuit-ai/darksuitai/pkg/agent"
	"github.com/darksuit-ai/darksuitai/pkg/tools"
)

/*
This script could handle the core logic and decision-making of this LLM agent.
*/

// CallLLMInterface is an interface that defines a single method _callLanguageModel
// which takes a map of strings to byte slices as input and returns an LLMResult.
type callLLMInterface interface {
	_callLanguageModel(queryToolResponsePrompt map[string][]byte) *LLMResult
}

// Constants for maximum iterations and final iteration
const (
	_maxIterations  int = 5
	_finalIteration int = 3
)

var wrongToolSelection = []byte(`You tried to use the tool {tool}, but it doesn't exist. You must use any of these available tools: [{name_of_tools}].`)

// _callLanguageModel is a method that belongs to the Synapse struct. It takes a map of strings to byte slices
// as input and returns an LLMResult. The method performs the following steps:
//
//  1. It calls the MultiModalAgent method to get an instance of the language model (llm), a prompt string,
//     and an optional error (mmerr). If an error occurs, it logs a warning using the System logger.
//
//  2. It creates a channel (llmResponse) to receive the response from the language model and a variable (message)
//     to store the formatted prompt.
//
//  3. It checks if the input map contains a key "question". If it does, it formats the prompt using the
//     CustomFormat function from the utilities package, replacing the "query" placeholder with the value
//     associated with the "question" key.
//
//  4. It calls the StreamChat method of the language model instance (llm) with the formatted prompt and the
//     InstructPrompt field of the Synapse struct as arguments. The StreamChat method likely streams the response
//     from the language model.
//
//  5. It returns an LLMResult struct containing the formatted prompt (message) and the response stream from
//     the language model (llmStream).
//
//  6. If the input map contains a key "plan" instead of "question", it formats the prompt using the
//     CustomFormat function, replacing the "flow_of_thought" placeholder with the value associated with the
//     "plan" key. It then calls the StreamChat method with the formatted prompt and the InstructPrompt field,
//     and returns an LLMResult struct with the formatted prompt and the response stream.
//
// 7. If neither "question" nor "plan" keys are present in the input map, it returns an empty LLMResult struct.
func (prePrompt *AgentPreProgram) _callLanguageModel(queryToolResponsePrompt map[string][]byte) *LLMResult {

	var (
		message   []byte
		llmStream chan string
	)

	// Check if the "question" key exists in the input map
	if question, ok := queryToolResponsePrompt["question"]; ok {
		message = utilities.CustomFormat(prePrompt.BasePrompt, map[string][]byte{"query": question})
		prePrompt.BaseRunnableCaller(message,llmStream)
		for response := range llmStream {
			print(response)
		}
		// Create a new LLMResult instance and return it
		return &LLMResult{Message: message, LLMResponse: llmStream}
	} else if plan, ok := queryToolResponsePrompt["plan"]; ok { // Check if the "plan" key exists in the input map

		// If it exists, format the initial message with the plan
		thoughtWithToolResponse := utilities.CustomFormat([]byte(queryToolResponsePrompt["initial_message"]), map[string][]byte{"flow_of_thought": plan})

		prePrompt.RunnableCaller(thoughtWithToolResponse,llmStream)
		// Create a new LLMResult instance and return it
		return &LLMResult{Message: message, LLMResponse: llmStream}
	}

	return &LLMResult{} // Return an empty LLMResult if neither "question" nor "plan" keys are present
}

func _getToolReturn(agentTools map[string]tools.BaseTool, toolNames, action, actionInput string, toolMeta map[string]interface{}) (string, any, string, error) {
	// Remove leading or trailing punctuation marks from action
	action = strings.Trim(action, ".,!?;:'")
	// Attempt to find the tool in the AllSnapshotTools map
	tool, found := agentTools[action]
	if !found {
		// If the tool is not found, return an error message
		return string(utilities.CustomFormat(wrongToolSelection, map[string][]byte{"tool": []byte(action), "name_of_tools": []byte(toolNames)})), nil, "", nil
	}
	// Execute the tool function with the given input and metadata
	result, rawToolResponse, toolErr := tool.ToolFunc(actionInput, tool.Name, toolMeta)
	if toolErr != nil {
		return "", nil, "", toolErr
	}
	// Return the result, raw response, note data, and the tool's intent ID
	return result, rawToolResponse, tool.Name, nil

}

func (prePrompt *AgentPreProgram) StreamExecutor(queryPrompt map[string][]byte, writer *StreamWriter, maxIterations int, verbose bool) error {

	var (
		initMessage           []byte
		agentThoughtProcesses []byte
		llmResponse           []byte
		toolResponseList      []map[string]interface{}
		clm                   callLLMInterface
		actionReady           bool
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer writer.Close() // Ensure we close the stream writer when done
	print("here")

	prePrompt.AIIdentity = []byte("\nAI: ")

	iterationCount := 0

	clm = prePrompt
	llmStreamData := clm._callLanguageModel(queryPrompt)
	initMessage = llmStreamData.Message
	llmResponse, actionReady = _streamDifferentiator(ctx, writer, llmStreamData)

	// Main processing loop with proper iteration control
	for actionReady && iterationCount < maxIterations {
		iterationCount++
		agentActionTypes, _, neupErr := agent.NeuralParser(llmResponse, true)

		if neupErr != nil {
			return neupErr
		}

		// Process agent actions
		if action, exists := agentActionTypes.AgentAction["Action"]; exists {

			// Get tool response
			toolResponse, rawToolResponse, toolName, err := _getToolReturn(prePrompt.Tools, prePrompt.ToolNames, string(action), string(agentActionTypes.AgentAction["Input"]), prePrompt.AdditionalToolsMeta)

			if err != nil {
				return err
			}

			if verbose {
				utilities.Printer("Observation: ", toolResponse, "purple")
			}

			// Process tool response
			if toolResponse != "" {

				// Append the raw tool response bytes to toolResponseList
				toolResponseList = append(toolResponseList, map[string]interface{}{toolName: rawToolResponse})

				if verbose {
					utilities.Printer("Observation: ", toolResponse, "purple")
				}

				// Build next prompt
				agentThoughtProcesses = append(agentThoughtProcesses, prePrompt.AIIdentity...)
				agentThoughtProcesses = append(agentThoughtProcesses, llmResponse...)
				agentThoughtProcesses = append(agentThoughtProcesses, []byte("\nObservation: ")...)
				agentThoughtProcesses = append(agentThoughtProcesses, []byte(toolResponse)...)
				agentThoughtProcesses = append(agentThoughtProcesses, prePrompt.AIIdentity...)

				// Get next LLM response
				agentPlan := agent.AgentActionTypes{
					AgentPlan: map[string][]byte{
						"plan":            agentThoughtProcesses,
						"initial_message": initMessage,
					},
				}

				llmStreamData = clm._callLanguageModel(agentPlan.AgentPlan)
				llmResponse, actionReady = _streamDifferentiator(ctx, writer, llmStreamData)

				// Break if no more actions needed
				if !actionReady {
					break
				}
			}
		} else {
			// No action found, break the loop
			break
		}
	}
return nil
}

// hashToolResponse generates a SHA-256 hash of the given toolResponse and returns it as a hexadecimal string.
func hashToolResponse(toolResponse string) string {
	// Create a new SHA-256 hash.
	h := sha256.New()
	// Write the thought byte slice to the hash.
	h.Write([]byte(toolResponse))
	// Compute the final hash and return it as a hexadecimal string.
	return fmt.Sprintf("%x", h.Sum(nil))
}


func (prePrompt *AgentPreProgram) SaveChatHistory(query,finishText,sessionId string){
	if prePrompt.ChatMemoryCollection != nil {
		var mongoMemory mongodb.ChatMemoryCollectionInterface = mongodb.NewMongoCollection(prePrompt.ChatMemoryCollection)
		mongoMemory.AddConversationToMemory(sessionId, query, finishText)
	}
}