package vm

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"vm-scanner/pkg/client"
	"vm-scanner/pkg/utils"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func GetVirtLauncherPodMemoryInfo(k8sClient *client.ClusterClient, ctx context.Context, VirtLauncherPodName string) (float64, error) {
	podData, err := client.GetVirtLauncherPodRawAPIData(ctx, k8sClient, VirtLauncherPodName)
	if err != nil {
		return 0, err
	}
	// Unmarshal JSON into a map
	var metricsData map[string]interface{}
	if err := json.Unmarshal(podData, &metricsData); err != nil {
		return 0, fmt.Errorf("failed to unmarshal pod metrics: %w", err)
	}

	// virt-launcher pods always have exactly one container named "compute"
	containers := metricsData["containers"].([]interface{})
	container := containers[0].(map[string]interface{})
	usage := container["usage"].(map[string]interface{})
	totalMemoryUsed := usage["memory"].(string)
	// totalMemoryUsed is a string with units like Ki, Mi, Gi, etc. Convert it to a number
	totalMemoryUsedMiB, err := utils.QuantityToMiB(totalMemoryUsed)
	if err != nil {
		return 0, fmt.Errorf("failed to convert total memory used to number: %w", err)
	}

	return totalMemoryUsedMiB, nil
}

func parseMetricValue(line string) float64 {
	// Find the last space, which separates labels from value
	lastSpace := strings.LastIndex(line, " ")
	if lastSpace == -1 {
		return 0
	}

	valueStr := strings.TrimSpace(line[lastSpace+1:])
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0
	}

	return value
}

// parse the memory used from the query data - returns (usedMiB, reservedMiB, hotPlugMaxMiB, error)
func ParseMemoryFromMonitoring(queryData string, vmiName string) (float64, float64, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(queryData)))

	targetName := fmt.Sprintf(`name="%s"`, vmiName)
	var availableBytes, usableBytes float64
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") || len(line) == 0 {
			continue
		}

		// Only process lines for this specific VMI
		if !strings.Contains(line, targetName) {
			continue
		}

		// kubevirt_vmi_memory_available_bytes - Total memory available to the guest VM (what the guest OS sees as total RAM)
		if strings.HasPrefix(line, "kubevirt_vmi_memory_available_bytes{") {
			availableBytes = parseMetricValue(line)
		} else if strings.HasPrefix(line, "kubevirt_vmi_memory_usable_bytes{") {
			// kubevirt_vmi_memory_usable_bytes - Memory that is free/unused from the guest's perspective
			usableBytes = parseMetricValue(line)
		}

		// Other available metrics (not currently collected):
		// - kubevirt_vmi_memory_domain_bytes: Total memory allocated to VM domain by libvirt (host perspective)
		// - kubevirt_vmi_memory_resident_bytes: Actual physical RAM in use (RSS) - real memory consumption
		// - kubevirt_vmi_memory_cached_bytes: Memory used for page cache - can be reclaimed
		// - kubevirt_vmi_memory_unused_bytes: Memory allocated but not being used
		// - kubevirt_vmi_memory_actual_balloon_bytes: Current balloon size (memory ballooning for dynamic adjustment)
		// - kubevirt_vmi_memory_swap_in/out_traffic_bytes: Swap activity - indicates memory pressure
		// - kubevirt_vmi_memory_pgmajfault_total: Major page faults - indicates disk I/O for memory

	}

	if err := scanner.Err(); err != nil {
		return 0.0, 0.0, fmt.Errorf("error filtering metrics: %w", err)
	}

	usedBytes := availableBytes - usableBytes
	usedMiB := utils.BytesToMiB(int64(usedBytes))
	freeMiB := utils.BytesToMiB(int64(usableBytes))
	return usedMiB, freeMiB, nil
}

// GetMemoryUsedFromMonitoring queries OpenShift's monitoring stack and returns memory metrics in MiB
func GetMemoryUsedFromMonitoring(ctx context.Context, k8sClient *client.ClusterClient, namespace string, vmiName string, nodeName string) (float64, float64, error) {
	queryData, err := client.GetMonitoringRawAPIData(ctx, k8sClient, namespace, vmiName, nodeName)
	if err != nil {
		return 0, 0, err
	}

	usedMiB, freeMiB, err := ParseMemoryFromMonitoring(string(queryData), vmiName)
	if err != nil {
		return 0, 0, err
	}
	return usedMiB, freeMiB, nil
}


func GetMemoryHotPlugMax(vmiUnstructured unstructured.Unstructured) (float64, error) {
	hotPlugMax, found, err := unstructured.NestedString(vmiUnstructured.Object, "spec", "domain", "memory", "maxGuest")
	if err != nil || !found {
		return 0, fmt.Errorf("failed to get hot plug max: %w", err)
	}
	hotPlugMaxMiB, err := utils.QuantityToMiB(hotPlugMax)
	if err != nil {
		return 0, fmt.Errorf("failed to convert hot plug max to number: %w", err)
	}
	return hotPlugMaxMiB, nil
}
