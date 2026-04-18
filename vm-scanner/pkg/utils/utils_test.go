package utils

import (
	"reflect"
	"testing"
)

func TestParseQuantityToBytes(t *testing.T) {
	tests := []struct {
		name        string
		quantityStr string
		want        int64
		wantErr     bool
	}{
		{name: "2Gi", quantityStr: "2Gi", want: 2147483648, wantErr: false},
		{name: "1024Mi", quantityStr: "1024Mi", want: 1073741824, wantErr: false},
		{name: "zero", quantityStr: "0", want: 0, wantErr: false},
		{name: "empty string", quantityStr: "", want: 0, wantErr: true},
		{name: "invalid", quantityStr: "invalid", want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQuantityToBytes(tt.quantityStr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseQuantityToBytes() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseQuantityToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBytesToMiB(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  float64
	}{
		{name: "one MiB", bytes: 1048576, want: 1.0},
		{name: "zero", bytes: 0, want: 0.0},
		{name: "five MiB", bytes: 5242880, want: 5.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BytesToMiB(tt.bytes); got != tt.want {
				t.Errorf("BytesToMiB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBytesToGiB(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  float64
	}{
		{name: "one GiB", bytes: 1073741824, want: 1.0},
		{name: "zero", bytes: 0, want: 0.0},
		{name: "five GiB", bytes: 5368709120, want: 5.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BytesToGiB(tt.bytes); got != tt.want {
				t.Errorf("BytesToGiB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuantityToMiB(t *testing.T) {
	tests := []struct {
		name        string
		quantityStr string
		want        float64
		wantErr     bool
	}{
		{name: "2Gi", quantityStr: "2Gi", want: 2048.0, wantErr: false},
		{name: "empty string", quantityStr: "", want: 0.0, wantErr: false},
		{name: "invalid", quantityStr: "invalid", want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := QuantityToMiB(tt.quantityStr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("QuantityToMiB() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("QuantityToMiB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuantityToGiB(t *testing.T) {
	tests := []struct {
		name        string
		quantityStr string
		want        float64
		wantErr     bool
	}{
		{name: "2Gi", quantityStr: "2Gi", want: 2.0, wantErr: false},
		{name: "empty string", quantityStr: "", want: 0.0, wantErr: false},
		{name: "invalid", quantityStr: "invalid", want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := QuantityToGiB(tt.quantityStr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("QuantityToGiB() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("QuantityToGiB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildPath(t *testing.T) {
	tests := []struct {
		name   string
		prefix []string
		keys   []string
		want   []string
	}{
		{
			name:   "nested domain cpu",
			prefix: []string{"spec", "template", "spec"},
			keys:   []string{"domain", "cpu"},
			want:   []string{"spec", "template", "spec", "domain", "cpu"},
		},
		{name: "empty prefix", prefix: []string{}, keys: []string{"a"}, want: []string{"a"}},
		{name: "no keys", prefix: []string{"spec"}, keys: []string{}, want: []string{"spec"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildPath(tt.prefix, tt.keys...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterfaceSliceToStringSlice(t *testing.T) {
	tests := []struct {
		name  string
		input []interface{}
		want  []string
	}{
		{name: "all strings", input: []interface{}{"a", "b", "c"}, want: []string{"a", "b", "c"}},
		{name: "skip non-string", input: []interface{}{"a", 42, "b"}, want: []string{"a", "b"}},
		{name: "empty slice", input: []interface{}{}, want: []string{}},
		{name: "nil input", input: nil, want: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InterfaceSliceToStringSlice(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InterfaceSliceToStringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoundToOneDecimal(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  float64
	}{
		{name: "pi", value: 3.14159, want: 3.1},
		{name: "half up", value: 2.05, want: 2.1},
		{name: "zero", value: 0.0, want: 0.0},
		{name: "negative", value: -1.55, want: -1.6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RoundToOneDecimal(tt.value); got != tt.want {
				t.Errorf("RoundToOneDecimal() = %v, want %v", got, tt.want)
			}
		})
	}
}
