package output

import (
	"vm-scanner/pkg/cluster"
	"vm-scanner/pkg/hardware"
	"vm-scanner/pkg/vm"
)

// Formatter provides methods to format information for output
type Formatter struct {
	outputFormat string
	outputFile   string
}

// ComprehensiveReport represents a comprehensive RVTools-like report
type ComprehensiveReport struct {
	Cluster     *cluster.ClusterSummary     `json:"cluster" yaml:"cluster"`
	DataVolumes []DataVolumeInfo            `json:"dataVolumes" yaml:"dataVolumes"`
	GeneratedAt string                      `json:"generatedAt" yaml:"generatedAt"`
	GeneratedBy string                      `json:"generatedBy" yaml:"generatedBy"`
	NADs        []NADInfo                   `json:"nads" yaml:"nads"`
	Nodes       []cluster.ClusterNodeInfo   `json:"nodes" yaml:"nodes"`
	PVCs        []PVCInventoryItem          `json:"pvcs" yaml:"pvcs"`
	Storage     []hardware.StorageClassInfo `json:"storage" yaml:"storage"`
	Summary     *ReportSummary              `json:"summary" yaml:"summary"`
	VMs         []vm.VMDetails              `json:"vms" yaml:"vms"`
}

// ReportSummary represents summary information for the report
type ReportSummary struct {
	ClusterSummary *cluster.ClusterSummary `json:"clusterInfo" yaml:"clusterInfo"`
	StorageClasses int                     `json:"storageClasses" yaml:"storageClasses"`
}

// Report-facing aliases for cluster inventory types (avoids cluster↔output import cycle).
type (
	PVCInventoryItem = cluster.PVCInventoryItem
	NADInfo          = cluster.NADInfo
	DataVolumeInfo   = cluster.DataVolumeInfo
)

// ViewType represents the report organization approach
type ViewType string

const (
	NodeCentric ViewType = "node"
	ClusterWide ViewType = "cluster"
)

// ComprehensiveData represents all collected cluster data
type ComprehensiveData struct {
	ClusterName string         `json:"clusterName" yaml:"clusterName"`
	GeneratedAt string         `json:"generatedAt" yaml:"generatedAt"`
	Nodes       []NodeData     `json:"nodes" yaml:"nodes"`
	Security    SecurityInfo   `json:"security" yaml:"security"`
	VMs         []vm.VMDetails `json:"vms" yaml:"vms"`
}

// NodeData represents per-node information
type NodeData struct {
	Hardware    cluster.ClusterNodeInfo `json:"hardware" yaml:"hardware"`
	Name        string                  `json:"name" yaml:"name"`
	Utilization ResourceMetrics         `json:"utilization" yaml:"utilization"`
	VMs         []vm.VMDetails          `json:"vms" yaml:"vms"`
}

// SecurityInfo represents cluster-wide security information
type SecurityInfo struct {
	ClusterRoleBindings []ClusterRoleBinding `json:"clusterRoleBindings" yaml:"clusterRoleBindings"`
	IdentityProviders   []IdentityProvider   `json:"identityProviders" yaml:"identityProviders"`
	ServiceAccounts     []ServiceAccount     `json:"serviceAccounts" yaml:"serviceAccounts"`
}

// ResourceMetrics represents node resource utilization
type ResourceMetrics struct {
	CPUUsagePercent     float64 `json:"cpuUsagePercent" yaml:"cpuUsagePercent"`
	MemoryUsagePercent  float64 `json:"memoryUsagePercent" yaml:"memoryUsagePercent"`
	NetworkBytesIn      int64   `json:"networkBytesIn" yaml:"networkBytesIn"`
	NetworkBytesOut     int64   `json:"networkBytesOut" yaml:"networkBytesOut"`
	StorageUsagePercent float64 `json:"storageUsagePercent" yaml:"storageUsagePercent"`
}

type ClusterRoleBinding struct {
	Name    string `json:"name" yaml:"name"`
	Role    string `json:"role" yaml:"role"`
	Subject string `json:"subject" yaml:"subject"`
}

type IdentityProvider struct {
	Name string `json:"name" yaml:"name"`
	Type string `json:"type" yaml:"type"`
}

type ServiceAccount struct {
	Name      string `json:"name" yaml:"name"`
	Namespace string `json:"namespace" yaml:"namespace"`
}

// MigrationAssessment captures per-VM live migration readiness scoring and blockers.
type MigrationAssessment struct {
	VMName             string
	VMNamespace        string
	PowerState         string
	LiveMigratable     string // "Yes", "No", "Unknown"
	RunStrategy        string
	EvictionStrategy   string
	HasHostDevices     string // "Yes", "No"
	HasNodeAffinity    string // "Yes", "No"
	PVCAccessModeIssue string // "Yes" (has RWO PVC), "No" (all RWX or no PVCs)
	HasDedicatedCPU    string // "Yes", "No"
	GuestAgentReady    string // "Yes", "No"
	Blockers           []string // human-readable blocker descriptions
	ReadinessScore     string   // "X/10"
}

// StorageAnalysisRow joins PVC inventory with VM disk usage for utilization reporting.
type StorageAnalysisRow struct {
	PVCName        string
	Namespace      string
	StorageClass   string
	CapacityGiB    float64
	AccessModes    string
	VolumeMode     string
	Status         string
	OwningVM       string
	VMPowerState   string // Running/Stopped/empty if no VM
	GuestUsedGiB   float64 // from guest agent, 0 if unavailable
	UtilizationPct float64 // (GuestUsed / Capacity) * 100
	Flag           string  // "Orphaned", "Overprovisioned", "Low Utilization", ""
}
