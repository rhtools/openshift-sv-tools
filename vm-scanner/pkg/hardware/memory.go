package hardware

import (
	"context"
	"encoding/json"
	"fmt"
	"vm-scanner/pkg/client"
	"vm-scanner/pkg/utils"

	corev1 "k8s.io/api/core/v1"
)

// GetMemoryInfo retrieves memory capacity and usage for a node
// Capacity comes from node.Status.Capacity
// Usage comes from rawStatsData (/proxy/stats/summary API)
func GetMemoryInfo(ctx context.Context, k8sClient *client.ClusterClient, node corev1.Node, rawStatsData []byte) (*MemoryInfo, error) {
	// Get capacity from node
	capacityBytes := node.Status.Capacity.Memory().Value()

	// Parse usage from stats API
	var statsInfo map[string]interface{}
	if err := json.Unmarshal(rawStatsData, &statsInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stats data: %w", err)
	}

	nodeData, ok := statsInfo["node"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("node data not found in stats response")
	}

	memData, ok := nodeData["memory"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("memory data not found in node stats")
	}

	// Use workingSetBytes as the "used" metric (matches kubectl top nodes)
	workingSet, ok := memData["workingSetBytes"].(float64)
	if !ok {
		return nil, fmt.Errorf("workingSetBytes not found or invalid type")
	}

	memoryUsedPercentage := utils.RoundToOneDecimal(workingSet / float64(capacityBytes) * 100)
	return &MemoryInfo{
		MemoryCapacityGiB:    utils.RoundToOneDecimal(utils.BytesToGiB(capacityBytes)),
		MemoryUsedGiB:        utils.RoundToOneDecimal(utils.BytesToGiB(int64(workingSet))),
		MemoryUsedPercentage: memoryUsedPercentage,
	}, nil
}
