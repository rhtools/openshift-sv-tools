package hardware

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterNetworkConfig represents cluster-wide network configuration
type ClusterNetworkConfig struct {
	ClusterNetwork     []string `json:"clusterNetwork" yaml:"clusterNetwork"`
	ClusterNetworkMTU  int      `json:"clusterNetworkMTU" yaml:"clusterNetworkMTU"`
	ClusterNetworkType string   `json:"clusterNetworkType" yaml:"clusterNetworkType"`
	ExternalIP         string   `json:"externalIP" yaml:"externalIP"`
	ServiceNetwork     []string `json:"serviceNetwork" yaml:"serviceNetwork"`
}

// CPUInfo represents CPU information for a node
type CPUInfo struct {
	CPUCores int64  `json:"cpuCores" yaml:"cpuCores"`
	CPUModel string `json:"cpuModel" yaml:"cpuModel"`
}

// InterfaceNames represents network interface names
type InterfaceNames struct {
	Name string `json:"name" yaml:"name"`
}

// NICInfo represents detailed information about a physical network interface
type NICInfo struct {
	Duplex     string `json:"duplex" yaml:"duplex"`
	IPAddress  string `json:"ipAddress,omitempty" yaml:"ipAddress,omitempty"`
	MACAddress string `json:"macAddress,omitempty" yaml:"macAddress,omitempty"`
	Model      string `json:"model,omitempty" yaml:"model,omitempty"`
	Name       string `json:"name" yaml:"name"`
	SpeedMbps  int    `json:"speedMbps" yaml:"speedMbps"`
	State      string `json:"state" yaml:"state"`
}

// L3GatewayAnnotation represents the OVN L3 gateway configuration annotation
type L3GatewayAnnotation struct {
	Default L3GatewayConfig `json:"default" yaml:"default"`
}

// L3GatewayConfig represents the OVN L3 gateway configuration
type L3GatewayConfig struct {
	BridgeID       string   `json:"bridge-id" yaml:"bridge-id"`
	InterfaceID    string   `json:"interface-id" yaml:"interface-id"`
	IPAddress      string   `json:"ip-address" yaml:"ip-address"`
	IPAddresses    []string `json:"ip-addresses" yaml:"ip-addresses"`
	MACAddress     string   `json:"mac-address" yaml:"mac-address"`
	Mode           string   `json:"mode" yaml:"mode"`
	NextHops       []string `json:"next-hops" yaml:"next-hops"`
	NodePortEnable bool     `json:"node-port-enable" yaml:"node-port-enable"`
	VLANID         int      `json:"vlan-id,omitempty" yaml:"vlan-id,omitempty"`
}

// MemoryInfo represents memory information for a node
type MemoryInfo struct {
	MemoryCapacityGiB    float64 `json:"memoryCapacity" yaml:"memoryCapacity"`
	MemoryUsedGiB        float64 `json:"memoryUsed" yaml:"memoryUsed"`
	MemoryUsedPercentage float64 `json:"memoryUsedPercentage" yaml:"memoryUsedPercentage"`
}

// NetworkInfo represents node network configuration
type NetworkInfo struct {
	BridgeID          string              `json:"bridgeID" yaml:"bridgeID"`
	HostCIDRs         []string            `json:"hostCIDRs" yaml:"hostCIDRs"`
	InterfaceID       string              `json:"interfaceID" yaml:"interfaceID"`
	IPAddresses       []string            `json:"ipAddresses" yaml:"ipAddresses"`
	MACAddress        string              `json:"macAddress" yaml:"macAddress"`
	Mode              string              `json:"mode" yaml:"mode"`
	NetworkInterfaces []NICInfo            `json:"networkInterfaces" yaml:"networkInterfaces"`
	NextHops          []string            `json:"nextHops" yaml:"nextHops"`
	NodePortEnable    bool                `json:"nodePortEnable" yaml:"nodePortEnable"`
	PodNetworkSubnet  map[string][]string `json:"nodeSubnets" yaml:"nodeSubnets"`
	VLANID            int                 `json:"vlanID,omitempty" yaml:"vlanID,omitempty"`
}

// NetworkStats represents network statistics from the node
type NetworkStats struct {
	Interfaces []InterfaceNames `json:"interfaces" yaml:"interfaces"`
}

// NodeFilesystemInfo represents filesystem usage information for a node (all values in GiB)
type NodeFilesystemInfo struct {
	FilesystemAvailable    float64 `json:"filesystemAvailable" yaml:"filesystemAvailable"`
	FilesystemCapacity     float64 `json:"filesystemCapacity" yaml:"filesystemCapacity"`
	FilesystemUsed         float64 `json:"filesystemUsed" yaml:"filesystemUsed"`
	FilesystemUsagePercent float64 `json:"filesystemUsagePercent" yaml:"filesystemUsagePercent"`
}

// NodeStatsResponse represents the response from the node stats API
type NodeStatsResponse struct {
	Node struct {
		Network NetworkStats `json:"network" yaml:"network"`
	} `json:"node" yaml:"node"`
}

// StorageClassInfo represents storage class configuration
type StorageClassInfo struct {
	AllowVolumeExpansion        bool              `json:"allowVolumeExpansion" yaml:"allowVolumeExpansion"`
	AllowedTopologies           []string          `json:"allowedTopologies" yaml:"allowedTopologies"`
	AllowedUnsafeToEvictVolumes []string          `json:"allowedUnsafeToEvictVolumes" yaml:"allowedUnsafeToEvictVolumes"`
	CreatedAt                   metav1.Time       `json:"createdAt" yaml:"createdAt"`
	IsDefault                   bool              `json:"isDefault" yaml:"isDefault"`
	MountOptions                []string          `json:"mountOptions" yaml:"mountOptions"`
	Name                        string            `json:"name" yaml:"name"`
	Parameters                  map[string]string `json:"parameters" yaml:"parameters"`
	Provisioner                 string            `json:"provisioner" yaml:"provisioner"`
	ReclaimPolicy               string            `json:"reclaimPolicy" yaml:"reclaimPolicy"`
	VolumeBindingMode           string            `json:"volumeBindingMode" yaml:"volumeBindingMode"`
}
