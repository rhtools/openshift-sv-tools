package cluster

import (
	"testing"

	"vm-scanner/pkg/hardware"
	"vm-scanner/pkg/vm"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetClusterTopology(t *testing.T) {
	tests := []struct {
		name                         string
		nodes                        []ClusterNodeInfo
		wantHasSchedulableCP         bool
		wantSchedulableCPCount       int
		wantWorkerNodes              int
	}{
		{
			name: "3 schedulable CP + 3 workers",
			nodes: append(
				repeatTopologyNodes(3, []string{"control-plane", "worker"}, "true"),
				repeatTopologyNodes(3, []string{"worker"}, "true")...,
			),
			wantHasSchedulableCP:   true,
			wantSchedulableCPCount: 3,
			wantWorkerNodes:        6,
		},
		{
			name: "3 non-schedulable CP + 3 workers",
			nodes: append(
				repeatTopologyNodes(3, []string{"control-plane"}, "false"),
				repeatTopologyNodes(3, []string{"worker"}, "true")...,
			),
			wantHasSchedulableCP:   false,
			wantSchedulableCPCount: 0,
			wantWorkerNodes:        0,
		},
		{
			name: "mixed: 1 schedulable CP, 2 non-schedulable CP, 2 workers",
			nodes: []ClusterNodeInfo{
				topologyNode([]string{"control-plane"}, "true"),
				topologyNode([]string{"control-plane"}, "false"),
				topologyNode([]string{"control-plane"}, "false"),
				topologyNode([]string{"worker"}, "true"),
				topologyNode([]string{"worker"}, "true"),
			},
			wantHasSchedulableCP:   true,
			wantSchedulableCPCount: 1,
			wantWorkerNodes:        2,
		},
		{
			name:                   "empty nodes",
			nodes:                  nil,
			wantHasSchedulableCP:   false,
			wantSchedulableCPCount: 0,
			wantWorkerNodes:        0,
		},
		{
			name: "node with both CP and worker roles, schedulable",
			nodes: []ClusterNodeInfo{
				topologyNode([]string{"control-plane", "worker"}, "true"),
			},
			wantHasSchedulableCP:   true,
			wantSchedulableCPCount: 1,
			wantWorkerNodes:        1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHas, gotCP, gotWorkers, err := GetClusterTopology(tt.nodes)
			if err != nil {
				t.Fatalf("GetClusterTopology() error = %v", err)
			}
			if gotHas != tt.wantHasSchedulableCP {
				t.Errorf("hasSchedulableControlPlane = %v, want %v", gotHas, tt.wantHasSchedulableCP)
			}
			if gotCP != tt.wantSchedulableCPCount {
				t.Errorf("schedulableControlPlaneCount = %d, want %d", gotCP, tt.wantSchedulableCPCount)
			}
			if gotWorkers != tt.wantWorkerNodes {
				t.Errorf("totalWorkerNodes = %d, want %d", gotWorkers, tt.wantWorkerNodes)
			}
		})
	}
}

func topologyNode(roles []string, schedulable string) ClusterNodeInfo {
	return ClusterNodeInfo{
		NodeRoles:       roles,
		NodeSchedulable: schedulable,
	}
}

func repeatTopologyNodes(n int, roles []string, schedulable string) []ClusterNodeInfo {
	out := make([]ClusterNodeInfo, n)
	for i := 0; i < n; i++ {
		out[i] = topologyNode(roles, schedulable)
	}
	return out
}

func TestCalculateClusterResources(t *testing.T) {
	tests := []struct {
		name    string
		nodes   []ClusterNodeInfo
		want    *ClusterResources
		wantErr bool
	}{
		{
			name: "two nodes",
			nodes: []ClusterNodeInfo{
				resourceNode(8, 64.0, 32.0, 500.0, 250.0),
				resourceNode(16, 128.0, 96.0, 1000.0, 750.0),
			},
			want: &ClusterResources{
				TotalCPU:                         24,
				TotalMemory:                      192.0,
				TotalLocalStorage:                1500.0,
				UsedMemory:                       128.0,
				TotalLocalStorageUsed:            1000.0,
				TotalApplicationRequestedStorage: 0,
				TotalApplicationUsedStorage:      0,
			},
		},
		{
			name:  "empty nodes",
			nodes: nil,
			want: &ClusterResources{
				TotalCPU:                         0,
				TotalMemory:                      0,
				TotalLocalStorage:                0,
				UsedMemory:                       0,
				TotalLocalStorageUsed:            0,
				TotalApplicationRequestedStorage: 0,
				TotalApplicationUsedStorage:      0,
			},
		},
		{
			name: "single node with fractional values",
			nodes: []ClusterNodeInfo{
				resourceNode(4, 31.75, 0, 249.33, 0),
			},
			want: &ClusterResources{
				TotalCPU:                         4,
				TotalMemory:                      31.8,
				TotalLocalStorage:                249.3,
				UsedMemory:                       0,
				TotalLocalStorageUsed:            0,
				TotalApplicationRequestedStorage: 0,
				TotalApplicationUsedStorage:      0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateClusterResources(tt.nodes)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CalculateClusterResources() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.TotalCPU != tt.want.TotalCPU {
				t.Errorf("TotalCPU = %d, want %d", got.TotalCPU, tt.want.TotalCPU)
			}
			if got.TotalMemory != tt.want.TotalMemory {
				t.Errorf("TotalMemory = %v, want %v", got.TotalMemory, tt.want.TotalMemory)
			}
			if got.TotalLocalStorage != tt.want.TotalLocalStorage {
				t.Errorf("TotalLocalStorage = %v, want %v", got.TotalLocalStorage, tt.want.TotalLocalStorage)
			}
			if got.UsedMemory != tt.want.UsedMemory {
				t.Errorf("UsedMemory = %v, want %v", got.UsedMemory, tt.want.UsedMemory)
			}
			if got.TotalLocalStorageUsed != tt.want.TotalLocalStorageUsed {
				t.Errorf("TotalLocalStorageUsed = %v, want %v", got.TotalLocalStorageUsed, tt.want.TotalLocalStorageUsed)
			}
			if got.TotalApplicationRequestedStorage != tt.want.TotalApplicationRequestedStorage {
				t.Errorf("TotalApplicationRequestedStorage = %d, want %d", got.TotalApplicationRequestedStorage, tt.want.TotalApplicationRequestedStorage)
			}
			if got.TotalApplicationUsedStorage != tt.want.TotalApplicationUsedStorage {
				t.Errorf("TotalApplicationUsedStorage = %d, want %d", got.TotalApplicationUsedStorage, tt.want.TotalApplicationUsedStorage)
			}
		})
	}
}

func resourceNode(cpu int64, memCapGiB, memUsedGiB, fsCapGiB, fsUsedGiB float64) ClusterNodeInfo {
	return ClusterNodeInfo{
		CPU: hardware.CPUInfo{CPUCores: cpu},
		Memory: hardware.MemoryInfo{
			MemoryCapacityGiB: memCapGiB,
			MemoryUsedGiB:     memUsedGiB,
		},
		Filesystem: hardware.NodeFilesystemInfo{
			FilesystemCapacity: fsCapGiB,
			FilesystemUsed:     fsUsedGiB,
		},
	}
}

func TestCalculateVMStorageTotals(t *testing.T) {
	const gib = 1024 * 1024 * 1024
	tests := []struct {
		name           string
		vms            []vm.VMDetails
		wantRequested  int64
		wantUsed       int64
	}{
		{
			name: "two VMs with disks",
			vms: []vm.VMDetails{
				{
					VMBaseInfo: vm.VMBaseInfo{
						Disks: map[string]vm.StorageInfo{
							"d1": {SizeBytes: 30 * gib, TotalStorageInUse: 10 * gib},
						},
					},
				},
				{
					VMBaseInfo: vm.VMBaseInfo{
						Disks: map[string]vm.StorageInfo{
							"d1": {SizeBytes: 10 * gib, TotalStorageInUse: 5 * gib},
						},
					},
				},
			},
			wantRequested: 40,
			wantUsed:      15,
		},
		{
			name: "VM with no disks",
			vms: []vm.VMDetails{
				{VMBaseInfo: vm.VMBaseInfo{Disks: map[string]vm.StorageInfo{}}},
			},
			wantRequested: 0,
			wantUsed:      0,
		},
		{
			name: "truncation",
			vms: []vm.VMDetails{
				{
					VMBaseInfo: vm.VMBaseInfo{
						Disks: map[string]vm.StorageInfo{
							"d1": {SizeBytes: 1073741823},
						},
					},
				},
			},
			wantRequested: 0,
			wantUsed:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReq, gotUsed := CalculateVMStorageTotals(tt.vms)
			if gotReq != tt.wantRequested {
				t.Errorf("requestedGiB = %d, want %d", gotReq, tt.wantRequested)
			}
			if gotUsed != tt.wantUsed {
				t.Errorf("usedGiB = %d, want %d", gotUsed, tt.wantUsed)
			}
		})
	}
}

func TestSortNamespaces(t *testing.T) {
	nsList := func(names ...string) []v1.Namespace {
		out := make([]v1.Namespace, len(names))
		for i, n := range names {
			out[i] = v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: n}}
		}
		return out
	}

	tests := []struct {
		name              string
		namespaces        []v1.Namespace
		wantProtected     int
		wantUser          int
		wantTotal         int
		wantErr           bool
	}{
		{
			name:          "mix of protected and user",
			namespaces:    nsList("openshift-cnv", "my-app", "kube-system", "production"),
			wantProtected: 2,
			wantUser:      2,
			wantTotal:     4,
		},
		{
			name:          "all protected",
			namespaces:    nsList("openshift-cnv", "default"),
			wantProtected: 2,
			wantUser:      0,
			wantTotal:     2,
		},
		{
			name:          "empty",
			namespaces:    nil,
			wantProtected: 0,
			wantUser:      0,
			wantTotal:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProt, gotUser, gotTotal, err := SortNamespaces(tt.namespaces)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SortNamespaces() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotProt != tt.wantProtected {
				t.Errorf("protected = %d, want %d", gotProt, tt.wantProtected)
			}
			if gotUser != tt.wantUser {
				t.Errorf("user = %d, want %d", gotUser, tt.wantUser)
			}
			if gotTotal != tt.wantTotal {
				t.Errorf("total = %d, want %d", gotTotal, tt.wantTotal)
			}
		})
	}
}
