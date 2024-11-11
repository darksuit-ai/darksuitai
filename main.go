package main

import (
	"fmt"
	"github.com/darksuit-ai/darksuitai/tsk"
)

/*to use put darksuitai.go in tsk folder. change package to tsk
// then move main.go from tsk folder to root and change to package main
ensure to reverse before PR
*/

func main() {
	db := tsk.NewMongoChatMemory("","")
	weatherReportTool :=tsk.NewTool(
		"weather report",
		"",
		func(query, toolName string, metaData map[string]interface{}) (string, []interface{},error) {
			// your API call to weather API like openweather
			rawWeatherResultFromAPI := `{"location": "San Francisco", "weather": "sunny", "high": "68°F", "low": "54°F"}`
			return "The weather in San Francisco is sunny with a high of 68°F and a low of 54°F.", []interface{}{rawWeatherResultFromAPI}, nil
		},
	)
	tsk.ToolNodes = append(tsk.ToolNodes, weatherReportTool)
	// toolName := googleTool.Name
	// fmt.Println("Tool Name:", toolName)
	// d,_,_:=tsk.ToolNodes["google search"].ToolFunc("about the US",nil)

	fmt.Printf("%v",tsk.ToolNodes)
	args := tsk.NewLLMArgs()

// 	args.SetChatInstruction([]byte(`Answer the following like a british gentleman from medieval era in {year}.
// Question:{query}
// Answer:`))
	args.SetMongoDBChatMemory(db)
	args.AddPromptKey("year", []byte(`1664`)) // pass variables to your prompt
	// args.SetModelType("openai", "gpt-4o") // set the model
	args.SetModelType("groq", "llama3-70b-8192") // set the model
	args.AddModelKwargs(5000, 0.8, true, []string{"\nObservation:"}) // set model keyword arguments
	agent,err := args.NewSuitedAgent()
	if err != nil{
		print(err.Error())
	}
	err = agent.Program(3,"ter",true)
	if err != nil{
		print(err.Error())
	}
	// resp,_,err:=agent.Chat("hello,what is my name?")
	// if err != nil{
	// 	print(err.Error())
	// }
	// fmt.Println(resp)

	streamChan, err := agent.Stream("what is the weather currently")
	if err != nil{
		print(err.Error())
	}

    for chunk := range streamChan {
		// Process each chunk as it arrives
		fmt.Print(chunk)
	}





// 	args := tsk.NewChatLLMArgs()

// 	args.SetChatInstruction([]byte(`Answer the following like a british gentleman from medieval era in {year}.
// Question:{query}
// Answer:`))
// 	args.AddPromptKey("year", []byte(`1664`)) // pass variables to your prompt
// 	// args.SetModelType("openai", "gpt-4o") // set the model
// 	args.SetModelType("groq", "llama3-70b-8192") // set the model
// 	args.AddModelKwargs(5, 0.8, true, []string{"\nObservation:"}) // set model keyword arguments
// 	llm,err := args.NewLLM()
// 	if err != nil{
// 		print(err.Error())
// 	}
// 	// resp,err:=llm.Chat("hello from earth🌍, what is your name?")
// 	// if err != nil{
// 	// 	print(err.Error())
// 	// }
// 	// fmt.Println(resp)
// 	respChan :=llm.Stream("hello from earth🌍, what is your name?")
// 	for r := range respChan{
// 		fmt.Println(r)
// 	}
	

}