package output

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"vm-scanner/pkg/client"
	"vm-scanner/pkg/cluster"
	"vm-scanner/pkg/hardware"
	"vm-scanner/pkg/utils"
	"vm-scanner/pkg/vm"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// TestConnection verifies the connection to the cluster and returns KubeVirt version info
func TestConnection(ctx context.Context, clusterClient *client.ClusterClient) (*cluster.KubeVirtVersion, error) {
	kubevirtVersion, err := cluster.GetKubeVirtVersion(ctx, clusterClient)
	if err != nil {
		fmt.Printf("❌ Connection test FAILED: %v\n", err)
		return nil, fmt.Errorf("connection test failed: %w", err)
	}

	fmt.Printf("✅ Connection test SUCCESSFUL\n")
	fmt.Printf("   Cluster is reachable\n")
	fmt.Printf("   KubeVirt Version: %s\n", kubevirtVersion.Version)
	fmt.Printf("   Status: %s\n\n", kubevirtVersion.Deployed)

	return kubevirtVersion, nil
}

// GenerateStorageVolumesReport collects storage volume information for all VMs
func GenerateStorageVolumesReport(ctx context.Context, clusterClient *client.ClusterClient) ([][]vm.StorageInfo, error) {
	vms, err := vm.GetAllVMs(ctx, clusterClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get VMs: %w", err)
	}

	allStorageVolumes := [][]vm.StorageInfo{}
	for _, vminfo := range vms.Items {
		storageVolumes, err := vm.GetVMStorageInfo(&vminfo, clusterClient, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get storage volumes for VM %s: %w", vminfo.GetName(), err)
		}
		if len(storageVolumes) > 0 {
			allStorageVolumes = append(allStorageVolumes, storageVolumes)
		}
	}

	return allStorageVolumes, nil
}

// clusterData holds raw data collected from the cluster
type clusterData struct {
	nodes          []cluster.ClusterNodeInfo
	storage        []hardware.StorageClassInfo
	allVMs         *unstructured.UnstructuredList
	vmInventoryMap map[string]vm.VMConsolidatedReport
	instanceTypes  map[string]vm.InstanceTypeSpecs
	pvcs           []cluster.PVCInventoryItem
	nads           []cluster.NADInfo
	dataVolumes    []cluster.DataVolumeInfo
	operators      []cluster.OperatorStatus
}

// vmCounts tracks running and stopped VM counts
type vmCounts struct {
	running int
	stopped int
}

// GenerateComprehensiveReport collects and aggregates all cluster data into a comprehensive report
func GenerateComprehensiveReport(ctx context.Context, clusterClient *client.ClusterClient) (*ComprehensiveReport, error) {
	report := &ComprehensiveReport{
		GeneratedAt: time.Now().Format(time.RFC3339),
		GeneratedBy: "vm-scanner",
	}

	data := collectClusterData(ctx, clusterClient)
	
	report.Storage = data.storage
	
	vms, counts := mergeVMsWithRuntime(ctx, clusterClient, data.allVMs, data.vmInventoryMap)
	report.VMs = vms
	report.VMs = resolveInstanceTypeSpecs(report.VMs, data.instanceTypes)

	report.PVCs = crossReferencePVCOwnership(data.pvcs, report.VMs)
	report.NADs = data.nads
	report.DataVolumes = data.dataVolumes

	clusterSummary, err := buildClusterSummary(ctx, clusterClient, data.nodes, vms, data.allVMs, counts, data.storage)
	if err != nil {
		return nil, err
	}
	
	report.Cluster = clusterSummary
	report.Cluster.Operators = data.operators
	report.Nodes = data.nodes
	report.Summary = &ReportSummary{
		ClusterSummary: clusterSummary,
		StorageClasses: len(data.storage),
	}

	return report, nil
}

// collectClusterData fetches raw data from cluster APIs
func collectClusterData(ctx context.Context, clusterClient *client.ClusterClient) *clusterData {
	data := &clusterData{}

	nodeHardwareInfo, err := cluster.GetClusterNodeInfo(ctx, clusterClient)
	if err != nil {
		log.Printf("Warning: Failed to get node hardware info: %v", err)
	}
	data.nodes = nodeHardwareInfo

	storageClasses, err := hardware.GetStorageClasses(ctx, clusterClient)
	if err != nil {
		log.Printf("Warning: Failed to get storage classes: %v", err)
	}
	data.storage = storageClasses

	allVMs, err := vm.GetAllVMs(ctx, clusterClient)
	if err != nil {
		log.Printf("Warning: Failed to get all VMs: %v", err)
	}
	data.allVMs = allVMs

	vmInventoryMap, err := vm.BuildVMIInventory(ctx, clusterClient)
	if err != nil {
		log.Printf("Warning: Failed to get VM runtime inventory: %v", err)
	}
	data.vmInventoryMap = vmInventoryMap

	instanceTypes, err := vm.BuildInstanceTypeMap(clusterClient.Dynamic)
	if err != nil {
		log.Printf("Warning: Failed to get instance types: %v", err)
	}
	data.instanceTypes = instanceTypes

	pvcs, err := cluster.GetPVCInventory(clusterClient.Dynamic)
	if err != nil {
		log.Printf("Warning: Failed to get PVC inventory: %v", err)
	}
	data.pvcs = pvcs

	nads, err := cluster.GetNADInventory(clusterClient.Dynamic)
	if err != nil {
		log.Printf("Warning: Failed to get NAD inventory: %v", err)
	}
	data.nads = nads

	dataVolumes, err := cluster.GetDataVolumeInventory(clusterClient.Dynamic)
	if err != nil {
		log.Printf("Warning: Failed to get DataVolume inventory: %v", err)
	}
	data.dataVolumes = dataVolumes

	operators, err := cluster.GetOperatorStatus(clusterClient.Dynamic)
	if err != nil {
		log.Printf("Warning: Failed to get operator status: %v", err)
	}
	data.operators = operators

	return data
}

// mergeVMsWithRuntime combines VM definitions with their runtime data
func mergeVMsWithRuntime(ctx context.Context, clusterClient *client.ClusterClient, allVMs *unstructured.UnstructuredList, vmInventoryMap map[string]vm.VMConsolidatedReport) ([]vm.VMDetails, vmCounts) {
	counts := vmCounts{}
	
	if allVMs == nil {
		return nil, counts
	}

	vms := make([]vm.VMDetails, 0, len(allVMs.Items))

	for _, vmObj := range allVMs.Items {
		vmName := vmObj.GetName()
		vmNamespace := vmObj.GetNamespace()
		key := vm.VMKey(vmNamespace, vmName)

		phase := extractVMPhase(vmObj)
		disksMap := getVMStorageInfo(ctx, clusterClient, vmObj, vmNamespace, vmName)

		consolidatedReport, hasRuntime := vmInventoryMap[key]

		var vmDetail vm.VMDetails
		if hasRuntime {
			vmDetail = buildRunningVMDetail(consolidatedReport, phase, disksMap, vmObj)
			if consolidatedReport.Runtime != nil && consolidatedReport.Runtime.PowerState == "Running" {
				counts.running++
			}
		} else {
			vmDetail = buildStoppedVMDetail(ctx, clusterClient, vmObj, vmName, vmNamespace, phase, disksMap)
			counts.stopped++
		}

		vms = append(vms, vmDetail)
	}

	return vms, counts
}

// extractVMPhase gets the printable status from VM object
func extractVMPhase(vmObj unstructured.Unstructured) string {
	phase := "Unknown"
	if statusPhase, found, _ := unstructured.NestedString(vmObj.Object, "status", "printableStatus"); found {
		phase = statusPhase
	}
	return phase
}

// getVMStorageInfo retrieves and converts storage info to map
func getVMStorageInfo(ctx context.Context, clusterClient *client.ClusterClient, vmObj unstructured.Unstructured, vmNamespace, vmName string) map[string]vm.StorageInfo {
	storageInfo, err := vm.GetVMStorageInfo(&vmObj, clusterClient, ctx)
	if err != nil {
		log.Printf("Warning: Failed to get storage info for VM %s/%s: %v", vmNamespace, vmName, err)
	}

	disksMap := make(map[string]vm.StorageInfo)
	for _, disk := range storageInfo {
		disksMap[disk.VolumeName] = disk
	}
	return disksMap
}

// mergeGuestAgentDiskUsage merges VMDiskInfo usage data into StorageInfo
func mergeGuestAgentDiskUsage(disksMap map[string]vm.StorageInfo, diskInfo []vm.VMDiskInfo) map[string]vm.StorageInfo {
	if len(diskInfo) == 0 {
		return disksMap
	}

	// Calculate total used bytes across all filesystems
	var totalUsedBytes int64
	for _, disk := range diskInfo {
		totalUsedBytes += disk.UsedBytes
	}

	// Update each StorageInfo entry with the aggregated usage data
	// Note: We sum all filesystem usage and apply it proportionally or to the first disk
	// since we can't reliably match individual filesystems to PVCs
	for volumeName, storage := range disksMap {
		storage.TotalStorageInUse = totalUsedBytes
		
		// Calculate human-readable format (GiB with 2 decimals)
		usedGiB := float64(totalUsedBytes) / (1024 * 1024 * 1024)
		storage.TotalStorageInUseHuman = fmt.Sprintf("%.2f", usedGiB)
		
		// Calculate percentage if we have total storage
		if storage.TotalStorage > 0 {
			storage.TotalStorageInUsePercentage = (float64(totalUsedBytes) / float64(storage.TotalStorage)) * 100
		}
		
		disksMap[volumeName] = storage
	}

	return disksMap
}

// buildRunningVMDetail creates VMDetails for a running VM
func buildRunningVMDetail(consolidatedReport vm.VMConsolidatedReport, phase string, disksMap map[string]vm.StorageInfo, vmObj unstructured.Unstructured) vm.VMDetails {
	// Merge guest agent disk usage data if available
	if consolidatedReport.Runtime != nil &&
		consolidatedReport.Runtime.GuestMetadata != nil &&
		len(consolidatedReport.Runtime.GuestMetadata.DiskInfo) > 0 {
		disksMap = mergeGuestAgentDiskUsage(disksMap, consolidatedReport.Runtime.GuestMetadata.DiskInfo)
	}

	vmDetail := vm.VMDetails{
		VMBaseInfo: consolidatedReport.VMBaseInfo,
		Runtime:    consolidatedReport.Runtime,
	}

	// Enrich with VM CR metadata that the VMI cannot provide
	enrichBaseInfoFromVM(&vmDetail.VMBaseInfo, vmObj)

	vmDetail.Phase = phase
	vmDetail.Disks = disksMap
	return vmDetail
}

// enrichBaseInfoFromVM overlays VM CR spec metadata onto VMBaseInfo.
// The VMI-built base info has empty RunStrategy/EvictionStrategy/InstanceType/Preference
// because those live on the VM CR spec, not the VMI. Labels from the VM CR are also
// preferred over VMI labels since they are more meaningful for filtering.
func enrichBaseInfoFromVM(baseInfo *vm.VMBaseInfo, vmObj unstructured.Unstructured) {
	if vmObj.Object == nil {
		return
	}

	// Run strategy -- empty string if not set
	if runStrategy, found, _ := unstructured.NestedString(vmObj.Object, "spec", "runStrategy"); found {
		baseInfo.RunStrategy = runStrategy
	}

	// Eviction strategy -- empty string if not set
	if evictionStrategy, found, _ := unstructured.NestedString(vmObj.Object, "spec", "evictionStrategy"); found {
		baseInfo.EvictionStrategy = evictionStrategy
	}

	// Instance type -- capture name for both cluster and namespaced kinds
	if instancetypeKind, found, _ := unstructured.NestedString(vmObj.Object, "spec", "instancetype", "kind"); found {
		if strings.EqualFold(instancetypeKind, "virtualmachineclusterinstancetype") ||
			strings.EqualFold(instancetypeKind, "virtualmachineinstancetype") {
			if instancetypeName, _, _ := unstructured.NestedString(vmObj.Object, "spec", "instancetype", "name"); instancetypeName != "" {
				baseInfo.InstanceType = instancetypeName
			}
		}
	}

	// Preference -- capture for both cluster and namespaced kinds
	if preferenceKind, found, _ := unstructured.NestedString(vmObj.Object, "spec", "preference", "kind"); found {
		if strings.EqualFold(preferenceKind, "virtualmachineclusterpreference") ||
			strings.EqualFold(preferenceKind, "virtualmachinepreference") {
			if preferenceName, _, _ := unstructured.NestedString(vmObj.Object, "spec", "preference", "name"); preferenceName != "" {
				baseInfo.Preference = preferenceName
			}
		}
	}

	// VM CR labels are more useful than VMI labels (pod-level) -- use VM labels
	if vmLabels := vmObj.GetLabels(); len(vmLabels) > 0 {
		baseInfo.Labels = vmLabels
	}

	// Annotations from VM CR
	if vmAnnotations := vmObj.GetAnnotations(); len(vmAnnotations) > 0 {
		baseInfo.Annotations = vmAnnotations
	}
}

// resolveInstanceTypeSpecs fills in CPU/memory for VMs using instance type references.
func resolveInstanceTypeSpecs(vms []vm.VMDetails, instanceTypes map[string]vm.InstanceTypeSpecs) []vm.VMDetails {
	if len(instanceTypes) == 0 {
		return vms
	}
	for i := range vms {
		if vms[i].InstanceType == "" || vms[i].CPUInfo.VCPUs != 0 {
			continue
		}
		specs, found := instanceTypes[vms[i].InstanceType]
		if !found {
			specs, found = instanceTypes[vms[i].Namespace+"/"+vms[i].InstanceType]
		}
		if !found {
			continue
		}
		vms[i].CPUInfo.VCPUs = specs.CPUGuest
		vms[i].CPUInfo.CPUCores = specs.CPUGuest
		vms[i].CPUInfo.CPUSockets = 1
		vms[i].CPUInfo.CPUThreads = 1
		memMiB, err := utils.QuantityToMiB(specs.MemoryGuest)
		if err != nil {
			log.Printf("Warning: failed to parse memory %q for instance type %q: %v", specs.MemoryGuest, vms[i].InstanceType, err)
			continue
		}
		vms[i].MemoryInfo.MemoryConfiguredMiB = memMiB
	}
	return vms
}

// crossReferencePVCOwnership matches PVCs to their owning VMs by checking disk volume names.
func crossReferencePVCOwnership(pvcs []cluster.PVCInventoryItem, vms []vm.VMDetails) []cluster.PVCInventoryItem {
	if len(pvcs) == 0 {
		return pvcs
	}
	vmDiskIndex := make(map[string]vm.VMDetails)
	for _, vmDetail := range vms {
		for diskName := range vmDetail.Disks {
			key := vmDetail.Namespace + "/" + diskName
			vmDiskIndex[key] = vmDetail
		}
	}
	for i := range pvcs {
		if pvcs[i].OwningVM != "" {
			continue
		}
		key := pvcs[i].Namespace + "/" + pvcs[i].Name
		if vmDetail, found := vmDiskIndex[key]; found {
			pvcs[i].OwningVM = vmDetail.Name
			pvcs[i].OwningVMNamespace = vmDetail.Namespace
		}
	}
	return pvcs
}

// buildStoppedVMDetail creates VMDetails for a stopped VM
func buildStoppedVMDetail(ctx context.Context, clusterClient *client.ClusterClient, vmObj unstructured.Unstructured, vmName, vmNamespace, phase string, disksMap map[string]vm.StorageInfo) vm.VMDetails {
	baseInfo, err := vm.BuildVMBaseInfoFromVM(ctx, clusterClient, vmObj)
	if err != nil {
		log.Printf("Warning: Failed to build base info for stopped VM %s/%s: %v", vmNamespace, vmName, err)
		baseInfo = &vm.VMBaseInfo{
			Name:      vmName,
			Namespace: vmNamespace,
			UID:       string(vmObj.GetUID()),
			Running:   false,
		}
	}
	baseInfo.Phase = phase
	baseInfo.Disks = disksMap
	
	return vm.VMDetails{
		VMBaseInfo: *baseInfo,
		Runtime:    nil,
	}
}

// buildClusterSummary calculates cluster-wide statistics
func buildClusterSummary(ctx context.Context, clusterClient *client.ClusterClient, nodes []cluster.ClusterNodeInfo, vms []vm.VMDetails, allVMs *unstructured.UnstructuredList, counts vmCounts, storageClasses []hardware.StorageClassInfo) (*cluster.ClusterSummary, error) {
	clusterResources, err := cluster.CalculateClusterResources(nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate cluster resources: %w", err)
	}

	requestedGiB, usedGiB := cluster.CalculateVMStorageTotals(vms)
	clusterResources.TotalApplicationRequestedStorage = requestedGiB
	clusterResources.TotalApplicationUsedStorage = usedGiB

	clusterInfo, err := cluster.GetClusterSummary(ctx, clusterClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster info: %w", err)
	}

	if allVMs != nil {
		clusterInfo.TotalVMs = len(allVMs.Items)
		clusterInfo.RunningVMs = counts.running
		clusterInfo.StoppedVMs = counts.stopped
	}
	clusterInfo.Resources = clusterResources

	return clusterInfo, nil
}

// assessMigrationReadiness evaluates migration readiness for each VM using PVC inventory cross-reference.
func assessMigrationReadiness(vms []vm.VMDetails, pvcs []PVCInventoryItem) []MigrationAssessment {
	pvcMap := make(map[string]PVCInventoryItem, len(pvcs))
	for _, p := range pvcs {
		pvcMap[p.Namespace+"/"+p.Name] = p
	}
	out := make([]MigrationAssessment, 0, len(vms))
	for i := range vms {
		out = append(out, assessSingleVM(vms[i], pvcMap))
	}
	return out
}

// assessSingleVM computes migration scoring and blockers for one VM (see MigrationAssessment).
func assessSingleVM(vmDetail vm.VMDetails, pvcMap map[string]PVCInventoryItem) MigrationAssessment {
	return buildMigrationAssessment(vmDetail, pvcMap)
}

func buildMigrationAssessment(vmDetail vm.VMDetails, pvcMap map[string]PVCInventoryItem) MigrationAssessment {
	a := MigrationAssessment{
		VMName:           vmDetail.Name,
		VMNamespace:      vmDetail.Namespace,
		RunStrategy:      vmDetail.RunStrategy,
		EvictionStrategy: vmDetail.EvictionStrategy,
		LiveMigratable:   "Unknown",
	}
	if vmDetail.Runtime != nil {
		a.PowerState = vmDetail.Runtime.PowerState
	} else {
		a.PowerState = vmDetail.Phase
	}

	ann := vmDetail.Annotations
	if ann == nil {
		ann = map[string]string{}
	}
	lab := vmDetail.Labels
	if lab == nil {
		lab = map[string]string{}
	}

	score := 0
	var blockers []string

	// 1. Power state
	if vmDetail.Runtime != nil && vmDetail.Runtime.PowerState == "Running" {
		score++
	} else {
		blockers = append(blockers, "VM is not running")
	}

	// 2. LiveMigratable (VMI conditions not stored on VMRuntimeInfo — unknown, no blocker, counts as pass)
	score++

	// 3. Run strategy
	if vmDetail.RunStrategy == "Always" || vmDetail.RunStrategy == "RerunOnFailure" {
		score++
	} else {
		blockers = append(blockers, "Run strategy does not support auto-restart")
	}

	// 4. Eviction strategy allows migration
	ev := vmDetail.EvictionStrategy
	if ev == "LiveMigrate" || ev == "" {
		score++
	} else {
		blockers = append(blockers, "Eviction strategy blocks live migration")
	}

	// 5. Host devices (annotation/label proxies)
	if migrationHasHostDeviceIndicators(ann, lab) {
		a.HasHostDevices = "Yes"
		blockers = append(blockers, "VM has host device passthrough")
	} else {
		a.HasHostDevices = "No"
		score++
	}

	// 6. Node affinity (annotation/label proxies)
	if migrationHasNodeAffinityIndicators(ann, lab) {
		a.HasNodeAffinity = "Yes"
		blockers = append(blockers, "VM has node affinity constraints")
	} else {
		a.HasNodeAffinity = "No"
		score++
	}

	// 7. PVC access modes (RWO blocks migration)
	pvcIssue := migrationPVCAccessModeIssue(vmDetail, pvcMap)
	a.PVCAccessModeIssue = pvcIssue
	if pvcIssue == "No" {
		score++
	} else {
		blockers = append(blockers, "PVC uses ReadWriteOnce (blocks live migration)")
	}

	// 8. Dedicated CPU
	if _, ok := ann["kubevirt.io/dedicated-cpu"]; ok {
		a.HasDedicatedCPU = "Yes"
		blockers = append(blockers, "VM uses dedicated CPU pinning")
	} else if _, ok := lab["kubevirt.io/dedicated-cpu"]; ok {
		a.HasDedicatedCPU = "Yes"
		blockers = append(blockers, "VM uses dedicated CPU pinning")
	} else {
		a.HasDedicatedCPU = "No"
		score++
	}

	// 9. Guest agent
	if vmDetail.Runtime != nil && vmDetail.Runtime.GuestMetadata != nil {
		a.GuestAgentReady = "Yes"
		score++
	} else {
		a.GuestAgentReady = "No"
		blockers = append(blockers, "No guest agent installed")
	}

	// 10. Explicit eviction strategy set
	if vmDetail.EvictionStrategy != "" {
		score++
	} else {
		blockers = append(blockers, "No explicit eviction strategy")
	}

	a.Blockers = blockers
	a.ReadinessScore = fmt.Sprintf("%d/10", score)
	return a
}

func migrationHasHostDeviceIndicators(ann, lab map[string]string) bool {
	if _, ok := ann["kubevirt.io/host-model-cpu"]; ok {
		return true
	}
	for k := range lab {
		lk := strings.ToLower(k)
		if strings.Contains(lk, "host-device") || strings.Contains(lk, "gpu") || strings.Contains(lk, "vfio") {
			return true
		}
	}
	return false
}

func migrationHasNodeAffinityIndicators(ann, lab map[string]string) bool {
	for k, v := range ann {
		ku, vu := strings.ToLower(k), strings.ToLower(v)
		if strings.Contains(ku, "nodeplacement") || strings.Contains(vu, "nodeplacement") ||
			strings.Contains(ku, "nodeselector") || strings.Contains(vu, "nodeaffinity") ||
			strings.Contains(vu, "requiredscheduling") {
			return true
		}
	}
	for k := range lab {
		if strings.Contains(strings.ToLower(k), "nodeplacement") {
			return true
		}
	}
	return false
}

func migrationPVCAccessModeIssue(vmDetail vm.VMDetails, pvcMap map[string]PVCInventoryItem) string {
	if len(vmDetail.Disks) == 0 {
		return "No"
	}
	for _, d := range vmDetail.Disks {
		if d.VolumeType != "pvc" && d.VolumeType != "dataVolumeTemplate" {
			continue
		}
		pvc, ok := pvcMap[vmDetail.Namespace+"/"+d.VolumeName]
		if !ok {
			continue
		}
		for _, mode := range pvc.AccessModes {
			if mode == "ReadWriteOnce" {
				return "Yes"
			}
		}
	}
	return "No"
}

func computeStorageAnalysis(pvcs []PVCInventoryItem, vms []vm.VMDetails) []StorageAnalysisRow {
	vmByKey := make(map[string]vm.VMDetails, len(vms))
	for i := range vms {
		v := vms[i]
		vmByKey[v.Namespace+"/"+v.Name] = v
	}
	out := make([]StorageAnalysisRow, 0, len(pvcs))
	for _, pvc := range pvcs {
		out = append(out, storageAnalysisRowFromPVC(pvc, vmByKey))
	}
	return out
}

func storageAnalysisRowFromPVC(pvc PVCInventoryItem, vmByKey map[string]vm.VMDetails) StorageAnalysisRow {
	row := StorageAnalysisRow{
		PVCName:      pvc.Name,
		Namespace:    pvc.Namespace,
		StorageClass: pvc.StorageClass,
		CapacityGiB:  float64(pvc.Capacity) / (1024 * 1024 * 1024),
		AccessModes:  strings.Join(pvc.AccessModes, ", "),
		VolumeMode:   pvc.VolumeMode,
		Status:       pvc.Status,
	}
	if pvc.OwningVM != "" {
		row.OwningVM = pvc.OwningVM
		key := pvc.OwningVMNamespace + "/" + pvc.OwningVM
		if vmDetail, ok := vmByKey[key]; ok {
			row.VMPowerState = vmPowerStateForStorageAnalysis(vmDetail)
			if vmRunningForGuestStorage(vmDetail) {
				if d, ok := vmDetail.Disks[pvc.Name]; ok {
					row.GuestUsedGiB = float64(d.TotalStorageInUse) / (1024 * 1024 * 1024)
				}
			}
		}
	}
	if row.CapacityGiB > 0 {
		row.UtilizationPct = (row.GuestUsedGiB / row.CapacityGiB) * 100
	}
	row.Flag = storageAnalysisComputeFlag(pvc, row)
	return row
}

func vmRunningForGuestStorage(vm vm.VMDetails) bool {
	if vm.Runtime != nil && vm.Runtime.PowerState == "Running" {
		return true
	}
	return vm.Running
}

func vmPowerStateForStorageAnalysis(vm vm.VMDetails) string {
	if vm.Runtime != nil && vm.Runtime.PowerState != "" {
		return vm.Runtime.PowerState
	}
	if vm.Running {
		return "Running"
	}
	return vm.Phase
}

func storageAnalysisComputeFlag(pvc PVCInventoryItem, row StorageAnalysisRow) string {
	if pvc.OwningVM == "" {
		return "Orphaned"
	}
	if row.GuestUsedGiB <= 0 {
		return ""
	}
	if row.UtilizationPct < 20 {
		return "Overprovisioned"
	}
	if row.UtilizationPct < 50 {
		return "Low Utilization"
	}
	return ""
}
