package _stream

// import (
// 	"context"
// 	"crypto/sha256"
// 	"fmt"
// 	"strconv"
// 	"strings"
// 	_ "sync"
// 	"time"
// )

// /*
// This script could handle the core logic and decision-making of this LLM agent.
// */

// // CallLLMInterface is an interface that defines a single method _callLanguageModel
// // which takes a map of strings to byte slices as input and returns an LLMResult.
// type callLLMInterface interface {
// 	_callLanguageModel(queryToolResponsePrompt map[string][]byte) *LLMResult
// }

// // Constants for maximum iterations and final iteration
// const (
// 	_maxIterations  int = 5
// 	_finalIteration int = 3
// )

// // _callLanguageModel is a method that belongs to the Synapse struct. It takes a map of strings to byte slices
// // as input and returns an LLMResult. The method performs the following steps:
// //
// //  1. It calls the MultiModalAgent method to get an instance of the language model (llm), a prompt string,
// //     and an optional error (mmerr). If an error occurs, it logs a warning using the System logger.
// //
// //  2. It creates a channel (llmResponse) to receive the response from the language model and a variable (message)
// //     to store the formatted prompt.
// //
// //  3. It checks if the input map contains a key "question". If it does, it formats the prompt using the
// //     CustomFormat function from the utilities package, replacing the "query" placeholder with the value
// //     associated with the "question" key.
// //
// //  4. It calls the StreamChat method of the language model instance (llm) with the formatted prompt and the
// //     InstructPrompt field of the Synapse struct as arguments. The StreamChat method likely streams the response
// //     from the language model.
// //
// //  5. It returns an LLMResult struct containing the formatted prompt (message) and the response stream from
// //     the language model (llmStream).
// //
// //  6. If the input map contains a key "plan" instead of "question", it formats the prompt using the
// //     CustomFormat function, replacing the "flow_of_thought" placeholder with the value associated with the
// //     "plan" key. It then calls the StreamChat method with the formatted prompt and the InstructPrompt field,
// //     and returns an LLMResult struct with the formatted prompt and the response stream.
// //
// // 7. If neither "question" nor "plan" keys are present in the input map, it returns an empty LLMResult struct.
// func (synapse *Synapse) _callLanguageModel(queryToolResponsePrompt map[string][]byte) *LLMResult {
// 	// Get the language model instance, prompt, and optional error
// 	args, prompt, mmerr := synapse.MultiModalAgent()
// 	if mmerr != nil {
// 		exp.Loggers.System.Warn(mmerr) // Log a warning if an error occurs
// 	}
// 	var message []byte // Variable to store the formatted prompt

// 	// Check if the input map contains the "question" key
// 	if question, ok := queryToolResponsePrompt["question"]; ok {
// 		message = utilities.CustomFormat([]byte(prompt), map[string][]byte{"query": question})
// 		args.SetChatInstruction(prompt)
// 		args.SetChatSystemInstruction(synapse.InstructPrompt)
// 		args.AddPromptKey("partner_data", synapse.RagData)
// 		llm, llmErr := args.NewLLM()
// 		if llmErr != nil {
// 			exp.Loggers.System.Warn(llmErr)
// 		}

// 		llmStream := llm.Stream(string(question))

// 		// Create a new LLMResult instance and return it
// 		return &LLMResult{Message: message, LLMResponse: llmStream}
// 	} else if plan, ok := queryToolResponsePrompt["plan"]; ok { // Check if the input map contains the "plan" key
// 		thoughtWithToolResponse := utilities.CustomFormat(queryToolResponsePrompt["initial_message"], map[string][]byte{"flow_of_thought": plan})
// 		args.SetChatInstruction(thoughtWithToolResponse)
// 		args.SetChatSystemInstruction(synapse.InstructPrompt)
// 		llm, llmErr := args.NewLLM()
// 		if llmErr != nil {
// 			exp.Loggers.System.Warn(llmErr)
// 		}
// 		// fmt.Println(string(thoughtWithToolResponse))
// 		llmStream := llm.Stream("")
// 		// Create a new LLMResult instance and return it
// 		return &LLMResult{Message: message, LLMResponse: llmStream}
// 	}

// 	return &LLMResult{} // Return an empty LLMResult if neither "question" nor "plan" keys are present
// }

// // _getToolReturn is a function that takes an action string, an actionInput string, and an actionMeta map[string]interface{}
// // as input parameters. It returns a string, an interface{}, a types.NotePad, and a string.
// //
// // The purpose of this function is to retrieve the appropriate tool function based on the provided action string
// // and execute it with the given actionInput and actionMeta. It then returns the result of the tool function
// // along with the raw tool response, any note data generated, and the intent ID of the tool.
// //
// // Here's a breakdown of the function's logic:
// //
// //  1. It checks if the provided action string exists in the multimodal.Alltools.AllTools map. If not found,
// //     it returns a custom formatted error message indicating that the selected tool is wrong, along with a list
// //     of available tool names. It also returns nil for the raw tool response, an empty types.NotePad, and an empty string
// //     for the intent ID.
// //
// // 2. If the tool is found, it retrieves the corresponding tool function from the map.
// //
// //  3. It executes the tool function by calling tool.ToolFunc(actionInput, actionMeta). This function likely performs
// //     some operation based on the provided actionInput and actionMeta, and returns the result, raw tool response,
// //     and any note data generated.
// //
// // 4. The function then returns the result, raw tool response, note data, and the intent ID of the tool.
// //
// // Parameters:
// // - action (string): The name or identifier of the tool to be executed.
// // - actionInput (string): The input data or prompt to be provided to the tool function.
// // - actionMeta (map[string]interface{}): Additional metadata or parameters required by the tool function.
// //
// // Returns:
// // - string: The result of executing the tool function.
// // - interface{}: The raw response from the tool function.
// // - types.NotePad: Any note data generated by the tool function.
// // - string: The intent ID of the selected tool.
// func _getToolReturn(partnerTools typesv2.agentTools, toolNames, action, actionInput string, actionMeta map[string]interface{}) (string, any, types.NotePad, string) {
// 	// Remove leading or trailing punctuation marks from action
// 	action = strings.Trim(action, ".,!?;:'")

// 	// Attempt to find the requested tool in the AllagentTools map
// 	tool, found := partnerTools.AllagentTools[action]
// 	if !found {
// 		// If the tool is not found, return the error message with the requested tool name and available tool names
// 		return string(utilities.CustomFormat(utilities.WrongToolSelection, map[string][]byte{`tool`: []byte(action), `name_of_tools`: []byte(toolNames)})), nil, types.NotePad{}, ""
// 	}
// 	// If the tool is found, call its ToolFunc with the provided input and metadata
// 	result, rawToolResponse, noteData := tool.ToolFunc(actionInput, tool.Name, actionMeta)
// 	// Return the result, raw tool response, note data, and the tool's intent ID
// 	return result, rawToolResponse, noteData, tool.IntentId
// }

// // _sendToolFeedback sends feedback to the provided agentResponseChan channel.
// // It creates a new FeedbackSender, uses a cancelable context, and handles any errors that occur during the process.
// func _sendToolFeedback(agentResponseChan chan<- string, feedback []byte) {
// 	sender := NewFeedbackSender()
// 	// Use a cancelable context
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()
// 	sender.SendToolFeedback(ctx, agentResponseChan, feedback)

// }

// // MultiModalCortex is a method that belongs to the Synapse struct. It takes a map of string to byte slice (query_prompt),
// // a channel to send 's responses (agentResponseChan), and a channel to send the list of tool responses (toolListChan).
// // The purpose of this method is to orchestrate the multi-modal interaction between the language model, the neural parser,
// // and the various tools available in the system.
// //
// // Here's a breakdown of the method's logic:
// //
// //  1. It initializes several variables to store intermediate data, such as the initial message, agent thought processes,
// //     agent telemetry thought, LLM response, tool response list, and an instance of the CallLLMInterface.
// //
// //  2. It calls the _callLanguageModel method with the provided query_prompt to obtain the initial LLM response.
// //     This response is then processed using the _streamDifferentiator function and printed to the console.
// //
// // 3. It enters a loop based on the iterationCount variable, which is initially set to 1.
// //
// //  4. Inside the loop, it calls the NeuralParser function with the LLM response to obtain the agent action types and thought.
// //     If an error occurs during this process, it is logged to the system logger.
// //
// //  5. If an "Action" is present in the agent action types, it performs the following steps:
// //     a. Updates the iterationCount based on the "IterationValue" provided by the agent action types.
// //     b. Sends a feedback message through the agentResponseChan.
// //     c. Appends the current thought to the agentTelemetryThought.
// //     d. Calls the _getToolReturn function with the specified action, input, and metadata to execute the corresponding tool.
// //     e. Marshals the raw tool response and appends it to the toolResponseList.
// //     f. Prints the tool response to the console.
// //     g. If the tool response is not empty, it appends the LLM response, observation, and thought to the agentThoughtProcesses.
// //     h. Calls the _callLanguageModel method with the updated agentPlan to obtain a new LLM response.
// //     i. Saves any note data generated by the tool using the cUT.SaveNote function.
// //
// // 6. After the loop, it saves the telemetry data using the SaveTelemetryData function from the utilities package.
// //
// //  7. Finally, it sends the tool response list through the toolListChan and closes the channel.
// //     It also prints the total execution time of the MultiModalCortex method.
// //
// // Parameters:
// // - queryPrompt (map[string][]byte): A map containing the initial query or prompt for the language model.
// // - agentResponseChan (chan<- string): A channel to send 's responses to the client.
// // - toolListChan (chan<- interface{}): A channel to send the list of tool responses to the client.
// func (synapse *Synapse) MultiModalCortex(queryPrompt map[string][]byte, agentResponseChan chan<- string, toolListChan chan<- []map[string]interface{}, debug bool) {

// 	var (
// 		// Initialize a byte slice to store the initial message
// 		initMessage []byte
// 		// Initialize a byte slice to store the agent's thought process
// 		agentThoughtProcesses []byte
// 		// Initialize a byte slice to store the agent's thought process for telemetry
// 		agentTelemetryThought []byte
// 		// Initialize a byte to store the llm response
// 		llmResponse []byte
// 		// Initialize a list to store the responses from the tools
// 		toolResponseList []map[string]interface{}
// 		// Initialize the llm interface
// 		clm         callLLMInterface
// 		actionReady bool
// 	)
// 	// Create a context with cancellation
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel() // Ensure all spawned goroutines are cancelled when we return

// 	startTime := time.Now()
// 	utilities.Printer("", "\n🧠 thinking through... ✨ 🌟 ✨\n", "")

// 	// Anti-looping safeguards
// 	lastToolResponseHash := ""

// 	llmStart := time.Now()
// 	clm = synapse

// 	llmStreamData := clm._callLanguageModel(queryPrompt)
	
// 	if debug {
// 		fmt.Printf("_callLanguageModel took >> %v\n", time.Since(llmStart))
// 	}
// 	initMessage = llmStreamData.Message
// 	llmResponse, actionReady = _streamDifferentiator(ctx, agentResponseChan, []string{}, llmStreamData)
// 	utilities.Printer("", string(llmResponse), "green")

// 	if actionReady {
// 		// begin roundtrip thought and action process
// 		// TODO: set a count on system prompt length, break loop into multi-goroutines and converge using channels
// 		// Initialize a counter to control agent action iteration
// 		for iterationCount := 1; iterationCount == 1; {
// 			// reset iteration count to zero to stop excessive looping
// 			iterationCount = 0

// 			var (
// 				agentActionTypes *types.AgentActionTypes
// 				thought          []byte
// 				neupErr          error
// 				llmStreamData    *LLMResult
// 				// wg               sync.WaitGroup
// 			)

// 			neuralStart := time.Now()
// 			agentActionTypes, thought, neupErr = agent.NeuralParser(llmResponse)
// 			if debug {
// 				fmt.Printf("NeuralParser took >> %v\n", time.Since(neuralStart))
// 			}
// 			if neupErr != nil {
// 				exp.Loggers.System.Warn(neupErr)
// 			}

// 			// Check if maxIterations is 7, stop the cortext from thinking
// 			if iterationCount >= _maxIterations {
// 				agentThoughtProcesses = append(agentThoughtProcesses, llmResponse...)
// 				agentThoughtProcesses = append(agentThoughtProcesses, []byte("\nObservation: ")...)
// 				agentThoughtProcesses = append(agentThoughtProcesses, utilities.SystemFailed...)
// 				agentThoughtProcesses = append(agentThoughtProcesses, []byte("\nThought: ")...)

// 				agentPlan := types.AgentActionTypes{
// 					AgentPlan: map[string][]byte{
// 						"plan":            agentThoughtProcesses,
// 						"initial_message": initMessage,
// 					},
// 				}
// 				llmStreamData = clm._callLanguageModel(agentPlan.AgentPlan)
// 				// llmStreamData := clm._callLanguageModelSim(agentPlan.AgentPlan)

// 				llmResponse, actionReady = _streamDifferentiator(ctx, agentResponseChan, []string{}, llmStreamData)
// 				if !actionReady {
// 					break
// 				}

// 			}

// 			// If the agent has an action to perform, execute it
// 			if action, exists := agentActionTypes.AgentAction["Action"]; exists {

// 				// the agent actions determine number of iterations
// 				// Update iteration count based on agent's response
// 				if newIterationValue, err := strconv.Atoi(string(agentActionTypes.AgentAction["IterationValue"])); err == nil {
// 					iterationCount = newIterationValue
// 				}

// 				utilities.Printer("\n", string(agentActionTypes.AgentAction["Feedback"]), "blue")

// 				// Send the agent's feedback to the agentResponseChan in a separate goroutine, now with context and WaitGroup
// 				go func() {
// 					// defer wg.Done()
// 					_sendToolFeedback(agentResponseChan, agentActionTypes.AgentAction["Feedback"])
// 				}()

// 				//time.Sleep(300 * time.Millisecond)

// 				//wg.Wait()
// 				// Append the thought to agentTelemetryThought for telemetry purposes
// 				agentTelemetryThought = append(agentTelemetryThought, thought...)

// 				toolStart := time.Now()
// 				toolResponse, rawToolResponse, NotepadData, toolIntentId := _getToolReturn(synapse.PartnerTools, synapse.ToolNames, string(action), string(agentActionTypes.AgentAction["Input"]), synapse.MetaData)
// 				if debug {
// 					fmt.Printf("\n_getToolReturn took >> %v\n", time.Since(toolStart))
// 				}

// 				// Append the raw tool response bytes to toolResponseList
// 				toolResponseList = append(toolResponseList, map[string]interface{}{toolIntentId: rawToolResponse})

// 				utilities.Printer("Observation: ", toolResponse, "purple")
// 				// Anti-looping check: Compare current toolResponse with the last one
// 				currentToolResponseHash := hashToolResponse(toolResponse)
// 				if currentToolResponseHash == lastToolResponseHash {
// 					exp.Loggers.System.Warn("Detected repeated toolResponse. Breaking loop.")
// 					agentResponseChan <- string(utilities.SystemFailed)
// 				}

// 				lastToolResponseHash = currentToolResponseHash

// 				if toolResponse != "" {
// 					// save notes if sessionId is provided
// 					if sessionId, ok := synapse.MetaData["systemInfo"].(map[string]string)["sessionId"]; ok {
// 						if NotepadData.Body != "" && NotepadData.Body != "No available data right now!" {
// 							rebasedNotepadData, rebaseErr := rebase.GenerateRebasedNotepadData(synapse.MetaData["systemInfo"].(map[string]string)["partnerId"],
// 								synapse.MetaData["querySentence"].(string),
// 								synapse.MetaData["notepadRebaseModelSelection"].(map[string]string),
// 								NotepadData)
// 							if rebaseErr != nil {
// 								exp.Loggers.System.Warn(rebaseErr)
// 							}
// 							go uT.SaveNote(rebasedNotepadData, synapse.MetaData["systemInfo"].(map[string]string)["partnerId"], sessionId)
// 						}
// 					}
// 					// agentThoughtProcesses = append(agentThoughtProcesses, llmResponse...)
// 					agentThoughtProcesses = append(agentThoughtProcesses, []byte("\nObservation: ")...)
// 					agentThoughtProcesses = append(agentThoughtProcesses, []byte(toolResponse)...)
// 					agentThoughtProcesses = append(agentThoughtProcesses, []byte("\nThought: ")...)

// 					agentPlan := types.AgentActionTypes{
// 						AgentPlan: map[string][]byte{
// 							"plan":            agentThoughtProcesses,
// 							"initial_message": initMessage,
// 						},
// 					}

// 					llmPlanStart := time.Now()
// 					llmStreamData = clm._callLanguageModel(agentPlan.AgentPlan)
// 					// llmStreamData := clm._callLanguageModelSim(agentPlan.AgentPlan)

// 					llmResponse, actionReady = _streamDifferentiator(ctx, agentResponseChan, []string{}, llmStreamData)
// 					llmPlanDuration := time.Since(llmPlanStart)
// 					if debug {
// 						fmt.Printf("\n_callLanguageModel (agentPlan) took >> %v", llmPlanDuration)
// 					}
// 					if !actionReady {
// 						break
// 					}
// 				}

// 			}
// 			// If the agent encounters an error, handle it
// 			if agentError, ok := agentActionTypes.AgentError["Error"]; ok {
// 				// the agent actions determine number of iterations
// 				iterationCheck, anyErr := strconv.Atoi(string(agentActionTypes.AgentError["IterationValue"]))
// 				if anyErr != nil {
// 					exp.Loggers.System.Warn(anyErr)
// 				}
// 				iterationCount = iterationCheck
// 				utilities.Printer("Observation: ", string(agentError), "purple")

// 				agentThoughtProcesses = append(agentThoughtProcesses, llmResponse...)
// 				agentThoughtProcesses = append(agentThoughtProcesses, []byte("\nObservation: ")...)
// 				agentThoughtProcesses = append(agentThoughtProcesses, agentError...)
// 				agentThoughtProcesses = append(agentThoughtProcesses, []byte("\nThought: ")...)

// 				agentPlan := types.AgentActionTypes{
// 					AgentPlan: map[string][]byte{
// 						"plan":            agentThoughtProcesses,
// 						"initial_message": initMessage,
// 					},
// 				}
// 				// Measure time taken by _callLanguageModel for agentPlan
// 				llmPlanStart := time.Now()
// 				llmStreamData := clm._callLanguageModel(agentPlan.AgentPlan)
// 				// llmStreamData := clm._callLanguageModelSim(agentPlan.AgentPlan)
// 				llmResponse, actionReady = _streamDifferentiator(ctx, agentResponseChan, []string{}, llmStreamData)
// 				llmPlanDuration := time.Since(llmPlanStart)
// 				if debug {
// 					fmt.Printf("_callLanguageModel (agentPlan) took >> %v\n", llmPlanDuration)
// 				}
// 				if !actionReady {
// 					break
// 				}

// 			}
// 			// Create telemetry data for analysis and model improvement
// 			telemetryData := &utilities.TelemetryData{
// 				UserId:     synapse.MetaData["systemInfo"].(map[string]string)["partnerId"],
// 				UserPrompt: string(queryPrompt["question"]),
// 				AiThoughts: string(agentTelemetryThought),
// 			}
// 			// Save the telemetry data
// 			telemetryData.SaveTelemetryData()
// 		}
// 	}
// 	if len(toolResponseList) > 0 {
// 		select {
// 		case toolListChan <- toolResponseList:
// 			// Successfully sent the tool responses
// 		default:
// 			// If the channel is not ready to receive, handle it here
// 			exp.Loggers.System.Warn("Failed to send tool responses: channel not ready")
// 		}
// 	} else {
// 		select {
// 		case toolListChan <- []map[string]interface{}{}:
// 			// Successfully sent the empty list
// 		default:
// 			// If the channel is not ready to receive, handle it here
// 			exp.Loggers.System.Warn("Failed to send empty tool responses: channel not ready")
// 		}
// 	}
// 	utilities.Printer("", fmt.Sprintf("\n\n🧠 thought completed in %v\n", time.Since(startTime)), "")
// 	if debug {
// 		fmt.Printf("\nMultiModalCortex took >> %v\n", time.Since(startTime))
// 	}
// }

// // hashToolResponse generates a SHA-256 hash of the given toolResponse and returns it as a hexadecimal string.
// func hashToolResponse(toolResponse string) string {
// 	// Create a new SHA-256 hash.
// 	h := sha256.New()
// 	// Write the thought byte slice to the hash.
// 	h.Write([]byte(toolResponse))
// 	// Compute the final hash and return it as a hexadecimal string.
// 	return fmt.Sprintf("%x", h.Sum(nil))
// }
