package output

import (
	"fmt"
	"time"
	"vm-scanner/pkg/vm"
)

// VMCSVHeaders defines the headers for VM CSV output
var VMCSVHeaders = []string{
	"Name",
	"Namespace",
	"Phase",
	"Power State",
	"Hostname",
	"OS Name",
	"OS Version",
	"Running On Node",
	"vCPUs Total",
	"CPU Cores",
	"CPU Sockets",
	"CPU Threads",
	"CPU Model",
	"Configured Memory (MiB)",
	"Memory Free (MiB)",
	"Memory Used Total (MiB)",
	"Memory Used (%)",
	"Memory Used by LibVirt (MiB)",
	"Memory Used by VMI (MiB)",
	"Memory Hot Plug Max (MiB)",
	"Disk Count",
	"Guest Agent Version",
	"Timezone",
	"Virt Launcher Pod",
	"UID",
	"Created At",
	"Creation Age (days)",
	"Machine Type",
	"Kernel Version",
	"Run Strategy",
	"Eviction Strategy",
	"Instance Type",
	"Preference",
	"Labels",
}

// VMCSVRowBuilder builds CSV rows from VM data
type VMCSVRowBuilder struct {
	vm vm.VMDetails
}

// NewVMCSVRowBuilder creates a builder for the given VM
func NewVMCSVRowBuilder(vmDetail vm.VMDetails) *VMCSVRowBuilder {
	return &VMCSVRowBuilder{vm: vmDetail}
}

// BuildRow constructs a complete CSV row
func (b *VMCSVRowBuilder) BuildRow() []string {
	row := make([]string, 34)

	// Populate sections
	b.populateBasicFields(row)
	b.populateRuntimeFields(row)
	b.populateCPUFields(row)
	b.populateMemoryFields(row)
	b.populateStorageFields(row)
	// Network fields removed - see Network Interfaces sheet
	b.populateMetadataFields(row)

	return row
}

// populateBasicFields fills basic VM information
func (b *VMCSVRowBuilder) populateBasicFields(row []string) {
	row[0] = b.vm.Name
	row[1] = b.vm.Namespace
	row[2] = b.vm.Phase
	row[5] = b.vm.OSName
}

// populateRuntimeFields fills runtime information
func (b *VMCSVRowBuilder) populateRuntimeFields(row []string) {
	powerState := "Stopped"
	runningOnNode := ""
	osVersion := ""
	hostname := ""
	kernelVersion := ""
	guestAgentVersion := ""
	timezone := ""
	virtLauncherPod := ""

	if b.vm.Runtime != nil {
		powerState = b.vm.Runtime.PowerState
		if b.vm.Runtime.GuestMetadata != nil {
			osVersion = b.vm.Runtime.GuestMetadata.OSVersion
			hostname = b.vm.Runtime.GuestMetadata.HostName
			kernelVersion = b.vm.Runtime.GuestMetadata.KernelVersion
			guestAgentVersion = b.vm.Runtime.GuestMetadata.GuestAgentVersion
			timezone = b.vm.Runtime.GuestMetadata.Timezone
			runningOnNode = b.vm.Runtime.GuestMetadata.RunningOnNode
			virtLauncherPod = b.vm.Runtime.GuestMetadata.VirtLauncherPod
		}
	}

	row[3] = powerState
	row[4] = hostname
	row[6] = osVersion
	row[7] = runningOnNode
	row[21] = guestAgentVersion
	row[22] = timezone
	row[23] = virtLauncherPod
	row[28] = kernelVersion
}

// populateCPUFields fills CPU information
func (b *VMCSVRowBuilder) populateCPUFields(row []string) {
	row[8] = fmt.Sprintf("%d", b.vm.CPUInfo.VCPUs)
	row[9] = fmt.Sprintf("%d", b.vm.CPUInfo.CPUCores)
	row[10] = fmt.Sprintf("%d", b.vm.CPUInfo.CPUSockets)
	row[11] = fmt.Sprintf("%d", b.vm.CPUInfo.CPUThreads)
	row[12] = b.vm.CPUInfo.CPUModel
}

// populateMemoryFields fills memory information
func (b *VMCSVRowBuilder) populateMemoryFields(row []string) {
	if b.vm.MemoryInfo.MemoryConfiguredMiB > 0 {
		row[13] = fmt.Sprintf("%.1f", b.vm.MemoryInfo.MemoryConfiguredMiB)
	}
	if b.vm.MemoryInfo.MemoryFree > 0 {
		row[14] = fmt.Sprintf("%.1f", b.vm.MemoryInfo.MemoryFree)
	}
	if b.vm.MemoryInfo.TotalMemoryUsed > 0 {
		row[15] = fmt.Sprintf("%.1f", b.vm.MemoryInfo.TotalMemoryUsed)
	}
	if b.vm.MemoryInfo.MemoryUsedPercentage > 0 {
		row[16] = fmt.Sprintf("%.1f", b.vm.MemoryInfo.MemoryUsedPercentage)
	}
	if b.vm.MemoryInfo.MemoryUsedByLibVirt > 0 {
		row[17] = fmt.Sprintf("%.1f", b.vm.MemoryInfo.MemoryUsedByLibVirt)
	}
	if b.vm.MemoryInfo.MemoryUsedByVMI > 0 {
		row[18] = fmt.Sprintf("%.1f", b.vm.MemoryInfo.MemoryUsedByVMI)
	}
	if b.vm.MemoryInfo.MemoryHotPlugMax > 0 {
		row[19] = fmt.Sprintf("%.1f", b.vm.MemoryInfo.MemoryHotPlugMax)
	}
}

// populateStorageFields fills storage information
func (b *VMCSVRowBuilder) populateStorageFields(row []string) {
	row[20] = fmt.Sprintf("%d", len(b.vm.Disks))
}

// populateNetworkFields -- removed, network data now in Network Interfaces sheet
func (b *VMCSVRowBuilder) populateNetworkFields(row []string) {
	// Network data moved to dedicated Network Interfaces sheet
}

// populateMetadataFields fills metadata information
func (b *VMCSVRowBuilder) populateMetadataFields(row []string) {
	// UID: Use VMI UID if running, otherwise VM UID
	uid := b.vm.UID
	if b.vm.Runtime != nil && b.vm.Runtime.VMIUID != "" {
		uid = b.vm.Runtime.VMIUID
	}
	row[24] = uid

	// Created At: Use VMI creation time if running, otherwise VM creation time
	if b.vm.Runtime != nil && !b.vm.Runtime.CreationTimestamp.IsZero() {
		row[25] = b.vm.Runtime.CreationTimestamp.Time.Format("2006-01-02 15:04:05")
	} else if !b.vm.CreatedAt.IsZero() {
		row[25] = b.vm.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}

	// Creation Age in days
	if !b.vm.CreatedAt.IsZero() {
		row[26] = fmt.Sprintf("%d", int(time.Since(b.vm.CreatedAt.Time).Hours()/24))
	}

	row[27] = b.vm.MachineType
	row[29] = b.vm.RunStrategy
	row[30] = b.vm.EvictionStrategy
	row[31] = b.vm.InstanceType
	row[32] = b.vm.Preference
	row[33] = formatLabelsForCell(b.vm.Labels)
}

// VMConsolidatedCSVRowBuilder builds CSV rows for VMConsolidatedReport
type VMConsolidatedCSVRowBuilder struct {
	vm vm.VMConsolidatedReport
}

// NewVMConsolidatedCSVRowBuilder creates a builder for VMConsolidatedReport
func NewVMConsolidatedCSVRowBuilder(vmReport vm.VMConsolidatedReport) *VMConsolidatedCSVRowBuilder {
	return &VMConsolidatedCSVRowBuilder{vm: vmReport}
}

// BuildRow constructs a CSV row for VM runtime info output
func (b *VMConsolidatedCSVRowBuilder) BuildRow() []string {
	row := make([]string, 23)
	
	b.extractRuntimeFields(row)
	b.formatMemoryFields(row)
	b.formatCPUFields(row)
	b.formatUID(row)
	
	return row
}

// extractRuntimeFields extracts all runtime metadata
func (b *VMConsolidatedCSVRowBuilder) extractRuntimeFields(row []string) {
	osVersion := ""
	hostname := ""
	kernelVersion := ""
	guestAgentVersion := ""
	timezone := ""
	runningOnNode := ""
	virtLauncherPod := ""
	powerState := "Stopped"
	
	if b.vm.Runtime != nil {
		powerState = b.vm.Runtime.PowerState
		if b.vm.Runtime.GuestMetadata != nil {
			osVersion = b.vm.Runtime.GuestMetadata.OSVersion
			hostname = b.vm.Runtime.GuestMetadata.HostName
			kernelVersion = b.vm.Runtime.GuestMetadata.KernelVersion
			guestAgentVersion = b.vm.Runtime.GuestMetadata.GuestAgentVersion
			timezone = b.vm.Runtime.GuestMetadata.Timezone
			runningOnNode = b.vm.Runtime.GuestMetadata.RunningOnNode
			virtLauncherPod = b.vm.Runtime.GuestMetadata.VirtLauncherPod
		}
	}
	
	row[0] = b.vm.Name
	row[1] = b.vm.Namespace
	row[3] = powerState
	row[4] = runningOnNode
	row[5] = b.vm.OSName
	row[6] = osVersion
	row[7] = b.vm.MachineType
	row[8] = hostname
	row[9] = kernelVersion
	row[10] = guestAgentVersion
	row[11] = timezone
	row[12] = virtLauncherPod
}

// formatMemoryFields formats memory information
func (b *VMConsolidatedCSVRowBuilder) formatMemoryFields(row []string) {
	if b.vm.MemoryInfo.MemoryUsedByVMI > 0 {
		row[13] = fmt.Sprintf("%.1f", b.vm.MemoryInfo.MemoryUsedByVMI)
	}
	if b.vm.MemoryInfo.MemoryFree > 0 {
		row[14] = fmt.Sprintf("%.1f", b.vm.MemoryInfo.MemoryFree)
	}
	if b.vm.MemoryInfo.TotalMemoryUsed > 0 {
		row[15] = fmt.Sprintf("%.1f", b.vm.MemoryInfo.TotalMemoryUsed)
	}
	if b.vm.MemoryInfo.MemoryUsedByLibVirt > 0 {
		row[16] = fmt.Sprintf("%.1f", b.vm.MemoryInfo.MemoryUsedByLibVirt)
	}
	if b.vm.MemoryInfo.MemoryUsedPercentage > 0 {
		row[17] = fmt.Sprintf("%.1f", b.vm.MemoryInfo.MemoryUsedPercentage)
	}
	if b.vm.MemoryInfo.MemoryHotPlugMax > 0 {
		row[18] = fmt.Sprintf("%.1f", b.vm.MemoryInfo.MemoryHotPlugMax)
	}
}

// formatCPUFields formats CPU information
func (b *VMConsolidatedCSVRowBuilder) formatCPUFields(row []string) {
	row[19] = fmt.Sprintf("%d", b.vm.CPUInfo.CPUCores)
	row[20] = fmt.Sprintf("%d", b.vm.CPUInfo.CPUSockets)
	row[21] = fmt.Sprintf("%d", b.vm.CPUInfo.CPUThreads)
}

// formatUID formats the UID field
func (b *VMConsolidatedCSVRowBuilder) formatUID(row []string) {
	uid := b.vm.UID
	if b.vm.Runtime != nil && b.vm.Runtime.VMIUID != "" {
		uid = b.vm.Runtime.VMIUID
	}
	row[2] = uid
}
