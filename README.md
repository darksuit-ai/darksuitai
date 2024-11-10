# 🕵️ DarkSuitAI

⚡ Blazing production-ready library for building scalable reasoning AI systems ✨

[![Release Notes](https://img.shields.io/github/release/darksuit-ai/darksuitai?style=flat-square)](https://github.com/darksuit-ai/darksuitai/releases)
[![CI](https://github.com/darksuit-ai/darksuitai/actions/workflows/check_diffs.yml/badge.svg)](https://github.com/darksuit-ai/darksuitai/actions/workflows/check_diffs.yml)
[![GitHub star chart](https://img.shields.io/github/stars/darksuit-ai/darksuitai?style=flat-square)](https://star-history.com/#darksuit-ai/darksuitai)
[![Open Issues](https://img.shields.io/github/issues-raw/darksuit-ai/darksuitai?style=flat-square)](https://github.com/darksuit-ai/darksuitai/issues)
[![Open in Dev Containers](https://img.shields.io/static/v1?label=Dev%20Containers&message=Open&color=blue&logo=visualstudiocode&style=flat-square)](https://vscode.dev/redirect?url=vscode://ms-vscode-remote.remote-containers/cloneInVolume?url=https://github.com/darksuit-ai/darksuitai)
[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://codespaces.new/darksuit-ai/darksuitai)



## Quick Install

```go
go get github.com/darksuit-ai/darksuitai@latest
```


## 🤔 What is DarkSuitAI?

**DarkSuitAI** is a framework for developing production-ready AI systems powered by large language models (LLMs).



## 🧱 What can you build with DarkSuitAI?


**🧱 Agent Powered AI System**

- [Documentation]()

**🤖 Chatbots**

- [Documentation]()

And much more!

## 🚀 How does DarkSuitAI bring you straight to production?
The main value props of the DarkSuitAI libraries are:
1. **Components**: composable building blocks, tools and integrations for working with language models. Components are modular and easy-to-use, and full scale production-ready for AI systems.
2. **Off-the-shelf chains**: built-in assemblages of components for accomplishing higher-level tasks

Off-the-shelf chains make it easy to get started. Components make it easy to customize existing chains and build new ones. 


## Components

Components fall into the following **modules**:

**📃 Model I/O**

This includes [prompt management](s), [prompt optimization](), a generic interface for [chat models](), and common utilities for working with [model outputs]().

```go

package main

import (
	"fmt"
	"github.com/darksuit-ai/darksuitai"
)

func main() {
	// either add apikey to your .env and darksuit picks it up or pass as argument
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: error loading .env file: %v", err)
	}

	args := darksuitai.NewChatLLMArgs()
	args.AddAPIKey([]byte(`your-api-key`)) // pass LLM API Key
	// args.SetChatInstruction([]byte(`Your chat instruction goes here`)) // uncomment to pass your own prompt instruction
	args.AddPromptKey("year", []byte(`2024`)) // pass variables to your prompt
	args.SetModelType("openai", "gpt-4o") // set the model
	args.AddModelKwargs(500, 0.8, true, []string{"\nObservation:"}) // set model keyword arguments
	llm,err := args.NewLLM()
	if err != nil{
		// handle the error as you wish
	}
	resp,err:=llm.Chat("hello, Sam Ayo from earth🌍. What is your name?")
	if err != nil{
		// handle the error as you wish
	}
	fmt.Println(resp)
	// to stream
	streamText :=llm.Stream("hello from earth🌍, what is your name?")
		for r := range streamText{
			fmt.Println(r)
		}
	}

```

```go

package main

import (
	"fmt"
	"context"
	"log"
	"github.com/darksuit-ai/darksuitai"
)

func main() {

	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: error loading .env file: %v", err)
	}
	user := "your_db_username"
	password := "your_db_password"
	host := "your_db_host"
	port := "your_db_port"
	databaseName := "your_atabase_name"

	args := darksuitai.NewChatLLMArgs()
	args.AddAPIKey([]byte(`your-api-key`)) // pass LLM API Key
	// Set up the MongoDB connection URL
	url := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?serverSelectionTimeoutMS=5000&authSource=mongo_staging&directConnection=true", user, password, host, port,databaseName)

	// Connect to the MongoDB database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		log.Fatal(err)
	}
	// Ping the primary
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	// Get a handle to the database and collections
	db := client.Database(databaseName)
	// args.SetChatInstruction([]byte(`Your chat instruction goes here`)) // uncomment to pass your own prompt instruction
	args.AddPromptKey("year", []byte(`2024`)) // pass variables to your prompt
	args.MongoDB(db) // add mongodb client
	args.SetModelType("openai", "gpt-4o") // set the model
	args.AddModelKwargs(500, 0.8, true, []string{"\nObservation:"}) // set model keyword arguments
	llm,err := args.NewLLM()
	if err != nil{
		// handle the error as you wish
	}
	resp,err:=llm.ConvChat("hello, Sam Ayo from earth🌍. What is your name?")
	if err != nil{
		// handle the error as you wish
	}
	fmt.Println(resp)

}

```
**🤖 Agents**

Agents allow an LLM autonomy over how a task is accomplished. Agents make decisions about which Actions to take, then take that Action, observe the result, and repeat until the task is complete. DarkSuitAI supercedes all other agentic frameworks/library through it's AI self-reflect action control.


## 🛠️ Tool

Tools in agentic AI are essential components that allow AI agents to interact with their environment and perform specific tasks. 
In the context of agentic AI, tools are functions or modules that the AI can utilize to achieve its goals. 
These tools can range from simple data processing functions to complex algorithms that enable decision-making and problem-solving. 
By leveraging tools, AI agents can extend their capabilities beyond their core functionalities, allowing them to adapt to various scenarios and challenges. 
In the DarkSuitAI framework, tools are defined using the `NewTool` function, which allows developers to create custom tools tailored to specific needs. 
These tools are then registered in the `ToolNodes` map, making them accessible to the AI agents for execution during their task completion processes.

### Creating a tool and testing it

To create the `a Tool`, you can use the `NewTool` function provided by the DarkSuitAI framework. Here's a sample code snippet to create and register the `a Tool`:
```go
func googleSearch(query string)(string,[]interface{}, error){
	// your logic
}
testingTool := tsk.NewTool(
    "google search", // tool name
    "this tool is useful for performing web search using Google.", // tool description
    func(query string, metaData []interface{}) (string, []interface{}, error) {
        return googleSearch(query string)
    },
)

// Register the google search Tool in the ToolNodes map
tsk.ToolNodes["google search"] = testingTool
result,toolMeta,_:=tsk.ToolNodes["google search"].ToolFunc("about the US",nil)
fmt.Printf("%s,%v",result,toolMeta)

```
### Building an agent
```go

package main

import (
	"fmt"
	"context"
	"log"
	"github.com/darksuit-ai/darksuitai"
)

func main() {
	databaseName := "your_database_name"
	databaseURL := "your_database_url either mongodb+srv:// or mongodb://"
	db := NewMongoChatMemory(data,databaseName) // Get the database pointer

	weatherReportTool :=darksuitai.NewTool(
		"weather report",
		"",
		func(query, toolName string, metaData map[string]interface{}) (string, []interface{},error) {
			// your API call to weather API like openweather
			rawWeatherResultFromAPI := `{"location": "San Francisco", "weather": "sunny", "high": "68°F", "low": "54°F"}`
			return "The weather in San Francisco is sunny with a high of 68°F and a low of 54°F.", []interface{}{rawWeatherResultFromAPI}, nil
		},
	)

	darksuitai.ToolNodes = append(darksuitai.ToolNodes, weatherReportTool)

	args := darksuitai.NewChatLLMArgs()
	args.AddAPIKey([]byte(`your-api-key`)) // pass LLM API Key

	// args.SetChatInstruction([]byte(`Your chat instruction goes here`)) // uncomment to pass your own prompt instruction
	args.SetMongoDBChatMemory(db) // set the database
	args.AddPromptKey("year", []byte(`2024`)) // pass variables to your prompt
	args.SetModelType("openai", "gpt-4o") // set the model
	args.AddModelKwargs(1000, 0.8, true, []string{"\nObservation:"}) // set model keyword arguments

	agent,err := args.NewSuitedAgent()
	if err != nil{
		// handle the error as you wish
	}
	err = agent.Program(3,"your-session-id",true)
	if err != nil{
		// handle the error as you wish
	}
	resp,_,err:=agent.Chat("hello what is the current weather?")
	if err != nil{
		// handle the error as you wish
	}
	fmt.Println(resp)

}

```

## 💁 Contributing

As an open-source project in a rapidly developing field, we are extremely open to contributions, whether it be in the form of a new feature, improved infrastructure, or better documentation.
