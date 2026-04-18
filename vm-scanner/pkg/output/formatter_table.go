package output

import (
	"fmt"
	"os"
	"strings"
	"vm-scanner/pkg/cluster"
	"vm-scanner/pkg/hardware"
	"vm-scanner/pkg/vm"

	"github.com/jedib0t/go-pretty/v6/table"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// TableFormatter formats data as human-readable tables to stdout.
type TableFormatter struct{}

// NewTableFormatter creates a new table formatter instance.
func NewTableFormatter() *TableFormatter {
	return &TableFormatter{}
}

// Format routes to the appropriate table formatter based on data type.
func (tf *TableFormatter) Format(data interface{}) error {
	switch v := data.(type) {
	case *cluster.ClusterNodeInfo:
		return tf.formatNodeInfoTable(v)
	case []cluster.ClusterNodeInfo:
		return tf.formatNodeListTable(v)
	case *cluster.ClusterSummary:
		return tf.formatClusterSummaryTable(v)
	case *ComprehensiveReport:
		return tf.formatComprehensiveReportTable(v)
	case *ComprehensiveData:
		return fmt.Errorf("ComprehensiveData table formatting not implemented")
	case hardware.StorageClassInfo:
		return tf.formatHardwareStorageClassTable(&v)
	case *hardware.StorageClassInfo:
		return tf.formatHardwareStorageClassTable(v)
	case []hardware.StorageClassInfo:
		return tf.formatHardwareStorageClassListTable(v)
	case *cluster.KubeVirtVersion:
		return tf.formatKubevirtVersionTable(v)
	case *unstructured.UnstructuredList:
		return tf.formatVMIListTable(v)
	case [][]vm.StorageInfo:
		return tf.formatStorageVolumeListTable(v)
	case map[string]vm.VMConsolidatedReport:
		return tf.formatVMRuntimeInfoMapTable(v)
	default:
		return fmt.Errorf("table format not supported for type %T", v)
	}
}

// formatNodeInfoTable formats a single NodeHardwareInfo as a categorized table
func (tf *TableFormatter) formatNodeInfoTable(nodeInfo *cluster.ClusterNodeInfo) error {
	formatter := NewNodePrettyTableFormatter(*nodeInfo)
	formatter.FormatSystemInfoTable()
	formatter.FormatResourceCapacityTable()
	formatter.FormatFilesystemTable()
	formatter.FormatNetworkTable()
	return nil
}

// formatNodeListTable formats multiple nodes using the same categorized format as single nodes
func (tf *TableFormatter) formatNodeListTable(nodeList []cluster.ClusterNodeInfo) error {
	if len(nodeList) == 0 {
		fmt.Println("No node hardware information available.")
		return nil
	}

	fmt.Printf("\n=== CLUSTER NODE HARDWARE INFORMATION ===\n")
	fmt.Printf("Total Nodes: %d\n\n", len(nodeList))

	for i, node := range nodeList {
		if i > 0 {
			fmt.Println("\n" + strings.Repeat("=", 80) + "\n")
		}

		fmt.Printf("Node: %s\n\n", node.ClusterNodeName)

		formatter := NewNodeTableFormatter(node)
		formatter.FormatSystemInfo()
		formatter.FormatResourceCapacity()
		formatter.FormatFilesystem()
		formatter.FormatNetwork()
	}

	return nil
}

// formatStorageClassTable formats a single StorageClassInfo as a detailed table (legacy - moved to hardware)
// This function is deprecated - use formatHardwareStorageClassTable instead
func (tf *TableFormatter) formatStorageClassTable(storageClass *hardware.StorageClassInfo) error {
	fmt.Println("\n=== BASIC INFORMATION ===")
	basicTable := table.NewWriter()
	basicTable.SetOutputMirror(os.Stdout)
	basicTable.SetStyle(table.StyleColoredBright)
	basicTable.AppendHeader(table.Row{"Property", "Value"})

	basicTable.AppendRow(table.Row{"Name", storageClass.Name})
	basicTable.AppendRow(table.Row{"Provisioner", storageClass.Provisioner})
	basicTable.AppendRow(table.Row{"Reclaim Policy", storageClass.ReclaimPolicy})
	basicTable.AppendRow(table.Row{"Volume Binding Mode", storageClass.VolumeBindingMode})

	defaultStatus := "No"
	if storageClass.IsDefault {
		defaultStatus = "Yes"
	}
	basicTable.AppendRow(table.Row{"Default Storage Class", defaultStatus})

	basicTable.Render()
	fmt.Println()

	fmt.Println("=== PARAMETERS ===")
	if len(storageClass.Parameters) > 0 {
		paramsTable := table.NewWriter()
		paramsTable.SetOutputMirror(os.Stdout)
		paramsTable.SetStyle(table.StyleColoredBright)
		paramsTable.AppendHeader(table.Row{"Parameter", "Value"})

		for key, value := range storageClass.Parameters {
			paramsTable.AppendRow(table.Row{key, value})
		}
		paramsTable.Render()
	} else {
		fmt.Println("No parameters configured")
	}

	return nil
}

// formatStorageClassListTable formats multiple StorageClassInfo using the same categorized format as single storage classes (legacy - moved to hardware)
// This function is deprecated - use formatHardwareStorageClassListTable instead
func (tf *TableFormatter) formatStorageClassListTable(storageClasses []hardware.StorageClassInfo) error {
	if len(storageClasses) == 0 {
		fmt.Println("No storage class information available.")
		return nil
	}

	fmt.Printf("\n=== CLUSTER STORAGE CLASS INFORMATION ===\n")
	fmt.Printf("Total Storage Classes: %d\n\n", len(storageClasses))

	for i, sc := range storageClasses {
		if i > 0 {
			fmt.Println("\n" + strings.Repeat("=", 80) + "\n")
		}

		fmt.Printf("Storage Class: %s\n\n", sc.Name)

		fmt.Println("  BASIC INFORMATION:")
		fmt.Printf("    Name: %s\n", sc.Name)
		fmt.Printf("    Provisioner: %s\n", sc.Provisioner)
		fmt.Printf("    Reclaim Policy: %s\n", sc.ReclaimPolicy)
		fmt.Printf("    Volume Binding Mode: %s\n", sc.VolumeBindingMode)

		defaultStatus := "No"
		if sc.IsDefault {
			defaultStatus = "Yes"
		}
		fmt.Printf("    Default Storage Class: %s\n", defaultStatus)
		fmt.Println()

		fmt.Println("  PARAMETERS:")
		if len(sc.Parameters) > 0 {
			for key, value := range sc.Parameters {
				fmt.Printf("    %s: %s\n", key, value)
			}
		} else {
			fmt.Printf("    No parameters configured\n")
		}
	}

	return nil
}

// formatHardwareStorageClassTable formats a single hardware.StorageClassInfo as a detailed table
func (tf *TableFormatter) formatHardwareStorageClassTable(storageClass *hardware.StorageClassInfo) error {
	fmt.Println("\n=== BASIC INFORMATION ===")
	basicTable := table.NewWriter()
	basicTable.SetOutputMirror(os.Stdout)
	basicTable.SetStyle(table.StyleColoredBright)
	basicTable.AppendHeader(table.Row{"Property", "Value"})

	basicTable.AppendRow(table.Row{"Name", storageClass.Name})
	basicTable.AppendRow(table.Row{"Provisioner", storageClass.Provisioner})
	basicTable.AppendRow(table.Row{"Reclaim Policy", storageClass.ReclaimPolicy})
	basicTable.AppendRow(table.Row{"Volume Binding Mode", storageClass.VolumeBindingMode})

	defaultStatus := "No"
	if storageClass.IsDefault {
		defaultStatus = "Yes"
	}
	basicTable.AppendRow(table.Row{"Default Storage Class", defaultStatus})

	expansionStatus := "No"
	if storageClass.AllowVolumeExpansion {
		expansionStatus = "Yes"
	}
	basicTable.AppendRow(table.Row{"Allow Volume Expansion", expansionStatus})
	basicTable.AppendRow(table.Row{"Created At", storageClass.CreatedAt.Format("2006-01-02 15:04:05")})

	basicTable.Render()
	fmt.Println()

	fmt.Println("=== PARAMETERS ===")
	if len(storageClass.Parameters) > 0 {
		paramsTable := table.NewWriter()
		paramsTable.SetOutputMirror(os.Stdout)
		paramsTable.SetStyle(table.StyleColoredBright)
		paramsTable.AppendHeader(table.Row{"Parameter", "Value"})

		for key, value := range storageClass.Parameters {
			paramsTable.AppendRow(table.Row{key, value})
		}
		paramsTable.Render()
	} else {
		fmt.Println("No parameters configured")
	}
	fmt.Println()

	fmt.Println("=== ADVANCED CONFIGURATION ===")
	advTable := table.NewWriter()
	advTable.SetOutputMirror(os.Stdout)
	advTable.SetStyle(table.StyleColoredBright)
	advTable.AppendHeader(table.Row{"Setting", "Value"})

	mountOptions := "None"
	if len(storageClass.MountOptions) > 0 {
		mountOptions = strings.Join(storageClass.MountOptions, ", ")
	}
	advTable.AppendRow(table.Row{"Mount Options", mountOptions})

	allowedTopologies := "None"
	if len(storageClass.AllowedTopologies) > 0 {
		allowedTopologies = strings.Join(storageClass.AllowedTopologies, ", ")
	}
	advTable.AppendRow(table.Row{"Allowed Topologies", allowedTopologies})

	unsafeVolumes := "None"
	if len(storageClass.AllowedUnsafeToEvictVolumes) > 0 {
		unsafeVolumes = strings.Join(storageClass.AllowedUnsafeToEvictVolumes, ", ")
	}
	advTable.AppendRow(table.Row{"Allowed Unsafe Evict Volumes", unsafeVolumes})

	advTable.Render()

	return nil
}

// formatHardwareStorageClassListTable formats multiple hardware.StorageClassInfo using the same categorized format as single storage classes
func (tf *TableFormatter) formatHardwareStorageClassListTable(storageClasses []hardware.StorageClassInfo) error {
	if len(storageClasses) == 0 {
		fmt.Println("No storage class information available.")
		return nil
	}

	fmt.Printf("\n=== CLUSTER STORAGE CLASS INFORMATION ===\n")
	fmt.Printf("Total Storage Classes: %d\n\n", len(storageClasses))

	for i, sc := range storageClasses {
		if i > 0 {
			fmt.Println("\n" + strings.Repeat("=", 80) + "\n")
		}

		fmt.Printf("Storage Class: %s\n\n", sc.Name)

		fmt.Println("  BASIC INFORMATION:")
		fmt.Printf("    Name: %s\n", sc.Name)
		fmt.Printf("    Provisioner: %s\n", sc.Provisioner)
		fmt.Printf("    Reclaim Policy: %s\n", sc.ReclaimPolicy)
		fmt.Printf("    Volume Binding Mode: %s\n", sc.VolumeBindingMode)

		defaultStatus := "No"
		if sc.IsDefault {
			defaultStatus = "Yes"
		}
		fmt.Printf("    Default Storage Class: %s\n", defaultStatus)

		expansionStatus := "No"
		if sc.AllowVolumeExpansion {
			expansionStatus = "Yes"
		}
		fmt.Printf("    Allow Volume Expansion: %s\n", expansionStatus)
		fmt.Printf("    Created At: %s\n", sc.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()

		fmt.Println("  PARAMETERS:")
		if len(sc.Parameters) > 0 {
			for key, value := range sc.Parameters {
				fmt.Printf("    %s: %s\n", key, value)
			}
		} else {
			fmt.Printf("    No parameters configured\n")
		}
		fmt.Println()

		fmt.Println("  ADVANCED CONFIGURATION:")

		mountOptions := "None"
		if len(sc.MountOptions) > 0 {
			mountOptions = strings.Join(sc.MountOptions, ", ")
		}
		fmt.Printf("    Mount Options: %s\n", mountOptions)

		allowedTopologies := "None"
		if len(sc.AllowedTopologies) > 0 {
			allowedTopologies = strings.Join(sc.AllowedTopologies, ", ")
		}
		fmt.Printf("    Allowed Topologies: %s\n", allowedTopologies)

		unsafeVolumes := "None"
		if len(sc.AllowedUnsafeToEvictVolumes) > 0 {
			unsafeVolumes = strings.Join(sc.AllowedUnsafeToEvictVolumes, ", ")
		}
		fmt.Printf("    Allowed Unsafe Evict Volumes: %s\n", unsafeVolumes)
	}

	return nil
}

// formatClusterSummaryTable formats cluster-wide information
func (tf *TableFormatter) formatClusterSummaryTable(clusterInfo *cluster.ClusterSummary) error {
	return fmt.Errorf("formatClusterSummaryTable not implemented")
}

// formatComprehensiveReportTable formats the complete report in table format
func (tf *TableFormatter) formatComprehensiveReportTable(report *ComprehensiveReport) error {
	return fmt.Errorf("formatComprehensiveReportTable not implemented")
}

// formatKubevirtVersionTable formats KubeVirt version information as a table
func (tf *TableFormatter) formatKubevirtVersionTable(version *cluster.KubeVirtVersion) error {
	fmt.Println("\n=== KUBEVIRT VERSION ===")
	versionTable := table.NewWriter()
	versionTable.SetOutputMirror(os.Stdout)
	versionTable.SetStyle(table.StyleColoredBright)
	versionTable.AppendHeader(table.Row{"Property", "Value"})

	versionTable.AppendRow(table.Row{"Version", version.Version})
	versionTable.AppendRow(table.Row{"Status", version.Deployed})

	versionTable.Render()
	fmt.Println()

	return nil
}

// formatVMIListTable formats Virtual Machine Instance list as a table
func (tf *TableFormatter) formatVMIListTable(vmiList *unstructured.UnstructuredList) error {
	if vmiList == nil || len(vmiList.Items) == 0 {
		fmt.Println("\n=== VIRTUAL MACHINE INSTANCES ===")
		fmt.Println("No VM instances found.")
		return nil
	}

	fmt.Printf("\n=== VIRTUAL MACHINE INSTANCES ===\n")
	fmt.Printf("Total VM Instances: %d\n\n", len(vmiList.Items))

	vmiTable := table.NewWriter()
	vmiTable.SetOutputMirror(os.Stdout)
	vmiTable.SetStyle(table.StyleColoredBright)
	vmiTable.AppendHeader(table.Row{"Name", "Namespace", "Phase", "Node"})

	for _, vmi := range vmiList.Items {
		name := vmi.GetName()
		namespace := vmi.GetNamespace()

		phase := "Unknown"
		if status, found, err := unstructured.NestedString(vmi.Object, "status", "phase"); found && err == nil {
			phase = status
		}

		nodeName := "N/A"
		if node, found, err := unstructured.NestedString(vmi.Object, "status", "nodeName"); found && err == nil {
			nodeName = node
		}

		vmiTable.AppendRow(table.Row{name, namespace, phase, nodeName})
	}

	vmiTable.Render()
	fmt.Println()

	return nil
}

// formatVMRuntimeInfoMapTable formats a map of VMRuntimeInfo as a table
func (tf *TableFormatter) formatVMRuntimeInfoMapTable(vmiMap map[string]vm.VMConsolidatedReport) error {
	if len(vmiMap) == 0 {
		fmt.Println("\n=== VM RUNTIME INVENTORY ===")
		fmt.Println("No running VMs found.")
		return nil
	}

	fmt.Printf("\n=== VM RUNTIME INVENTORY ===\n")
	fmt.Printf("Total Running VMs: %d\n\n", len(vmiMap))

	vmiTable := table.NewWriter()
	vmiTable.SetOutputMirror(os.Stdout)
	vmiTable.SetStyle(table.StyleColoredBright)
	vmiTable.AppendHeader(table.Row{"Name", "Namespace", "Power State", "Node", "OS", "Memory Used"})

	for _, vmReport := range vmiMap {
		osName := "N/A"
		runningOnNode := "N/A"
		memoryUsed := "N/A"
		powerState := "Stopped"

		if vmReport.OSName != "" {
			osName = vmReport.OSName
		}

		if vmReport.Runtime != nil {
			powerState = vmReport.Runtime.PowerState
			if vmReport.Runtime.GuestMetadata != nil {
				runningOnNode = vmReport.Runtime.GuestMetadata.RunningOnNode
			}
		}

		if vmReport.MemoryInfo.MemoryUsedByVMI > 0 {
			memoryUsed = fmt.Sprintf("%.1f MiB", vmReport.MemoryInfo.MemoryUsedByVMI)
		}

		vmiTable.AppendRow(table.Row{
			vmReport.Name,
			vmReport.Namespace,
			powerState,
			runningOnNode,
			osName,
			memoryUsed,
		})
	}

	vmiTable.Render()
	fmt.Println()

	return nil
}

// formatStorageVolumeListTable formats storage volumes for all VMs as a table
func (tf *TableFormatter) formatStorageVolumeListTable(allVolumes [][]vm.StorageInfo) error {
	if len(allVolumes) == 0 {
		fmt.Println("\n=== STORAGE VOLUMES ===")
		fmt.Println("No storage volumes found.")
		return nil
	}

	fmt.Println("\n=== STORAGE VOLUMES ===")

	totalVolumes := 0
	for _, volumes := range allVolumes {
		totalVolumes += len(volumes)
	}

	fmt.Printf("Total VMs: %d\n", len(allVolumes))
	fmt.Printf("Total Volumes: %d\n\n", totalVolumes)

	volumeTable := table.NewWriter()
	volumeTable.SetOutputMirror(os.Stdout)
	volumeTable.SetStyle(table.StyleColoredBright)
	volumeTable.AppendHeader(table.Row{"Volume Name", "Size", "Storage Class", "Type", "Total Storage"})

	for _, volumes := range allVolumes {
		for _, vol := range volumes {
			volumeTable.AppendRow(table.Row{
				vol.VolumeName,
				vol.SizeHuman,
				vol.StorageClass,
				vol.VolumeType,
				vol.TotalStorageHuman,
			})
		}
	}

	volumeTable.Render()
	fmt.Println()

	return nil
}
