package vm

import (
	"testing"
)

func TestGetCPUInfo(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		prefix []string
		want   CPUInfo
	}{
		{
			name:   "rhel8 VM with explicit CPU",
			path:   "../../testdata/vms/vm_stopped_rhel8.yaml",
			prefix: []string{"spec", "template", "spec"},
			want: CPUInfo{
				CPUCores:   1,
				CPUSockets: 2,
				CPUThreads: 1,
				VCPUs:      2,
				CPUModel:   "host-model",
			},
		},
		{
			name:   "centos VM with no CPU block",
			path:   "../../testdata/vms/vm_stopped_centos.yaml",
			prefix: []string{"spec", "template", "spec"},
			want:   CPUInfo{},
		},
		{
			name:   "migrated VM with partial CPU",
			path:   "../../testdata/vms/vm_migrated_example.yaml",
			prefix: []string{"spec", "template", "spec"},
			want:   CPUInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := loadFixture(t, tt.path)
			got, err := GetCPUInfo(u, tt.prefix...)
			if err != nil {
				t.Fatalf("GetCPUInfo: %v", err)
			}
			if got == nil {
				t.Fatal("got nil *CPUInfo")
			}
			assertCPUInfoFields(t, tt.want, got)
		})
	}
}

func assertCPUInfoFields(t *testing.T, want CPUInfo, got *CPUInfo) {
	t.Helper()
	if got.CPUCores != want.CPUCores {
		t.Errorf("CPUCores: got %d, want %d", got.CPUCores, want.CPUCores)
	}
	if got.CPUSockets != want.CPUSockets {
		t.Errorf("CPUSockets: got %d, want %d", got.CPUSockets, want.CPUSockets)
	}
	if got.CPUThreads != want.CPUThreads {
		t.Errorf("CPUThreads: got %d, want %d", got.CPUThreads, want.CPUThreads)
	}
	if got.VCPUs != want.VCPUs {
		t.Errorf("VCPUs: got %d, want %d", got.VCPUs, want.VCPUs)
	}
	if got.CPUModel != want.CPUModel {
		t.Errorf("CPUModel: got %q, want %q", got.CPUModel, want.CPUModel)
	}
}

func TestGetCPUInfoFromVM(t *testing.T) {
	u := loadFixture(t, "../../testdata/vms/vm_stopped_rhel8.yaml")
	got, err := GetCPUInfoFromVM(u)
	if err != nil {
		t.Fatalf("GetCPUInfoFromVM: %v", err)
	}
	want := CPUInfo{
		CPUCores:   1,
		CPUSockets: 2,
		CPUThreads: 1,
		VCPUs:      2,
		CPUModel:   "host-model",
	}
	assertCPUInfoFields(t, want, got)
}

func TestGetCPUInfoFromVMI(t *testing.T) {
	u := loadFixture(t, "../../testdata/vmis/vmi_running_example.yaml")
	got, err := GetCPUInfoFromVMI(u)
	if err != nil {
		t.Fatalf("GetCPUInfoFromVMI: %v", err)
	}
	want := CPUInfo{
		CPUCores:   2,
		CPUSockets: 1,
		CPUThreads: 1,
		VCPUs:      2,
		CPUModel:   "host-model",
	}
	assertCPUInfoFields(t, want, got)
}
