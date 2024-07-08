package utilities

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v2"
)


type Prompts struct {
	ChatPrompt string `yaml:"chat_prompt"`
}

func LoadPrompts(filename string, promptFS embed.FS) (*Prompts, error) {
	// List all embedded files
	entries, err := promptFS.ReadDir("internal/prompts")
	if err != nil {
		return nil, fmt.Errorf("error reading embedded directory: %w", err)
	}

	// Print available files for debugging
	fmt.Println("Available embedded files:")
	for _, entry := range entries {
		fmt.Println(entry.Name())
	}

	// Construct the full path
	fullPath := filepath.Join("internal/prompts", filename)

	// Read the file
	data, err := promptFS.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", fullPath, err)
	}

	var prompts Prompts
	err = yaml.Unmarshal(data, &prompts)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling YAML: %w", err)
	}

	return &prompts, nil
}

// CustomFormat is a placeholder for your existing CustomFormat function
// Implement this function according to your needs
func CustomFormat(s []byte, kwargs map[string][]byte) []byte {
	for k := range kwargs {
		s = bytes.ReplaceAll(s, []byte("{"+k+"}"), kwargs[k])
	}
	return s
}
