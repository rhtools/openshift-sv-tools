package hardware

import (
	"reflect"
	"testing"
)

func TestIsPhysicalInterface(t *testing.T) {
	tests := []struct {
		name string
		iface string
		want bool
	}{
		{"ethernet enp", "enp1s0", true},
		{"ethernet eno", "eno1", true},
		{"ethernet ens", "ens3", true},
		{"traditional eth", "eth0", true},
		{"openvswitch bridge br-", "br-int", true},
		{"wireless wlp", "wlp2s0", true},
		{"wireless wlan", "wlan0", true},
		{"external bridge br-ex", "br-ex", true},
		{"veth virtual", "veth1234", false},
		{"ovs virtual", "ovs-system", false},
		{"loopback lo", "lo", false},
		{"docker bridge", "docker0", false},
		{"cni bridge", "cni0", false},
		{"ovn managed", "ovn-k8s-mp0", false},
		{"long hex container id", "abcdef012345", false},
		{"short non-physical", "abc", false},
		{"eth without digits", "eth", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPhysicalInterface(tt.iface); got != tt.want {
				t.Errorf("isPhysicalInterface(%q) = %v, want %v", tt.iface, got, tt.want)
			}
		})
	}
}

func TestIsHexString(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"lowercase hex", "abcdef0123", true},
		{"digits only", "0123456789", true},
		{"non-hex letters", "xyz", false},
		{"uppercase hex rejected", "12AB", false},
		{"contains space", "ab cd", false},
		{"empty vacuously true", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isHexString(tt.s); got != tt.want {
				t.Errorf("isHexString(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestFilterPhysicalInterfaces(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			"mixed physical and virtual",
			[]string{"enp1s0", "veth123", "eth0", "lo"},
			[]string{"enp1s0", "eth0"},
		},
		{
			"empty input yields nil",
			[]string{},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterPhysicalInterfaces(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterPhysicalInterfaces() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParsePhysicalInterfaceNames(t *testing.T) {
	tests := []struct {
		name    string
		rawJSON []byte
		want    []string
		wantErr bool
	}{
		{
			"valid json filters physical",
			[]byte(`{"node":{"network":{"interfaces":[{"name":"enp1s0"},{"name":"veth123"},{"name":"eth0"}]}}}`),
			[]string{"enp1s0", "eth0"},
			false,
		},
		{
			"invalid json",
			[]byte(`{bad`),
			nil,
			true,
		},
		{
			"empty interfaces",
			[]byte(`{"node":{"network":{"interfaces":[]}}}`),
			nil,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePhysicalInterfaceNames(tt.rawJSON)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParsePhysicalInterfaceNames() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePhysicalInterfaceNames() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseLinkSpeedString(t *testing.T) {
	// parseLinkSpeedString strips a trailing " Mb/s" / "Mb/s" suffix and parses Mbps (see nic_resolver.go).
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"25 Gbps as Mbps", "25000 Mb/s", 25000},
		{"10 Gbps as Mbps", "10000 Mb/s", 10000},
		{"1 Gbps as Mbps", "1000 Mb/s", 1000},
		{"100 Mb/s", "100 Mb/s", 100},
		{"compact Mb/s", "100Mb/s", 100},
		{"empty", "", 0},
		{"unknown text", "unknown", 0},
		{"Gb/s not stripped by implementation", "25 Gb/s", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseLinkSpeedString(tt.in); got != tt.want {
				t.Errorf("parseLinkSpeedString(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
