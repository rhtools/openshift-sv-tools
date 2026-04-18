package vm

import (
	"math"
	"testing"

	"vm-scanner/pkg/utils"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestParseMetricValue(t *testing.T) {
	tests := []struct {
		name string
		line string
		want float64
	}{
		{
			name: "normal metric line",
			line: `kubevirt_vmi_memory_available_bytes{name="test"} 8589934592`,
			want: 8589934592.0,
		},
		{
			name: "no space",
			line: "nospace",
			want: 0,
		},
		{
			name: "non-numeric value",
			line: "metric value abc",
			want: 0,
		},
		{
			name: "empty string",
			line: "",
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseMetricValue(tt.line)
			if got != tt.want {
				t.Errorf("parseMetricValue(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestParseMemoryFromMonitoring(t *testing.T) {
	validMetrics := `# HELP kubevirt_vmi_memory_available_bytes
# TYPE kubevirt_vmi_memory_available_bytes gauge
kubevirt_vmi_memory_available_bytes{name="test-vm"} 8589934592
kubevirt_vmi_memory_usable_bytes{name="test-vm"} 4294967296
`
	wantUsed := utils.BytesToMiB(8589934592 - 4294967296)
	wantFree := utils.BytesToMiB(4294967296)

	tests := []struct {
		name      string
		queryData string
		vmiName   string
		wantUsed  float64
		wantFree  float64
	}{
		{
			name:      "valid metrics",
			queryData: validMetrics,
			vmiName:   "test-vm",
			wantUsed:  wantUsed,
			wantFree:  wantFree,
		},
		{
			name:      "wrong vmi name",
			queryData: validMetrics,
			vmiName:   "other-vm",
			wantUsed:  0,
			wantFree:  0,
		},
		{
			name:      "empty input",
			queryData: "",
			vmiName:   "test",
			wantUsed:  0,
			wantFree:  0,
		},
		{
			name:      "comments only",
			queryData: "# comment\n# another",
			vmiName:   "test",
			wantUsed:  0,
			wantFree:  0,
		},
	}
	const eps = 1e-9
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUsed, gotFree, err := ParseMemoryFromMonitoring(tt.queryData, tt.vmiName)
			if err != nil {
				t.Fatalf("ParseMemoryFromMonitoring() unexpected error: %v", err)
			}
			if math.Abs(gotUsed-tt.wantUsed) > eps {
				t.Errorf("usedMiB = %v, want %v", gotUsed, tt.wantUsed)
			}
			if math.Abs(gotFree-tt.wantFree) > eps {
				t.Errorf("freeMiB = %v, want %v", gotFree, tt.wantFree)
			}
		})
	}
}

func TestGetMemoryHotPlugMax(t *testing.T) {
	withMaxGuest := unstructured.Unstructured{Object: map[string]interface{}{
		"spec": map[string]interface{}{
			"domain": map[string]interface{}{
				"memory": map[string]interface{}{
					"maxGuest": "16Gi",
				},
			},
		},
	}}

	tests := []struct {
		name    string
		obj     unstructured.Unstructured
		want    float64
		wantErr bool
	}{
		{
			name:    "with maxGuest",
			obj:     withMaxGuest,
			want:    16384.0,
			wantErr: false,
		},
		{
			name:    "without maxGuest",
			obj:     unstructured.Unstructured{Object: map[string]interface{}{}},
			want:    0,
			wantErr: true,
		},
	}
	const eps = 1e-6
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMemoryHotPlugMax(tt.obj)
			if tt.wantErr {
				if err == nil {
					t.Fatal("GetMemoryHotPlugMax() error = nil, want non-nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("GetMemoryHotPlugMax() unexpected error: %v", err)
			}
			if math.Abs(got-tt.want) > eps {
				t.Errorf("GetMemoryHotPlugMax() = %v, want %v", got, tt.want)
			}
		})
	}
}
