package internal

import "fmt"

var isExecuted bool

type DarkSuitAgent interface {
	WakeDarkSuitAgent() func()
}

type darkSuitAgentImpl struct {
	AgentName string `json:"agent_name"`
}

func (d *darkSuitAgentImpl) WakeDarkSuitAgent() func() {
	return _wakeDarkSuitAgent(d.AgentName)
}

func _wakeDarkSuitAgent(agentName string) func() {

	agentName = "Sam Ayo"

	return func() {
		if !isExecuted {
			fmt.Print("... STARTING DARKSUIT\n")
			fmt.Printf("Deploying your Agent 🕵️: %s\n", agentName)
			isExecuted = true
		}
	}
}

// NewDarkSuitAgent returns an instance of the DarkSuitAgent interface
func NewDarkSuitAgent() DarkSuitAgent {
	return &darkSuitAgentImpl{}
}
