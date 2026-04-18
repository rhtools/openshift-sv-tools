# VM Scanner Test Data

This directory contains test fixtures for the vm-scanner project. These files allow testing without requiring a live OpenShift cluster.

## Directory Structure

```
testdata/
├── vms/                    # VirtualMachine resources
│   ├── vm_stopped_*.yaml   # VMs in stopped state (from test cluster)
│   └── vm_migrated_*.yaml  # Sanitized migrated VMs
├── vmis/                   # VirtualMachineInstance resources
│   └── vmi_running_*.yaml  # Running VM instances
├── pvcs/                   # PersistentVolumeClaim resources
│   ├── pvc_*_volume.yaml   # PVCs from test cluster
│   └── pvc_migrated_*.yaml # Sanitized migrated PVCs
└── raw/                    # Raw API responses
    └── vm_list_all.yaml    # Full VM list from cluster
```

## Data Sources

### From Test Cluster (openshift-cnv namespace)
Files extracted from the lab cluster at api.example.cluster.com:
- `vms/vm_stopped_*.yaml` - Various stopped VMs (CentOS, RHEL, Fedora)
- `pvcs/pvc_*_volume.yaml` - PVCs associated with test VMs
- `raw/vm_list_all.yaml` - Complete VM list from cluster

**Note**: Test cluster data is NOT sanitized as it's a non-production environment.

### Sanitized Migration Files
Files derived from production migrations with all identifying information removed:
- `vms/vm_migrated_example.yaml` - Example migrated VM
- `pvcs/pvc_migrated_example.yaml` - Example migrated PVC

**Sanitization includes**:
- Generic UUIDs (00000000-0000-0000-0000-00000000000X)
- Generic names (test-vm-migrated, test-namespace)
- Generic network info (test-vlan, 02:00:00:00:00:XX MACs)
- Generic hostnames (test-node-01.example.com)
- Removed internal plan/migration IDs
- Removed internal datastore paths

## Using Test Data

### In Go Tests

```go
func TestParseVM(t *testing.T) {
    // Load test fixture
    data, err := os.ReadFile("testdata/vms/vm_stopped_centos.yaml")
    if err != nil {
        t.Fatalf("Failed to read test data: %v", err)
    }
    
    // Parse and test your code
    var vm unstructured.Unstructured
    err = yaml.Unmarshal(data, &vm)
    // ... your test logic
}
```

### Table-Driven Tests

```go
func TestVMParsing(t *testing.T) {
    tests := []struct {
        name     string
        fixture  string
        expected VMInfo
    }{
        {"stopped centos", "testdata/vms/vm_stopped_centos.yaml", /* ... */},
        {"migrated vm", "testdata/vms/vm_migrated_example.yaml", /* ... */},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

## Updating Test Data

To refresh test data from the lab cluster:

```bash
export KUBECONFIG=/path/to/kubeconfig-lab

# Get VMs
oc get vm <vm-name> -n openshift-cnv -o yaml > testdata/vms/vm_<name>.yaml

# Get VMIs (if running)
oc get vmi <vmi-name> -n openshift-cnv -o yaml > testdata/vmis/vmi_<name>.yaml

# Get PVCs
oc get pvc <pvc-name> -n openshift-cnv -o yaml > testdata/pvcs/pvc_<name>.yaml

# Get full lists
oc get vms -n openshift-cnv -o yaml > testdata/raw/vm_list_all.yaml
```

## Test Coverage

These fixtures support testing:
- VM parsing and metadata extraction
- Storage configuration parsing (disks, PVCs)
- Network interface parsing
- CPU and memory resource parsing
- VM state detection (running, stopped)
- Migration metadata handling
- Multi-disk configurations
- Various Linux distributions (CentOS, RHEL, Fedora)
