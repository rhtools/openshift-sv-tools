package hardware

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GetCPUMakeAndModel extracts the CPU model string from a BareMetalHost.
func GetCPUMakeAndModel(bareMetalHost *unstructured.Unstructured) string {
	status, ok := bareMetalHost.Object["status"].(map[string]interface{})
	if !ok {
		return ""
	}
	hw, ok := status["hardware"].(map[string]interface{})
	if !ok {
		return ""
	}
	cpu, ok := hw["cpu"].(map[string]interface{})
	if !ok {
		return ""
	}
	model, _ := cpu["model"].(string)
	return model
}

// GetCPUInfo retrieves CPU information from a node.
func GetCPUInfo(node corev1.Node, bmh *unstructured.Unstructured) (*CPUInfo, error) {
	cpuModel := ""
	if bmh != nil {
		cpuModel = GetCPUMakeAndModel(bmh)
	}
	return &CPUInfo{
		CPUCores: node.Status.Capacity.Cpu().Value(),
		CPUModel: cpuModel,
	}, nil
}
