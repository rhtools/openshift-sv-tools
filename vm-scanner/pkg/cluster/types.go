package cluster

import (
	"vm-scanner/pkg/hardware"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CNIConfiguration represents CNI configuration
type CNIConfiguration struct {
	CNIVersion    string                 `json:"cniVersion" yaml:"cniVersion"`
	Configuration map[string]interface{} `json:"configuration" yaml:"configuration"`
	Name          string                 `json:"name" yaml:"name"`
	Plugins       []CNIPlugin            `json:"plugins" yaml:"plugins"`
	Type          string                 `json:"type" yaml:"type"`
}

// CNIPlugin represents CNI plugin information
type CNIPlugin struct {
	Config map[string]interface{} `json:"config" yaml:"config"`
	Name   string                 `json:"name" yaml:"name"`
	Type   string                 `json:"type" yaml:"type"`
}

// ClusterSummary represents cluster summary information
type ClusterSummary struct {
	CNIConfiguration             *CNIConfiguration `json:"cniConfiguration" yaml:"cniConfiguration"`
	ClusterID                    string            `json:"clusterID" yaml:"clusterID"`
	ClusterName                  string            `json:"clusterName" yaml:"clusterName"`
	ClusterVersion               string            `json:"clusterVersion" yaml:"clusterVersion"`
	HasSchedulableControlPlane   bool              `json:"hasSchedulableControlPlane,omitempty" yaml:"hasSchedulableControlPlane,omitempty"`
	KubernetesVersion            string            `json:"kubernetesVersion" yaml:"kubernetesVersion"`
	KubeVirtVersion              *KubeVirtVersion  `json:"kubeVirtVersion" yaml:"kubeVirtVersion"`
	Operators                    []OperatorStatus  `json:"operators" yaml:"operators"`
	ProtectedNamespaces          int               `json:"protectedNamespaces" yaml:"protectedNamespaces"`
	Resources                    *ClusterResources `json:"resources" yaml:"resources"`
	RunningVMs                   int               `json:"runningVMs" yaml:"runningVMs"`
	SchedulableControlPlaneCount int               `json:"schedulableControlPlaneCount,omitempty" yaml:"schedulableControlPlaneCount,omitempty"`
	StoppedVMs                   int               `json:"stoppedVMs" yaml:"stoppedVMs"`
	TotalNamespaces              int               `json:"totalNamespaces" yaml:"totalNamespaces"`
	TotalVMs                     int               `json:"totalVMs" yaml:"totalVMs"`
	UserNamespaces               int               `json:"userNamespaces" yaml:"userNamespaces"`
	WorkerNodesCount             int               `json:"workerNodesCount" yaml:"workerNodesCount"`
}

type ClusterNodeInfo struct {
	ClusterNodeName   string                      `json:"clusterNodeName" yaml:"clusterNodeName"`
	CoreOSVersion     string                      `json:"coreOSVersion" yaml:"coreOSVersion"`
	CPU               hardware.CPUInfo            `json:"cpu" yaml:"cpu"`
	Filesystem        hardware.NodeFilesystemInfo `json:"filesystem" yaml:"filesystem"`
	Memory            hardware.MemoryInfo         `json:"memory" yaml:"memory"`
	Network           hardware.NetworkInfo        `json:"network" yaml:"network"`
	NodeKernelVersion string                      `json:"nodeKernelVersion" yaml:"nodeKernelVersion"`
	NodePodLimits     int64                       `json:"nodePodLimits" yaml:"nodePodLimits"`
	NodeRoles         []string                    `json:"nodeRoles" yaml:"nodeRoles"`
	NodeSchedulable   string                      `json:"nodeSchedulable" yaml:"nodeSchedulable"`
	StorageCapacity   float64                     `json:"storageCapacity" yaml:"storageCapacity"`
}

// ClusterResources represents cluster resource information
type ClusterResources struct {
	CPUUtilization                   float64 `json:"cpuUtilization" yaml:"cpuUtilization"`
	MemoryUtilization                float64 `json:"memoryUtilization" yaml:"memoryUtilization"`
	StorageUtilization               float64 `json:"storageUtilization" yaml:"storageUtilization"`
	TotalCPU                         int64   `json:"totalCpu" yaml:"totalCpu"`
	TotalMemory                      float64 `json:"totalMemoryGiB" yaml:"totalMemoryGiB"`
	TotalLocalStorage                float64 `json:"totalLocalStorageGiB" yaml:"totalLocalStorageGiB"`
	TotalLocalStorageUsed            float64 `json:"totalLocalStorageUsedGiB" yaml:"totalLocalStorageUsedGiB"`
	TotalApplicationRequestedStorage int64   `json:"totalApplicationRequestedStorageGiB" yaml:"totalApplicationRequestedStorageGiB"`
	TotalApplicationUsedStorage      int64   `json:"totalApplicationUsedStorageGiB" yaml:"totalApplicationUsedStorageGiB"`
	UsedCPU                          int64   `json:"usedCpu" yaml:"usedCpu"`
	UsedMemory                       float64 `json:"usedMemoryGiB" yaml:"usedMemoryGiB"`
	UsedStorage                      int64   `json:"usedStorageGiB" yaml:"usedStorageGiB"`
}

// KubeVirtVersion represents KubeVirt version information
type KubeVirtVersion struct {
	Deployed string `json:"deployed" yaml:"deployed"`
	Version  string `json:"version" yaml:"version"`
}

// NodeCondition represents a node condition
type NodeCondition struct {
	Message string `json:"message" yaml:"message"`
	Reason  string `json:"reason" yaml:"reason"`
	Status  string `json:"status" yaml:"status"`
	Type    string `json:"type" yaml:"type"`
}

type OperatorStatus struct {
	CreatedAt metav1.Time       `json:"createdAt" yaml:"createdAt"`
	Health    string            `json:"health" yaml:"health"`
	Labels    map[string]string `json:"labels" yaml:"labels"`
	Name      string            `json:"name" yaml:"name"`
	Namespace string            `json:"namespace" yaml:"namespace"`
	Source    string            `json:"source" yaml:"source"`
	Status    string            `json:"status" yaml:"status"`
	Version   string            `json:"version" yaml:"version"`
}

// PVCInventoryItem represents a PersistentVolumeClaim in the cluster inventory.
// output.PVCInventoryItem is a type alias so report types stay stable without an import cycle.
type PVCInventoryItem struct {
	AccessModes       []string    `json:"accessModes" yaml:"accessModes"`
	Capacity          int64       `json:"capacity" yaml:"capacity"`
	CapacityHuman     string      `json:"capacityHuman" yaml:"capacityHuman"`
	CreatedAt         metav1.Time `json:"createdAt" yaml:"createdAt"`
	Name              string      `json:"name" yaml:"name"`
	Namespace         string      `json:"namespace" yaml:"namespace"`
	OwningVM          string      `json:"owningVM" yaml:"owningVM"`
	OwningVMNamespace string      `json:"owningVMNamespace" yaml:"owningVMNamespace"`
	Status            string      `json:"status" yaml:"status"`
	StorageClass      string      `json:"storageClass" yaml:"storageClass"`
	VolumeMode        string      `json:"volumeMode" yaml:"volumeMode"`
	VolumeName        string      `json:"volumeName" yaml:"volumeName"`
}

// NADInfo represents a NetworkAttachmentDefinition in the cluster inventory.
type NADInfo struct {
	Config       string      `json:"config" yaml:"config"`
	CreatedAt    metav1.Time `json:"createdAt" yaml:"createdAt"`
	Name         string      `json:"name" yaml:"name"`
	Namespace    string      `json:"namespace" yaml:"namespace"`
	ResourceName string      `json:"resourceName" yaml:"resourceName"`
	Type         string      `json:"type" yaml:"type"`
	VLAN         string      `json:"vlan" yaml:"vlan"`
}

// DataVolumeInfo represents a CDI DataVolume in the cluster inventory.
type DataVolumeInfo struct {
	CreatedAt    metav1.Time `json:"createdAt" yaml:"createdAt"`
	Name         string      `json:"name" yaml:"name"`
	Namespace    string      `json:"namespace" yaml:"namespace"`
	OwningVM     string      `json:"owningVM" yaml:"owningVM"`
	Phase        string      `json:"phase" yaml:"phase"`
	Progress     string      `json:"progress" yaml:"progress"`
	SourceType   string      `json:"sourceType" yaml:"sourceType"`
	StorageClass string      `json:"storageClass" yaml:"storageClass"`
	StorageHuman string      `json:"storageHuman" yaml:"storageHuman"`
	StorageSize  int64       `json:"storageSize" yaml:"storageSize"`
}
