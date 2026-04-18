package vm

import (
	"testing"
)

func TestParseVMIMachineInfo(t *testing.T) {
	vmi := loadFixture(t, "../../testdata/vmis/vmi_running_example.yaml")
	prettyName, node, version, err := ParseVMIMachineInfo(vmi)
	if err != nil {
		t.Fatalf("ParseVMIMachineInfo: %v", err)
	}
	// Fixture vmi_running_example.yaml: status.guestOSInfo.prettyName
	wantPretty := "Red Hat Enterprise Linux 9.3 (Plow)"
	if prettyName != wantPretty {
		t.Errorf("prettyName = %q, want %q", prettyName, wantPretty)
	}
	if node != "test-node-01" {
		t.Errorf("runningOnNode = %q, want %q", node, "test-node-01")
	}
	if version != "9.3" {
		t.Errorf("guestOSVersion = %q, want %q", version, "9.3")
	}
}

func TestParseVMMachineInfo(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		want    string
	}{
		{
			name:    "rhel8 os annotation",
			fixture: "../../testdata/vms/vm_stopped_rhel8.yaml",
			want:    "rhel8",
		},
		{
			name:    "no os annotation",
			fixture: "../../testdata/vms/vm_migrated_example.yaml",
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := loadFixture(t, tt.fixture)
			got, err := ParseVMMachineInfo(vm)
			if err != nil {
				t.Fatalf("ParseVMMachineInfo: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseGuestMetadataFromGuestAgentInfo(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *GuestMetadata
		wantErr bool
	}{
		{
			name:  "valid with timezone",
			input: `{"hostname":"test-host","guestAgentVersion":"5.2","timezone":"EST, -0500","os":{"kernelRelease":"5.14.0"}}`,
			want: &GuestMetadata{
				HostName:          "test-host",
				GuestAgentVersion: "5.2",
				KernelVersion:     "5.14.0",
				Timezone:          "EST",
			},
			wantErr: false,
		},
		{
			name:  "valid without timezone",
			input: `{"hostname":"test-host","guestAgentVersion":"5.2","timezone":"","os":{"kernelRelease":"5.14.0"}}`,
			want: &GuestMetadata{
				HostName:          "test-host",
				GuestAgentVersion: "5.2",
				KernelVersion:     "5.14.0",
				Timezone:          "",
			},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{invalid`,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGuestMetadataFromGuestAgentInfo(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseGuestMetadataFromGuestAgentInfo: %v", err)
			}
			if got.HostName != tt.want.HostName {
				t.Errorf("HostName = %q, want %q", got.HostName, tt.want.HostName)
			}
			if got.GuestAgentVersion != tt.want.GuestAgentVersion {
				t.Errorf("GuestAgentVersion = %q, want %q", got.GuestAgentVersion, tt.want.GuestAgentVersion)
			}
			if got.KernelVersion != tt.want.KernelVersion {
				t.Errorf("KernelVersion = %q, want %q", got.KernelVersion, tt.want.KernelVersion)
			}
			if got.Timezone != tt.want.Timezone {
				t.Errorf("Timezone = %q, want %q", got.Timezone, tt.want.Timezone)
			}
		})
	}
}
