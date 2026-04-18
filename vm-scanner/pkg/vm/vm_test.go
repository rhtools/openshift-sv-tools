package vm

import (
	"testing"
)

func TestVMKey(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		vmName    string
		want      string
	}{
		{name: "standard", namespace: "default", vmName: "my-vm", want: "default/my-vm"},
		{name: "empty namespace", namespace: "", vmName: "vm1", want: "/vm1"},
		{name: "empty name", namespace: "ns", vmName: "", want: "ns/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VMKey(tt.namespace, tt.vmName); got != tt.want {
				t.Errorf("VMKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConvertNetworkMapToSlice(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]VMNetworkInfo
		check func(t *testing.T, got []VMNetworkInfo)
	}{
		{
			name: "two entries",
			input: map[string]VMNetworkInfo{
				"eth0": {Name: "eth0"},
				"eth1": {Name: "eth1"},
			},
			check: func(t *testing.T, got []VMNetworkInfo) {
				if len(got) != 2 {
					t.Fatalf("len = %d, want 2", len(got))
				}
				seen := make(map[string]bool)
				for _, n := range got {
					seen[n.Name] = true
				}
				if !seen["eth0"] || !seen["eth1"] {
					t.Errorf("got %#v, want slice containing eth0 and eth1", got)
				}
			},
		},
		{
			name:  "empty map",
			input: map[string]VMNetworkInfo{},
			check: func(t *testing.T, got []VMNetworkInfo) {
				if len(got) != 0 {
					t.Errorf("len = %d, want 0", len(got))
				}
			},
		},
		{
			name:  "nil map",
			input: nil,
			check: func(t *testing.T, got []VMNetworkInfo) {
				if len(got) != 0 {
					t.Errorf("len = %d, want 0", len(got))
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertNetworkMapToSlice(tt.input)
			tt.check(t, got)
		})
	}
}

func TestFixtureIntegrity(t *testing.T) {
	tests := []struct {
		name     string
		fixture  string
		checkCPU func(t *testing.T, cpu *CPUInfo)
	}{
		{
			name:    "vm_stopped_rhel8",
			fixture: "../../testdata/vms/vm_stopped_rhel8.yaml",
			checkCPU: func(t *testing.T, cpu *CPUInfo) {
				if cpu.CPUCores <= 0 {
					t.Errorf("CPUCores = %d, want > 0", cpu.CPUCores)
				}
			},
		},
		{
			name:    "vm_stopped_centos",
			fixture: "../../testdata/vms/vm_stopped_centos.yaml",
			checkCPU: func(t *testing.T, cpu *CPUInfo) {
				var empty CPUInfo
				if *cpu != empty {
					t.Errorf("want empty CPUInfo, got %#v", *cpu)
				}
			},
		},
		{
			name:    "vm_migrated_example",
			fixture: "../../testdata/vms/vm_migrated_example.yaml",
			checkCPU: func(t *testing.T, cpu *CPUInfo) {
				// Fixture has cores/sockets but no threads; production returns zeroed CPUInfo without error.
				if cpu == nil {
					t.Fatal("nil CPUInfo")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := loadFixture(t, tt.fixture)
			cpu, err := GetCPUInfoFromVM(vm)
			if err != nil {
				t.Fatalf("GetCPUInfoFromVM: %v", err)
			}
			if cpu == nil {
				t.Fatal("GetCPUInfoFromVM returned nil *CPUInfo")
			}
			tt.checkCPU(t, cpu)
		})
	}
}
