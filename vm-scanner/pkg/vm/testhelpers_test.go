package vm

import (
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func loadFixture(t *testing.T, path string) unstructured.Unstructured {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", path, err)
	}
	var obj unstructured.Unstructured
	if err := yaml.Unmarshal(data, &obj); err != nil {
		t.Fatalf("failed to unmarshal fixture %s: %v", path, err)
	}
	return obj
}
