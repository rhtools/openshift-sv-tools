package vm

import (
	"context"
	"fmt"
	"log"
	"strings"
	"vm-scanner/pkg/client"
	"vm-scanner/pkg/utils"

	"k8s.io/apimachinery/pkg/runtime/schema"
	// kubevirtv1 "kubevirt.io/api/core/v1" // Temporarily disabled due to version conflicts
	// kubevirt "kubevirt.io/client-go/kubecli" // Temporarily disabled due to version conflicts
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// VMIInfo is deprecated - use VMRuntimeInfo from types.go instead
// This type alias maintains backward compatibility
type VMIInfo = VMRuntimeInfo

func VMKey(namespace, name string) string {
	return namespace + "/" + name
}

// convertNetworkMapToSlice converts a map of network interfaces to a slice
func convertNetworkMapToSlice(networkMap map[string]VMNetworkInfo) []VMNetworkInfo {
	result := make([]VMNetworkInfo, 0, len(networkMap))
	for _, netInfo := range networkMap {
		result = append(result, netInfo)
	}
	return result
}

func BuildVMIInventory(ctx context.Context, k8sClient *client.ClusterClient) (map[string]VMConsolidatedReport, error) {

	// get the vmi list
	vmis, err := GetAllVMIs(ctx, k8sClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get VMIs: %w", err)
	}
	// create an empty map to store the VMConsolidatedReport
	// remember that any value in a struct that is undeclared will get a zero value

	vmiInventory := make(map[string]VMConsolidatedReport)
	for _, vmiUnstructured := range vmis.Items {
		name := vmiUnstructured.GetName()
		namespace := vmiUnstructured.GetNamespace()

		// Check status.phase which shows the actual VMI runtime state
		// status.phase can be: "Pending", "Scheduling", "Scheduled", "Running", "Succeeded", "Failed", "Unknown"
		powerState := "Running" // Default to Running since VMI exists
		if phase, found, _ := unstructured.NestedString(vmiUnstructured.Object, "status", "phase"); found {
			powerState = phase
		}
		machineType, uid, timestamp, prettyName, runningOnNode, guestOSVersion, launcherPodName, err := GetMachineMetadata(k8sClient, ctx, vmiUnstructured, true, "spec")
		if err != nil {
			return nil, fmt.Errorf("failed to get guest metadata: %w", err)
		}

		// Get network interface information: MAC/IP from status, Type/Model/NAD from spec
		networkInterfaces, err := GetVMNetworkInterfaces(vmiUnstructured)
		if err != nil {
			fmt.Printf("Warning: Failed to get network interfaces for %s/%s: %v\n", namespace, name, err)
			networkInterfaces = make(map[string]VMNetworkInfo)
		}

		// Get the guest agent info and parse metadata from it
		guestAgentInfo, err := client.GetGuestAgentInfoRawAPIData(ctx, k8sClient, namespace, name)
		if err != nil {
			return nil, fmt.Errorf("failed to get guest agent info: %w", err)
		}

		// Parse Guest Agent JSON into GuestMetadata struct (using intermediate struct)
		guestMetadata, err := parseGuestMetadataFromGuestAgentInfo(guestAgentInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to parse guest metadata: %w", err)
		}

		// Parse disk info from guest agent fsInfo data
		diskInfo, err := parseStorageInfoFromGuestAgentInfo(guestAgentInfo)
		if err != nil {
			log.Printf("Warning: Failed to parse disk info from guest agent for %s/%s: %v", namespace, name, err)
		} else {
			guestMetadata.DiskInfo = diskInfo
		}

		// Enrich GuestMetadata with additional data from VMI object and cluster queries
		if launcherPodName != "" {
			guestMetadata.VirtLauncherPod = launcherPodName
			guestMetadata.RunningOnNode = runningOnNode
			guestMetadata.OSVersion = guestOSVersion
		}
		memoryHotPlugMaxByVMIMiB, err := GetMemoryHotPlugMax(vmiUnstructured)
		if err != nil {
			return nil, fmt.Errorf("failed to get memory hot plug max: %w", err)
		}
		// Get memory and CPU information (errors ignored - optional data)
		// All values returned as float64 in MiB
		memoryUsedByVMIMiB, memoryFreeByVMIMiB, _ := GetMemoryUsedFromMonitoring(ctx, k8sClient, namespace, name, runningOnNode)
		memoryUsedByNodeMiB, _ := GetVirtLauncherPodMemoryInfo(k8sClient, ctx, launcherPodName)

		// Calculate LibVirt overhead (difference between node and VMI memory)
		var memoryUsedByLibVirt float64
		if memoryUsedByVMIMiB > 0 && memoryUsedByNodeMiB > 0 {
			memoryUsedByLibVirt = utils.RoundToOneDecimal(memoryUsedByNodeMiB - memoryUsedByVMIMiB)
		}

		// calculate the percentage of memory used by the vmi
		memoryUsedPercentage := utils.RoundToOneDecimal(memoryUsedByVMIMiB / (memoryFreeByVMIMiB + memoryUsedByVMIMiB) * 100)
		var memInfo *MemoryInfo
		if memoryUsedByVMIMiB > 0 || memoryUsedByNodeMiB > 0 {
			memInfo = &MemoryInfo{
				MemoryUsedByVMI:      utils.RoundToOneDecimal(memoryUsedByVMIMiB),
				TotalMemoryUsed:      utils.RoundToOneDecimal(memoryUsedByNodeMiB),
				MemoryUsedByLibVirt:  memoryUsedByLibVirt,
				MemoryFree:           utils.RoundToOneDecimal(memoryFreeByVMIMiB),
				MemoryUsedPercentage: memoryUsedPercentage,
				MemoryHotPlugMax:     utils.RoundToOneDecimal(memoryHotPlugMaxByVMIMiB),
			}
		}

		// Get CPU configuration from VMI spec
		vmiCPUInfo, err := GetCPUInfoFromVMI(vmiUnstructured)
		if err != nil {
			return nil, fmt.Errorf("failed to get CPU info: %w", err)
		}

		// Get memory guest size from VMI spec
		memoryGuest, _, _ := unstructured.NestedString(vmiUnstructured.Object, "spec", "domain", "memory", "guest")

		// Convert memory guest to MiB and round to 1 decimal
		memoryConfiguredMiB, err := utils.QuantityToMiB(memoryGuest)
		if err != nil {
			log.Printf("Warning: Failed to convert memory guest to MiB for %s: %v", name, err)
			memoryConfiguredMiB = 0
		}
		memoryConfiguredMiB = utils.RoundToOneDecimal(memoryConfiguredMiB)

		// Build consolidated memory info with both configuration and usage metrics
		memoryInfoWithConfig := MemoryInfo{
			MemoryConfiguredMiB: memoryConfiguredMiB,
			MemoryHotPlugMax:    utils.RoundToOneDecimal(memoryHotPlugMaxByVMIMiB),
		}
		if memInfo != nil {
			memoryInfoWithConfig.MemoryFree = memInfo.MemoryFree
			memoryInfoWithConfig.MemoryUsedByLibVirt = memInfo.MemoryUsedByLibVirt
			memoryInfoWithConfig.MemoryUsedByVMI = memInfo.MemoryUsedByVMI
			memoryInfoWithConfig.MemoryUsedPercentage = memInfo.MemoryUsedPercentage
			memoryInfoWithConfig.TotalMemoryUsed = memInfo.TotalMemoryUsed
		}

		// Build VMBaseInfo with configuration from VMI spec + usage metrics
		baseInfo := VMBaseInfo{
			Name:              name,
			Namespace:         namespace,
			Labels:            vmiUnstructured.GetLabels(),
			Annotations:       vmiUnstructured.GetAnnotations(),
			Running:           powerState == "Running",
			MachineType:       machineType,
			OSName:            prettyName,
			CPUInfo:           *vmiCPUInfo,
			MemoryInfo:        memoryInfoWithConfig,
			Disks:             make(map[string]StorageInfo), // Will be populated later
			NetworkInterfaces: convertNetworkMapToSlice(networkInterfaces),
		}

		// Build runtime-only info
		runtimeInfo := VMRuntimeInfo{
			VMIUID:            uid,
			CreationTimestamp: timestamp,
			PowerState:        powerState,
			GuestMetadata:     guestMetadata,
		}

		// Create consolidated report
		report := VMConsolidatedReport{
			VMBaseInfo: baseInfo,
			Runtime:    &runtimeInfo,
		}

		// generate the key for the vmi
		key := VMKey(namespace, name)
		vmiInventory[key] = report
	}
	return vmiInventory, nil
}

// get all vmis in the cluster
func GetAllVMIs(ctx context.Context, k8sClient *client.ClusterClient) (*unstructured.UnstructuredList, error) {
	dynamicClient := k8sClient.Dynamic
	vmiGroupVersionResource := schema.GroupVersionResource{
		Group:    "kubevirt.io",
		Version:  "v1",
		Resource: "virtualmachineinstances",
	}
	vmis, err := dynamicClient.Resource(vmiGroupVersionResource).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var runningVMIs []unstructured.Unstructured

	for _, vmi := range vmis.Items {
		runningVMIs = append(runningVMIs, vmi)
	}
	return vmis, nil
}

func GetAllVMs(ctx context.Context, k8sClient *client.ClusterClient) (*unstructured.UnstructuredList, error) {
	dynamicClient := k8sClient.Dynamic
	vmGroupVersionResource := schema.GroupVersionResource{
		Group:    "kubevirt.io",
		Version:  "v1",
		Resource: "virtualmachines",
	}
	vms, err := dynamicClient.Resource(vmGroupVersionResource).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return vms, nil
}

// BuildVMBaseInfoFromVM extracts base configuration from a VM object (for stopped VMs)
// This includes CPU, memory, disk, and network specs from the VM template
func BuildVMBaseInfoFromVM(ctx context.Context, k8sClient *client.ClusterClient, vmUnstructured unstructured.Unstructured) (*VMBaseInfo, error) {
	name := vmUnstructured.GetName()
	namespace := vmUnstructured.GetNamespace()
	createdAt := vmUnstructured.GetCreationTimestamp()
	labels := vmUnstructured.GetLabels()
	annotations := vmUnstructured.GetAnnotations()

	// Get CPU info from VM spec.template.spec
	cpuInfo, err := GetCPUInfoFromVM(vmUnstructured)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}

	// Get memory info from VM spec.template.spec
	memoryGuest, _, _ := unstructured.NestedString(vmUnstructured.Object, "spec", "template", "spec", "domain", "memory", "guest")

	// Convert memory guest to MiB and round to 1 decimal
	memoryConfiguredMiB, err := utils.QuantityToMiB(memoryGuest)
	if err != nil {
		log.Printf("Warning: Failed to convert memory guest to MiB for VM %s: %v", name, err)
		memoryConfiguredMiB = 0
	}
	memoryConfiguredMiB = utils.RoundToOneDecimal(memoryConfiguredMiB)

	memoryInfo := MemoryInfo{
		MemoryConfiguredMiB: memoryConfiguredMiB,
	}

	machineType, uid, _, _, _, guestOSVersion, _, err := GetMachineMetadata(k8sClient, ctx, vmUnstructured, false, "spec", "template", "spec")
	if err != nil {
		return nil, fmt.Errorf("failed to get machine metadata: %w", err)
	}

	runStrategy, _, _ := unstructured.NestedString(vmUnstructured.Object, "spec", "runStrategy")

	evictionStrategy, _, _ := unstructured.NestedString(vmUnstructured.Object, "spec", "evictionStrategy")

	instancetypeKind, _, _ := unstructured.NestedString(vmUnstructured.Object, "spec", "instancetype", "kind")
	instancetypeName, _, _ := unstructured.NestedString(vmUnstructured.Object, "spec", "instancetype", "name")
	var instanceType string
	if instancetypeName != "" &&
		(strings.EqualFold(instancetypeKind, "virtualmachineclusterinstancetype") ||
			strings.EqualFold(instancetypeKind, "virtualmachineinstancetype")) {
		instanceType = instancetypeName
	}

	preferenceKind, _, _ := unstructured.NestedString(vmUnstructured.Object, "spec", "preference", "kind")
	preferenceName, _, _ := unstructured.NestedString(vmUnstructured.Object, "spec", "preference", "name")
	var preference string
	if preferenceName != "" &&
		(strings.EqualFold(preferenceKind, "virtualmachineclusterpreference") ||
			strings.EqualFold(preferenceKind, "virtualmachinepreference")) {
		preference = preferenceName
	}

	networkInterfaces, err := GetNetworkInterfaces(vmUnstructured, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}
	return &VMBaseInfo{
		Name:              name,
		Namespace:         namespace,
		CreatedAt:         createdAt,
		Labels:            labels,
		Annotations:       annotations,
		Running:           false, // This is a stopped VM
		CPUInfo:           *cpuInfo,
		MemoryInfo:        memoryInfo,
		Disks:             make(map[string]StorageInfo),
		NetworkInterfaces: convertNetworkMapToSlice(networkInterfaces),
		MachineType:       machineType,
		OSName:            guestOSVersion,
		UID:               uid,
		RunStrategy:       runStrategy,
		EvictionStrategy:  evictionStrategy,
		InstanceType:      instanceType,
		Preference:        preference,
	}, nil
}
