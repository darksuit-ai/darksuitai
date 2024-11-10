package prompts

import (
	"embed"
	"fmt"
	"gopkg.in/yaml.v2"
)

var (
	//go:embed *.yaml
	configFiles  embed.FS
	yamlFilePath string = ""
)

// PromptConfig holds the structure of our YAML file
type promptConfig struct {
	AGENTCHATINSTRUCTION []byte `yaml:"ReAct"`
	AGENTSYSTEMINSTRUCTION []byte `yaml:"SYSTEMPROMPT"`
	CHATINSTRUCTION []byte `yaml:"PromptTemplate"`
}
type intermediatePromptConfig struct {
	AGENTCHATINSTRUCTION string `yaml:"ReAct"`
	AGENTSYSTEMINSTRUCTION string `yaml:"SYSTEMPROMPT"`
	CHATINSTRUCTION string `yaml:"PromptTemplate"`
}

func LoadPromptConfigs() (*promptConfig, error) {
	config := &promptConfig{}
	var errs []error
	filenames := []string{"agent.yaml","aichat.yaml"}
	for _, filename := range filenames {
		// Read the YAML file
		data, err := configFiles.ReadFile(yamlFilePath + filename)
		if err != nil {
			errs = append(errs, fmt.Errorf("error reading config file %s: %v", filename, err))
			continue
		}

		// Create a temporary config to hold this file's data
		var tempConfig intermediatePromptConfig
		err = yaml.Unmarshal(data, &tempConfig)
		if err != nil {
			errs = append(errs, fmt.Errorf("error parsing config file %s: %v", filename, err))
			continue
		}

		// Merge the temp config into the main config
		mergeConfigs(config, &tempConfig)
	}

	// Check if we encountered any errors
	if len(errs) > 0 {
		return config, fmt.Errorf("encountered errors while loading configs: %v", errs)
	}

	return config, nil
}

// mergeConfigs merges the source config into the destination config
func mergeConfigs(dst *promptConfig, src *intermediatePromptConfig) {

	if src.CHATINSTRUCTION != "" {
		dst.CHATINSTRUCTION = []byte(src.CHATINSTRUCTION)
	}
	if src.AGENTCHATINSTRUCTION != "" {
		dst.AGENTCHATINSTRUCTION = []byte(src.AGENTCHATINSTRUCTION)
	}
	if src.AGENTSYSTEMINSTRUCTION != "" {
		dst.AGENTSYSTEMINSTRUCTION = []byte(src.AGENTSYSTEMINSTRUCTION)
	}
	// Add any other fields here
}
