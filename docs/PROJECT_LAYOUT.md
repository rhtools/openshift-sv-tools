# VM Scanner Project Layout

## Overview
This project provides a Go-based tool for extracting VirtualMachine information from OpenShift Virtualization or KubeVirt clusters, similar to RVTools for VMware. The tool uses the Kubernetes client-go library and KubeVirt client to interact with cluster resources.

## Project Structure (Enhanced for RVTools Equivalent)

```
vm-scanner/
├── cmd/
│   └── main.go                 # Application entry point with CLI flags
├── pkg/
│   ├── api/                   # API endpoint definitions (empty stub for future use)
│   ├── client/
│   │   ├── client.go          # Kubernetes/KubeVirt client management
│   │   └── raw_api.go         # Raw Kubernetes API data collection
│   ├── cluster/
│   │   ├── cluster.go         # Cluster-wide information with modular hardware collection
│   │   └── types.go           # Cluster-related type definitions
│   ├── config/
│   │   └── config.go          # Configuration and CLI flag parsing
│   ├── error_handling/
│   │   └── error_handling.go # Common error handling utilities
│   ├── hardware/
│   │   ├── baremetal.go       # Bare metal host integration (Metal3)
│   │   ├── cpu.go             # CPU information extraction
│   │   ├── memory.go          # Memory information extraction
│   │   ├── network.go         # Network hardware and configuration extraction
│   │   ├── storage.go         # Enhanced storage class and filesystem collection
│   │   └── types.go           # Hardware-related type definitions
│   ├── output/
│   │   ├── csv_helpers.go     # CSV formatting utilities
│   │   ├── formatter.go       # Main formatter and routing logic
│   │   ├── formatter_csv.go   # Single CSV file output
│   │   ├── formatter_json.go  # JSON output formatting
│   │   ├── formatter_multi_csv.go  # Multi-CSV file output
│   │   ├── formatter_table.go # Table output formatting
│   │   ├── formatter_xlsx.go  # Excel (XLSX) workbook output
│   │   ├── formatter_yaml.go  # YAML output formatting
│   │   ├── report_generator.go # Comprehensive report orchestration
│   │   ├── table_helpers.go   # Table formatting utilities
│   │   ├── types.go           # Output-related type definitions
│   │   └── xlsx_helpers.go    # Excel formatting utilities
│   ├── rbac/
│   │   └── permissions.go     # RBAC configuration
│   ├── software/              # Software inventory (empty, planned for future)
│   ├── utils/
│   │   ├── quantity.go        # Kubernetes quantity parsing and conversions
│   │   └── shared_functions.go # Common utility functions
│   └── vm/
│       ├── cpu.go             # CPU-specific VM information extraction
│       ├── memory.go          # Memory-specific VM information extraction
│       ├── metadata.go        # VM metadata and annotation handling
│       ├── metrics.go         # Resource utilization collection
│       ├── network.go         # VM network interface and configuration
│       ├── storage.go         # VM storage volume information
│       ├── types.go           # Comprehensive VM data structures
│       ├── vm.go              # VirtualMachine operations and scanning
│       └── vm_test.go         # Unit tests for VM operations
├── docs/
│   └── PROJECT_LAYOUT.md      # This comprehensive documentation
├── go.mod                     # Go module dependencies
├── go.sum                     # Go module checksums
├── Makefile                   # Build automation
├── README.md                  # Project overview
```

## Component Details

### cmd/main.go
**Purpose**: Application entry point and CLI interface
**Key Features**:
- Command-line argument parsing
- Client initialization and connection testing
- Operation routing (list VMs, get VM details, cluster info)
- Error handling and logging

**API Endpoints Used**:
- No direct API calls - orchestrates other components

### pkg/client/client.go
**Purpose**: Kubernetes and KubeVirt client management
**Key Features**:
- Client initialization with kubeconfig support
- In-cluster and out-of-cluster configuration
- Connection testing
- Namespace listing

**API Endpoints Used**:
- `k8s.io/api/core/v1` - Core Kubernetes resources (Nodes, Namespaces)
- `k8s.io/client-go` - Kubernetes client library
- `kubevirt.io/client-go` - KubeVirt client library

### pkg/client/raw_api.go
**Purpose**: Raw Kubernetes API data collection
**Key Features**:
- Direct access to node statistics summary endpoints
- Raw API data collection for filesystem and network information
- Support for advanced metrics collection

**API Endpoints Used**:
- `/api/v1/nodes/{node}/proxy/stats/summary` - Node statistics endpoint

### pkg/vm/vm.go
**Purpose**: VirtualMachine operations and data extraction
**Key Features**:
- List all VMs across namespaces
- List VMs in specific namespace
- Get detailed VM information
- Get cluster-wide VM statistics
- Extract related resources (Pods, Services, PVCs, Events)
- Collect resource utilization metrics
- Gather storage and network information
- Extract node and cluster information

**API Endpoints Used**:
- `kubevirt.io/api/core/v1` - VirtualMachine, VirtualMachineInstance resources
- `k8s.io/api/core/v1` - Pod, Service, PersistentVolumeClaim, Event, Node resources
- `k8s.io/api/apps/v1` - Deployment, ReplicaSet resources
- `k8s.io/api/storage/v1` - PersistentVolume, StorageClass resources
- `k8s.io/apimachinery/pkg/apis/meta/v1` - ListOptions, GetOptions, metav1 types
- `metrics.k8s.io/v1beta1` - NodeMetrics, PodMetrics (if metrics server available)

### pkg/vm/types.go
**Purpose**: Data structures and type definitions
**Key Features**:
- VMInfo - Basic VM information structure
- VMDetails - Detailed VM information with related resources
- VMList - Collection of VMs with summary statistics
- ClusterInfo - Cluster-wide VM information
- DiskInfo, InterfaceInfo - VM resource details
- ResourceMetrics - CPU and memory utilization data
- StorageMetrics - Disk usage and capacity information
- NetworkMetrics - Network interface statistics
- NodeInfo - Node hardware and resource information

**API Endpoints Used**:
- Type definitions based on Kubernetes and KubeVirt API schemas

### pkg/vm/metrics.go
**Purpose**: Resource utilization collection and metrics gathering
**Key Features**:
- CPU and memory utilization collection
- Storage I/O metrics gathering
- Network interface statistics
- Node resource allocation tracking
- Historical metrics collection (if available)

**API Endpoints Used**:
- `metrics.k8s.io/v1beta1` - NodeMetrics, PodMetrics
- `custom.metrics.k8s.io/v1beta1` - Custom metrics
- `k8s.io/api/core/v1` - Node status and resources

### pkg/vm/storage.go
**Purpose**: VM storage volume information extraction
**Key Features**:
- Extract storage configuration from VMs
- Handle multiple storage patterns (instanceType, template, generic PVC)
- Parse dataVolumeTemplates and PVC references
- Detect storage classes and volume types

**Functions**:
- `GetVMStorageInfo()` - Main entry point for VM storage extraction
- `extractInstanceTypeStorage()` - Extracts storage from instanceType-based VMs
- `extractTemplateStorage()` - Extracts storage from template-based VMs
- `extractGenericPVCStorage()` - Extracts storage from external PVC references
- `hasDataVolumeTemplates()` - Detects presence of dataVolumeTemplates

**API Endpoints Used**:
- Works with unstructured VM data from dynamic client
- Parses nested JSON structures for volume definitions

### pkg/vm/cpu.go
**Purpose**: CPU-specific VM information extraction
**Key Features**:
- Extract CPU allocation from VM specifications
- Parse CPU cores, sockets, and threads configuration
- Handle instanceType CPU references
- CPU request and limit extraction

**API Endpoints Used**:
- `kubevirt.io/api/core/v1` - VirtualMachine CPU specifications
- Works with unstructured VM data

### pkg/vm/memory.go
**Purpose**: Memory-specific VM information extraction
**Key Features**:
- Extract memory allocation from VM specifications
- Parse memory requests and limits
- Handle instanceType memory references
- Convert memory quantities to standard units (using pkg/utils)

**API Endpoints Used**:
- `kubevirt.io/api/core/v1` - VirtualMachine memory specifications
- Works with unstructured VM data

### pkg/vm/metadata.go
**Purpose**: VM metadata and annotation handling
**Key Features**:
- Extract VM labels and annotations
- Parse VM naming and namespace information
- Handle custom metadata fields
- Extract creation timestamps and owner references

**API Endpoints Used**:
- `k8s.io/apimachinery/pkg/apis/meta/v1` - Metadata structures
- Works with unstructured VM data

### pkg/vm/network.go
**Purpose**: VM network interface and configuration extraction
**Key Features**:
- Extract network interface definitions from VMs
- Parse pod network and multus network attachments
- IP address extraction (pod IP and guest IP if available)
- Network interface models and MAC addresses

**API Endpoints Used**:
- `kubevirt.io/api/core/v1` - VirtualMachineInstance network status
- `k8s.io/api/core/v1` - Pod network information
- Works with unstructured VM and VMI data

### pkg/hardware/storage.go
**Purpose**: Enhanced storage class information and filesystem collection
**Key Features**:
- Comprehensive storage class information with extended metadata
- Storage class creation timestamps and expansion capabilities
- Mount options and topology constraints
- Filesystem statistics parsing from raw API data
- Storage class utilization tracking

**Data Structures**:
- `NodeFilesystemInfo` - Filesystem capacity, usage, and availability metrics
- `StorageClassInfo` - Enhanced storage class details including creation time, volume expansion, mount options, and topology constraints
- `GetStorageClasses()` - Retrieves comprehensive storage class information from cluster
- `ParseFileSystemStats()` - Parses raw filesystem data from node stats endpoint

**API Endpoints Used**:
- `k8s.io/api/storage/v1` - StorageClass (enhanced collection)
- Node stats summary endpoint - `/api/v1/nodes/{node}/proxy/stats/summary`

### pkg/hardware/cpu.go
**Purpose**: CPU information extraction from node resources
**Key Features**:
- CPU core count extraction
- Modular CPU information collection
- CPU capacity parsing from node status

**Data Structures**:
- `CPUInfo` - CPU core information structure
- `GetCPUInfo()` - Extracts CPU information from node object

**API Endpoints Used**:
- `k8s.io/api/core/v1` - Node status and capacity information

### pkg/hardware/memory.go
**Purpose**: Memory information extraction from node resources
**Key Features**:
- Memory capacity extraction and conversion to GiB
- Modular memory information collection
- Memory capacity parsing from node status

**Data Structures**:
- `MemoryInfo` - Memory capacity information structure
- `GetMemoryInfo()` - Extracts memory information from node object

**API Endpoints Used**:
- `k8s.io/api/core/v1` - Node status and capacity information

### pkg/hardware/network.go
**Purpose**: Network hardware information and configuration extraction
**Key Features**:
- Network interface detection and filtering
- OVN/OVS network configuration parsing
- Physical interface identification
- L3 gateway configuration extraction
- Host CIDR and pod network subnet parsing

**Data Structures**:
- `NetworkInfo` - Comprehensive network configuration structure
- `L3GatewayConfig` - OVN L3 gateway configuration
- `GetPodNetworkSubnets()` - Extracts pod network subnet information
- `GetHostCIDRs()` - Extracts host CIDR information
- `GetL3GatewayConfig()` - Parses L3 gateway configuration
- `ParseNetworkInterfaces()` - Parses network interfaces from raw API data

**API Endpoints Used**:
- Node annotations for OVN configuration
- Node stats summary endpoint for interface information

### pkg/hardware/baremetal.go
**Purpose**: Bare metal host integration for Metal3-enabled clusters
**Key Features**:
- Retrieve BareMetalHost custom resources associated with nodes
- Integration with Metal3.io bare metal provisioning
- Enriches node information with hardware details from BMH

**Functions**:
- `GetBareMetalHost()` - Retrieves BareMetalHost CR for a given node

**Data Structures**:
- Uses `unstructured.Unstructured` for dynamic BareMetalHost access

**API Endpoints Used**:
- `metal3.io/v1alpha1` - BareMetalHost custom resource
- Queries openshift-machine-api namespace for BMH resources

### pkg/hardware/types.go
**Purpose**: Hardware-related type definitions
**Key Features**:
- StorageClassInfo structure with comprehensive metadata
- NodeFilesystemInfo for filesystem statistics
- CPUInfo and MemoryInfo structures
- NetworkInfo and L3GatewayConfig structures

### pkg/config/config.go
**Purpose**: Configuration and CLI flag parsing
**Key Features**:
- CLI flag definition and parsing using Go's flag package
- Configuration struct with embedded AuthOptions (like inheritance in Python)
- Operation flags for different data collection modes
- Output format and file destination configuration

**CLI Flags**:
- Auth flags: `auth-method`, `kube-config`, `token`, `api-url`
- Operation flags: `node-hardware-info`, `kubevirt-version`, `storage-classes`, `storage-volumes`, `vm-info`, `vm-inventory`
- Testing flags: `test-connection`, `test-all-outputs`, `test-current-feat`
- Output flags: `output-format` (stdout, json, yaml, csv, multi-csv, xlsx), `output-file`

**Functions**:
- `FromCLIFlags()` - Parses CLI arguments and returns Config struct

**API Endpoints Used**:
- No direct API calls - configuration management only

### pkg/cluster/cluster.go
**Purpose**: Cluster-wide information collection with modular hardware extraction
**Key Features**:
- KubeVirt version and configuration using dynamic client
- Modular node hardware information collection
- Enhanced ClusterNodeInfo structure with comprehensive network and filesystem data
- Integration with hardware package functions for CPU, memory, and network data
- Operator status and health (stubbed)

**Functions**:
- `GetKubeVirtVersion()` - Uses dynamic client to retrieve KubeVirt operator information
- `GetClusterNodeInfo()` - Collects hardware information using modular hardware functions

**API Endpoints Used**:
- `k8s.io/api/core/v1` - Node, Namespace
- `k8s.io/apimachinery/pkg/apis/meta/v1/unstructured` - Dynamic client operations
- `kubevirt.io/api/core/v1` - KubeVirt Custom Resources via dynamic client
- Node stats summary endpoint for raw data collection

### pkg/cluster/types.go
**Purpose**: Cluster-related type definitions
**Key Features**:
- ClusterNodeInfo - Comprehensive node hardware information including network configuration
- KubeVirtVersion - KubeVirt deployment status and version
- ClusterInfo - Overall cluster information structure
- ClusterSummary - Aggregated cluster statistics
- OperatorStatus - KubeVirt operator health status

### pkg/error_handling/error_handling.go
**Purpose**: Common error handling utilities for the vm-scanner
**Key Features**:
- Standardized error handling for required field validation
- Consistent error formatting and messaging
- Error handling for nested JSON/YAML structure access

**Functions**:
- `GetRequiredString()` - Validates required string fields from unstructured data

### pkg/rbac/permissions.go
**Purpose**: RBAC and Service Account management
**Key Features**:
- Programmatic creation of ClusterRole for vm-scanner
- Programmatic creation of ClusterRoleBinding
- Programmatic creation of ServiceAccount

**API Endpoints Used**:
- `k8s.io/api/core/v1` - ServiceAccount
- `k8s.io/api/rbac/v1` - ClusterRole, ClusterRoleBinding

### pkg/utils/quantity.go
**Purpose**: Kubernetes quantity parsing and unit conversion utilities
**Key Features**:
- Parse Kubernetes quantity strings (e.g., "2Gi", "1024Mi") to bytes
- Convert bytes to MiB and GiB (base-2: 1024)
- Convenience functions for quantity-to-unit conversions

**Functions**:
- `ParseQuantityToBytes()` - Converts quantity string to bytes (similar to Python's int conversion)
- `BytesToMiB()` - Converts bytes to mebibytes
- `BytesToGiB()` - Converts bytes to gibibytes
- `QuantityToMiB()` - Combined parsing and MiB conversion
- `QuantityToGiB()` - Combined parsing and GiB conversion

**API Endpoints Used**:
- No direct API calls - utility functions for data manipulation

### pkg/utils/shared_functions.go
**Purpose**: Common utility functions shared across packages
**Key Features**:
- Path building for nested data structures
- Type conversions for interface slices
- Decimal rounding for consistent API output

**Functions**:
- `BuildPath()` - Constructs nested path arrays for data access
- `InterfaceSliceToStringSlice()` - Converts []interface{} to []string (like Python list comprehension)
- `RoundToOneDecimal()` - Rounds float64 to 1 decimal place (like Python's round(value, 1))

**API Endpoints Used**:
- No direct API calls - utility functions for data manipulation

### pkg/api/
**Purpose**: API endpoint definitions (placeholder for future functionality)
**Key Features**:
- Currently an empty package stub
- Reserved for future REST API or gRPC endpoint implementations
- Planned for exposing vm-scanner as a service

**API Endpoints Used**:
- Not yet implemented

### pkg/output/formatter.go
**Purpose**: Main formatter and routing logic for all output formats
**Key Features**:
- Routes data to appropriate format-specific formatter based on output type
- Type detection and dispatch for different data structures
- Orchestrates comprehensive report generation
- Supports multiple output destinations (stdout, file)

**Supported Output Formats**:
- Table (stdout) - via formatter_table.go
- JSON (file or stdout) - via formatter_json.go
- YAML (file or stdout) - via formatter_yaml.go
- CSV (single file) - via formatter_csv.go
- Multi-CSV (multiple files) - via formatter_multi_csv.go
- XLSX (Excel workbook) - via formatter_xlsx.go

**API Endpoints Used**:
- No direct API calls - formats data from other components

### pkg/output/formatter_table.go
**Purpose**: Table output formatting for terminal display
**Key Features**:
- Categorized tabular presentation with section-based organization
- Enhanced storage class formatting with metadata
- Detailed node hardware information formatting
- Consistent table styling with `table.StyleColoredBright`
- ViewType support (NodeCentric vs ClusterWide)

**Formatting Patterns**:
- Single items: Categorized sections (System Information, Resource Capacity, etc.)
- Multiple items: Text-based format with indentation and separators

### pkg/output/formatter_json.go
**Purpose**: JSON output formatting
**Key Features**:
- Structured data with consistent camelCase naming
- Pretty-printed JSON for readability
- Support for all data types (VMs, nodes, storage, cluster info)

### pkg/output/formatter_yaml.go
**Purpose**: YAML output formatting
**Key Features**:
- Human-readable structured data
- Identical field names as JSON format
- Support for all data types

### pkg/output/formatter_csv.go
**Purpose**: Single CSV file output
**Key Features**:
- Exports data to single CSV file
- Flattened data structure for spreadsheet consumption
- Header row with column names

### pkg/output/formatter_multi_csv.go
**Purpose**: Multi-CSV file output (multiple related CSV files)
**Key Features**:
- Exports different data types to separate CSV files
- Related files with consistent naming (e.g., report_vms.csv, report_nodes.csv)
- Maintains relationships through common identifiers

### pkg/output/formatter_xlsx.go
**Purpose**: Excel (XLSX) workbook generation
**Key Features**:
- Multi-sheet Excel workbooks with separate sheets for different data types
- Formatted headers and cells
- Supports comprehensive reports with VMs, nodes, storage, and summary data
- Cell styling and column width optimization

### pkg/output/csv_helpers.go
**Purpose**: CSV formatting utility functions
**Key Features**:
- Common CSV writing operations
- Field escaping and quoting
- Row and column management utilities

### pkg/output/xlsx_helpers.go
**Purpose**: Excel formatting utility functions
**Key Features**:
- Sheet creation and management
- Cell styling and formatting
- Column width calculation
- Header row formatting

### pkg/output/table_helpers.go
**Purpose**: Table formatting utility functions
**Key Features**:
- Common table rendering operations
- Column alignment and padding
- Section header formatting
- Separator line generation

### pkg/output/report_generator.go
**Purpose**: Comprehensive report orchestration and data collection
**Key Features**:
- TestConnection() - Verifies cluster connectivity and KubeVirt availability
- GenerateStorageVolumesReport() - Collects storage volume information for all VMs
- GenerateComprehensiveReport() - Orchestrates collection of all cluster data (nodes, storage, VMs)
- Merges data from multiple sources into unified reports

**Data Flow**:
- Calls cluster, hardware, and vm packages to collect data
- Aggregates and correlates information across different resources
- Produces ComprehensiveReport structure for formatting

**API Endpoints Used**:
- Indirectly uses all API endpoints through domain packages (cluster, hardware, vm)

### pkg/output/types.go
**Purpose**: Output-related type definitions
**Key Features**:
- Formatter struct and configuration
- ComprehensiveReport structure
- ComprehensiveData container
- ViewType enumeration (NodeCentric, ClusterWide)
- ReportSummary structure
- SecurityInfo, ResourceMetrics, and supporting types

## Report Organization and Output Structure

### Report Categories and Information Architecture

The vm-scanner organizes collected data into logical categories that match how administrators think about cluster information:

#### **1. Hardware Information**
- **Scope**: Per-node data with cluster-wide aggregation
- **Content**: Node specifications, capacity, physical resources
- **Views**: 
  - Node-Centric: Detailed hardware per node
  - Cluster-Wide: Hardware summary table across all nodes

#### **2. Security Information**
- **Scope**: Cluster-wide (RBAC is cluster-scoped)
- **Content**: RBAC policies, identity providers, security configurations
- **Views**: 
  - Single view only (security policies apply cluster-wide)
  - Organized by: ClusterRoleBindings, IdentityProviders, SecurityContextConstraints

#### **3. Resource Utilization**
- **Scope**: Per-node metrics with cluster aggregation
- **Content**: CPU/Memory usage percentages, performance metrics
- **Views**:
  - Node-Centric: Utilization details per node
  - Cluster-Wide: Resource utilization summary and trends

#### **4. VM Information**
- **Scope**: Most flexible - can be organized by node or cluster-wide
- **Content**: Virtual machine details, configurations, status
- **Views**:
  - Node-Centric: VMs grouped under their host nodes
  - Cluster-Wide: All VMs in table format with node as attribute

### Output View Types

#### **Node-Centric View**
Organizes information by node first, then by category:
```
Node-1:
  Hardware:
    CPU Cores: 16
    Memory: 64GB
  Resource Utilization:
    CPU Usage: 45%
    Memory Usage: 60%
  Virtual Machines:
    VM-1: {...}
    VM-2: {...}

Node-2:
  Hardware: {...}
  Resource Utilization: {...}
  Virtual Machines: {...}

Security (Cluster-Wide):
  RBAC: {...}
  Identity Providers: {...}
```

#### **Cluster-Wide View**
Organizes information by category first, then by scope:
```
Hardware Summary:
  Node-1: 16 cores, 64GB RAM
  Node-2: 12 cores, 32GB RAM
  Total: 28 cores, 96GB RAM

Resource Utilization:
  CPU: 45% cluster average
  Memory: 52% cluster average

Security Configuration:
  RBAC Policies: 15 ClusterRoles, 8 ServiceAccounts
  Identity Providers: LDAP, OAuth

Virtual Machines:
  | VM Name | Node | CPU | Memory | Status |
  |---------|------|-----|---------|--------|
  | VM-1    | Node-1| 4   | 8GB    | Running|
  | VM-2    | Node-1| 2   | 4GB    | Running|
  | VM-3    | Node-2| 8   | 16GB   | Stopped|
```

### Formatter Architecture

#### **Section-Based Functions**
Each major category has dedicated formatting functions:

```go
// Main coordinators
func (f *Formatter) FormatNodeCentricReport(data *ComprehensiveData)
func (f *Formatter) FormatClusterWideReport(data *ComprehensiveData)

// Section specialists
func (f *Formatter) formatHardwareSection(nodes []NodeHardwareInfo, viewType ViewType)
func (f *Formatter) formatSecuritySection(security SecurityInfo)
func (f *Formatter) formatUtilizationSection(metrics []NodeMetrics, viewType ViewType)  
func (f *Formatter) formatVMSection(vms []VMInfo, viewType ViewType)
```

#### **Output Format Support**
Each section supports multiple output formats:
- **JSON**: Structured data with consistent camelCase naming (file or stdout)
- **YAML**: Human-readable structured data with identical field names as JSON (file or stdout)
- **Table (stdout)**: Categorized, formatted tables optimized for terminal viewing with consistent section headers and proper spacing
- **CSV**: Single CSV file output with flattened data structure
- **Multi-CSV**: Multiple related CSV files (e.g., vms.csv, nodes.csv, storage.csv)
- **XLSX**: Excel workbook with multiple sheets for different data types

#### **CLI Integration**
Users can control report organization through command-line flags:
```bash
# Connection testing
./vm-scanner -test-connection                            # Test cluster connection

# Section filtering
./vm-scanner -node-hardware-info                         # Node hardware information
./vm-scanner -storage-classes                            # Storage class information
./vm-scanner -vm-info                                    # VM instance information
./vm-scanner -vm-inventory                               # VM inventory report
./vm-scanner -storage-volumes                            # Storage volume information
./vm-scanner -kubevirt-version                           # KubeVirt version info

# Format selection
./vm-scanner -storage-classes -output-format=json -output-file=storage.json # JSON to file
./vm-scanner -node-hardware-info -output-format=yaml     # YAML to stdout
./vm-scanner -vm-info -output-format=csv -output-file=vms.csv # CSV to file
./vm-scanner -test-all-outputs -output-format=xlsx -output-file=report.xlsx # Excel workbook

# Combined operations
./vm-scanner -storage-classes -node-hardware-info        # Multiple sections
```

### Data Structure Design

#### **Comprehensive Data Container**
```go
type ComprehensiveData struct {
    // Per-node collections
    Nodes []NodeData `json:"nodes" yaml:"nodes"`
    
    // Cluster-wide collections  
    Security SecurityInfo `json:"security" yaml:"security"`
    
    // Flexible collections (can be viewed either way)
    VMs []VMInfo `json:"vms" yaml:"vms"`
    
    // Metadata
    GeneratedAt string `json:"generatedAt" yaml:"generatedAt"`
    ClusterName string `json:"clusterName" yaml:"clusterName"`
}

type NodeData struct {
    Name string `json:"name" yaml:"name"`
    Hardware NodeHardwareInfo `json:"hardware" yaml:"hardware"`
    Utilization ResourceMetrics `json:"utilization" yaml:"utilization"`
    VMs []VMInfo `json:"vms" yaml:"vms"` // For node-centric view
}
```

#### **View Type Enumeration**
```go
type ViewType string

const (
    NodeCentric ViewType = "node"
    ClusterWide ViewType = "cluster"
)
```

### Information Hierarchy Examples

#### **Hardware Section Organization**
```
SYSTEM INFORMATION:
  Node Name: worker-1
  OS Version: Red Hat Enterprise Linux CoreOS
  Kernel Version: 5.14.0-284.25.1.el9_2.x86_64

RESOURCE CAPACITY:
  CPU Cores: 16
  Memory: 64.00 GB
  Storage: 500.00 GB
  Pod Limits: 250

FILESYSTEM STATUS:
  Available: 387.50 GB
  Capacity: 450.00 GB
  Used: 62.50 GB
  Usage: 13.89%

NETWORK CONFIGURATION:
  Host CIDRs: 10.0.0.1/24, 192.168.1.0/24
  Schedulable: true
```

#### **Storage Class Section Organization**
**Single Storage Class:**
```
BASIC INFORMATION:
  Name: lvms-vg1
  Provisioner: topolvm.io
  Reclaim Policy: Delete
  Volume Binding Mode: WaitForFirstConsumer
  Default Storage Class: No
  Allow Volume Expansion: No
  Created At: 2025-10-06 10:16:07

PARAMETERS:
  csi.storage.k8s.io/fstype: xfs
  topolvm.io/device-class: vg1

ADVANCED CONFIGURATION:
  Mount Options: None
  Allowed Topologies: None
  Allowed Unsafe Evict Volumes: None
```

**Multiple Storage Classes:**
```
=== CLUSTER STORAGE CLASS INFORMATION ===
Total Storage Classes: 2

Storage Class: lvms-vg1

  BASIC INFORMATION:
    Name: lvms-vg1
    Provisioner: topolvm.io
    Reclaim Policy: Delete
    Volume Binding Mode: WaitForFirstConsumer
    Default Storage Class: No
    Allow Volume Expansion: No
    Created At: 2025-10-06 10:16:07

  PARAMETERS:
    csi.storage.k8s.io/fstype: xfs
    topolvm.io/device-class: vg1

  ADVANCED CONFIGURATION:
    Mount Options: None
    Allowed Topologies: None
    Allowed Unsafe Evict Volumes: None

================================================================================

Storage Class: gp2

  BASIC INFORMATION:
    Name: gp2
    Provisioner: ebs.csi.aws.com
    Reclaim Policy: Delete
    Volume Binding Mode: WaitForFirstConsumer
    Default Storage Class: Yes
    Allow Volume Expansion: Yes
    Created At: 2025-10-05 14:30:22

  PARAMETERS:
    type: gp2
    fsType: ext4

  ADVANCED CONFIGURATION:
    Mount Options: None
    Allowed Topologies: None
    Allowed Unsafe Evict Volumes: None
```

#### **VM Section Organization**
**Node-Centric View:**
```
Node worker-1:
  VM-web-app:
    Status: Running
    CPU: 4 cores
    Memory: 8GB
    
Node worker-2:
  VM-database:
    Status: Running  
    CPU: 8 cores
    Memory: 16GB
```

**Cluster-Wide View:**
```
| VM Name     | Node     | CPU | Memory | Status  | Uptime |
|-------------|----------|-----|--------|---------|--------|
| VM-web-app  | worker-1 | 4   | 8GB    | Running | 5d 3h  |
| VM-database | worker-2 | 8   | 16GB   | Running | 12d 6h |
```

## Key Data Points to Collect (RVTools Equivalent)

### Virtual Machine Information
- **Basic VM Data**: VM Name, Namespace, Labels, Annotations
- **Power State**: Running, Stopped, Pending, Failed, etc.
- **CPU Allocation**: Cores, sockets, threads, CPU model
- **Memory Allocation**: Requested memory, limits, usage
- **Operating System**: OS type, version, architecture
- **Timestamps**: Creation time, last modified, uptime
- **Node Assignment**: Node name, scheduling constraints, affinity rules

### Storage Information
- **Disk Configuration**: Disk names, sizes, types (ContainerDisk, PVC, etc.)
- **Volume Types**: ContainerDisk, DataVolume, PVC, HostDisk
- **Storage Classes**: Provisioner, parameters, reclaim policy, creation timestamp, volume expansion capability
- **Storage Class Metadata**: Mount options, allowed topologies, unsafe evict volume policies
- **Disk Usage**: Capacity, available space, usage percentage
- **Node Filesystem**: Available bytes, capacity bytes, used bytes, usage percentages
- **Snapshot Information**: Snapshot status, retention policies
- **Persistent Volumes**: PV details, access modes, storage class

### Network Information
- **Network Interfaces**: Interface names, models, MAC addresses
- **IP Addresses**: Pod IP, VM guest IP (if available)
- **Network Policies**: Security groups, ingress/egress rules
- **Service Exposure**: Service types, ports, load balancer configuration
- **CNI Configuration**: Network plugin, bridge configuration

### Resource Utilization
- **CPU Metrics**: Usage percentage, throttling, CPU time
- **Memory Metrics**: Usage, RSS, working set, page faults
- **Storage Metrics**: I/O operations, throughput, latency
- **Network Metrics**: Bytes in/out, packet rates, errors
- **Node Resources**: Available vs allocated CPU/memory

### Cluster Information
- **KubeVirt Version**: Operator version, API version, feature gates
- **Node Hardware**: CPU model, memory, storage, network interfaces
- **Storage Classes**: Available classes, default class, parameters
- **Network CNI**: CNI plugin, network policies, service mesh
- **Cluster Resources**: Total capacity, utilization, limits

## API Endpoints Summary

### Core Kubernetes APIs
- **`k8s.io/api/core/v1`**: Pod, Service, PersistentVolumeClaim, Event, Node, Namespace
- **`k8s.io/api/apps/v1`**: Deployment, ReplicaSet (for related resources)
- **`k8s.io/api/storage/v1`**: PersistentVolume, StorageClass (for storage information)
- **`k8s.io/api/networking/v1`**: NetworkPolicy, Ingress (for network policies)

### KubeVirt APIs
- **`kubevirt.io/api/core/v1`**: VirtualMachine, VirtualMachineInstance
- **`kubevirt.io/api/core/v1alpha3`**: VirtualMachineInstanceReplicaSet (if needed)
- **`kubevirt.io/api/core/v1`**: DataVolume, VirtualMachineSnapshot (for storage)

### Metrics APIs
- **`metrics.k8s.io/v1beta1`**: NodeMetrics, PodMetrics (if metrics server available)
- **`custom.metrics.k8s.io/v1beta1`**: Custom metrics (if available)

### Client Libraries
- **`k8s.io/client-go`**: Core Kubernetes client
- **`kubevirt.io/client-go`**: KubeVirt-specific client
- **`k8s.io/apimachinery`**: Kubernetes API machinery and types

## Usage Examples

### Test connection to cluster
```bash
./vm-scanner -test-connection
```

### Get VM information
```bash
./vm-scanner -vm-info
```

### Get VM inventory report
```bash
./vm-scanner -vm-inventory
```

### Get node hardware information
```bash
./vm-scanner -node-hardware-info
```

### Get storage class information
```bash
./vm-scanner -storage-classes
```

### Get storage volumes for all VMs
```bash
./vm-scanner -storage-volumes
```

### Get KubeVirt version
```bash
./vm-scanner -kubevirt-version
```

### Generate comprehensive Excel report
```bash
./vm-scanner -test-all-outputs -output-format=xlsx -output-file=cluster-report.xlsx
```

### Output to JSON file
```bash
./vm-scanner -vm-info -output-format=json -output-file=vms.json
```

### Output to YAML
```bash
./vm-scanner -node-hardware-info -output-format=yaml
```

### Output to CSV
```bash
./vm-scanner -storage-volumes -output-format=csv -output-file=volumes.csv
```

### Multiple operations with custom output
```bash
./vm-scanner -storage-classes -node-hardware-info -output-format=json -output-file=cluster-info.json
```

## Dependencies

### Required Go Modules
- `k8s.io/api v0.28.4` - Kubernetes API types
- `k8s.io/apimachinery v0.28.4` - Kubernetes API machinery
- `k8s.io/client-go v0.28.4` - Kubernetes client library
- `kubevirt.io/api v1.1.0` - KubeVirt API types
- `kubevirt.io/client-go v1.1.0` - KubeVirt client library

### Build Requirements
- Go 1.21 or later
- Access to a Kubernetes cluster with KubeVirt/OpenShift Virtualization
- Proper kubeconfig configuration
- RBAC permissions to read VirtualMachine resources

## RBAC Requirements

The service account running this tool needs the following permissions for comprehensive data collection:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: vm-scanner
rules:
# KubeVirt resources
- apiGroups: ["kubevirt.io"]
  resources: ["virtualmachines", "virtualmachineinstances", "datavolumes", "virtualmachinesnapshots"]
  verbs: ["get", "list"]
# Core Kubernetes resources
- apiGroups: [""]
  resources: ["pods", "services", "persistentvolumeclaims", "persistentvolumes", "events", "nodes", "namespaces"]
  verbs: ["get", "list"]
# Application resources
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets"]
  verbs: ["get", "list"]
# Storage resources
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses", "volumeattachments"]
  verbs: ["get", "list"]
# Network resources
- apiGroups: ["networking.k8s.io"]
  resources: ["networkpolicies", "ingresses"]
  verbs: ["get", "list"]
# Metrics resources (if metrics server is available)
- apiGroups: ["metrics.k8s.io"]
  resources: ["nodes", "pods"]
  verbs: ["get", "list"]
# Custom metrics (if available)
- apiGroups: ["custom.metrics.k8s.io"]
  resources: ["*"]
  verbs: ["get", "list"]
# Node information
- apiGroups: [""]
  resources: ["nodes/status"]
  verbs: ["get", "list"]
```

## Future Enhancements

### Phase 2 Features
- [x] Export to CSV format (completed)
- [x] Export to Excel/XLSX format (completed)
- [ ] VM performance metrics collection (partial - basic metrics implemented)
- [ ] Historical VM state tracking
- [ ] Custom resource definitions support
- [ ] Web-based dashboard
- [ ] Prometheus metrics export
- [ ] REST API endpoint (pkg/api prepared)

### Phase 3 Features
- [ ] VM migration planning
- [ ] Resource optimization recommendations
- [ ] Cost analysis and reporting
- [ ] Integration with external monitoring systems
- [ ] Automated VM lifecycle management
- [ ] Software inventory tracking (pkg/software prepared)

## Development Guidelines

### Code Style
- Follow Go standard formatting (`gofmt`)
- Use meaningful variable and function names
- Add comprehensive error handling
- Include unit tests for all public functions
- Document all exported functions and types

### Testing
- Unit tests for individual components
- Integration tests with mock Kubernetes API
- End-to-end tests with real cluster (optional)
- Performance tests for large VM counts

### Error Handling
- Graceful degradation when resources are unavailable
- Clear error messages for common issues
- Proper logging for debugging
- Timeout handling for long-running operations
