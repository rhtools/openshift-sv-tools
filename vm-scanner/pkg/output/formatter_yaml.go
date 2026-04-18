package output

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

// YAMLFormatter handles YAML output format
type YAMLFormatter struct {
	outputFile string
}

// NewYAMLFormatter creates a new YAML formatter
func NewYAMLFormatter(outputFile string) *YAMLFormatter {
	return &YAMLFormatter{
		outputFile: outputFile,
	}
}

// Format implements the formatting for YAML output
func (f *YAMLFormatter) Format(data interface{}) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Println(string(yamlData))
	return nil
}
