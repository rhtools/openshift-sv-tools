package output

import (
	"fmt"
	"os"
	"strings"
	"vm-scanner/pkg/cluster"
	"vm-scanner/pkg/hardware"

	"github.com/jedib0t/go-pretty/v6/table"
)

// NodeTableFormatter formats node information for table display
type NodeTableFormatter struct {
	node cluster.ClusterNodeInfo
}

// NewNodeTableFormatter creates a formatter for node table display
func NewNodeTableFormatter(node cluster.ClusterNodeInfo) *NodeTableFormatter {
	return &NodeTableFormatter{node: node}
}

// FormatSystemInfo prints system information section
func (n *NodeTableFormatter) FormatSystemInfo() {
	fmt.Println("  SYSTEM INFORMATION:")
	fmt.Printf("    Node Name: %s\n", n.node.ClusterNodeName)
	fmt.Printf("    Node Roles: %s\n", strings.Join(n.node.NodeRoles, "; "))
	fmt.Printf("    OS Version: %s\n", n.node.CoreOSVersion)
	fmt.Printf("    Kernel Version: %s\n", n.node.NodeKernelVersion)
	fmt.Printf("    Schedulable: %s\n", n.node.NodeSchedulable)
	fmt.Println()
}

// FormatResourceCapacity prints resource capacity section
func (n *NodeTableFormatter) FormatResourceCapacity() {
	fmt.Println("  RESOURCE CAPACITY:")
	fmt.Printf("    CPU Cores: %d\n", n.node.CPU.CPUCores)
	fmt.Printf("    CPU Model: %s\n", n.node.CPU.CPUModel)
	fmt.Printf("    Memory Capacity (GiB): %.1f\n", n.node.Memory.MemoryCapacityGiB)
	fmt.Printf("    Memory Used (GiB): %.1f\n", n.node.Memory.MemoryUsedGiB)

	memoryUsedPercent := 0.0
	if n.node.Memory.MemoryCapacityGiB > 0 {
		memoryUsedPercent = (n.node.Memory.MemoryUsedGiB / n.node.Memory.MemoryCapacityGiB) * 100
	}
	fmt.Printf("    Memory Used (%%): %.1f\n", memoryUsedPercent)
	fmt.Printf("    Storage Capacity (GiB): %.2f\n", n.node.StorageCapacity)
	fmt.Printf("    Pod Limits: %d\n", n.node.NodePodLimits)
	fmt.Println()
}

// FormatFilesystem prints filesystem section
func (n *NodeTableFormatter) FormatFilesystem() {
	fmt.Println("  FILESYSTEM STATUS:")
	fmt.Printf("    Available (GiB): %.2f\n", n.node.Filesystem.FilesystemAvailable)
	fmt.Printf("    Capacity (GiB): %.2f\n", n.node.Filesystem.FilesystemCapacity)
	fmt.Printf("    Used (GiB): %.2f\n", n.node.Filesystem.FilesystemUsed)
	fmt.Printf("    Usage Percentage: %.2f%%\n", n.node.Filesystem.FilesystemUsagePercent)
	fmt.Println()
}

// FormatNetwork prints network configuration section
func (n *NodeTableFormatter) FormatNetwork() {
	fmt.Println("  NETWORK CONFIGURATION:")

	network := n.node.Network

	fmt.Printf("    Bridge ID: %s\n", formatOptionalString(network.BridgeID))
	fmt.Printf("    Interface ID: %s\n", formatOptionalString(network.InterfaceID))
	fmt.Printf("    IP Addresses: %s\n", formatStringSlice(network.IPAddresses))
	fmt.Printf("    MAC Address: %s\n", formatOptionalString(network.MACAddress))
	fmt.Printf("    Mode: %s\n", formatOptionalString(network.Mode))
	fmt.Printf("    Network Interfaces: %s\n", FormatNICNames(network.NetworkInterfaces))
	fmt.Printf("    Network Interface Speeds: %s\n", FormatNICSpeedList(network.NetworkInterfaces))
	fmt.Printf("    Next Hops: %s\n", formatStringSlice(network.NextHops))
	fmt.Printf("    Node Port Enable: %t\n", network.NodePortEnable)
	fmt.Printf("    VLAN ID: %d\n", network.VLANID)
	fmt.Printf("    Host CIDRs: %s\n", formatStringSlice(network.HostCIDRs))

	if len(network.PodNetworkSubnet) > 0 {
		for key, values := range network.PodNetworkSubnet {
			subnetValue := strings.Join(values, ", ")
			fmt.Printf("    Subnets (%s): %s\n", key, subnetValue)
		}
	} else {
		fmt.Printf("    Subnets: None\n")
	}
}

// formatOptionalString returns the string or "None" if empty
func formatOptionalString(s string) string {
	if len(s) > 0 {
		return s
	}
	return "None"
}

// formatStringSlice returns joined strings or "None" if empty
func formatStringSlice(slice []string) string {
	if len(slice) > 0 {
		return strings.Join(slice, ", ")
	}
	return "None"
}

// NodePrettyTableFormatter formats node information using go-pretty tables
type NodePrettyTableFormatter struct {
	node cluster.ClusterNodeInfo
}

// NewNodePrettyTableFormatter creates a formatter for go-pretty table display
func NewNodePrettyTableFormatter(node cluster.ClusterNodeInfo) *NodePrettyTableFormatter {
	return &NodePrettyTableFormatter{node: node}
}

// FormatSystemInfoTable creates and renders system information table
func (n *NodePrettyTableFormatter) FormatSystemInfoTable() {
	fmt.Println("\n=== SYSTEM INFORMATION ===")
	sysTable := table.NewWriter()
	sysTable.SetOutputMirror(os.Stdout)
	sysTable.SetStyle(table.StyleColoredBright)
	sysTable.AppendHeader(table.Row{"Property", "Value"})

	sysTable.AppendRow(table.Row{"Node Name", n.node.ClusterNodeName})
	sysTable.AppendRow(table.Row{"Node Roles", strings.Join(n.node.NodeRoles, "; ")})
	sysTable.AppendRow(table.Row{"Core OS Version", n.node.CoreOSVersion})
	sysTable.AppendRow(table.Row{"Kernel Version", n.node.NodeKernelVersion})
	sysTable.AppendRow(table.Row{"Schedulable", n.node.NodeSchedulable})

	sysTable.Render()
	fmt.Println()
}

// FormatResourceCapacityTable creates and renders resource capacity table
func (n *NodePrettyTableFormatter) FormatResourceCapacityTable() {
	fmt.Println("=== RESOURCE CAPACITY ===")
	resTable := table.NewWriter()
	resTable.SetOutputMirror(os.Stdout)
	resTable.SetStyle(table.StyleColoredBright)
	resTable.AppendHeader(table.Row{"Resource", "Value"})

	resTable.AppendRow(table.Row{"CPU Cores", fmt.Sprintf("%d", n.node.CPU.CPUCores)})
	resTable.AppendRow(table.Row{"CPU Model", n.node.CPU.CPUModel})
	resTable.AppendRow(table.Row{"Memory Capacity (GiB)", fmt.Sprintf("%.1f", n.node.Memory.MemoryCapacityGiB)})
	resTable.AppendRow(table.Row{"Memory Used (GiB)", fmt.Sprintf("%.1f", n.node.Memory.MemoryUsedGiB)})

	memoryUsedPercent := 0.0
	if n.node.Memory.MemoryCapacityGiB > 0 {
		memoryUsedPercent = (n.node.Memory.MemoryUsedGiB / n.node.Memory.MemoryCapacityGiB) * 100
	}
	resTable.AppendRow(table.Row{"Memory Used (%)", fmt.Sprintf("%.1f", memoryUsedPercent)})
	resTable.AppendRow(table.Row{"Storage Capacity (GiB)", fmt.Sprintf("%.2f", n.node.StorageCapacity)})
	resTable.AppendRow(table.Row{"Pod Limits", fmt.Sprintf("%d", n.node.NodePodLimits)})

	resTable.Render()
	fmt.Println()
}

// FormatFilesystemTable creates and renders filesystem table
func (n *NodePrettyTableFormatter) FormatFilesystemTable() {
	fmt.Println("=== FILESYSTEM STATUS ===")
	fsTable := table.NewWriter()
	fsTable.SetOutputMirror(os.Stdout)
	fsTable.SetStyle(table.StyleColoredBright)
	fsTable.AppendHeader(table.Row{"Metric", "Value"})

	fsTable.AppendRow(table.Row{"Available (GiB)", fmt.Sprintf("%.2f", n.node.Filesystem.FilesystemAvailable)})
	fsTable.AppendRow(table.Row{"Capacity (GiB)", fmt.Sprintf("%.2f", n.node.Filesystem.FilesystemCapacity)})
	fsTable.AppendRow(table.Row{"Used (GiB)", fmt.Sprintf("%.2f", n.node.Filesystem.FilesystemUsed)})
	fsTable.AppendRow(table.Row{"Usage Percentage", fmt.Sprintf("%.2f%%", n.node.Filesystem.FilesystemUsagePercent)})

	fsTable.Render()
	fmt.Println()
}

// FormatNetworkTable creates and renders network configuration table
func (n *NodePrettyTableFormatter) FormatNetworkTable() {
	fmt.Println("=== NETWORK CONFIGURATION ===")
	netTable := table.NewWriter()
	netTable.SetOutputMirror(os.Stdout)
	netTable.SetStyle(table.StyleColoredBright)
	netTable.AppendHeader(table.Row{"Configuration", "Value"})

	network := n.node.Network

	netTable.AppendRow(table.Row{"Bridge ID", formatOptionalString(network.BridgeID)})
	netTable.AppendRow(table.Row{"Interface ID", formatOptionalString(network.InterfaceID)})
	netTable.AppendRow(table.Row{"IP Addresses", formatStringSlice(network.IPAddresses)})
	netTable.AppendRow(table.Row{"MAC Address", formatOptionalString(network.MACAddress)})
	netTable.AppendRow(table.Row{"Mode", formatOptionalString(network.Mode)})
	netTable.AppendRow(table.Row{"Network Interfaces", FormatNICNames(network.NetworkInterfaces)})
	netTable.AppendRow(table.Row{"Network Interface Speeds", FormatNICSpeedList(network.NetworkInterfaces)})
	netTable.AppendRow(table.Row{"Next Hops", formatStringSlice(network.NextHops)})
	netTable.AppendRow(table.Row{"Node Port Enable", fmt.Sprintf("%t", network.NodePortEnable)})
	netTable.AppendRow(table.Row{"VLAN ID", fmt.Sprintf("%d", network.VLANID)})
	netTable.AppendRow(table.Row{"Host CIDRs", formatStringSlice(network.HostCIDRs)})

	if len(network.PodNetworkSubnet) > 0 {
		for key, values := range network.PodNetworkSubnet {
			subnetValue := strings.Join(values, ", ")
			netTable.AppendRow(table.Row{fmt.Sprintf("Subnets (%s)", key), subnetValue})
		}
	} else {
		netTable.AppendRow(table.Row{"Subnets", "None"})
	}

	netTable.Render()
}

// formatNICSpeed returns a human-readable speed string (e.g. "25 Gbps", "1000 Mbps", "unknown").
func formatNICSpeed(nic hardware.NICInfo) string {
	if nic.SpeedMbps == 0 {
		return "unknown"
	}
	if nic.SpeedMbps >= 1000 && nic.SpeedMbps%1000 == 0 {
		return fmt.Sprintf("%d Gbps", nic.SpeedMbps/1000)
	}
	return fmt.Sprintf("%d Mbps", nic.SpeedMbps)
}

// formatNICSummary returns "name (speed)" for display (e.g. "eno1 (25 Gbps)").
func formatNICSummary(nic hardware.NICInfo) string {
	return fmt.Sprintf("%s (%s)", nic.Name, formatNICSpeed(nic))
}

// FormatNICNames returns a semicolon-separated list of NIC name+speed summaries.
func FormatNICNames(nics []hardware.NICInfo) string {
	if len(nics) == 0 {
		return "None"
	}
	parts := make([]string, len(nics))
	for i, nic := range nics {
		parts[i] = formatNICSummary(nic)
	}
	return strings.Join(parts, "; ")
}

// FormatNICSpeedList returns a semicolon-separated list of speeds only.
func FormatNICSpeedList(nics []hardware.NICInfo) string {
	if len(nics) == 0 {
		return "None"
	}
	parts := make([]string, len(nics))
	for i, nic := range nics {
		parts[i] = formatNICSpeed(nic)
	}
	return strings.Join(parts, "; ")
}
