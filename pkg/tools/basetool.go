package tools

import 	 "github.com/darksuit-ai/darksuitai/pkg/tools/google"
// Define ToolFunc to return a slice of interface{} to hold any number of values
type ToolFunc func(string,string, map[string]interface{}) (string, []interface{},error)

type BaseTool struct {
	Name        string   // Name of the tool
	Description string   // Description of the tool
	ToolFunc    ToolFunc // Function of the tool
}

var ToolNodes = []BaseTool{}

var ToolNodesMeta = make(map[string]interface{})

var GoogleTool = BaseTool{
	Name: "Google Search",
	Description: `This tool is useful ONLY in the following circumstances:
	- The user is asking about recent/current events or something that requires real-time information (sports scores, news, latest happenings/events, any information deemed recent; beyond your knowledge cutoff year. Recently, the current year is %d).
	- When you need to look up information about topics, these topics can be a wide range of topics, especially about humans, stocks, e.g., who won the Super Bowl? Is Jim Simons still alive? Who is in Scotland's Euro group this year?
	- The user is asking about some term you are totally unfamiliar with (it might be new).
	- The user explicitly asks you to browse, show images, or provide links to references.
Always rephrase the input word in the best search term to get results.`,
	ToolFunc: func(query string,toolName string, meta map[string]interface{}) (string, []interface{}, error) {
		return google.GoogleSearchAndImages(query, nil)
	},
}

