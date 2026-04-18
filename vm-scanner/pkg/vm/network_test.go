package vm

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// --- interfaceTypeFromMap tests ---

func TestInterfaceTypeFromMap(t *testing.T) {
	tests := []struct {
		name  string
		iface map[string]interface{}
		want  string
	}{
		{"bridge", map[string]interface{}{"bridge": map[string]interface{}{}, "name": "eth0"}, "bridge"},
		{"masquerade", map[string]interface{}{"masquerade": map[string]interface{}{}, "name": "eth0"}, "masquerade"},
		{"sriov", map[string]interface{}{"sriov": map[string]interface{}{}, "name": "eth0"}, "sriov"},
		{"slirp", map[string]interface{}{"slirp": map[string]interface{}{}, "name": "eth0"}, "slirp"},
		{"no type", map[string]interface{}{"name": "eth0"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := interfaceTypeFromMap(tt.iface); got != tt.want {
				t.Errorf("interfaceTypeFromMap() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- modelFromInterface tests ---

func TestModelFromInterface(t *testing.T) {
	tests := []struct {
		name  string
		iface map[string]interface{}
		want  string
	}{
		{"plain string", map[string]interface{}{"model": "virtio", "name": "eth0"}, "virtio"},
		{"nested object", map[string]interface{}{
			"model": map[string]interface{}{"type": "e1000e"},
			"name":  "eth0",
		}, "e1000e"},
		{"empty", map[string]interface{}{"name": "eth0"}, ""},
		{"nil model", map[string]interface{}{"model": nil, "name": "eth0"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := modelFromInterface(tt.iface); got != tt.want {
				t.Errorf("modelFromInterface() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- buildNetworkToNADMap tests ---

func TestBuildNetworkToNADMap(t *testing.T) {
	tests := []struct {
		name     string
		networks []interface{}
		assert   func(t *testing.T, result map[string]string)
	}{
		{
			name: "pod network",
			networks: []interface{}{
				map[string]interface{}{"name": "default", "pod": map[string]interface{}{}},
			},
			assert: func(t *testing.T, result map[string]string) {
				if nad := result["default"]; nad != "None" {
					t.Errorf("default = %q, want None", nad)
				}
			},
		},
		{
			name: "multus network",
			networks: []interface{}{
				map[string]interface{}{"name": "bridge-network", "multus": map[string]interface{}{"networkName": "my-nad"}},
			},
			assert: func(t *testing.T, result map[string]string) {
				if nad := result["bridge-network"]; nad != "my-nad" {
					t.Errorf("bridge-network = %q, want my-nad", nad)
				}
			},
		},
		{
			name: "mixed",
			networks: []interface{}{
				map[string]interface{}{"name": "pod-net", "pod": map[string]interface{}{}},
				map[string]interface{}{"name": "multus-net", "multus": map[string]interface{}{"networkName": "production-nad"}},
			},
			assert: func(t *testing.T, result map[string]string) {
				if nad := result["pod-net"]; nad != "None" {
					t.Errorf("pod-net = %q, want None", nad)
				}
				if nad := result["multus-net"]; nad != "production-nad" {
					t.Errorf("multus-net = %q, want production-nad", nad)
				}
			},
		},
		{
			name:     "empty networks",
			networks: []interface{}{},
			assert: func(t *testing.T, result map[string]string) {
				if len(result) != 0 {
					t.Errorf("len = %d, want 0", len(result))
				}
			},
		},
		{
			name: "non-map entry",
			networks: []interface{}{
				"not a map",
				map[string]interface{}{"name": "valid-net", "pod": map[string]interface{}{}},
			},
			assert: func(t *testing.T, result map[string]string) {
				if nad := result["valid-net"]; nad != "None" {
					t.Errorf("valid-net = %q, want None", nad)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildNetworkToNADMap(tt.networks)
			tt.assert(t, result)
		})
	}
}

// --- enrichInterfaceFromSpec tests ---

func TestEnrichInterfaceFromSpec(t *testing.T) {
	tests := []struct {
		name         string
		ifaceMap     map[string]interface{}
		networkToNAD map[string]string
		assert       func(t *testing.T, ifaceType, model, netName, nad string)
	}{
		{
			name: "bridge pod",
			ifaceMap: map[string]interface{}{
				"bridge": map[string]interface{}{},
				"name":   "eth0",
			},
			networkToNAD: map[string]string{"eth0": "None"},
			assert: func(t *testing.T, ifaceType, model, netName, nad string) {
				if ifaceType != "bridge" {
					t.Errorf("ifaceType = %q, want bridge", ifaceType)
				}
				if model != "" {
					t.Errorf("model = %q, want empty", model)
				}
				if netName != "eth0" {
					t.Errorf("netName = %q, want eth0", netName)
				}
				if nad != "None" {
					t.Errorf("nad = %q, want None", nad)
				}
			},
		},
		{
			name: "masquerade multus",
			ifaceMap: map[string]interface{}{
				"masquerade": map[string]interface{}{},
				"model":      "virtio",
				"network":    "primary",
			},
			networkToNAD: map[string]string{"primary": "production-nad"},
			assert: func(t *testing.T, ifaceType, model, netName, nad string) {
				if ifaceType != "masquerade" {
					t.Errorf("ifaceType = %q, want masquerade", ifaceType)
				}
				if model != "virtio" {
					t.Errorf("model = %q, want virtio", model)
				}
				if netName != "primary" {
					t.Errorf("netName = %q, want primary", netName)
				}
				if nad != "production-nad" {
					t.Errorf("nad = %q, want production-nad", nad)
				}
			},
		},
		{
			name: "no network field falls back to name",
			ifaceMap: map[string]interface{}{
				"bridge": map[string]interface{}{},
				"name":   "default",
			},
			networkToNAD: map[string]string{"default": "None"},
			assert: func(t *testing.T, _, _, netName, nad string) {
				if netName != "default" {
					t.Errorf("netName = %q, want default", netName)
				}
				if nad != "None" {
					t.Errorf("nad = %q, want None", nad)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ifaceType, model, netName, nad := enrichInterfaceFromSpec(tt.ifaceMap, tt.networkToNAD)
			tt.assert(t, ifaceType, model, netName, nad)
		})
	}
}

// --- integration tests via ReturnNetworkInterfaceMap (VM/stopped path) ---

func TestReturnNetworkInterfaceMap_StoppedVM_Centos(t *testing.T) {
	vm := loadFixture(t, "../../testdata/vms/vm_stopped_centos.yaml")
	interfaces, _, _ := unstructured.NestedSlice(vm.Object, "spec", "template", "spec", "domain", "devices", "interfaces")
	networks, _, _ := unstructured.NestedSlice(vm.Object, "spec", "template", "spec", "networks")

	result, err := ReturnNetworkInterfaceMap(interfaces, networks, false)
	if err != nil {
		t.Fatalf("ReturnNetworkInterfaceMap failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(result))
	}

	iface, ok := result["default"]
	if !ok {
		t.Fatal("interface default not found")
	}
	if iface.Type != "masquerade" {
		t.Errorf("Type = %q, want masquerade", iface.Type)
	}
	if iface.MACAddress != "02:f3:f3:00:00:03" {
		t.Errorf("MACAddress = %q, want 02:f3:f3:00:00:03", iface.MACAddress)
	}
	if iface.NetworkAttachmentDefinition != "None" {
		t.Errorf("NetworkAttachmentDefinition = %q, want None (pod network)", iface.NetworkAttachmentDefinition)
	}
}

func TestReturnNetworkInterfaceMap_StoppedVM_RHEL8(t *testing.T) {
	vm := loadFixture(t, "../../testdata/vms/vm_stopped_rhel8.yaml")
	interfaces, _, _ := unstructured.NestedSlice(vm.Object, "spec", "template", "spec", "domain", "devices", "interfaces")
	networks, _, _ := unstructured.NestedSlice(vm.Object, "spec", "template", "spec", "networks")

	result, err := ReturnNetworkInterfaceMap(interfaces, networks, false)
	if err != nil {
		t.Fatalf("ReturnNetworkInterfaceMap failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(result))
	}

	iface, ok := result["default"]
	if !ok {
		t.Fatal("interface default not found")
	}
	if iface.Type != "masquerade" {
		t.Errorf("Type = %q, want masquerade", iface.Type)
	}
	if iface.Model != "virtio" {
		t.Errorf("Model = %q, want virtio", iface.Model)
	}
	if iface.Network != "default" {
		t.Errorf("Network = %q, want default", iface.Network)
	}
	if iface.NetworkAttachmentDefinition != "None" {
		t.Errorf("NetworkAttachmentDefinition = %q, want None", iface.NetworkAttachmentDefinition)
	}
}

func TestReturnNetworkInterfaceMap_MultusBridge(t *testing.T) {
	vm := loadFixture(t, "../../testdata/vms/vm_migrated_example.yaml")
	interfaces, _, _ := unstructured.NestedSlice(vm.Object, "spec", "template", "spec", "domain", "devices", "interfaces")
	networks, _, _ := unstructured.NestedSlice(vm.Object, "spec", "template", "spec", "networks")

	result, err := ReturnNetworkInterfaceMap(interfaces, networks, false)
	if err != nil {
		t.Fatalf("ReturnNetworkInterfaceMap failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(result))
	}

	var foundNonNoneNAD bool
	for _, iface := range result {
		if iface.NetworkAttachmentDefinition != "None" && iface.NetworkAttachmentDefinition != "" {
			foundNonNoneNAD = true
			break
		}
	}
	if !foundNonNoneNAD {
		t.Fatal("expected at least one interface with NetworkAttachmentDefinition != \"None\"")
	}

	iface, ok := result["net-0"]
	if !ok {
		t.Fatal("interface net-0 not found")
	}
	if iface.Type != "bridge" {
		t.Errorf("Type = %q, want bridge", iface.Type)
	}
	if iface.Model != "virtio" {
		t.Errorf("Model = %q, want virtio", iface.Model)
	}
	if iface.Network != "net-0" {
		t.Errorf("Network = %q, want net-0", iface.Network)
	}
	if iface.NetworkAttachmentDefinition != "default/test-vlan" {
		t.Errorf("NetworkAttachmentDefinition = %q, want default/test-vlan", iface.NetworkAttachmentDefinition)
	}
	if iface.MACAddress != "02:00:00:00:00:01" {
		t.Errorf("MACAddress = %q, want 02:00:00:00:00:01", iface.MACAddress)
	}
}

// --- integration tests via GetVMNetworkInterfaces (VMI/running path) ---

func TestGetVMNetworkInterfaces_RunningVMI(t *testing.T) {
	vmi := loadFixture(t, "../../testdata/vmis/vmi_running_example.yaml")

	result, err := GetVMNetworkInterfaces(vmi)
	if err != nil {
		t.Fatalf("GetVMNetworkInterfaces failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(result))
	}

	iface, ok := result["default"]
	if !ok {
		t.Fatal("interface default not found")
	}
	if iface.MACAddress != "02:00:00:00:00:05" {
		t.Errorf("MACAddress = %q, want 02:00:00:00:00:05", iface.MACAddress)
	}
	if iface.Type != "bridge" {
		t.Errorf("Type = %q, want bridge", iface.Type)
	}
	if iface.Model != "virtio" {
		t.Errorf("Model = %q, want virtio", iface.Model)
	}
	if iface.Network != "default" {
		t.Errorf("Network = %q, want default", iface.Network)
	}
	if iface.NetworkAttachmentDefinition != "None" {
		t.Errorf("NetworkAttachmentDefinition = %q, want None (pod network)", iface.NetworkAttachmentDefinition)
	}
}

// --- integration tests via GetNetworkInterfaces (top-level VM dispatcher) ---

func TestGetNetworkInterfaces_VM_Path(t *testing.T) {
	vm := loadFixture(t, "../../testdata/vms/vm_stopped_rhel8.yaml")

	result, err := GetNetworkInterfaces(vm, false)
	if err != nil {
		t.Fatalf("GetNetworkInterfaces(vmi=false) failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(result))
	}

	iface, ok := result["default"]
	if !ok {
		t.Fatal("interface default not found")
	}
	if iface.MACAddress != "02:f3:f3:00:00:00" {
		t.Errorf("MACAddress = %q, want 02:f3:f3:00:00:00", iface.MACAddress)
	}
	if iface.Type != "masquerade" {
		t.Errorf("Type = %q, want masquerade", iface.Type)
	}
	if iface.Model != "virtio" {
		t.Errorf("Model = %q, want virtio", iface.Model)
	}
}
