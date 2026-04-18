package hardware

import (
	"math"
	"testing"

	"vm-scanner/pkg/utils"
)

func TestParseFileSystemStats(t *testing.T) {
	validJSON := `{"node":{"fs":{"availableBytes":536870912000.0,"capacityBytes":1073741824000.0,"usedBytes":536870912000.0}}}`
	var availableBytes int64 = 536870912000
	var capacityBytes int64 = 1073741824000
	var usedBytes int64 = 536870912000
	wantAvail := utils.BytesToGiB(availableBytes)
	wantCap := utils.BytesToGiB(capacityBytes)
	wantUsed := utils.BytesToGiB(usedBytes)
	wantPct := float64(usedBytes) / float64(capacityBytes) * 100

	tests := []struct {
		name    string
		raw     []byte
		wantA   float64
		wantC   float64
		wantU   float64
		wantPct float64
		wantErr string
	}{
		{
			name:    "valid",
			raw:     []byte(validJSON),
			wantA:   wantAvail,
			wantC:   wantCap,
			wantU:   wantUsed,
			wantPct: wantPct,
			wantErr: "",
		},
		{
			name:    "missing node",
			raw:     []byte(`{"other":{}}`),
			wantErr: "node data not found",
		},
		{
			name:    "missing fs",
			raw:     []byte(`{"node":{"other":{}}}`),
			wantErr: "filesystem data not found",
		},
		{
			name:    "invalid JSON",
			raw:     []byte(`{bad`),
			wantErr: "json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, c, u, pct, err := ParseFileSystemStats(tt.raw)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("ParseFileSystemStats() err = nil, want error")
				}
				switch tt.wantErr {
				case "json":
					// any JSON unmarshal error
				default:
					if err.Error() != tt.wantErr {
						t.Fatalf("ParseFileSystemStats() err = %v, want %q", err, tt.wantErr)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFileSystemStats() unexpected err: %v", err)
			}
			const eps = 1e-9
			if math.Abs(a-tt.wantA) > eps {
				t.Errorf("available GiB = %v, want %v", a, tt.wantA)
			}
			if math.Abs(c-tt.wantC) > eps {
				t.Errorf("capacity GiB = %v, want %v", c, tt.wantC)
			}
			if math.Abs(u-tt.wantU) > eps {
				t.Errorf("used GiB = %v, want %v", u, tt.wantU)
			}
			if math.Abs(pct-tt.wantPct) > eps {
				t.Errorf("usage %% = %v, want %v", pct, tt.wantPct)
			}
		})
	}
}
