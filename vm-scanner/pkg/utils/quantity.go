package utils

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

// ParseQuantityToBytes converts a Kubernetes quantity string to bytes.
// Handles strings like "2Gi", "1024Mi", "1073741824", etc.
// Similar to parsing K8s quantity strings in Python and converting to int.
func ParseQuantityToBytes(quantityStr string) (int64, error) {
	quantity, err := resource.ParseQuantity(quantityStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse quantity '%s': %w", quantityStr, err)
	}
	return quantity.Value(), nil
}

// BytesToMiB converts bytes to mebibytes (base-2: 1024^2).
// Like dividing by 1024**2 in Python.
func BytesToMiB(bytes int64) float64 {
	return float64(bytes) / (1024 * 1024)
}

// BytesToGiB converts bytes to gibibytes (base-2: 1024^3).
// Like dividing by 1024**3 in Python.
func BytesToGiB(bytes int64) float64 {
	return float64(bytes) / (1024 * 1024 * 1024)
}

// QuantityToMiB is a convenience function that combines parsing and conversion.
// This is like a Python function that does parse().convert() in one call.
func QuantityToMiB(quantityStr string) (float64, error) {
	if quantityStr == "" {
		return 0, nil
	}
	bytes, err := ParseQuantityToBytes(quantityStr)
	if err != nil {
		return 0, err
	}
	return BytesToMiB(bytes), nil
}

// QuantityToGiB is a convenience function that combines parsing and conversion.
func QuantityToGiB(quantityStr string) (float64, error) {
	if quantityStr == "" {
		return 0, nil
	}
	bytes, err := ParseQuantityToBytes(quantityStr)
	if err != nil {
		return 0, err
	}
	return BytesToGiB(bytes), nil
}
