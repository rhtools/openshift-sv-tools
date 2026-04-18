package error_handling

import (
	"fmt"
)

func GetRequiredString(err error, found bool, path string) error {
	if err != nil {
		return fmt.Errorf("error accessing %v: %w", path, err)
	}
	if !found {
		return fmt.Errorf("required field %v not found", path)
	}
	return nil
}
