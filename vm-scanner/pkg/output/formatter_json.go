package output

import (
	"encoding/json"
	"fmt"
)

// JSONFormatter handles JSON output format
type JSONFormatter struct {
	outputFile string
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter(outputFile string) *JSONFormatter {
	return &JSONFormatter{
		outputFile: outputFile,
	}
}

// Format implements the formatting for JSON output
func (f *JSONFormatter) Format(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}
