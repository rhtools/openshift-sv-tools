package vm

import (
	"fmt"

	"vm-scanner/pkg/utils"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type CPUInfo struct {
	CPUCores   int64  `json:"cpuCores" yaml:"cpuCores"`
	CPUModel   string `json:"cpuModel" yaml:"cpuModel"`
	CPUSockets int64  `json:"cpuSockets" yaml:"cpuSockets"`
	CPUThreads int64  `json:"cpuThreads" yaml:"cpuThreads"`
	VCPUs      int64  `json:"vCPUsTotal" yaml:"vCPUsTotal"`
}

func GetCPUInfoFromVM(vmUnstructured unstructured.Unstructured) (*CPUInfo, error) {
	return GetCPUInfo(vmUnstructured, "spec", "template", "spec")
}

func GetCPUInfoFromVMI(vmiUnstructured unstructured.Unstructured) (*CPUInfo, error) {
	return GetCPUInfo(vmiUnstructured, "spec")
}

func GetCPUInfo(vmiUnstructured unstructured.Unstructured, prefix ...string) (*CPUInfo, error) {
	// vCPUs is the total number of cores * sockets * threads
	// Here is an example VMI spec layout
	// spec:
	// 	domain:
	// 		cpu:
	// 		cores: 1
	// 		maxSockets: 8
	// 		model: host-model
	// 		sockets: 2
	// 		threads: 1
	// The VM layout is prefixed with spec.template otherwise is exactly the same

	// I want to make this flexible so I am going to take a prefix in
	// I need to build the full path for each thing so I need a helper
	// hopefully StackOverflow is correct... I have no idea how to do this in Go
	// path := func(keys ...string) []string {
	// 	full := make([]string, len(prefix), len(prefix)+len(keys))
	// 	copy(full, prefix)
	// 	return append(full, keys...)
	// }

	CPUCores, found, err := unstructured.NestedInt64(vmiUnstructured.Object, utils.BuildPath(prefix, "domain", "cpu", "cores")...)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU cores: %w", err)
	}
	if !found {
		// Stopped VMs may not have explicit CPU config -- return zeroed values rather than error
		return &CPUInfo{}, nil
	}
	CPUSockets, found, err := unstructured.NestedInt64(vmiUnstructured.Object, utils.BuildPath(prefix, "domain", "cpu", "sockets")...)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU sockets: %w", err)
	}
	if !found {
		return &CPUInfo{}, nil
	}
	CPUThreads, found, err := unstructured.NestedInt64(vmiUnstructured.Object, utils.BuildPath(prefix, "domain", "cpu", "threads")...)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU threads: %w", err)
	}
	if !found {
		return &CPUInfo{}, nil
	}
	// total vCPUs is cores * sockets * threads per core
	vCPUs := CPUCores * CPUSockets * CPUThreads

	// CPU model is optional - if not specified, use empty string
	CPUModel, _, _ := unstructured.NestedString(vmiUnstructured.Object, utils.BuildPath(prefix, "domain", "cpu", "model")...)
	if CPUModel == "" {
		CPUModel = "host-model" // Default for VMs without explicit model
	}

	cpus := CPUInfo{
		CPUCores:   CPUCores,
		CPUSockets: CPUSockets,
		CPUThreads: CPUThreads,
		VCPUs:      vCPUs,
		CPUModel:   CPUModel,
	}

	return &cpus, nil
}
