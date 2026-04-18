package vm

import (
	"slices"
	"testing"
)

func TestGetPVCList(t *testing.T) {
	migrated := loadFixture(t, "../../testdata/vms/vm_migrated_example.yaml")
	vmSpec, ok := migrated.Object["spec"].(map[string]interface{})
	if !ok {
		t.Fatal("fixture spec is not a map")
	}

	tests := []struct {
		name     string
		vmSpec   map[string]interface{}
		wantOK   bool
		wantPVCs []string
	}{
		{
			name:     "migrated example volumes",
			vmSpec:   vmSpec,
			wantOK:   true,
			wantPVCs: []string{"test-vm-disk-0", "test-vm-disk-1"},
		},
		{
			name:     "empty spec without template",
			vmSpec:   map[string]interface{}{},
			wantOK:   false,
			wantPVCs: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, got := getPVCList(tt.vmSpec)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !slices.Equal(got, tt.wantPVCs) {
				t.Errorf("PVCs = %#v, want %#v", got, tt.wantPVCs)
			}
		})
	}
}

func TestSumTotalStorage(t *testing.T) {
	tests := []struct {
		name string
		in   []StorageInfo
		want int64
	}{
		{
			name: "two items",
			in: []StorageInfo{
				{SizeBytes: 100},
				{SizeBytes: 200},
			},
			want: 300,
		},
		{
			name: "empty slice",
			in:   nil,
			want: 0,
		},
		{
			name: "single 1 GiB",
			in: []StorageInfo{
				{SizeBytes: 1073741824},
			},
			want: 1073741824,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sumTotalStorage(tt.in)
			if got != tt.want {
				t.Errorf("sumTotalStorage = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseStorageInfoFromGuestAgentInfo(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []VMDiskInfo
		wantErr bool
	}{
		{
			name: "valid",
			input: `{"fsInfo":{"disks":[{"diskName":"vda1","fileSystemType":"xfs","mountPoint":"/","totalBytes":10737418240,"usedBytes":5368709120}]}}`,
			want: []VMDiskInfo{
				{
					DiskName:   "vda1",
					FsType:     "xfs",
					MountPoint: "/",
					TotalBytes: 10737418240,
					UsedBytes:  5368709120,
				},
			},
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   "{bad",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStorageInfoFromGuestAgentInfo(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseStorageInfoFromGuestAgentInfo: %v", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
