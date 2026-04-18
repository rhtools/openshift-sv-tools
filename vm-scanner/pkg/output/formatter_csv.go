package output

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"vm-scanner/pkg/cluster"
	"vm-scanner/pkg/hardware"
	"vm-scanner/pkg/vm"
)

// CSVFormatter handles CSV output format
type CSVFormatter struct {
	outputFile string
}

// NewCSVFormatter creates a new CSV formatter
func NewCSVFormatter(outputFile string) *CSVFormatter {
	return &CSVFormatter{
		outputFile: outputFile,
	}
}

// Format routes to appropriate CSV formatter based on data type
func (f *CSVFormatter) Format(data interface{}) error {
	switch v := data.(type) {
	case *cluster.ClusterNodeInfo:
		return f.formatNodeInfo(v)
	case []cluster.ClusterNodeInfo:
		return f.formatNodeList(v)
	case *cluster.ClusterSummary:
		return f.formatClusterSummary(v)
	case *ComprehensiveReport:
		return f.formatComprehensiveReport(v)
	case *ComprehensiveData:
		return fmt.Errorf("ComprehensiveData CSV formatting not implemented")
	case hardware.StorageClassInfo:
		return f.formatHardwareStorageClass(&v)
	case *hardware.StorageClassInfo:
		return f.formatHardwareStorageClass(v)
	case []hardware.StorageClassInfo:
		return f.formatHardwareStorageClassList(v)
	case map[string]vm.VMConsolidatedReport:
		return f.formatVMRuntimeInfoMap(v)
	default:
		return fmt.Errorf("CSV format not supported for type %T", v)
	}
}

// formatNodeInfo formats a single NodeHardwareInfo as CSV (vertical format)
func (f *CSVFormatter) formatNodeInfo(nodeInfo *cluster.ClusterNodeInfo) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	writer.Write([]string{"Property", "Value"})

	writer.Write([]string{"Node Name", nodeInfo.ClusterNodeName})
	writer.Write([]string{"Node Roles", strings.Join(nodeInfo.NodeRoles, "; ")})
	writer.Write([]string{"OS Version", nodeInfo.CoreOSVersion})
	writer.Write([]string{"Kernel Version", nodeInfo.NodeKernelVersion})
	writer.Write([]string{"Schedulable", nodeInfo.NodeSchedulable})

	writer.Write([]string{"CPU Cores", fmt.Sprintf("%d", nodeInfo.CPU.CPUCores)})
	writer.Write([]string{"CPU Model", nodeInfo.CPU.CPUModel})
	writer.Write([]string{"Memory", fmt.Sprintf("%.1f", nodeInfo.Memory.MemoryCapacityGiB)})
	writer.Write([]string{"Storage", fmt.Sprintf("%.2f", nodeInfo.StorageCapacity)})
	writer.Write([]string{"Pod Limits", fmt.Sprintf("%d", nodeInfo.NodePodLimits)})

	writer.Write([]string{"Filesystem Available", fmt.Sprintf("%.2f", nodeInfo.Filesystem.FilesystemAvailable)})
	writer.Write([]string{"Filesystem Capacity", fmt.Sprintf("%.2f", nodeInfo.Filesystem.FilesystemCapacity)})
	writer.Write([]string{"Filesystem Used", fmt.Sprintf("%.2f", nodeInfo.Filesystem.FilesystemUsed)})
	writer.Write([]string{"Filesystem Usage %", fmt.Sprintf("%.2f", nodeInfo.Filesystem.FilesystemUsagePercent)})

	writer.Write([]string{"Host CIDRs", strings.Join(nodeInfo.Network.HostCIDRs, "; ")})
	writer.Write([]string{"MAC Address", nodeInfo.Network.MACAddress})
	writer.Write([]string{"Network Interfaces", FormatNICNames(nodeInfo.Network.NetworkInterfaces)})
	writer.Write([]string{"Network Interface Speeds", FormatNICSpeedList(nodeInfo.Network.NetworkInterfaces)})
	writer.Write([]string{"Next Hops", strings.Join(nodeInfo.Network.NextHops, "; ")})

	return nil
}

// formatNodeList formats multiple nodes as horizontal CSV
func (f *CSVFormatter) formatNodeList(nodeList []cluster.ClusterNodeInfo) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	writer.Write([]string{
		"Node Name",
		"Node Roles",
		"CPU Cores",
		"CPU Model",
		"Memory Capacity (GiB)",
		"Memory Used (GiB)",
		"Memory Used (%)",
		"Storage",
		"Filesystem Available (GB)",
		"Filesystem Used (GB)",
		"Filesystem Usage (%)",
		"OS Version",
		"Kernel Version",
		"Schedulable",
		"Pod Limits",
		"Host CIDRs",
		"MAC Address",
		"Network Interfaces",
		"Network Interface Speeds",
		"Next Hops",
	})

	for _, node := range nodeList {
		memoryUsedPercent := 0.0
		if node.Memory.MemoryCapacityGiB > 0 {
			memoryUsedPercent = (node.Memory.MemoryUsedGiB / node.Memory.MemoryCapacityGiB) * 100
		}
		writer.Write([]string{
			node.ClusterNodeName,
			strings.Join(node.NodeRoles, "; "),
			fmt.Sprintf("%d", node.CPU.CPUCores),
			node.CPU.CPUModel,
			fmt.Sprintf("%.1f", node.Memory.MemoryCapacityGiB),
			fmt.Sprintf("%.1f", node.Memory.MemoryUsedGiB),
			fmt.Sprintf("%.1f", memoryUsedPercent),
			fmt.Sprintf("%.2f", node.StorageCapacity),
			fmt.Sprintf("%.2f", node.Filesystem.FilesystemAvailable),
			fmt.Sprintf("%.2f", node.Filesystem.FilesystemUsed),
			fmt.Sprintf("%.2f", node.Filesystem.FilesystemUsagePercent),
			node.CoreOSVersion,
			node.NodeKernelVersion,
			node.NodeSchedulable,
			fmt.Sprintf("%d", node.NodePodLimits),
			strings.Join(node.Network.HostCIDRs, "; "),
			node.Network.MACAddress,
			FormatNICNames(node.Network.NetworkInterfaces),
			FormatNICSpeedList(node.Network.NetworkInterfaces),
			strings.Join(node.Network.NextHops, "; "),
		})
	}

	return nil
}

// formatVMRuntimeInfoMap formats VM runtime inventory map as CSV
func (f *CSVFormatter) formatVMRuntimeInfoMap(vmiMap map[string]vm.VMConsolidatedReport) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	writer.Write([]string{
		"Name",
		"Namespace",
		"UID",
		"Power State",
		"Running On Node",
		"OS Name",
		"OS Version",
		"Machine Type",
		"Hostname",
		"Kernel Version",
		"Guest Agent Version",
		"Timezone",
		"Virt Launcher Pod",
		"Memory Used By VMI (MiB)",
		"Memory Free (MiB)",
		"Total Memory Used (MiB)",
		"Memory Used By LibVirt (MiB)",
		"Memory Used Percentage (%)",
		"Memory HotPlug Max (MiB)",
		"CPU Cores",
		"CPU Sockets",
		"CPU Threads",
	})

	for _, vmReport := range vmiMap {
		builder := NewVMConsolidatedCSVRowBuilder(vmReport)
		row := builder.BuildRow()
		writer.Write(row)
	}

	return nil
}

// formatClusterSummary formats cluster-wide information as CSV
func (f *CSVFormatter) formatClusterSummary(clusterInfo *cluster.ClusterSummary) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	writer.Write([]string{"Metric", "Value"})

	writer.Write([]string{"Cluster Name", clusterInfo.ClusterName})
	writer.Write([]string{"Kubernetes Version", clusterInfo.KubernetesVersion})
	writer.Write([]string{"Total VMs", fmt.Sprintf("%d", clusterInfo.TotalVMs)})
	writer.Write([]string{"Running VMs", fmt.Sprintf("%d", clusterInfo.RunningVMs)})
	writer.Write([]string{"Stopped VMs", fmt.Sprintf("%d", clusterInfo.StoppedVMs)})

	if clusterInfo.KubeVirtVersion != nil {
		writer.Write([]string{"KubeVirt Version", clusterInfo.KubeVirtVersion.Version})
		writer.Write([]string{"KubeVirt Deployed", clusterInfo.KubeVirtVersion.Deployed})
	}

	return nil
}

// formatComprehensiveReport formats the complete report as multiple CSV sections
func (f *CSVFormatter) formatComprehensiveReport(report *ComprehensiveReport) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	writer.Write([]string{"# SUMMARY"})
	writer.Write([]string{"Metric", "Value"})
	if report.Summary != nil && report.Summary.ClusterSummary != nil {
		writer.Write([]string{"Total VMs", fmt.Sprintf("%d", report.Summary.ClusterSummary.TotalVMs)})
		writer.Write([]string{"Running VMs", fmt.Sprintf("%d", report.Summary.ClusterSummary.RunningVMs)})
		writer.Write([]string{"Stopped VMs", fmt.Sprintf("%d", report.Summary.ClusterSummary.StoppedVMs)})
		if report.Summary.ClusterSummary.Resources != nil {
			writer.Write([]string{"Total CPU", fmt.Sprintf("%d", report.Summary.ClusterSummary.Resources.TotalCPU)})
			writer.Write([]string{"Total Memory (GiB)", fmt.Sprintf("%.1f", report.Summary.ClusterSummary.Resources.TotalMemory)})
			writer.Write([]string{"Used Memory (GiB)", fmt.Sprintf("%.1f", report.Summary.ClusterSummary.Resources.UsedMemory)})
			writer.Write([]string{"Total Local Storage (GiB)", fmt.Sprintf("%.1f", float64(report.Summary.ClusterSummary.Resources.TotalLocalStorage))})
			writer.Write([]string{"Total Application Requested Storage (GiB)", fmt.Sprintf("%.1f", float64(report.Summary.ClusterSummary.Resources.TotalApplicationRequestedStorage))})
			writer.Write([]string{"Total Application Used Storage (GiB)", fmt.Sprintf("%.1f", float64(report.Summary.ClusterSummary.Resources.TotalApplicationUsedStorage))})
		}
		writer.Write([]string{"Total Namespaces", fmt.Sprintf("%d", report.Summary.ClusterSummary.TotalNamespaces)})
		writer.Write([]string{"User Created Namespaces", fmt.Sprintf("%d", report.Summary.ClusterSummary.UserNamespaces)})
		writer.Write([]string{"Nodes", fmt.Sprintf("%d", len(report.Nodes))})
		writer.Write([]string{"Storage Classes", fmt.Sprintf("%d", report.Summary.StorageClasses)})
	}

	writer.Write([]string{})
	writer.Write([]string{"# VIRTUAL MACHINES"})
	writer.Write([]string{"Name", "Namespace", "Phase", "Configured CPU", "Configured Memory (MiB)"})
	for _, vmDetail := range report.VMs {
		writer.Write([]string{
			vmDetail.Name,
			vmDetail.Namespace,
			vmDetail.Phase,
			fmt.Sprintf("%d", vmDetail.CPUInfo.VCPUs),
			fmt.Sprintf("%.1f", vmDetail.MemoryInfo.MemoryConfiguredMiB),
		})
	}

	writer.Write([]string{})
	writer.Write([]string{"# STORAGE CLASSES"})
	writer.Write([]string{"Name", "Provisioner", "Reclaim Policy", "Volume Binding Mode", "Default", "Allow Expansion"})
	for _, sc := range report.Storage {
		writer.Write([]string{
			sc.Name,
			sc.Provisioner,
			sc.ReclaimPolicy,
			sc.VolumeBindingMode,
			fmt.Sprintf("%t", sc.IsDefault),
			fmt.Sprintf("%t", sc.AllowVolumeExpansion),
		})
	}

	return nil
}

// formatStorageClass formats a single StorageClassInfo as CSV (vertical format - deprecated)
func (f *CSVFormatter) formatStorageClass(storageClass *hardware.StorageClassInfo) error {
	fmt.Println("Property,Value")

	fmt.Printf("Name,%s\n", storageClass.Name)
	fmt.Printf("Provisioner,%s\n", storageClass.Provisioner)
	fmt.Printf("Reclaim Policy,%s\n", storageClass.ReclaimPolicy)
	fmt.Printf("Volume Binding Mode,%s\n", storageClass.VolumeBindingMode)

	defaultStatus := "No"
	if storageClass.IsDefault {
		defaultStatus = "Yes"
	}
	fmt.Printf("Default Storage Class,%s\n", defaultStatus)

	if len(storageClass.Parameters) > 0 {
		for key, value := range storageClass.Parameters {
			fmt.Printf("Parameter: %s,%s\n", key, escapeCSVValue(value))
		}
	} else {
		fmt.Println("Parameters,None")
	}

	return nil
}

// formatStorageClassList formats multiple StorageClasses as horizontal CSV (deprecated)
func (f *CSVFormatter) formatStorageClassList(storageClasses []hardware.StorageClassInfo) error {
	if len(storageClasses) == 0 {
		fmt.Println("No storage classes found.")
		return nil
	}

	fmt.Println("Name,Provisioner,Reclaim Policy,Volume Binding Mode,Default,Parameter Count")

	for _, sc := range storageClasses {
		defaultStatus := "No"
		if sc.IsDefault {
			defaultStatus = "Yes"
		}

		fmt.Printf("%s,%s,%s,%s,%s,%d\n",
			escapeCSVValue(sc.Name),
			escapeCSVValue(sc.Provisioner),
			escapeCSVValue(sc.ReclaimPolicy),
			escapeCSVValue(sc.VolumeBindingMode),
			defaultStatus,
			len(sc.Parameters))
	}

	hasParameters := false
	for _, sc := range storageClasses {
		if len(sc.Parameters) > 0 {
			hasParameters = true
			break
		}
	}

	if hasParameters {
		fmt.Println("\n# Storage Class Parameters")
		fmt.Println("Storage Class,Parameter,Value")
		for _, sc := range storageClasses {
			if len(sc.Parameters) > 0 {
				for key, value := range sc.Parameters {
					fmt.Printf("%s,%s,%s\n",
						escapeCSVValue(sc.Name),
						escapeCSVValue(key),
						escapeCSVValue(value))
				}
			}
		}
	}

	return nil
}

// formatHardwareStorageClass formats a single hardware.StorageClassInfo as CSV (vertical format)
func (f *CSVFormatter) formatHardwareStorageClass(storageClass *hardware.StorageClassInfo) error {
	fmt.Println("Property,Value")

	fmt.Printf("Name,%s\n", escapeCSVValue(storageClass.Name))
	fmt.Printf("Provisioner,%s\n", escapeCSVValue(storageClass.Provisioner))
	fmt.Printf("Reclaim Policy,%s\n", escapeCSVValue(storageClass.ReclaimPolicy))
	fmt.Printf("Volume Binding Mode,%s\n", escapeCSVValue(storageClass.VolumeBindingMode))

	defaultStatus := "No"
	if storageClass.IsDefault {
		defaultStatus = "Yes"
	}
	fmt.Printf("Default Storage Class,%s\n", defaultStatus)

	expansionStatus := "No"
	if storageClass.AllowVolumeExpansion {
		expansionStatus = "Yes"
	}
	fmt.Printf("Allow Volume Expansion,%s\n", expansionStatus)

	fmt.Printf("Created At,%s\n", storageClass.CreatedAt.Format("2006-01-02 15:04:05"))

	mountOptions := "None"
	if len(storageClass.MountOptions) > 0 {
		mountOptions = strings.Join(storageClass.MountOptions, "; ")
	}
	fmt.Printf("Mount Options,%s\n", escapeCSVValue(mountOptions))

	allowedTopologies := "None"
	if len(storageClass.AllowedTopologies) > 0 {
		allowedTopologies = strings.Join(storageClass.AllowedTopologies, "; ")
	}
	fmt.Printf("Allowed Topologies,%s\n", escapeCSVValue(allowedTopologies))

	if len(storageClass.Parameters) > 0 {
		for key, value := range storageClass.Parameters {
			fmt.Printf("Parameter: %s,%s\n", escapeCSVValue(key), escapeCSVValue(value))
		}
	} else {
		fmt.Println("Parameters,None")
	}

	return nil
}

// formatHardwareStorageClassList formats multiple hardware.StorageClasses as horizontal CSV
func (f *CSVFormatter) formatHardwareStorageClassList(storageClasses []hardware.StorageClassInfo) error {
	if len(storageClasses) == 0 {
		fmt.Println("No storage classes found.")
		return nil
	}

	fmt.Println("Name,Provisioner,Reclaim Policy,Volume Binding Mode,Default,Allow Volume Expansion,Parameter Count,Created At")

	for _, sc := range storageClasses {
		defaultStatus := "No"
		if sc.IsDefault {
			defaultStatus = "Yes"
		}

		expansionStatus := "No"
		if sc.AllowVolumeExpansion {
			expansionStatus = "Yes"
		}

		fmt.Printf("%s,%s,%s,%s,%s,%s,%d,%s\n",
			escapeCSVValue(sc.Name),
			escapeCSVValue(sc.Provisioner),
			escapeCSVValue(sc.ReclaimPolicy),
			escapeCSVValue(sc.VolumeBindingMode),
			defaultStatus,
			expansionStatus,
			len(sc.Parameters),
			sc.CreatedAt.Format("2006-01-02"))
	}

	hasParameters := false
	for _, sc := range storageClasses {
		if len(sc.Parameters) > 0 {
			hasParameters = true
			break
		}
	}

	if hasParameters {
		fmt.Println("\n# Storage Class Parameters")
		fmt.Println("Storage Class,Parameter,Value")
		for _, sc := range storageClasses {
			if len(sc.Parameters) > 0 {
				for key, value := range sc.Parameters {
					fmt.Printf("%s,%s,%s\n",
						escapeCSVValue(sc.Name),
						escapeCSVValue(key),
						escapeCSVValue(value))
				}
			}
		}
	}

	return nil
}

// escapeCSVValue escapes special characters for CSV format
func escapeCSVValue(value string) string {
	if strings.Contains(value, ",") || strings.Contains(value, "\"") || strings.Contains(value, "\n") || strings.Contains(value, "\r") {
		escaped := strings.ReplaceAll(value, "\"", "\"\"")
		return fmt.Sprintf("\"%s\"", escaped)
	}
	return value
}
