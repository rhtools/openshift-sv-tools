package output

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"vm-scanner/pkg/cluster"
	"vm-scanner/pkg/hardware"
	"vm-scanner/pkg/vm"

	"github.com/xuri/excelize/v2"
)

// XLSXFormatter handles XLSX output format
type XLSXFormatter struct {
	outputFile string
}

// NewXLSXFormatter creates a new XLSX formatter
func NewXLSXFormatter(outputFile string) *XLSXFormatter {
	return &XLSXFormatter{
		outputFile: outputFile,
	}
}

// Format routes to appropriate XLSX formatter based on data type
func (f *XLSXFormatter) Format(data interface{}) error {
	switch v := data.(type) {
	case *ComprehensiveReport:
		return f.formatComprehensiveReport(v)
	case []cluster.ClusterNodeInfo:
		return f.formatNodeList(v)
	case map[string]vm.VMConsolidatedReport:
		return f.formatVMRuntimeInfoMap(v)
	default:
		return fmt.Errorf("XLSX format not supported for type %T", v)
	}
}

// formatComprehensiveReport creates an Excel workbook with multiple sheets
func (f *XLSXFormatter) formatComprehensiveReport(report *ComprehensiveReport) error {
	xlsxFile := excelize.NewFile()
	defer xlsxFile.Close()

	summarySheet := "Summary"
	xlsxFile.SetSheetName("Sheet1", summarySheet)
	if err := f.writeComprehensiveReportSummarySheet(xlsxFile, summarySheet, report); err != nil {
		return fmt.Errorf("failed to write summary sheet: %w", err)
	}

	if len(report.Nodes) > 0 {
		nodeSheet := "Node Hardware"
		_, err := xlsxFile.NewSheet(nodeSheet)
		if err != nil {
			return fmt.Errorf("failed to create node hardware sheet: %w", err)
		}
		if err := f.writeNodeHardwareSheetDetailed(xlsxFile, nodeSheet, report.Nodes); err != nil {
			return fmt.Errorf("failed to write node hardware sheet: %w", err)
		}
	}

	vmsSheet := "Virtual Machines"
	_, err := xlsxFile.NewSheet(vmsSheet)
	if err != nil {
		return fmt.Errorf("failed to create VMs sheet: %w", err)
	}
	if err := f.writeVMsSheet(xlsxFile, vmsSheet, report.VMs); err != nil {
		return fmt.Errorf("failed to write VMs sheet: %w", err)
	}

	storageSheet := "Storage Classes"
	_, err = xlsxFile.NewSheet(storageSheet)
	if err != nil {
		return fmt.Errorf("failed to create storage sheet: %w", err)
	}
	if err := f.writeStorageSheet(xlsxFile, storageSheet, report.Storage); err != nil {
		return fmt.Errorf("failed to write storage sheet: %w", err)
	}

	vmsDiskSheet := "VM Disks"
	_, err = xlsxFile.NewSheet(vmsDiskSheet)
	if err != nil {
		return fmt.Errorf("failed to create VM disks sheet: %w", err)
	}
	if err := f.writeVMDisksSheet(xlsxFile, vmsDiskSheet, report.VMs); err != nil {
		return fmt.Errorf("failed to write VM disks sheet: %w", err)
	}

	networkIfSheet := "Network Interfaces"
	_, err = xlsxFile.NewSheet(networkIfSheet)
	if err != nil {
		return fmt.Errorf("failed to create network interfaces sheet: %w", err)
	}
	if err := f.writeNetworkInterfacesSheet(xlsxFile, networkIfSheet, report.VMs); err != nil {
		return fmt.Errorf("failed to write network interfaces sheet: %w", err)
	}

	capacitySheet := "Capacity Planning"
	_, err = xlsxFile.NewSheet(capacitySheet)
	if err != nil {
		return fmt.Errorf("failed to create capacity planning sheet: %w", err)
	}
	if err := f.writeCapacityPlanningSheet(xlsxFile, capacitySheet, report.Nodes, report.VMs); err != nil {
		return fmt.Errorf("failed to write capacity planning sheet: %w", err)
	}

	assessmentSheet := "VM Assessment"
	_, err = xlsxFile.NewSheet(assessmentSheet)
	if err != nil {
		return fmt.Errorf("failed to create VM assessment sheet: %w", err)
	}
	if err := f.writeVMAssessmentSheet(xlsxFile, assessmentSheet, report.VMs); err != nil {
		return fmt.Errorf("failed to write VM assessment sheet: %w", err)
	}

	pvcSheet := "PVC Inventory"
	_, err = xlsxFile.NewSheet(pvcSheet)
	if err != nil {
		return fmt.Errorf("failed to create PVC inventory sheet: %w", err)
	}
	if err := f.writePVCInventorySheet(xlsxFile, pvcSheet, report.PVCs); err != nil {
		return fmt.Errorf("failed to write PVC inventory sheet: %w", err)
	}

	nadSheet := "NAD Inventory"
	_, err = xlsxFile.NewSheet(nadSheet)
	if err != nil {
		return fmt.Errorf("failed to create NAD inventory sheet: %w", err)
	}
	if err := f.writeNADInventorySheet(xlsxFile, nadSheet, report.NADs); err != nil {
		return fmt.Errorf("failed to write NAD inventory sheet: %w", err)
	}

	dvSheet := "DataVolumes"
	_, err = xlsxFile.NewSheet(dvSheet)
	if err != nil {
		return fmt.Errorf("failed to create DataVolumes sheet: %w", err)
	}
	if err := f.writeDataVolumesSheet(xlsxFile, dvSheet, report.DataVolumes); err != nil {
		return fmt.Errorf("failed to write DataVolumes sheet: %w", err)
	}

	migrationSheet := "Migration Readiness"
	_, err = xlsxFile.NewSheet(migrationSheet)
	if err != nil {
		return fmt.Errorf("failed to create migration readiness sheet: %w", err)
	}
	if err := f.writeMigrationReadinessSheet(xlsxFile, migrationSheet, report.VMs, report.PVCs); err != nil {
		return fmt.Errorf("failed to write migration readiness sheet: %w", err)
	}

	storageAnalysisSheet := "Storage Analysis"
	_, err = xlsxFile.NewSheet(storageAnalysisSheet)
	if err != nil {
		return fmt.Errorf("failed to create storage analysis sheet: %w", err)
	}
	if err := f.writeStorageAnalysisSheet(xlsxFile, storageAnalysisSheet, report.PVCs, report.VMs); err != nil {
		return fmt.Errorf("failed to write storage analysis sheet: %w", err)
	}

	operatorSheet := "Operator Status"
	_, err = xlsxFile.NewSheet(operatorSheet)
	if err != nil {
		return fmt.Errorf("failed to create operator status sheet: %w", err)
	}
	if err := f.writeOperatorStatusSheet(xlsxFile, operatorSheet, report.Cluster.Operators); err != nil {
		return fmt.Errorf("failed to write operator status sheet: %w", err)
	}

	outputFile := f.outputFile
	if outputFile == "" {
		outputFile = "report.xlsx"
	}
	if err := xlsxFile.SaveAs(outputFile); err != nil {
		return fmt.Errorf("failed to save XLSX file: %w", err)
	}

	fmt.Printf("Excel report saved to: %s\n", outputFile)
	return nil
}

// formatNodeList formats node hardware info as Excel workbook
func (f *XLSXFormatter) formatNodeList(nodes []cluster.ClusterNodeInfo) error {
	xlsxFile := excelize.NewFile()
	defer xlsxFile.Close()

	sheetName := "Node Hardware"
	xlsxFile.SetSheetName("Sheet1", sheetName)

	headers := []string{"Node Name", "Node Roles", "CPU Cores", "CPU Model", "Memory Capacity (GiB)", "Memory Used (GiB)", "Memory Used (%)", "Storage", "Filesystem Available (GB)",
		"Filesystem Used (GB)", "OS Version", "Kernel Version", "Schedulable"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheetName, cell, header)
	}

	for rowIdx, node := range nodes {
		row := rowIdx + 2
		memoryUsedPercent := 0.0
		if node.Memory.MemoryCapacityGiB > 0 {
			memoryUsedPercent = (node.Memory.MemoryUsedGiB / node.Memory.MemoryCapacityGiB) * 100
		}
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("A%d", row), node.ClusterNodeName)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", row), strings.Join(node.NodeRoles, "; "))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", row), node.CPU.CPUCores)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("D%d", row), node.CPU.CPUModel)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("E%d", row), node.Memory.MemoryCapacityGiB)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("F%d", row), node.Memory.MemoryUsedGiB)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("G%d", row), memoryUsedPercent)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("H%d", row), node.StorageCapacity)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("I%d", row), node.Filesystem.FilesystemAvailable)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("J%d", row), node.Filesystem.FilesystemUsed)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("K%d", row), node.CoreOSVersion)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("L%d", row), node.NodeKernelVersion)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("M%d", row), node.NodeSchedulable)
	}

	outputFile := f.outputFile
	if outputFile == "" {
		outputFile = "node-hardware.xlsx"
	}
	if err := xlsxFile.SaveAs(outputFile); err != nil {
		return fmt.Errorf("failed to save XLSX file: %w", err)
	}

	fmt.Printf("Excel report saved to: %s\n", outputFile)
	return nil
}

// formatVMRuntimeInfoMap formats VM runtime inventory map as Excel workbook
func (f *XLSXFormatter) formatVMRuntimeInfoMap(vmiMap map[string]vm.VMConsolidatedReport) error {
	xlsxFile := excelize.NewFile()
	defer xlsxFile.Close()

	sheetName := "VM Runtime Inventory"
	xlsxFile.SetSheetName("Sheet1", sheetName)

	for i, header := range VMXLSXHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheetName, cell, header)
	}

	rowIdx := 2
	for _, vmReport := range vmiMap {
		writer := NewVMXLSXRowWriter(vmReport, xlsxFile, sheetName, rowIdx)
		if err := writer.WriteRow(); err != nil {
			return fmt.Errorf("failed to write VM row: %w", err)
		}
		rowIdx++
	}

	outputFile := f.outputFile
	if outputFile == "" {
		outputFile = "vm-runtime-inventory.xlsx"
	}
	if err := xlsxFile.SaveAs(outputFile); err != nil {
		return fmt.Errorf("failed to save XLSX file: %w", err)
	}

	fmt.Printf("Excel report saved to: %s\n", outputFile)
	return nil
}

func (f *XLSXFormatter) writeComprehensiveReportSummarySheet(xlsxFile *excelize.File, sheetName string, report *ComprehensiveReport) error {
	if report == nil || report.Summary == nil {
		return nil
	}
	summary := report.Summary
	clusterInfo := report.Cluster
	nodes := report.Nodes
	vms := report.VMs

	data := [][]string{}

	if clusterInfo != nil {
		data = append(data, []string{"Cluster Name", clusterInfo.ClusterName})
		data = append(data, []string{"Cluster ID", clusterInfo.ClusterID})
		data = append(data, []string{"OpenShift Version", clusterInfo.ClusterVersion})
		data = append(data, []string{"Kubernetes Version", clusterInfo.KubernetesVersion})
		if clusterInfo.KubeVirtVersion != nil {
			data = append(data, []string{"KubeVirt Version", clusterInfo.KubeVirtVersion.Version})
			data = append(data, []string{"KubeVirt Deployed", clusterInfo.KubeVirtVersion.Deployed})
		}
		data = append(data, []string{"Schedulable Control Plane Count", fmt.Sprintf("%d", clusterInfo.SchedulableControlPlaneCount)})
		data = append(data, []string{"Control Plane Schedulable", fmt.Sprintf("%t", clusterInfo.HasSchedulableControlPlane)})
		data = append(data, []string{"Worker Nodes Count", fmt.Sprintf("%d", clusterInfo.WorkerNodesCount)})
		data = append(data, []string{"", ""})
	}

	if summary.ClusterSummary != nil {
		data = append(data, []string{"Total VMs", fmt.Sprintf("%d", summary.ClusterSummary.TotalVMs)})
		data = append(data, []string{"Running VMs", fmt.Sprintf("%d", summary.ClusterSummary.RunningVMs)})
		data = append(data, []string{"Stopped VMs", fmt.Sprintf("%d", summary.ClusterSummary.StoppedVMs)})
		if summary.ClusterSummary.Resources != nil {
			data = append(data, []string{"Total CPU", fmt.Sprintf("%d", summary.ClusterSummary.Resources.TotalCPU)})
			data = append(data, []string{"Total Memory (GiB)", fmt.Sprintf("%.1f", summary.ClusterSummary.Resources.TotalMemory)})
			data = append(data, []string{"Used Memory (GiB)", fmt.Sprintf("%.1f", summary.ClusterSummary.Resources.UsedMemory)})
			data = append(data, []string{"Total Local Storage (GiB)", fmt.Sprintf("%.1f", float64(summary.ClusterSummary.Resources.TotalLocalStorage))})
			data = append(data, []string{"Total Application Requested Storage (GiB)", fmt.Sprintf("%.1f", float64(summary.ClusterSummary.Resources.TotalApplicationRequestedStorage))})
			data = append(data, []string{"Total Application Used Storage (GiB)", fmt.Sprintf("%.1f", float64(summary.ClusterSummary.Resources.TotalApplicationUsedStorage))})
		}
		data = append(data, []string{"Total Namespaces", fmt.Sprintf("%d", summary.ClusterSummary.TotalNamespaces)})
		data = append(data, []string{"User Created Namespaces", fmt.Sprintf("%d", summary.ClusterSummary.UserNamespaces)})
		data = append(data, []string{"Nodes", fmt.Sprintf("%d", len(nodes))})
	}
	data = append(data, []string{"Storage Classes", fmt.Sprintf("%d", summary.StorageClasses)})

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
	data = append(data, []string{"Guest Agent Coverage (%)", fmt.Sprintf("%.1f%%", agentCoverage)})

	// Phase 3 inventory counts
	data = append(data, []string{"", ""})
	data = append(data, []string{"PVC Count", fmt.Sprintf("%d", len(report.PVCs))})
	var totalPVCStorageGiB float64
	for _, pvc := range report.PVCs {
		totalPVCStorageGiB += float64(pvc.Capacity) / (1024 * 1024 * 1024)
	}
	data = append(data, []string{"Total PVC Storage (GiB)", fmt.Sprintf("%.1f", totalPVCStorageGiB)})
	data = append(data, []string{"NAD Count", fmt.Sprintf("%d", len(report.NADs))})
	data = append(data, []string{"DataVolume Count", fmt.Sprintf("%d", len(report.DataVolumes))})
	dvInProgress := 0
	for _, dv := range report.DataVolumes {
		if dv.Phase != "Succeeded" {
			dvInProgress++
		}
	}
	data = append(data, []string{"DataVolumes In Progress", fmt.Sprintf("%d", dvInProgress)})

	// Phase 4 operator and migration counts
	data = append(data, []string{"", ""})
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
	data = append(data, []string{"Operator Count", fmt.Sprintf("%d", operatorCount)})
	data = append(data, []string{"Healthy Operators", fmt.Sprintf("%d", healthyOperators)})

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
	data = append(data, []string{"Migration Ready VMs", fmt.Sprintf("%d", migrationReady)})
	data = append(data, []string{"Migration Blocked VMs", fmt.Sprintf("%d", migrationBlocked)})

	for i, row := range data {
		rowNum := i + 1
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), row[0])
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), row[1])
	}

	return nil
}

func (f *XLSXFormatter) writeVMsSheet(xlsxFile *excelize.File, sheetName string, vms []vm.VMDetails) error {
	for i, header := range VMXLSXHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheetName, cell, header)
	}

	for rowIdx, vmDetail := range vms {
		rowNum := rowIdx + 2
		writer := NewVMDetailsXLSXRowWriter(vmDetail, xlsxFile, sheetName, rowNum)
		if err := writer.WriteRow(); err != nil {
			return fmt.Errorf("failed to write VM row: %w", err)
		}
	}

	return nil
}

func (f *XLSXFormatter) writeStorageSheet(xlsxFile *excelize.File, sheetName string, storageClasses []hardware.StorageClassInfo) error {
	headers := []string{"Name", "Provisioner", "Reclaim Policy", "Volume Binding Mode", "Default", "Allow Expansion"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheetName, cell, header)
	}

	for rowIdx, sc := range storageClasses {
		row := rowIdx + 2
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("A%d", row), sc.Name)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", row), sc.Provisioner)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", row), sc.ReclaimPolicy)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("D%d", row), sc.VolumeBindingMode)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("E%d", row), sc.IsDefault)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("F%d", row), sc.AllowVolumeExpansion)
	}

	return nil
}

func (f *XLSXFormatter) writeClusterSheet(xlsxFile *excelize.File, sheetName string, clusterInfo *cluster.ClusterSummary) error {
	data := [][]string{
		{"Cluster Name", clusterInfo.ClusterName},
		{"Kubernetes Version", clusterInfo.KubernetesVersion},
		{"Total VMs", fmt.Sprintf("%d", clusterInfo.TotalVMs)},
		{"Running VMs", fmt.Sprintf("%d", clusterInfo.RunningVMs)},
		{"Stopped VMs", fmt.Sprintf("%d", clusterInfo.StoppedVMs)},
	}

	if clusterInfo.KubeVirtVersion != nil {
		data = append(data, []string{"KubeVirt Version", clusterInfo.KubeVirtVersion.Version})
		data = append(data, []string{"KubeVirt Deployed", clusterInfo.KubeVirtVersion.Deployed})
	}

	for i, row := range data {
		rowNum := i + 1
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), row[0])
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), row[1])
	}

	return nil
}

func (f *XLSXFormatter) writeNodeHardwareSheetDetailed(xlsxFile *excelize.File, sheetName string, nodes []cluster.ClusterNodeInfo) error {
	headers := []string{
		"Node Name",
		"Node Roles",
		"OS Version",
		"Kernel Version",
		"Schedulable",
		"Storage Capacity",
		"Pod Limits",
		"CPU Cores",
		"CPU Model",
		"Filesystem Available (GB)",
		"Filesystem Capacity (GB)",
		"Filesystem Used (GB)",
		"Filesystem Usage (%)",
		"Memory Capacity (GiB)",
		"Memory Used (GiB)",
		"Memory Used (%)",
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
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheetName, cell, header)
	}

	for rowIdx, node := range nodes {
		row := rowIdx + 2

		podSubnets := "None"
		if len(node.Network.PodNetworkSubnet) > 0 {
			var subnetParts []string
			for key, values := range node.Network.PodNetworkSubnet {
				subnetParts = append(subnetParts, fmt.Sprintf("%s: %s", key, strings.Join(values, ", ")))
			}
			podSubnets = strings.Join(subnetParts, "; ")
		}

		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("A%d", row), node.ClusterNodeName)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("B%d", row), strings.Join(node.NodeRoles, "; "))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("C%d", row), node.CoreOSVersion)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("D%d", row), node.NodeKernelVersion)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("E%d", row), node.NodeSchedulable)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("%.1f", node.StorageCapacity))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("G%d", row), node.NodePodLimits)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("H%d", row), node.CPU.CPUCores)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("I%d", row), node.CPU.CPUModel)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("J%d", row), fmt.Sprintf("%.1f", node.Filesystem.FilesystemAvailable))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("K%d", row), fmt.Sprintf("%.1f", node.Filesystem.FilesystemCapacity))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("L%d", row), fmt.Sprintf("%.1f", node.Filesystem.FilesystemUsed))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("M%d", row), fmt.Sprintf("%.1f", node.Filesystem.FilesystemUsagePercent))

		memoryUsedPercent := 0.0
		if node.Memory.MemoryCapacityGiB > 0 {
			memoryUsedPercent = (node.Memory.MemoryUsedGiB / node.Memory.MemoryCapacityGiB) * 100
		}
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("N%d", row), fmt.Sprintf("%.1f", node.Memory.MemoryCapacityGiB))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("O%d", row), fmt.Sprintf("%.1f", node.Memory.MemoryUsedGiB))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("P%d", row), fmt.Sprintf("%.1f", memoryUsedPercent))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("Q%d", row), node.Network.BridgeID)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("R%d", row), node.Network.InterfaceID)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("S%d", row), strings.Join(node.Network.IPAddresses, "; "))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("T%d", row), node.Network.MACAddress)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("U%d", row), node.Network.Mode)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("V%d", row), FormatNICNames(node.Network.NetworkInterfaces))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("W%d", row), FormatNICSpeedList(node.Network.NetworkInterfaces))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("X%d", row), strings.Join(node.Network.NextHops, "; "))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("Y%d", row), node.Network.NodePortEnable)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("Z%d", row), node.Network.VLANID)
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("AA%d", row), strings.Join(node.Network.HostCIDRs, "; "))
		xlsxFile.SetCellValue(sheetName, fmt.Sprintf("AB%d", row), podSubnets)
	}

	return nil
}

func (f *XLSXFormatter) writeVMDisksSheet(xlsxFile *excelize.File, sheetName string, vms []vm.VMDetails) error {
	headers := []string{
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
	}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheetName, cell, header)
	}

	rowNum := 2
	for _, vmDetail := range vms {
		for volName, disk := range vmDetail.Disks {
			pvcSizeGiB := float64(disk.SizeBytes) / (1024 * 1024 * 1024)
			guestTotalGiB := float64(disk.TotalStorage) / (1024 * 1024 * 1024)
			guestUsedGiB := float64(disk.TotalStorageInUse) / (1024 * 1024 * 1024)
			guestFreeGiB := guestTotalGiB - guestUsedGiB

			cells := []interface{}{
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
			}
			for colIdx, val := range cells {
				cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowNum)
				xlsxFile.SetCellValue(sheetName, cell, val)
			}
			rowNum++
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

				cells := []interface{}{
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
				}
				for colIdx, val := range cells {
					cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowNum)
					xlsxFile.SetCellValue(sheetName, cell, val)
				}
				rowNum++
			}
		}
	}

	return nil
}

func (f *XLSXFormatter) writeNetworkInterfacesSheet(xlsxFile *excelize.File, sheetName string, vms []vm.VMDetails) error {
	headers := []string{
		"VM Name",
		"VM Namespace",
		"Interface Name",
		"MAC Address",
		"IP Addresses",
		"Type",
		"Model",
		"Network Name",
		"NAD Name",
	}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheetName, cell, header)
	}

	rowNum := 2
	for _, vmDetail := range vms {
		for _, netIf := range vmDetail.NetworkInterfaces {
			ipStr := strings.Join(netIf.IPAddresses, ", ")
			cells := []interface{}{
				vmDetail.Name,
				vmDetail.Namespace,
				netIf.Name,
				netIf.MACAddress,
				ipStr,
				netIf.Type,
				netIf.Model,
				netIf.Network,
				netIf.NetworkAttachmentDefinition,
			}
			for colIdx, val := range cells {
				cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowNum)
				xlsxFile.SetCellValue(sheetName, cell, val)
			}
			rowNum++
		}
	}

	return nil
}

func (f *XLSXFormatter) writeCapacityPlanningSheet(xlsxFile *excelize.File, sheetName string, nodes []cluster.ClusterNodeInfo, vms []vm.VMDetails) error {
	headers := []string{
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
	}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheetName, cell, header)
	}

	nodeToVMs := make(map[string][]vm.VMDetails)
	for _, vm := range vms {
		if vm.Runtime != nil && vm.Runtime.GuestMetadata != nil && vm.Runtime.GuestMetadata.RunningOnNode != "" {
			nodeToVMs[vm.Runtime.GuestMetadata.RunningOnNode] = append(nodeToVMs[vm.Runtime.GuestMetadata.RunningOnNode], vm)
		}
	}

	rowNum := 2
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

		cells := []interface{}{
			node.ClusterNodeName,
			node.CPU.CPUCores,
			cpuAllocated,
			fmt.Sprintf("%.1f:1", cpuRatio),
			fmt.Sprintf("%.1f", memNodeGiB),
			fmt.Sprintf("%.1f", memAllocatedGiB),
			fmt.Sprintf("%.1f:1", memRatio),
			len(vmList),
			fmt.Sprintf("%.1f", node.Filesystem.FilesystemUsagePercent),
			fmt.Sprintf("%.1f", memUsedPct),
		}
		for colIdx, val := range cells {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowNum)
			xlsxFile.SetCellValue(sheetName, cell, val)
		}
		rowNum++
	}

	return nil
}

func (f *XLSXFormatter) writeVMAssessmentSheet(xlsxFile *excelize.File, sheetName string, vms []vm.VMDetails) error {
	headers := []string{
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
	}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheetName, cell, header)
	}

	rowNum := 2
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

		cells := []interface{}{
			vm.Name,
			vm.Namespace,
			vm.Phase,
			hasAgent,
			memConfigured,
			memUsed,
			memUtilPct,
			wasteFlag,
			vCPUs,
			diskCount,
			fmt.Sprintf("%.2f", diskAllocGiB),
			fmt.Sprintf("%.2f", diskUsedGiB),
			fmt.Sprintf("%.2f", storageUtilPct),
			osDetected,
			vm.RunStrategy,
		}
		for colIdx, val := range cells {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowNum)
			xlsxFile.SetCellValue(sheetName, cell, val)
		}
		rowNum++
	}

	return nil
}

func (f *XLSXFormatter) writePVCInventorySheet(xlsxFile *excelize.File, sheetName string, pvcs []PVCInventoryItem) error {
	headers := []string{
		"PVC Name", "Namespace", "Status", "Capacity (GiB)", "Access Modes",
		"Storage Class", "Volume Mode", "Bound PV", "Owning VM",
		"Owning VM Namespace", "Created",
	}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheetName, cell, header)
	}

	for rowIdx, pvc := range pvcs {
		row := rowIdx + 2
		capGiB := float64(pvc.Capacity) / (1024 * 1024 * 1024)
		cells := []interface{}{
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
		}
		for colIdx, val := range cells {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			xlsxFile.SetCellValue(sheetName, cell, val)
		}
	}
	return nil
}

func (f *XLSXFormatter) writeNADInventorySheet(xlsxFile *excelize.File, sheetName string, nads []NADInfo) error {
	headers := []string{
		"Name", "Namespace", "Type", "VLAN", "Resource Name", "Created",
	}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheetName, cell, header)
	}

	for rowIdx, nad := range nads {
		row := rowIdx + 2
		cells := []interface{}{
			nad.Name,
			nad.Namespace,
			nad.Type,
			nad.VLAN,
			nad.ResourceName,
			nad.CreatedAt.Format("2006-01-02"),
		}
		for colIdx, val := range cells {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			xlsxFile.SetCellValue(sheetName, cell, val)
		}
	}
	return nil
}

func (f *XLSXFormatter) writeDataVolumesSheet(xlsxFile *excelize.File, sheetName string, dvs []DataVolumeInfo) error {
	headers := []string{
		"Name", "Namespace", "Phase", "Progress", "Source Type",
		"Storage Size (GiB)", "Storage Class", "Owning VM", "Created",
	}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheetName, cell, header)
	}

	for rowIdx, dv := range dvs {
		row := rowIdx + 2
		sizeGiB := float64(dv.StorageSize) / (1024 * 1024 * 1024)
		cells := []interface{}{
			dv.Name,
			dv.Namespace,
			dv.Phase,
			dv.Progress,
			dv.SourceType,
			fmt.Sprintf("%.2f", sizeGiB),
			dv.StorageClass,
			dv.OwningVM,
			dv.CreatedAt.Format("2006-01-02"),
		}
		for colIdx, val := range cells {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			xlsxFile.SetCellValue(sheetName, cell, val)
		}
	}
	return nil
}

func (f *XLSXFormatter) writeMigrationReadinessSheet(xlsxFile *excelize.File, sheetName string, vms []vm.VMDetails, pvcs []PVCInventoryItem) error {
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
	headers := []string{"VM Name", "Namespace", "Power State", "Live Migratable", "Run Strategy", "Eviction Strategy", "Host Devices", "Node Affinity", "PVC Access Mode Issue", "Dedicated CPU", "Guest Agent", "Blockers", "Readiness Score"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheetName, cell, header)
	}
	for rowIdx, a := range assessments {
		row := rowIdx + 2
		cells := []interface{}{
			a.VMName, a.VMNamespace, a.PowerState, a.LiveMigratable, a.RunStrategy, a.EvictionStrategy,
			a.HasHostDevices, a.HasNodeAffinity, a.PVCAccessModeIssue, a.HasDedicatedCPU, a.GuestAgentReady,
			strings.Join(a.Blockers, "; "), a.ReadinessScore,
		}
		for colIdx, val := range cells {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			xlsxFile.SetCellValue(sheetName, cell, val)
		}
	}
	return nil
}

func (f *XLSXFormatter) writeStorageAnalysisSheet(xlsxFile *excelize.File, sheetName string, pvcs []PVCInventoryItem, vms []vm.VMDetails) error {
	rows := computeStorageAnalysis(pvcs, vms)
	headers := []string{
		"PVC Name", "Namespace", "Storage Class", "Capacity (GiB)", "Access Modes",
		"Volume Mode", "Status", "Owning VM", "VM Power State", "Guest Used (GiB)",
		"Utilization (%)", "Flag",
	}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheetName, cell, header)
	}
	for rowIdx, r := range rows {
		row := rowIdx + 2
		cells := []interface{}{
			r.PVCName,
			r.Namespace,
			r.StorageClass,
			r.CapacityGiB,
			r.AccessModes,
			r.VolumeMode,
			r.Status,
			r.OwningVM,
			r.VMPowerState,
			r.GuestUsedGiB,
			r.UtilizationPct,
			r.Flag,
		}
		for colIdx, val := range cells {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			xlsxFile.SetCellValue(sheetName, cell, val)
		}
	}
	return nil
}

func (f *XLSXFormatter) writeOperatorStatusSheet(xlsxFile *excelize.File, sheetName string, operators []cluster.OperatorStatus) error {
	headers := []string{
		"Operator Name", "Source", "Namespace", "Version", "Status", "Health", "Created",
	}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		xlsxFile.SetCellValue(sheetName, cell, header)
	}
	for rowIdx, op := range operators {
		row := rowIdx + 2
		cells := []interface{}{
			op.Name,
			op.Source,
			op.Namespace,
			op.Version,
			op.Status,
			op.Health,
			op.CreatedAt.Format("2006-01-02"),
		}
		for colIdx, val := range cells {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			xlsxFile.SetCellValue(sheetName, cell, val)
		}
	}
	return nil
}
