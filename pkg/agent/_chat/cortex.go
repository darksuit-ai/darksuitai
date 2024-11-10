package _chat

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"

	"github.com/darksuit-ai/darksuitai/internal/memory/mongodb"
	"github.com/darksuit-ai/darksuitai/internal/utilities"
	"github.com/darksuit-ai/darksuitai/pkg/agent"
	"github.com/darksuit-ai/darksuitai/pkg/tools"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

/*
This script could handle the core logic and decision-making of this LLM agent.
*/

// CallLLMInterface is an interface that defines a single method _callLanguageModel
// which takes a map of strings to byte slices as input and returns an LLMResult.
type callLLMInterface interface {
	_callLanguageModel(queryToolResponsePrompt map[string][]byte) ([]byte, []byte, error)
}

var wrongToolSelection = []byte(`You tried to use the tool {tool}, but it doesn't exist. You must use any of these available tools: [{name_of_tools}].`)

// _callLanguageModel is a helper function that calls the LLM with the appropriate prompt.
// It takes a queryToolResponsePrompt map[string]string as input, which contains either the user's question or the agent's plan.
// It returns the initial message and the LLM's response.
// _callLanguageModel is a method of the Synapse struct that takes a map of strings to byte slices as input.
// It returns two byte slices.
func (prePrompt *AgentPreProgram) _callLanguageModel(queryToolResponsePrompt map[string][]byte) ([]byte, []byte, error) {

	var (
		message     []byte
		llmResponse string
		llmErr      error
	)

	// Check if the "question" key exists in the input map
	if question, ok := queryToolResponsePrompt["question"]; ok {
		message = utilities.CustomFormat(prePrompt.BasePrompt, map[string][]byte{"query": question})
		llmResponse, llmErr = prePrompt.BaseRunnableCaller(message)
		if llmErr != nil {
			return nil, nil, llmErr
		}
		// Return the formatted message and the llmResponse
		return []byte(message), []byte(llmResponse), nil
	} else if plan, ok := queryToolResponsePrompt["plan"]; ok { // Check if the "plan" key exists in the input map

		// If it exists, format the initial message with the plan
		thoughtWithToolResponse := utilities.CustomFormat([]byte(queryToolResponsePrompt["initial_message"]), map[string][]byte{"flow_of_thought": plan})

		llmResponse, llmErr = prePrompt.RunnableCaller(thoughtWithToolResponse)
		if llmErr != nil {
			return nil, nil, llmErr
		}
		// Return empty byte slices and the llmPlanResponse
		return []byte{}, []byte(llmResponse), nil
	}

	// If neither "question" nor "plan" keys exist in the input map, return empty byte slices
	return []byte{}, []byte{}, nil
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

func (prePrompt *AgentPreProgram) Executor(query_prompt map[string][]byte, sessionId string, maxIterations int, verbose bool) ([]byte, any, error) {

	var (
		wg sync.WaitGroup
		// Initialize a byte slice to store the agent's thought processes
		agentThoughtProcesses []byte

		// Initialize a byte slice to store the LLM's response
		llmResponse []byte

		// Initialize a byte slice to store the initial message
		initMessage []byte

		// Initialize a list to store the responses from the tools
		toolResponseList []interface{}

		callErr error
		// Initialize the llm interface
		clm callLLMInterface
	)

	prePrompt.AIIdentity = []byte("\nAI: ")

	// Anti-looping safeguards
	lastToolResponseHash := ""

	clm = prePrompt
	// Call the LLM with the query prompt and store the initial message and LLM's response
	initMessage, llmResponse, callErr = clm._callLanguageModel(query_prompt)

	if callErr != nil {
		return nil, nil, callErr
	}

	// TODO: Allow LLM call for action determine iteration count by updating the maxIteraion var everytime
	//  it calls action and reset to zero when it calls final answer
	for i := 0; i < maxIterations; i++ {
		var agentActionTypes *agent.AgentActionTypes

		agentActionTypes, _, neupErr := agent.NeuralParser(llmResponse, true)

		if neupErr != nil {
			return nil, nil, neupErr
		}

		// Check if maxIterations is 3 and either AgentFinish or AgentAction is nil, stop the cortext from thinking
		if i == 3 && (agentActionTypes.AgentFinish == nil || agentActionTypes.AgentAction == nil) {
			return []byte(`🥺 I missed this one. Can you ask me again?`), nil, nil
		}

		if finish, ok := agentActionTypes.AgentFinish["Output"]; ok {
			if verbose {
				utilities.Printer("", string(finish), "green")
			}

			finish = bytes.ReplaceAll(finish, []byte("<answer>"), []byte(""))
			finish = bytes.ReplaceAll(finish, []byte("</answer>"), []byte(""))

			if prePrompt.ChatMemoryCollection != nil {
				// Create local copies of variables needed in goroutine
				memoryCollection := prePrompt.ChatMemoryCollection
				questionCopy := string(query_prompt["question"])
				finishCopy := string(finish)
				// Save the conversation to memory in a separate goroutine
				wg.Add(1)
				go func(collection *mongo.Collection, question, finishText string) {
					defer wg.Done()
					var mongoMemory mongodb.ChatMemoryCollectionInterface = mongodb.NewMongoCollection(collection)
					mongoMemory.AddConversationToMemory(sessionId, question, finishText)

				}(memoryCollection, questionCopy, finishCopy)
			}

			if toolResponseList != nil {
				return finish, toolResponseList, nil
			}

			return finish, []string{}, nil
		}

		// TODO: allow the agentActionTypes.AgentAction["Action"] determine number of iterations
		if action, ok := agentActionTypes.AgentAction["Action"]; ok {

			toolResponse, rawToolResponse, toolName, err := _getToolReturn(prePrompt.Tools, prePrompt.ToolNames, string(action), string(agentActionTypes.AgentAction["Input"]), prePrompt.AdditionalToolsMeta)

			if err != nil {
				return nil, nil, err
			}
			// Append the raw tool response bytes to toolResponseList
			toolResponseList = append(toolResponseList, map[string]interface{}{toolName: rawToolResponse})

			if verbose {
				utilities.Printer("Observation: ", toolResponse, "purple")
			}

			// Anti-looping check: Compare current toolResponse with the last one
			currentToolResponseHash := hashToolResponse(toolResponse)
			if currentToolResponseHash == lastToolResponseHash {
				log.Println("Detected repeated toolResponse. Breaking loop.")
				agentThoughtProcesses = append(agentThoughtProcesses, llmResponse...)
				agentThoughtProcesses = append(agentThoughtProcesses, []byte("Observation: ")...)
				agentThoughtProcesses = append(agentThoughtProcesses, []byte(`Sorry about that I got abit distracted. Please ask me again`)...)
				agentThoughtProcesses = append(agentThoughtProcesses, prePrompt.AIIdentity...)
				agentPlan := agent.AgentActionTypes{
					AgentPlan: map[string][]byte{
						"plan":            agentThoughtProcesses,
						"initial_message": initMessage,
					},
				}

				_, llmResponse, callErr = prePrompt._callLanguageModel(agentPlan.AgentPlan)

				if callErr != nil {
					return nil, nil, callErr
				}
			}

			lastToolResponseHash = currentToolResponseHash
			if toolResponse != "" {
				// Build next prompt
				agentThoughtProcesses = append(agentThoughtProcesses, prePrompt.AIIdentity...)
				agentThoughtProcesses = append(agentThoughtProcesses, llmResponse...)
				agentThoughtProcesses = append(agentThoughtProcesses, []byte("\nObservation: ")...)
				agentThoughtProcesses = append(agentThoughtProcesses, []byte(toolResponse)...)
				agentThoughtProcesses = append(agentThoughtProcesses, prePrompt.AIIdentity...)

				agentPlan := agent.AgentActionTypes{
					AgentPlan: map[string][]byte{
						"plan":            agentThoughtProcesses,
						"initial_message": initMessage,
					},
				}
				_, llmResponse, callErr = prePrompt._callLanguageModel(agentPlan.AgentPlan)

				if callErr != nil {
					return nil, nil, callErr
				}
			}
		}
		// TODO: allow the agentActionTypes.AgentAction["Action"] determine number of iterations
		if errorMessage, ok := agentActionTypes.AgentError["Error"]; ok {
			if verbose {
				utilities.Printer("Warning: ", string(errorMessage), "orange")
			}

			// Build next prompt
			agentThoughtProcesses = append(agentThoughtProcesses, prePrompt.AIIdentity...)
			agentThoughtProcesses = append(agentThoughtProcesses, llmResponse...)
			agentThoughtProcesses = append(agentThoughtProcesses, []byte("\nSystemError: ")...)
			agentThoughtProcesses = append(agentThoughtProcesses, errorMessage...)
			agentThoughtProcesses = append(agentThoughtProcesses, prePrompt.AIIdentity...)

			agentPlan := agent.AgentActionTypes{
				AgentPlan: map[string][]byte{
					"plan":            agentThoughtProcesses,
					"initial_message": initMessage,
				},
			}
			_, llmResponse, callErr = prePrompt._callLanguageModel(agentPlan.AgentPlan)

			if callErr != nil {
				return nil, nil, callErr
			}

		}
	}
	wg.Wait()
	fmt.Println("Iteration limit exceeded")
	return nil, nil, nil
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
