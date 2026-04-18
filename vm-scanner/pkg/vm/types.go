package vm

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// kubevirtv1 "kubevirt.io/api/core/v1" // Temporarily disabled due to version conflicts
)

// GuestMetadata represents guest agent information and guest-level metadata
type GuestMetadata struct {
	DiskInfo          []VMDiskInfo `json:"diskInfo" yaml:"diskInfo"`
	GuestAgentVersion string       `json:"guestAgentVersion" yaml:"guestAgentVersion"`
	HostName          string       `json:"hostName" yaml:"hostName"`
	KernelVersion     string       `json:"kernelVersion" yaml:"kernelVersion"`
	OSVersion         string       `json:"osVersion" yaml:"osVersion"`
	RunningOnNode     string       `json:"runningOnNode" yaml:"runningOnNode"`
	Timezone          string       `json:"timezone" yaml:"timezone"`
	VirtLauncherPod   string       `json:"virtLauncherPod" yaml:"virtLauncherPod"`
}

// MemoryInfo represents memory allocation and usage information
// All values are in MiB with 1 decimal precision for API consistency
type MemoryInfo struct {
	// Configuration (always available from spec)
	MemoryConfiguredMiB  float64 `json:"memoryConfiguredMiB" yaml:"memoryConfiguredMiB"` // Configured memory in MiB
	MemoryHotPlugMax     float64 `json:"memoryHotPlugMax,omitempty" yaml:"memoryHotPlugMax,omitempty"` // Max memory for hotplug in MiB
	// Runtime metrics (only available when VM is running)
	MemoryFree           float64 `json:"memoryFree,omitempty" yaml:"memoryFree,omitempty"`
	MemoryUsedByLibVirt  float64 `json:"memoryUsedByLibVirt,omitempty" yaml:"memoryUsedByLibVirt,omitempty"`
	MemoryUsedByVMI      float64 `json:"memoryUsedByVMI,omitempty" yaml:"memoryUsedByVMI,omitempty"`
	MemoryUsedPercentage float64 `json:"memoryUsedPercentage,omitempty" yaml:"memoryUsedPercentage,omitempty"`
	TotalMemoryUsed      float64 `json:"totalMemoryUsed,omitempty" yaml:"totalMemoryUsed,omitempty"`
}

// StorageInfo represents storage volume configuration and usage information
type StorageInfo struct {
	// Configuration (always available from spec/PVC)
	SizeBytes    int64  `json:"sizeBytes" yaml:"sizeBytes"`
	SizeHuman    string `json:"sizeHuman" yaml:"sizeHuman"`
	StorageClass string `json:"storageClass,omitempty" yaml:"storageClass,omitempty"`
	VolumeName   string `json:"volumeName" yaml:"volumeName"`
	VolumeType   string `json:"volumeType" yaml:"volumeType"`

	// Runtime metrics (only available when VM is running and guest agent is active)
	TotalStorage                int64   `json:"totalStorage,omitempty" yaml:"totalStorage,omitempty"`
	TotalStorageHuman           string  `json:"totalStorageHuman,omitempty" yaml:"totalStorageHuman,omitempty"`
	TotalStorageInUse           int64   `json:"totalStorageInUse,omitempty" yaml:"totalStorageInUse,omitempty"`
	TotalStorageInUseHuman      string  `json:"totalStorageInUseHuman,omitempty" yaml:"totalStorageInUseHuman,omitempty"`
	TotalStorageInUsePercentage float64 `json:"totalStorageInUsePercentage,omitempty" yaml:"totalStorageInUsePercentage,omitempty"`
}

// VMBaseInfo represents base information about a VirtualMachine (always available from VM definition)
// This data is extracted from the VM spec and is available for both stopped and running VMs.
type VMBaseInfo struct {
	// Metadata
	Annotations     map[string]string       `json:"annotations" yaml:"annotations"`
	CreatedAt       metav1.Time             `json:"createdAt" yaml:"createdAt"`
	Labels          map[string]string       `json:"labels" yaml:"labels"`
	Name            string                  `json:"name" yaml:"name"`
	Namespace       string                  `json:"namespace" yaml:"namespace"`
	OwnerReferences []metav1.OwnerReference `json:"ownerReferences" yaml:"ownerReferences"`
	Phase           string                  `json:"phase" yaml:"phase"`
	Ready           bool                    `json:"ready" yaml:"ready"`
	Running         bool                    `json:"running" yaml:"running"`
	UID             string                  `json:"uid" yaml:"uid"`
	UpdatedAt       metav1.Time             `json:"updatedAt" yaml:"updatedAt"`

	// Machine Configuration (common to both VMs and VMIs)
	MachineType string `json:"machineType" yaml:"machineType"` // Machine type (e.g., "q35", "pc")
	OSName      string `json:"osName" yaml:"osName"`           // Operating system name

	// VM Configuration (from spec - only available for VM, not VMI)
	RunStrategy    string `json:"runStrategy" yaml:"runStrategy"`       // Run strategy (Running, Halted, Manual, Once)
	EvictionStrategy string `json:"evictionStrategy" yaml:"evictionStrategy"` // Eviction strategy
	InstanceType   string `json:"instanceType" yaml:"instanceType"`   // Instance type name
	Preference     string `json:"preference" yaml:"preference"`       // Preference name

	// Resource Configuration (from spec.template.spec)
	CPUInfo           CPUInfo                `json:"cpuInfo" yaml:"cpuInfo"`                     // CPU configuration
	Disks             map[string]StorageInfo `json:"disks" yaml:"disks"`                         // Disk configuration and usage
	MemoryInfo        MemoryInfo             `json:"memoryInfo" yaml:"memoryInfo"`               // Memory configuration and usage
	NetworkInterfaces []VMNetworkInfo        `json:"networkInterfaces" yaml:"networkInterfaces"` // Network configuration
}

// VMRuntimeInfo represents runtime-only information about a VirtualMachineInstance (only when VM is running)
// This data is only available when the VM is powered on and running.
type VMRuntimeInfo struct {
	// Runtime identifiers
	CreationTimestamp metav1.Time `json:"creationTimestamp" yaml:"creationTimestamp"`
	PowerState        string      `json:"powerState" yaml:"powerState"`
	VMIUID            string      `json:"vmiUID" yaml:"vmiUID"`

	// Guest agent data (requires guest agent running inside VM)
	GuestMetadata *GuestMetadata `json:"guestMetadata,omitempty" yaml:"guestMetadata,omitempty"`
}

// VMConsolidatedReport represents a unified view of a VM with both base and runtime information
// VMBaseInfo contains all configuration data (CPU, memory, disks) available for both stopped and running VMs.
// The configuration fields in VMBaseInfo (CPUInfo, MemoryInfo, Disks) will show allocated resources.
// For running VMs, these same fields will also contain usage metrics populated from monitoring/guest agent.
// Runtime contains additional runtime-only data (power state, VMI UID, guest metadata) - nil if VM is stopped.
type VMConsolidatedReport struct {
	VMBaseInfo                // Embedded base info (always available)
	Runtime    *VMRuntimeInfo `json:"runtime,omitempty" yaml:"runtime,omitempty"` // Runtime info (nil if VM is stopped)
}

// VMDetails represents detailed information about a specific VM
type VMDetails struct {
	VMBaseInfo
	Events   []corev1.Event                 `json:"events,omitempty" yaml:"events,omitempty"`
	Pods     []corev1.Pod                   `json:"pods,omitempty" yaml:"pods,omitempty"`
	PVCs     []corev1.PersistentVolumeClaim `json:"pvcs,omitempty" yaml:"pvcs,omitempty"`
	Runtime  *VMRuntimeInfo                 `json:"runtime,omitempty" yaml:"runtime,omitempty"`
	Services []corev1.Service               `json:"services,omitempty" yaml:"services,omitempty"`
}

// DiskInfo represents information about a VM disk
type VMDiskInfo struct {
	BusType    string `json:"busType" yaml:"busType"`
	DiskName   string `json:"diskName" yaml:"diskName"`
	FsType     string `json:"fsType" yaml:"fsType"`
	MountPoint string `json:"mountPoint" yaml:"mountPoint"`
	TotalBytes int64  `json:"totalBytes" yaml:"totalBytes"`
	UsedBytes  int64  `json:"usedBytes" yaml:"usedBytes"`
}

// VMList represents a collection of VMs with summary information
type VMList struct {
	Namespaces []string               `json:"namespaces" yaml:"namespaces"`
	RunningVMs int                    `json:"runningVMs" yaml:"runningVMs"`
	StoppedVMs int                    `json:"stoppedVMs" yaml:"stoppedVMs"`
	TotalVMs   int                    `json:"totalVMs" yaml:"totalVMs"`
	VMs        []VMConsolidatedReport `json:"vms" yaml:"vms"`
}

// VMNetworkInfo represents information about a VM network interface
type VMNetworkInfo struct {
	IPAddresses                 []string `json:"IPAddresses" yaml:"IPAddresses"`
	MACAddress                  string   `json:"macAddress" yaml:"macAddress"`
	Model                       string   `json:"model" yaml:"model"`
	Name                        string   `json:"name" yaml:"name"`
	Network                     string   `json:"network" yaml:"network"`
	NetworkAttachmentDefinition string   `json:"networkAttachmentDefinition" yaml:"networkAttachmentDefinition"`
	Type                        string   `json:"type" yaml:"type"`
}
