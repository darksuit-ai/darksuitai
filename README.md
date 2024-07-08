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

	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: error loading .env file: %v", err)
	}

	args := darksuitai.NewChatLLMArgs()

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


## 💁 Contributing

As an open-source project in a rapidly developing field, we are extremely open to contributions, whether it be in the form of a new feature, improved infrastructure, or better documentation.
