package output

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"vm-scanner/pkg/vm"
)

// VMXLSXHeaders defines the headers for VM XLSX output
var VMXLSXHeaders = []string{
	"Name", "Namespace", "Phase", "Power State", "Hostname", "OS Name", "OS Version",
	"Running On Node",
	"vCPUs Total", "CPU Cores", "CPU Sockets", "CPU Threads", "CPU Model",
	"Configured Memory (MiB)", "Memory Free (MiB)", "Memory Used Total (MiB)", "Memory Used (%)",
	"Memory Used by LibVirt (MiB)", "Memory Used by VMI (MiB)", "Memory Hot Plug Max (MiB)",
	"Disk Count",
	"Guest Agent Version", "Timezone", "Virt Launcher Pod",
	"UID", "Created At", "Creation Age (days)", "Machine Type", "Kernel Version",
	"Run Strategy", "Eviction Strategy", "Instance Type", "Preference",
	"Labels",
}

// VMXLSXRowWriter writes VM data to Excel cells
type VMXLSXRowWriter struct {
	vm       vm.VMConsolidatedReport
	file     *excelize.File
	sheet    string
	rowIdx   int
	colIdx   int
}

// NewVMXLSXRowWriter creates a writer for Excel rows
func NewVMXLSXRowWriter(vmReport vm.VMConsolidatedReport, file *excelize.File, sheet string, rowIdx int) *VMXLSXRowWriter {
	return &VMXLSXRowWriter{
		vm:     vmReport,
		file:   file,
		sheet:  sheet,
		rowIdx: rowIdx,
		colIdx: 1,
	}
}

// WriteRow writes all VM data to the Excel sheet
func (w *VMXLSXRowWriter) WriteRow() error {
	w.colIdx = 1
	
	w.writeBasicFields()
	w.writeRuntimeFields()
	w.writeCPUFields()
	w.writeMemoryFields()
	w.writeStorageFields()
	// Network fields removed - see Network Interfaces sheet
	w.writeMetadataFields()
	
	return nil
}

// setCellValue writes a value and increments column
func (w *VMXLSXRowWriter) setCellValue(value interface{}) {
	cell, _ := excelize.CoordinatesToCellName(w.colIdx, w.rowIdx)
	w.file.SetCellValue(w.sheet, cell, value)
	w.colIdx++
}

// writeBasicFields writes basic VM information
func (w *VMXLSXRowWriter) writeBasicFields() {
	w.setCellValue(w.vm.Name)
	w.setCellValue(w.vm.Namespace)
	w.setCellValue(w.vm.Phase)
}

// writeRuntimeFields writes runtime information
func (w *VMXLSXRowWriter) writeRuntimeFields() {
	powerState := "Stopped"
	runningOnNode := ""
	osVersion := ""
	hostname := ""
	
	if w.vm.Runtime != nil {
		powerState = w.vm.Runtime.PowerState
		if w.vm.Runtime.GuestMetadata != nil {
			osVersion = w.vm.Runtime.GuestMetadata.OSVersion
			hostname = w.vm.Runtime.GuestMetadata.HostName
			runningOnNode = w.vm.Runtime.GuestMetadata.RunningOnNode
		}
	}
	
	w.setCellValue(powerState)
	w.setCellValue(hostname)
	w.setCellValue(w.vm.OSName)
	w.setCellValue(osVersion)
	w.setCellValue(runningOnNode)
}

// writeCPUFields writes CPU information
func (w *VMXLSXRowWriter) writeCPUFields() {
	w.setCellValue(w.vm.CPUInfo.VCPUs)
	w.setCellValue(w.vm.CPUInfo.CPUCores)
	w.setCellValue(w.vm.CPUInfo.CPUSockets)
	w.setCellValue(w.vm.CPUInfo.CPUThreads)
	w.setCellValue(w.vm.CPUInfo.CPUModel)
}

// writeMemoryFields writes memory information
func (w *VMXLSXRowWriter) writeMemoryFields() {
	w.setCellValue(fmt.Sprintf("%.1f", w.vm.MemoryInfo.MemoryConfiguredMiB))
	
	if w.vm.MemoryInfo.MemoryFree > 0 {
		w.setCellValue(fmt.Sprintf("%.1f", w.vm.MemoryInfo.MemoryFree))
	} else {
		w.setCellValue("")
	}
	
	if w.vm.MemoryInfo.TotalMemoryUsed > 0 {
		w.setCellValue(fmt.Sprintf("%.1f", w.vm.MemoryInfo.TotalMemoryUsed))
	} else {
		w.setCellValue("")
	}
	
	if w.vm.MemoryInfo.MemoryUsedPercentage > 0 {
		w.setCellValue(fmt.Sprintf("%.1f", w.vm.MemoryInfo.MemoryUsedPercentage))
	} else {
		w.setCellValue("")
	}
	
	if w.vm.MemoryInfo.MemoryUsedByLibVirt > 0 {
		w.setCellValue(fmt.Sprintf("%.1f", w.vm.MemoryInfo.MemoryUsedByLibVirt))
	} else {
		w.setCellValue("")
	}
	
	if w.vm.MemoryInfo.MemoryUsedByVMI > 0 {
		w.setCellValue(fmt.Sprintf("%.1f", w.vm.MemoryInfo.MemoryUsedByVMI))
	} else {
		w.setCellValue("")
	}
	
	if w.vm.MemoryInfo.MemoryHotPlugMax > 0 {
		w.setCellValue(fmt.Sprintf("%.1f", w.vm.MemoryInfo.MemoryHotPlugMax))
	} else {
		w.setCellValue("")
	}
}

// writeStorageFields writes storage information (Phase 1: simplified to disk count only)
func (w *VMXLSXRowWriter) writeStorageFields() {
	w.setCellValue(fmt.Sprintf("%d", len(w.vm.Disks)))
}

// writeNetworkFields removed -- network data moved to Network Interfaces sheet

// writeMetadataFields writes metadata information (Phase 1: no kernel version column; Phase 2: adds runStrategy/evictionStrategy/instanceType/preference)
func (w *VMXLSXRowWriter) writeMetadataFields() {
	var guestAgentVersion, timezone, virtLauncherPod string

	if w.vm.Runtime != nil && w.vm.Runtime.GuestMetadata != nil {
		guestAgentVersion = w.vm.Runtime.GuestMetadata.GuestAgentVersion
		timezone = w.vm.Runtime.GuestMetadata.Timezone
		virtLauncherPod = w.vm.Runtime.GuestMetadata.VirtLauncherPod
	}

	w.setCellValue(guestAgentVersion)
	w.setCellValue(timezone)
	w.setCellValue(virtLauncherPod)

	// UID: Use VMI UID if running, otherwise VM UID
	uid := w.vm.UID
	if w.vm.Runtime != nil && w.vm.Runtime.VMIUID != "" {
		uid = w.vm.Runtime.VMIUID
	}
	w.setCellValue(uid)

	// Created At: Use VMI creation time if running, otherwise VM creation time
	if w.vm.Runtime != nil && !w.vm.Runtime.CreationTimestamp.IsZero() {
		w.setCellValue(w.vm.Runtime.CreationTimestamp.Time.Format("2006-01-02 15:04:05"))
	} else if !w.vm.CreatedAt.IsZero() {
		w.setCellValue(w.vm.CreatedAt.Time.Format("2006-01-02 15:04:05"))
	} else {
		w.colIdx++
	}

	w.setCellValue(w.vm.MachineType)
	// Kernel Version column removed in Phase 1 -- see Network Interfaces sheet

	// Phase 2 fields from VM CR spec
	w.setCellValue(w.vm.RunStrategy)
	w.setCellValue(w.vm.EvictionStrategy)
	w.setCellValue(w.vm.InstanceType)
	w.setCellValue(w.vm.Preference)
	// Labels not available on VMConsolidatedReport -- not written here
}

// VMDetailsXLSXRowWriter writes VM.VMDetails to Excel cells
type VMDetailsXLSXRowWriter struct {
	vm     vm.VMDetails
	file   *excelize.File
	sheet  string
	rowNum int
	colNum int
}

// NewVMDetailsXLSXRowWriter creates a writer for VMDetails Excel rows
func NewVMDetailsXLSXRowWriter(vmDetail vm.VMDetails, file *excelize.File, sheet string, rowNum int) *VMDetailsXLSXRowWriter {
	return &VMDetailsXLSXRowWriter{
		vm:     vmDetail,
		file:   file,
		sheet:  sheet,
		rowNum: rowNum,
		colNum: 1,
	}
}

// WriteRow writes all VM data to the Excel sheet
func (w *VMDetailsXLSXRowWriter) WriteRow() error {
	w.colNum = 1

	w.writeBasicFields()
	w.writeRuntimeFields()
	w.writeCPUFields()
	w.writeMemoryFields()
	w.writeStorageFields()
	// Network fields removed - see Network Interfaces sheet
	w.writeMetadataFields()

	return nil
}

// setCellValue writes a value and increments column
func (w *VMDetailsXLSXRowWriter) setCellValue(value interface{}) {
	cell, _ := excelize.CoordinatesToCellName(w.colNum, w.rowNum)
	w.file.SetCellValue(w.sheet, cell, value)
	w.colNum++
}

// writeBasicFields writes basic VM information
func (w *VMDetailsXLSXRowWriter) writeBasicFields() {
	w.setCellValue(w.vm.Name)
	w.setCellValue(w.vm.Namespace)
	w.setCellValue(w.vm.Phase)
}

// writeRuntimeFields writes runtime information
func (w *VMDetailsXLSXRowWriter) writeRuntimeFields() {
	powerState := "Stopped"
	runningOnNode := ""
	osVersion := ""
	hostname := ""
	
	if w.vm.Runtime != nil {
		powerState = w.vm.Runtime.PowerState
		if w.vm.Runtime.GuestMetadata != nil {
			osVersion = w.vm.Runtime.GuestMetadata.OSVersion
			hostname = w.vm.Runtime.GuestMetadata.HostName
			runningOnNode = w.vm.Runtime.GuestMetadata.RunningOnNode
		}
	}
	
	w.setCellValue(powerState)
	w.setCellValue(hostname)
	w.setCellValue(w.vm.OSName)
	w.setCellValue(osVersion)
	w.setCellValue(runningOnNode)
}

// writeCPUFields writes CPU information
func (w *VMDetailsXLSXRowWriter) writeCPUFields() {
	w.setCellValue(w.vm.CPUInfo.VCPUs)
	w.setCellValue(w.vm.CPUInfo.CPUCores)
	w.setCellValue(w.vm.CPUInfo.CPUSockets)
	w.setCellValue(w.vm.CPUInfo.CPUThreads)
	w.setCellValue(w.vm.CPUInfo.CPUModel)
}

// writeMemoryFields writes memory information
func (w *VMDetailsXLSXRowWriter) writeMemoryFields() {
	w.setCellValue(fmt.Sprintf("%.1f", w.vm.MemoryInfo.MemoryConfiguredMiB))
	
	if w.vm.MemoryInfo.MemoryFree > 0 {
		w.setCellValue(fmt.Sprintf("%.1f", w.vm.MemoryInfo.MemoryFree))
	} else {
		w.setCellValue("")
	}
	
	if w.vm.MemoryInfo.TotalMemoryUsed > 0 {
		w.setCellValue(fmt.Sprintf("%.1f", w.vm.MemoryInfo.TotalMemoryUsed))
	} else {
		w.setCellValue("")
	}
	
	if w.vm.MemoryInfo.MemoryUsedPercentage > 0 {
		w.setCellValue(fmt.Sprintf("%.1f", w.vm.MemoryInfo.MemoryUsedPercentage))
	} else {
		w.setCellValue("")
	}
	
	if w.vm.MemoryInfo.MemoryUsedByLibVirt > 0 {
		w.setCellValue(fmt.Sprintf("%.1f", w.vm.MemoryInfo.MemoryUsedByLibVirt))
	} else {
		w.setCellValue("")
	}
	
	if w.vm.MemoryInfo.MemoryUsedByVMI > 0 {
		w.setCellValue(fmt.Sprintf("%.1f", w.vm.MemoryInfo.MemoryUsedByVMI))
	} else {
		w.setCellValue("")
	}
	
	if w.vm.MemoryInfo.MemoryHotPlugMax > 0 {
		w.setCellValue(fmt.Sprintf("%.1f", w.vm.MemoryInfo.MemoryHotPlugMax))
	} else {
		w.setCellValue("")
	}
}

// writeStorageFields writes storage information
func (w *VMDetailsXLSXRowWriter) writeStorageFields() {
	w.setCellValue(fmt.Sprintf("%d", len(w.vm.Disks)))
}

// writeNetworkFields writes network information -- no longer used, kept for interface compatibility
func (w *VMDetailsXLSXRowWriter) writeNetworkFields() {
	// Network data moved to dedicated Network Interfaces sheet
}

// writeMetadataFields writes metadata information
func (w *VMDetailsXLSXRowWriter) writeMetadataFields() {
	var guestAgentVersion, timezone, virtLauncherPod, kernelVersion string

	if w.vm.Runtime != nil && w.vm.Runtime.GuestMetadata != nil {
		guestAgentVersion = w.vm.Runtime.GuestMetadata.GuestAgentVersion
		timezone = w.vm.Runtime.GuestMetadata.Timezone
		virtLauncherPod = w.vm.Runtime.GuestMetadata.VirtLauncherPod
		kernelVersion = w.vm.Runtime.GuestMetadata.KernelVersion
	}

	w.setCellValue(guestAgentVersion)
	w.setCellValue(timezone)
	w.setCellValue(virtLauncherPod)

	// UID: Use VMI UID if running, otherwise VM UID
	uid := w.vm.UID
	if w.vm.Runtime != nil && w.vm.Runtime.VMIUID != "" {
		uid = w.vm.Runtime.VMIUID
	}
	w.setCellValue(uid)

	// Created At: Use VMI creation time if running, otherwise VM creation time
	createdAt := ""
	if w.vm.Runtime != nil && !w.vm.Runtime.CreationTimestamp.IsZero() {
		createdAt = w.vm.Runtime.CreationTimestamp.Time.Format("2006-01-02 15:04:05")
	} else if !w.vm.CreatedAt.IsZero() {
		createdAt = w.vm.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	w.setCellValue(createdAt)

	// Creation Age in days
	ageDays := ""
	if !w.vm.CreatedAt.IsZero() {
		ageDays = fmt.Sprintf("%d", int(time.Since(w.vm.CreatedAt.Time).Hours()/24))
	}
	w.setCellValue(ageDays)

	w.setCellValue(w.vm.MachineType)
	w.setCellValue(kernelVersion)
	w.setCellValue(w.vm.RunStrategy)
	w.setCellValue(w.vm.EvictionStrategy)
	w.setCellValue(w.vm.InstanceType)
	w.setCellValue(w.vm.Preference)
	w.setCellValue(formatLabelsForCell(w.vm.Labels))
}

// formatLabelsForCell formats a label map as "key=value" pairs joined by "; ".
// Keys are sorted for deterministic output.
func formatLabelsForCell(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + "=" + labels[k]
	}
	return strings.Join(parts, "; ")
}
