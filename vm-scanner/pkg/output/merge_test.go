package output

import (
	"math"
	"os"
	"testing"

	"vm-scanner/pkg/vm"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

const floatEpsilon = 1e-9

func loadTestFixture(t *testing.T, path string) unstructured.Unstructured {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", path, err)
	}
	var obj unstructured.Unstructured
	if err := yaml.Unmarshal(data, &obj); err != nil {
		t.Fatalf("failed to unmarshal fixture %s: %v", path, err)
	}
	return obj
}

func TestExtractVMPhase(t *testing.T) {
	tests := []struct {
		name     string
		vmObj    unstructured.Unstructured
		wantPhase string
	}{
		{
			name: "with printableStatus",
			vmObj: unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"printableStatus": "Running",
					},
				},
			},
			wantPhase: "Running",
		},
		{
			name: "without status",
			vmObj: unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			wantPhase: "Unknown",
		},
		{
			name: "with Stopped status",
			vmObj: unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"printableStatus": "Stopped",
					},
				},
			},
			wantPhase: "Stopped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractVMPhase(tt.vmObj)
			if got != tt.wantPhase {
				t.Fatalf("extractVMPhase() = %q, want %q", got, tt.wantPhase)
			}
		})
	}
}

func TestMergeGuestAgentDiskUsage(t *testing.T) {
	tests := []struct {
		name       string
		disksMap   map[string]vm.StorageInfo
		diskInfo   []vm.VMDiskInfo
		assertFunc func(t *testing.T, got map[string]vm.StorageInfo)
	}{
		{
			name: "empty diskInfo",
			disksMap: map[string]vm.StorageInfo{
				"disk-0": {TotalStorage: 10737418240},
			},
			diskInfo: nil,
			assertFunc: func(t *testing.T, got map[string]vm.StorageInfo) {
				t.Helper()
				d := got["disk-0"]
				if d.TotalStorageInUse != 0 {
					t.Fatalf("TotalStorageInUse = %d, want 0", d.TotalStorageInUse)
				}
				if d.TotalStorage != 10737418240 {
					t.Fatalf("TotalStorage = %d, want 10737418240", d.TotalStorage)
				}
			},
		},
		{
			name: "single disk usage",
			disksMap: map[string]vm.StorageInfo{
				"disk-0": {TotalStorage: 10737418240},
			},
			diskInfo: []vm.VMDiskInfo{{UsedBytes: 1073741824}},
			assertFunc: func(t *testing.T, got map[string]vm.StorageInfo) {
				t.Helper()
				d := got["disk-0"]
				if d.TotalStorageInUse != 1073741824 {
					t.Fatalf("TotalStorageInUse = %d, want 1073741824", d.TotalStorageInUse)
				}
				if math.Abs(d.TotalStorageInUsePercentage-10.0) > floatEpsilon {
					t.Fatalf("TotalStorageInUsePercentage = %v, want 10.0", d.TotalStorageInUsePercentage)
				}
			},
		},
		{
			name: "multiple disk usage summed",
			disksMap: map[string]vm.StorageInfo{
				"disk-0": {TotalStorage: 21474836480},
			},
			diskInfo: []vm.VMDiskInfo{
				{UsedBytes: 1073741824},
				{UsedBytes: 2147483648},
			},
			assertFunc: func(t *testing.T, got map[string]vm.StorageInfo) {
				t.Helper()
				d := got["disk-0"]
				if d.TotalStorageInUse != 3221225472 {
					t.Fatalf("TotalStorageInUse = %d, want 3221225472", d.TotalStorageInUse)
				}
				if math.Abs(d.TotalStorageInUsePercentage-15.0) > floatEpsilon {
					t.Fatalf("TotalStorageInUsePercentage = %v, want 15.0", d.TotalStorageInUsePercentage)
				}
			},
		},
		{
			name: "zero total storage no divide by zero",
			disksMap: map[string]vm.StorageInfo{
				"disk-0": {TotalStorage: 0},
			},
			diskInfo: []vm.VMDiskInfo{{UsedBytes: 100}},
			assertFunc: func(t *testing.T, got map[string]vm.StorageInfo) {
				t.Helper()
				d := got["disk-0"]
				if d.TotalStorageInUse != 100 {
					t.Fatalf("TotalStorageInUse = %d, want 100", d.TotalStorageInUse)
				}
				if math.Abs(d.TotalStorageInUsePercentage-0.0) > floatEpsilon {
					t.Fatalf("TotalStorageInUsePercentage = %v, want 0.0", d.TotalStorageInUsePercentage)
				}
			},
		},
		{
			name:     "empty disksMap + non-empty diskInfo",
			disksMap: map[string]vm.StorageInfo{},
			diskInfo: []vm.VMDiskInfo{{UsedBytes: 100}},
			assertFunc: func(t *testing.T, got map[string]vm.StorageInfo) {
				t.Helper()
				if len(got) != 0 {
					t.Fatalf("len(got) = %d, want 0", len(got))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeGuestAgentDiskUsage(tt.disksMap, tt.diskInfo)
			tt.assertFunc(t, got)
		})
	}
}

func TestEnrichBaseInfoFromVM(t *testing.T) {
	tests := []struct {
		name       string
		baseInfo   *vm.VMBaseInfo
		vmObj      unstructured.Unstructured
		assertFunc func(t *testing.T, bi *vm.VMBaseInfo)
	}{
		{
			name:     "centos VM with instancetype",
			baseInfo: &vm.VMBaseInfo{},
			vmObj:    loadTestFixture(t, "../../testdata/vms/vm_stopped_centos.yaml"),
			assertFunc: func(t *testing.T, bi *vm.VMBaseInfo) {
				t.Helper()
				if bi.InstanceType != "u1.xlarge" {
					t.Fatalf("InstanceType = %q, want u1.xlarge", bi.InstanceType)
				}
				if bi.Preference != "centos.stream9" {
					t.Fatalf("Preference = %q, want centos.stream9", bi.Preference)
				}
			},
		},
		{
			name:     "rhel8 VM without instancetype",
			baseInfo: &vm.VMBaseInfo{},
			vmObj:    loadTestFixture(t, "../../testdata/vms/vm_stopped_rhel8.yaml"),
			assertFunc: func(t *testing.T, bi *vm.VMBaseInfo) {
				t.Helper()
				if bi.InstanceType != "" {
					t.Fatalf("InstanceType = %q, want empty", bi.InstanceType)
				}
				if bi.RunStrategy != "RerunOnFailure" {
					t.Fatalf("RunStrategy = %q, want RerunOnFailure", bi.RunStrategy)
				}
				if bi.EvictionStrategy != "" {
					t.Fatalf("EvictionStrategy = %q, want empty (not in fixture)", bi.EvictionStrategy)
				}
			},
		},
		{
			name:     "nil vmObj.Object",
			baseInfo: &vm.VMBaseInfo{RunStrategy: "unchanged"},
			vmObj:    unstructured.Unstructured{Object: nil},
			assertFunc: func(t *testing.T, bi *vm.VMBaseInfo) {
				t.Helper()
				if bi.RunStrategy != "unchanged" {
					t.Fatalf("RunStrategy was modified: %q", bi.RunStrategy)
				}
			},
		},
		{
			name:     "wrong instancetype kind",
			baseInfo: &vm.VMBaseInfo{},
			vmObj: unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"instancetype": map[string]interface{}{
							"kind": "SomeInvalidKind",
							"name": "small",
						},
					},
				},
			},
			assertFunc: func(t *testing.T, bi *vm.VMBaseInfo) {
				t.Helper()
				if bi.InstanceType != "" {
					t.Fatalf("InstanceType = %q, want empty", bi.InstanceType)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enrichBaseInfoFromVM(tt.baseInfo, tt.vmObj)
			tt.assertFunc(t, tt.baseInfo)
		})
	}
}
