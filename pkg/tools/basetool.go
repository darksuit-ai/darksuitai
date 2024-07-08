package tools

// Define ToolFunc to return a slice of interface{} to hold any number of values
type ToolFunc func(string, []interface{}) (string, []interface{})

type BaseTool struct {
	Name        string   // Name of the tool
	Description string   // Description of the tool
	ToolFunc    ToolFunc // Function of the tool
}
