package utils

import "math"

func BuildPath(prefix []string, keys ...string) []string {
	full := make([]string, len(prefix), len(prefix)+len(keys))
	copy(full, prefix)
	return append(full, keys...)
}

// InterfaceSliceToStringSlice converts []interface{} to []string
func InterfaceSliceToStringSlice(slice []interface{}) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if str, ok := item.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

// RoundToOneDecimal rounds a float64 to 1 decimal place for API consistency
// Similar to Python's round(value, 1)
func RoundToOneDecimal(value float64) float64 {
	return math.Round(value*10) / 10
}
