package internal

import (
	"fmt"
	"strings"
	"time"
)

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

func printBoxedText(text string, width int) {
	border := strings.Repeat("═", width)
	fmt.Printf("╔%s╗\n", border)
	fmt.Printf("║ %-*s ║\n", width-2, text)
	fmt.Printf("╚%s╝\n", border)
}

func _wakeDarkSuitAgent(agentName string) func() {
	if agentName == "" {
		agentName = "AlbusDD"
	}

	clearanceLevel := "ALPHA-1"
	specialization := "AI/Cyber Operations"
	agentID := "DS-847-SAY"
	status := "ACTIVE"
	return func() {
		if !isExecuted {
			// Clear terminal sequence
			fmt.Print("\033[H\033[2J")

			// Top secret banner
			printBoxedText("TOP SECRET - LEVEL 5 CLEARANCE REQUIRED", 50)
			fmt.Println()

			// Initialize sequence
			fmt.Println("INITIALIZING DARKSUIT PROTOCOL...")
			time.Sleep(time.Millisecond)
			fmt.Println("[■■■■■■■■■■] 100% COMPLETE")
			fmt.Println()

			// Agent profile header
			fmt.Println("╔══════════════════════════════════════════════════════╗")
			fmt.Println("║             DARKSUIT AGENT PROFILE                   ║")
			fmt.Println("╠══════════════════════════════════════════════════════╣")

			// Agent details
			fmt.Printf("║ AGENT NAME: %-41s║\n", agentName)
			fmt.Printf("║ AGENT ID: %-43s║\n", agentID)
			fmt.Printf("║ CLEARANCE: %-41s║\n", clearanceLevel)
			fmt.Printf("║ SPECIALIZATION: %-37s║\n", specialization)
			fmt.Printf("║ STATUS: %-44s║\n", status)
			fmt.Println("╠══════════════════════════════════════════════════════╣")

			// Timestamp
			currentTime := time.Now().Format("2006-01-02 15:04:05 MST")
			fmt.Printf("║ ACCESSED: %-42s║\n", currentTime)
			fmt.Println("╚══════════════════════════════════════════════════════╝")

			// Warning message
			fmt.Println()
			// printBoxedText("WARNING: UNAUTHORIZED ACCESS WILL BE PROSECUTED", 50)
			isExecuted = true
		}
	}
}

// NewDarkSuitAgent returns an instance of the DarkSuitAgent interface
func NewDarkSuitAgent() DarkSuitAgent {
	return &darkSuitAgentImpl{}
}
