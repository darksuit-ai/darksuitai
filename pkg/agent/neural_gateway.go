package agent

import (
	"bytes"
	"regexp"
)

var actionRegex = regexp.MustCompile(`\s*Action:\s*(.*)`)
var inputRegex = regexp.MustCompile(`\s*Input:\s*(.*)`)
var emojiRegex = regexp.MustCompile(`[^a-zA-Z0-9\s/]`)


// NeuralParser is a function that parses the output from the agent cortex.
// It takes a byte `data` as input, which represents the output from the agent.
// It returns an `AgentActionTypes` struct, which contains the agent's action and input,
// or an `AgentActionTypes` struct with the agent's final output, if the output contains "Final Answer".
// If the output does not contain the expected format, it returns an empty `AgentActionTypes` struct.
func NeuralParser(data []byte) (AgentActionTypes, []byte, error) {
	substringFinalAnswer := []byte("Final Answer:")
	//fmt.Printf("%s",data)
	if bytes.Contains(data, []byte("Action:")) && bytes.Contains(data, []byte("Input:")) {
		// Extract the words after "Action:" 
		action := actionRegex.FindSubmatch(data)[1]

		// Remove any emoji in the new variables
		action = emojiRegex.ReplaceAll(action, []byte(""))

		action = bytes.Trim(action, "\n")

		// Extract the words after "Input:"
		input := inputRegex.FindSubmatch(data)[1]
		//input := bytes.SplitAfter(data, []byte(`Input:`))[1]
		// input = emojiRegex.ReplaceAll(input, []byte(""))
		input = bytes.Trim(input, "\n")

		// Create an `AgentActionTypes` struct with the extracted action and input
		actionTypes := AgentActionTypes{
			AgentAction: map[string]string{
				"Action": string(bytes.TrimSpace(action)),
				"Input":  string(input),
			},
		}
		
		var rawThought []byte
		thoughtParts := bytes.SplitAfter(data, []byte("Thought:"))
		if len(thoughtParts) > 1 {
			rawThought = bytes.TrimPrefix(bytes.TrimSuffix(thoughtParts[1], []byte("\nAction:")), []byte(" "))
		}

		return actionTypes,rawThought, nil
	} else if bytes.HasPrefix(data, substringFinalAnswer) || bytes.Contains(data, substringFinalAnswer) {

		// Extract the final answer from the output
		parts := bytes.Split(data, substringFinalAnswer)
		finalAnswerOutput := bytes.Join(parts[1:], nil)
		agentFinish := AgentActionTypes{AgentFinish: map[string][]byte{"Output": finalAnswerOutput}}
		return agentFinish, data, nil
	} else {
		
		// If the output does not match the expected formats, return an error `AgentActionTypes` struct
		agentError := AgentActionTypes{AgentError: map[string][]byte{"Error":[]byte(`Oops! I used a wrong format. I will look at the response formats and try again.`)}}
		return agentError, nil, nil
	}
}
