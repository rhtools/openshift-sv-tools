package output

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"vm-scanner/pkg/cluster"
	"vm-scanner/pkg/hardware"
	"vm-scanner/pkg/vm"
)

// MultiCSVFormatter handles multi-file CSV output format (one CSV per category)
type MultiCSVFormatter struct {
	outputFile string
}

// NewMultiCSVFormatter creates a new MultiCSV formatter
func NewMultiCSVFormatter(outputFile string) *MultiCSVFormatter {
	return &MultiCSVFormatter{
		outputFile: outputFile,
	}
}

// Format routes to appropriate multi-CSV formatter based on data type
func (f *MultiCSVFormatter) Format(data interface{}) error {
	switch v := data.(type) {
	case *ComprehensiveReport:
		return f.formatComprehensiveReportMultiCSV(v)
	case []cluster.ClusterNodeInfo:
		return f.formatNodeListMultiCSV(v)
	case map[string]vm.VMConsolidatedReport:
		return f.formatVMRuntimeInfoMapCSV(v)
	default:
		return fmt.Errorf("Multi-CSV format not supported for type %T", v)
	}
}

// formatComprehensiveReportMultiCSV formats a comprehensive report as multiple CSV files
func (f *MultiCSVFormatter) formatComprehensiveReportMultiCSV(report *ComprehensiveReport) error {
	baseDir, baseName := f.basePathAndName("report")

	// Write Summary CSV (merged with cluster info)
	summaryFile := filepath.Join(baseDir, fmt.Sprintf("%s_summary.csv", baseName))
	if err := f.writeComprehensiveReportSummaryCSV(summaryFile, report); err != nil {
		return fmt.Errorf("failed to write summary CSV: %w", err)
	}
	fmt.Printf("Summary CSV saved to: %s\n", summaryFile)

	// Write Node Hardware CSV with FULL details
	if len(report.Nodes) > 0 {
		nodeHardwareFile := filepath.Join(baseDir, fmt.Sprintf("%s_node-hardware.csv", baseName))
		if err := f.writeNodeHardwareCSV(nodeHardwareFile, report.Nodes); err != nil {
			return fmt.Errorf("failed to write node hardware CSV: %w", err)
		}
		fmt.Printf("Node Hardware CSV saved to: %s\n", nodeHardwareFile)
	}

	// Write VMs CSV
	vmsFile := filepath.Join(baseDir, fmt.Sprintf("%s_vms.csv", baseName))
	if err := f.writeVMsCSV(vmsFile, report.VMs); err != nil {
		return fmt.Errorf("failed to write VMs CSV: %w", err)
	}
	fmt.Printf("VMs CSV saved to: %s\n", vmsFile)

	// Write Storage CSV
	storageFile := filepath.Join(baseDir, fmt.Sprintf("%s_storage.csv", baseName))
	if err := f.writeStorageCSV(storageFile, report.Storage); err != nil {
		return fmt.Errorf("failed to write storage CSV: %w", err)
	}
	fmt.Printf("Storage CSV saved to: %s\n", storageFile)

	// Write VM Disks CSV
	vmDisksFile := filepath.Join(baseDir, fmt.Sprintf("%s_vm-disks.csv", baseName))
	if err := f.writeVMDisksCSV(vmDisksFile, report.VMs); err != nil {
		return fmt.Errorf("failed to write VM disks CSV: %w", err)
	}
	fmt.Printf("VM Disks CSV saved to: %s\n", vmDisksFile)

	// Write Network Interfaces CSV
	netIfFile := filepath.Join(baseDir, fmt.Sprintf("%s_network-interfaces.csv", baseName))
	if err := f.writeNetworkInterfacesCSV(netIfFile, report.VMs); err != nil {
		return fmt.Errorf("failed to write network interfaces CSV: %w", err)
	}
	fmt.Printf("Network Interfaces CSV saved to: %s\n", netIfFile)

	// Write Capacity Planning CSV
	capFile := filepath.Join(baseDir, fmt.Sprintf("%s_capacity-planning.csv", baseName))
	if err := f.writeCapacityPlanningCSV(capFile, report.Nodes, report.VMs); err != nil {
		return fmt.Errorf("failed to write capacity planning CSV: %w", err)
	}
	fmt.Printf("Capacity Planning CSV saved to: %s\n", capFile)

	// Write VM Assessment CSV
	assessFile := filepath.Join(baseDir, fmt.Sprintf("%s_vm-assessment.csv", baseName))
	if err := f.writeVMAssessmentCSV(assessFile, report.VMs); err != nil {
		return fmt.Errorf("failed to write VM assessment CSV: %w", err)
	}
	fmt.Printf("VM Assessment CSV saved to: %s\n", assessFile)

	// Write PVC Inventory CSV
	pvcFile := filepath.Join(baseDir, fmt.Sprintf("%s_pvc-inventory.csv", baseName))
	if err := f.writePVCInventoryCSV(pvcFile, report.PVCs); err != nil {
		return fmt.Errorf("failed to write PVC inventory CSV: %w", err)
	}
	fmt.Printf("PVC Inventory CSV saved to: %s\n", pvcFile)

	// Write NAD Inventory CSV
	nadFile := filepath.Join(baseDir, fmt.Sprintf("%s_nad-inventory.csv", baseName))
	if err := f.writeNADInventoryCSV(nadFile, report.NADs); err != nil {
		return fmt.Errorf("failed to write NAD inventory CSV: %w", err)
	}
	fmt.Printf("NAD Inventory CSV saved to: %s\n", nadFile)

	// Write DataVolumes CSV
	dvFile := filepath.Join(baseDir, fmt.Sprintf("%s_datavolumes.csv", baseName))
	if err := f.writeDataVolumesCSV(dvFile, report.DataVolumes); err != nil {
		return fmt.Errorf("failed to write DataVolumes CSV: %w", err)
	}
	fmt.Printf("DataVolumes CSV saved to: %s\n", dvFile)

	migrationFile := filepath.Join(baseDir, fmt.Sprintf("%s_migration-readiness.csv", baseName))
	if err := f.writeMigrationReadinessCSV(migrationFile, report.VMs, report.PVCs); err != nil {
		return fmt.Errorf("failed to write migration readiness CSV: %w", err)
	}
	fmt.Printf("Migration Readiness CSV saved to: %s\n", migrationFile)

	storageAnalysisFile := filepath.Join(baseDir, fmt.Sprintf("%s_storage-analysis.csv", baseName))
	if err := f.writeStorageAnalysisCSV(storageAnalysisFile, report.PVCs, report.VMs); err != nil {
		return fmt.Errorf("failed to write storage analysis CSV: %w", err)
	}
	fmt.Printf("Storage Analysis CSV saved to: %s\n", storageAnalysisFile)

	operatorFile := filepath.Join(baseDir, fmt.Sprintf("%s_operator-status.csv", baseName))
	if err := f.writeOperatorStatusCSV(operatorFile, report.Cluster.Operators); err != nil {
		return fmt.Errorf("failed to write operator status CSV: %w", err)
	}
	fmt.Printf("Operator Status CSV saved to: %s\n", operatorFile)

	return nil
}

// formatNodeListMultiCSV formats node list as multiple CSV files
func (f *MultiCSVFormatter) formatNodeListMultiCSV(nodes []cluster.ClusterNodeInfo) error {
	baseDir, baseName := f.basePathAndName("nodes")

	// Write Node Hardware CSV
	hardwareFile := filepath.Join(baseDir, fmt.Sprintf("%s_hardware.csv", baseName))
	if err := f.writeNodeHardwareCSV(hardwareFile, nodes); err != nil {
		return fmt.Errorf("failed to write node hardware CSV: %w", err)
	}
	fmt.Printf("Node Hardware CSV saved to: %s\n", hardwareFile)

	// Write Network Configuration CSV
	networkFile := filepath.Join(baseDir, fmt.Sprintf("%s_network.csv", baseName))
	if err := f.writeNodeNetworkCSV(networkFile, nodes); err != nil {
		return fmt.Errorf("failed to write node network CSV: %w", err)
	}
	fmt.Printf("Node Network CSV saved to: %s\n", networkFile)

	return nil
}

// formatVMRuntimeInfoMapCSV formats VM runtime inventory map as CSV (single file)
func (f *MultiCSVFormatter) formatVMRuntimeInfoMapCSV(vmiMap map[string]vm.VMConsolidatedReport) error {
	baseDir, baseName := f.basePathAndName("vm-runtime")
	filename := filepath.Join(baseDir, fmt.Sprintf("%s.csv", baseName))

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
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

	fmt.Printf("VM Runtime CSV saved to: %s\n", filename)
	return nil
}

// basePathAndName returns base directory and base name for output files
func (f *MultiCSVFormatter) basePathAndName(defaultName string) (string, string) {
	baseDir := filepath.Dir(f.outputFile)
	if baseDir == "" || baseDir == "." {
		baseDir = "."
	}
	baseName := strings.TrimSuffix(filepath.Base(f.outputFile), filepath.Ext(f.outputFile))
	if baseName == "" {
		baseName = defaultName
	}
	return baseDir, baseName
}

// writeComprehensiveReportSummaryCSV writes summary CSV with cluster info
func (f *MultiCSVFormatter) writeComprehensiveReportSummaryCSV(filename string, report *ComprehensiveReport) error {
	if report == nil || report.Summary == nil {
		return nil
	}
	summary := report.Summary
	clusterInfo := report.Cluster
	nodes := report.Nodes
	vms := report.VMs

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write cluster-level information first (no header row)
	if clusterInfo != nil {
		writer.Write([]string{"Cluster Name", clusterInfo.ClusterName})
		writer.Write([]string{"Cluster ID", clusterInfo.ClusterID})
		writer.Write([]string{"OpenShift Version", clusterInfo.ClusterVersion})
		writer.Write([]string{"Kubernetes Version", clusterInfo.KubernetesVersion})
		if clusterInfo.KubeVirtVersion != nil {
			writer.Write([]string{"KubeVirt Version", clusterInfo.KubeVirtVersion.Version})
			writer.Write([]string{"KubeVirt Deployed", clusterInfo.KubeVirtVersion.Deployed})
		}
		writer.Write([]string{"Schedulable Control Plane Count", fmt.Sprintf("%d", clusterInfo.SchedulableControlPlaneCount)})
		writer.Write([]string{"Control Plane Schedulable", fmt.Sprintf("%t", clusterInfo.HasSchedulableControlPlane)})
		writer.Write([]string{"Worker Nodes Count", fmt.Sprintf("%d", clusterInfo.WorkerNodesCount)})
		writer.Write([]string{"", ""})
	}

	// Write summary statistics
	if summary.ClusterSummary != nil {
		writer.Write([]string{"Total VMs", fmt.Sprintf("%d", summary.ClusterSummary.TotalVMs)})
		writer.Write([]string{"Running VMs", fmt.Sprintf("%d", summary.ClusterSummary.RunningVMs)})
		writer.Write([]string{"Stopped VMs", fmt.Sprintf("%d", summary.ClusterSummary.StoppedVMs)})
		if summary.ClusterSummary.Resources != nil {
			writer.Write([]string{"Total CPU", fmt.Sprintf("%d", summary.ClusterSummary.Resources.TotalCPU)})
			writer.Write([]string{"Total Memory (GiB)", fmt.Sprintf("%.1f", summary.ClusterSummary.Resources.TotalMemory)})
			writer.Write([]string{"Used Memory (GiB)", fmt.Sprintf("%.1f", summary.ClusterSummary.Resources.UsedMemory)})
			writer.Write([]string{"Total Local Storage (GiB)", fmt.Sprintf("%.1f", float64(summary.ClusterSummary.Resources.TotalLocalStorage))})
			writer.Write([]string{"Total Application Requested Storage (GiB)", fmt.Sprintf("%.1f", float64(summary.ClusterSummary.Resources.TotalApplicationRequestedStorage))})
			writer.Write([]string{"Total Application Used Storage (GiB)", fmt.Sprintf("%.1f", float64(summary.ClusterSummary.Resources.TotalApplicationUsedStorage))})
		}
		writer.Write([]string{"Total Namespaces", fmt.Sprintf("%d", summary.ClusterSummary.TotalNamespaces)})
		writer.Write([]string{"User Created Namespaces", fmt.Sprintf("%d", summary.ClusterSummary.UserNamespaces)})
		writer.Write([]string{"Nodes", fmt.Sprintf("%d", len(nodes))})
	}
	writer.Write([]string{"Storage Classes", fmt.Sprintf("%d", summary.StorageClasses)})

	// Guest Agent Coverage
	agentCount := 0
	for _, vm := range vms {
		if vm.Runtime != nil && vm.Runtime.GuestMetadata != nil && vm.Runtime.GuestMetadata.GuestAgentVersion != "" {
			agentCount++
		}
	}
	runningVMs := 0
	if summary.ClusterSummary != nil {
		runningVMs = summary.ClusterSummary.RunningVMs
	}
	agentCoverage := 0.0
	if runningVMs > 0 {
		agentCoverage = (float64(agentCount) / float64(runningVMs)) * 100
	}
	writer.Write([]string{"Guest Agent Coverage (%)", fmt.Sprintf("%.1f%%", agentCoverage)})

	// Phase 3 inventory counts
	writer.Write([]string{"", ""})
	writer.Write([]string{"PVC Count", fmt.Sprintf("%d", len(report.PVCs))})
	var totalPVCStorageGiB float64
	for _, pvc := range report.PVCs {
		totalPVCStorageGiB += float64(pvc.Capacity) / (1024 * 1024 * 1024)
	}
	writer.Write([]string{"Total PVC Storage (GiB)", fmt.Sprintf("%.1f", totalPVCStorageGiB)})
	writer.Write([]string{"NAD Count", fmt.Sprintf("%d", len(report.NADs))})
	writer.Write([]string{"DataVolume Count", fmt.Sprintf("%d", len(report.DataVolumes))})
	dvInProgress := 0
	for _, dv := range report.DataVolumes {
		if dv.Phase != "Succeeded" {
			dvInProgress++
		}
	}
	writer.Write([]string{"DataVolumes In Progress", fmt.Sprintf("%d", dvInProgress)})

	// Phase 4 operator and migration counts
	writer.Write([]string{"", ""})
	operatorCount := 0
	healthyOperators := 0
	if report.Cluster != nil {
		operatorCount = len(report.Cluster.Operators)
		for _, op := range report.Cluster.Operators {
			if op.Health == "Healthy" {
				healthyOperators++
			}
		}
	}
	writer.Write([]string{"Operator Count", fmt.Sprintf("%d", operatorCount)})
	writer.Write([]string{"Healthy Operators", fmt.Sprintf("%d", healthyOperators)})

	assessments := assessMigrationReadiness(report.VMs, report.PVCs)
	migrationReady := 0
	migrationBlocked := 0
	for _, a := range assessments {
		if len(a.Blockers) == 0 {
			migrationReady++
		} else {
			migrationBlocked++
		}
	}
	writer.Write([]string{"Migration Ready VMs", fmt.Sprintf("%d", migrationReady)})
	writer.Write([]string{"Migration Blocked VMs", fmt.Sprintf("%d", migrationBlocked)})

	return nil
}

// writeVMsCSV writes VMs to a CSV file
func (f *MultiCSVFormatter) writeVMsCSV(filename string, vms []vm.VMDetails) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write(VMCSVHeaders)

	for _, vmDetail := range vms {
		builder := NewVMCSVRowBuilder(vmDetail)
		row := builder.BuildRow()
		writer.Write(row)
	}

	return nil
}

// writeStorageCSV writes storage classes to a CSV file
func (f *MultiCSVFormatter) writeStorageCSV(filename string, storageClasses []hardware.StorageClassInfo) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Name", "Provisioner", "Reclaim Policy", "Volume Binding Mode", "Default", "Allow Expansion"})

	for _, sc := range storageClasses {
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

// writeVMDisksCSV writes VM disks to a CSV file
func (f *MultiCSVFormatter) writeVMDisksCSV(filename string, vms []vm.VMDetails) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{
		"VM Name",
		"VM Namespace",
		"Volume Name",
		"Volume Type",
		"PVC Size (GiB)",
		"Storage Class",
		"Guest Mount Point",
		"Guest FS Type",
		"Guest Total (GiB)",
		"Guest Used (GiB)",
		"Guest Free (GiB)",
		"Guest Usage (%)",
	})

	for _, vmDetail := range vms {
		for volName, disk := range vmDetail.Disks {
			pvcSizeGiB := float64(disk.SizeBytes) / (1024 * 1024 * 1024)
			guestTotalGiB := float64(disk.TotalStorage) / (1024 * 1024 * 1024)
			guestUsedGiB := float64(disk.TotalStorageInUse) / (1024 * 1024 * 1024)
			guestFreeGiB := guestTotalGiB - guestUsedGiB

			writer.Write([]string{
				vmDetail.Name,
				vmDetail.Namespace,
				volName,
				disk.VolumeType,
				fmt.Sprintf("%.2f", pvcSizeGiB),
				disk.StorageClass,
				"",
				"",
				fmt.Sprintf("%.2f", guestTotalGiB),
				fmt.Sprintf("%.2f", guestUsedGiB),
				fmt.Sprintf("%.2f", guestFreeGiB),
				fmt.Sprintf("%.2f", disk.TotalStorageInUsePercentage),
			})
		}

		if vmDetail.Runtime != nil && vmDetail.Runtime.GuestMetadata != nil {
			for _, guestDisk := range vmDetail.Runtime.GuestMetadata.DiskInfo {
				guestTotalGiB := float64(guestDisk.TotalBytes) / (1024 * 1024 * 1024)
				guestUsedGiB := float64(guestDisk.UsedBytes) / (1024 * 1024 * 1024)
				guestFreeGiB := guestTotalGiB - guestUsedGiB
				guestUsage := 0.0
				if guestDisk.TotalBytes > 0 {
					guestUsage = (float64(guestDisk.UsedBytes) / float64(guestDisk.TotalBytes)) * 100
				}

				writer.Write([]string{
					vmDetail.Name,
					vmDetail.Namespace,
					guestDisk.DiskName,
					"guest-agent",
					"",
					"",
					guestDisk.MountPoint,
					guestDisk.FsType,
					fmt.Sprintf("%.2f", guestTotalGiB),
					fmt.Sprintf("%.2f", guestUsedGiB),
					fmt.Sprintf("%.2f", guestFreeGiB),
					fmt.Sprintf("%.2f", guestUsage),
				})
			}
		}
	}

	return nil
}

// writeClusterCSV writes cluster summary to a CSV file
func (f *MultiCSVFormatter) writeClusterCSV(filename string, clusterInfo *cluster.ClusterSummary) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
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

// writeNodeHardwareCSV writes node hardware details to a CSV file
func (f *MultiCSVFormatter) writeNodeHardwareCSV(filename string, nodes []cluster.ClusterNodeInfo) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{
		"Node Name",
		"Node Roles",
		"OS Version",
		"Kernel Version",
		"Schedulable",
		"Storage Capacity",
		"Pod Limits",
		// CPU
		"CPU Cores",
		"CPU Model",
		// Filesystem
		"Filesystem Available (GB)",
		"Filesystem Capacity (GB)",
		"Filesystem Used (GB)",
		"Filesystem Usage (%)",
		// Memory
		"Memory Capacity (GiB)",
		"Memory Used (GiB)",
		"Memory Used (%)",
		// Network
		"Network Bridge ID",
		"Network Interface ID",
		"Network IP Addresses",
		"Network MAC Address",
		"Network Mode",
		"Network Interfaces",
		"Network Interface Speeds",
		"Network Next Hops",
		"Network Node Port Enable",
		"Network VLAN ID",
		"Network Host CIDRs",
		"Network Pod Subnets",
	})

	for _, node := range nodes {
		podSubnets := "None"
		if len(node.Network.PodNetworkSubnet) > 0 {
			var subnetParts []string
			for key, values := range node.Network.PodNetworkSubnet {
				subnetParts = append(subnetParts, fmt.Sprintf("%s: %s", key, strings.Join(values, ", ")))
			}
			podSubnets = strings.Join(subnetParts, "; ")
		}

		writer.Write([]string{
			node.ClusterNodeName,
			strings.Join(node.NodeRoles, "; "),
			node.CoreOSVersion,
			node.NodeKernelVersion,
			node.NodeSchedulable,
			fmt.Sprintf("%.1f", node.StorageCapacity),
			fmt.Sprintf("%d", node.NodePodLimits),
			fmt.Sprintf("%d", node.CPU.CPUCores),
			node.CPU.CPUModel,
			fmt.Sprintf("%.1f", node.Filesystem.FilesystemAvailable),
			fmt.Sprintf("%.1f", node.Filesystem.FilesystemCapacity),
			fmt.Sprintf("%.1f", node.Filesystem.FilesystemUsed),
			fmt.Sprintf("%.1f", node.Filesystem.FilesystemUsagePercent),
			fmt.Sprintf("%.1f", node.Memory.MemoryCapacityGiB),
			fmt.Sprintf("%.1f", node.Memory.MemoryUsedGiB),
			fmt.Sprintf("%.1f", func() float64 {
				if node.Memory.MemoryCapacityGiB > 0 {
					return (node.Memory.MemoryUsedGiB / node.Memory.MemoryCapacityGiB) * 100
				}
				return 0.0
			}()),
			node.Network.BridgeID,
			node.Network.InterfaceID,
			strings.Join(node.Network.IPAddresses, "; "),
		node.Network.MACAddress,
		node.Network.Mode,
		FormatNICNames(node.Network.NetworkInterfaces),
		FormatNICSpeedList(node.Network.NetworkInterfaces),
		strings.Join(node.Network.NextHops, "; "),
		fmt.Sprintf("%t", node.Network.NodePortEnable),
		fmt.Sprintf("%d", node.Network.VLANID),
		strings.Join(node.Network.HostCIDRs, "; "),
			podSubnets,
		})
	}

	return nil
}

// writeNodeNetworkCSV writes node network configuration to a CSV file
func (f *MultiCSVFormatter) writeNodeNetworkCSV(filename string, nodes []cluster.ClusterNodeInfo) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Node Name", "Host CIDRs", "MAC Address", "Network Interfaces", "Network Interface Speeds", "Next Hops"})

	for _, node := range nodes {
		writer.Write([]string{
			node.ClusterNodeName,
			strings.Join(node.Network.HostCIDRs, "; "),
			node.Network.MACAddress,
			FormatNICNames(node.Network.NetworkInterfaces),
		FormatNICSpeedList(node.Network.NetworkInterfaces),
			strings.Join(node.Network.NextHops, "; "),
		})
	}

	return nil
}

// writeNetworkInterfacesCSV writes VM network interfaces to a CSV file
func (f *MultiCSVFormatter) writeNetworkInterfacesCSV(filename string, vms []vm.VMDetails) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{
		"VM Name",
		"VM Namespace",
		"Interface Name",
		"MAC Address",
		"IP Addresses",
		"Type",
		"Model",
		"Network Name",
		"NAD Name",
	})

	for _, vm := range vms {
		for _, netIf := range vm.NetworkInterfaces {
			ipStr := strings.Join(netIf.IPAddresses, ", ")
			writer.Write([]string{
				vm.Name,
				vm.Namespace,
				netIf.Name,
				netIf.MACAddress,
				ipStr,
				netIf.Type,
				netIf.Model,
				netIf.Network,
				netIf.NetworkAttachmentDefinition,
			})
		}
	}

	return nil
}

// writeCapacityPlanningCSV writes capacity planning data to a CSV file
func (f *MultiCSVFormatter) writeCapacityPlanningCSV(filename string, nodes []cluster.ClusterNodeInfo, vms []vm.VMDetails) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{
		"Node Name",
		"CPU Cores",
		"CPU Allocated to VMs",
		"CPU Overcommit Ratio",
		"Memory Capacity (GiB)",
		"Memory Allocated to VMs (GiB)",
		"Memory Overcommit Ratio",
		"VM Count",
		"Filesystem Used (%)",
		"Memory Used (%)",
	})

	nodeToVMs := make(map[string][]vm.VMDetails)
	for _, vm := range vms {
		if vm.Runtime != nil && vm.Runtime.GuestMetadata != nil && vm.Runtime.GuestMetadata.RunningOnNode != "" {
			nodeToVMs[vm.Runtime.GuestMetadata.RunningOnNode] = append(nodeToVMs[vm.Runtime.GuestMetadata.RunningOnNode], vm)
		}
	}

	for _, node := range nodes {
		vmList := nodeToVMs[node.ClusterNodeName]
		cpuAllocated := int64(0)
		memAllocatedGiB := 0.0
		for _, vm := range vmList {
			cpuAllocated += vm.CPUInfo.VCPUs
			memAllocatedGiB += float64(vm.MemoryInfo.MemoryConfiguredMiB) / 1024.0
		}

		cpuRatio := 0.0
		if node.CPU.CPUCores > 0 {
			cpuRatio = float64(cpuAllocated) / float64(node.CPU.CPUCores)
		}
		memNodeGiB := node.Memory.MemoryCapacityGiB
		memRatio := 0.0
		if memNodeGiB > 0 {
			memRatio = memAllocatedGiB / memNodeGiB
		}

		memUsedPct := node.Memory.MemoryUsedPercentage
		if memUsedPct == 0 && node.Memory.MemoryCapacityGiB > 0 {
			memUsedPct = (node.Memory.MemoryUsedGiB / node.Memory.MemoryCapacityGiB) * 100
		}

		writer.Write([]string{
			node.ClusterNodeName,
			fmt.Sprintf("%d", node.CPU.CPUCores),
			fmt.Sprintf("%d", cpuAllocated),
			fmt.Sprintf("%.1f:1", cpuRatio),
			fmt.Sprintf("%.1f", memNodeGiB),
			fmt.Sprintf("%.1f", memAllocatedGiB),
			fmt.Sprintf("%.1f:1", memRatio),
			fmt.Sprintf("%d", len(vmList)),
			fmt.Sprintf("%.1f", node.Filesystem.FilesystemUsagePercent),
			fmt.Sprintf("%.1f", memUsedPct),
		})
	}

	return nil
}

// writeVMAssessmentCSV writes VM assessment data to a CSV file
func (f *MultiCSVFormatter) writeVMAssessmentCSV(filename string, vms []vm.VMDetails) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{
		"VM Name",
		"Namespace",
		"Power State",
		"Guest Agent",
		"Memory Configured (MiB)",
		"Memory Used (MiB)",
		"Memory Utilization (%)",
		"Memory Waste Flag",
		"vCPUs",
		"Disk Count",
		"Total Disk Allocated (GiB)",
		"Total Disk Used (GiB)",
		"Storage Utilization (%)",
		"OS Detected",
		"Run Strategy",
	})

	for _, vm := range vms {
		hasAgent := "No"
		if vm.Runtime != nil && vm.Runtime.GuestMetadata != nil && vm.Runtime.GuestMetadata.GuestAgentVersion != "" {
			hasAgent = "Yes"
		}

		memConfigured := vm.MemoryInfo.MemoryConfiguredMiB
		memUsed := vm.MemoryInfo.TotalMemoryUsed
		memUtilPct := vm.MemoryInfo.MemoryUsedPercentage

		wasteFlag := ""
		if memConfigured > 0 && memUsed > 0 && memConfigured > (2 * memUsed) {
			wasteFlag = "OVERSIZED"
		}

		vCPUs := vm.CPUInfo.VCPUs
		diskCount := len(vm.Disks)

		var diskAllocGiB, diskUsedGiB float64
		for _, disk := range vm.Disks {
			diskAllocGiB += float64(disk.SizeBytes) / (1024 * 1024 * 1024)
			diskUsedGiB += float64(disk.TotalStorageInUse) / (1024 * 1024 * 1024)
		}

		storageUtilPct := 0.0
		if diskAllocGiB > 0 {
			storageUtilPct = (diskUsedGiB / diskAllocGiB) * 100
		}

		osDetected := vm.OSName
		if osDetected == "" && vm.Runtime != nil && vm.Runtime.GuestMetadata != nil {
			osDetected = vm.Runtime.GuestMetadata.OSVersion
		}

		writer.Write([]string{
			vm.Name,
			vm.Namespace,
			vm.Phase,
			hasAgent,
			fmt.Sprintf("%.1f", memConfigured),
			fmt.Sprintf("%.1f", memUsed),
			fmt.Sprintf("%.1f", memUtilPct),
			wasteFlag,
			fmt.Sprintf("%d", vCPUs),
			fmt.Sprintf("%d", diskCount),
			fmt.Sprintf("%.2f", diskAllocGiB),
			fmt.Sprintf("%.2f", diskUsedGiB),
			fmt.Sprintf("%.2f", storageUtilPct),
			osDetected,
			vm.RunStrategy,
		})
	}

	return nil
}

func (f *MultiCSVFormatter) writePVCInventoryCSV(filename string, pvcs []PVCInventoryItem) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{
		"PVC Name", "Namespace", "Status", "Capacity (GiB)", "Access Modes",
		"Storage Class", "Volume Mode", "Bound PV", "Owning VM",
		"Owning VM Namespace", "Created",
	})

	for _, pvc := range pvcs {
		capGiB := float64(pvc.Capacity) / (1024 * 1024 * 1024)
		writer.Write([]string{
			pvc.Name,
			pvc.Namespace,
			pvc.Status,
			fmt.Sprintf("%.2f", capGiB),
			strings.Join(pvc.AccessModes, ", "),
			pvc.StorageClass,
			pvc.VolumeMode,
			pvc.VolumeName,
			pvc.OwningVM,
			pvc.OwningVMNamespace,
			pvc.CreatedAt.Format("2006-01-02"),
		})
	}
	return nil
}

func (f *MultiCSVFormatter) writeNADInventoryCSV(filename string, nads []NADInfo) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{
		"Name", "Namespace", "Type", "VLAN", "Resource Name", "Created",
	})

	for _, nad := range nads {
		writer.Write([]string{
			nad.Name,
			nad.Namespace,
			nad.Type,
			nad.VLAN,
			nad.ResourceName,
			nad.CreatedAt.Format("2006-01-02"),
		})
	}
	return nil
}

func (f *MultiCSVFormatter) writeDataVolumesCSV(filename string, dvs []DataVolumeInfo) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{
		"Name", "Namespace", "Phase", "Progress", "Source Type",
		"Storage Size (GiB)", "Storage Class", "Owning VM", "Created",
	})

	for _, dv := range dvs {
		sizeGiB := float64(dv.StorageSize) / (1024 * 1024 * 1024)
		writer.Write([]string{
			dv.Name,
			dv.Namespace,
			dv.Phase,
			dv.Progress,
			dv.SourceType,
			fmt.Sprintf("%.2f", sizeGiB),
			dv.StorageClass,
			dv.OwningVM,
			dv.CreatedAt.Format("2006-01-02"),
		})
	}
	return nil
}

func (f *MultiCSVFormatter) writeMigrationReadinessCSV(filename string, vms []vm.VMDetails, pvcs []PVCInventoryItem) error {
	assessments := assessMigrationReadiness(vms, pvcs)
	parseScore := func(s string) int {
		parts := strings.SplitN(s, "/", 2)
		if len(parts) < 1 {
			return 0
		}
		n, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return 0
		}
		return n
	}
	sort.Slice(assessments, func(i, j int) bool {
		return parseScore(assessments[i].ReadinessScore) < parseScore(assessments[j].ReadinessScore)
	})
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.Write([]string{"VM Name", "Namespace", "Power State", "Live Migratable", "Run Strategy", "Eviction Strategy", "Host Devices", "Node Affinity", "PVC Access Mode Issue", "Dedicated CPU", "Guest Agent", "Blockers", "Readiness Score"})
	for _, a := range assessments {
		writer.Write([]string{
			a.VMName, a.VMNamespace, a.PowerState, a.LiveMigratable, a.RunStrategy, a.EvictionStrategy,
			a.HasHostDevices, a.HasNodeAffinity, a.PVCAccessModeIssue, a.HasDedicatedCPU, a.GuestAgentReady,
			strings.Join(a.Blockers, "; "), a.ReadinessScore,
		})
	}
	return nil
}

func (f *MultiCSVFormatter) writeStorageAnalysisCSV(filename string, pvcs []PVCInventoryItem, vms []vm.VMDetails) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{
		"PVC Name", "Namespace", "Storage Class", "Capacity (GiB)", "Access Modes",
		"Volume Mode", "Status", "Owning VM", "VM Power State", "Guest Used (GiB)",
		"Utilization (%)", "Flag",
	})

	rows := computeStorageAnalysis(pvcs, vms)
	for _, r := range rows {
		writer.Write([]string{
			r.PVCName,
			r.Namespace,
			r.StorageClass,
			fmt.Sprintf("%.2f", r.CapacityGiB),
			r.AccessModes,
			r.VolumeMode,
			r.Status,
			r.OwningVM,
			r.VMPowerState,
			fmt.Sprintf("%.2f", r.GuestUsedGiB),
			fmt.Sprintf("%.2f", r.UtilizationPct),
			r.Flag,
		})
	}
	return nil
}

func (f *MultiCSVFormatter) writeOperatorStatusCSV(filename string, operators []cluster.OperatorStatus) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{
		"Operator Name", "Source", "Namespace", "Version", "Status", "Health", "Created",
	})
	for _, op := range operators {
		writer.Write([]string{
			op.Name,
			op.Source,
			op.Namespace,
			op.Version,
			op.Status,
			op.Health,
			op.CreatedAt.Format("2006-01-02"),
		})
	}
	return nil
}
