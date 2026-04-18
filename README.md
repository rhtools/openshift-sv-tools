# VM Scanner

A comprehensive Go-based tool for extracting VirtualMachine information from OpenShift Virtualization and KubeVirt clusters, providing **RVTools-equivalent functionality** for Kubernetes environments.

## Features

- **Comprehensive VM Information**: Power state, CPU/memory allocation, OS details, guest agent data, timestamps, labels
- **Storage Analysis**: Disk configuration, PVC inventory, volume types, storage classes, guest-reported usage metrics
- **Network Information**: Interface details, IP addresses, MAC addresses, NAD inventory
- **Resource Utilization**: CPU/memory metrics, storage I/O, capacity planning with overcommit ratios
- **Cluster Overview**: KubeVirt version, node hardware, storage classes, operator status, CNI configuration
- **Migration Readiness**: Per-VM assessment with blocker detection and readiness scoring
- **Multiple Output Formats**: Table, JSON, YAML, CSV, Multi-CSV, Excel (XLSX) with file export support

## Quick Start

### Prerequisites

- Go 1.24.0 or later
- Access to a Kubernetes cluster with KubeVirt/OpenShift Virtualization
- Valid kubeconfig or bearer token for authentication

### Installation

```bash
# Clone the repository
git clone https://github.com/<your-org>/openshift-sv-tools.git
cd openshift-sv-tools/vm-scanner

# Build the binary
make build

# Verify
./build/vm-scanner -help
```

### First Run

Test your connection to the cluster:

```bash
# Test connection with default kubeconfig
./build/vm-scanner -test-connection

# Test connection with specific kubeconfig
./build/vm-scanner -test-connection -kube-config=/path/to/kubeconfig
```

Generate a comprehensive Excel report:

```bash
./build/vm-scanner -test-all-outputs -output-format=xlsx -output-file=cluster-report.xlsx
```

## Example Output

### VM Inventory

```
=== VM RUNTIME INVENTORY ===
Total Running VMs: 5

┌──────────────────────┬────────────────┬─────────────┬──────────────────────────┬───────────────────────┬──────────────┐
│ NAME                 │ NAMESPACE      │ POWER STATE │ NODE                     │ OS                    │ MEMORY USED  │
├──────────────────────┼────────────────┼─────────────┼──────────────────────────┼───────────────────────┼──────────────┤
│ rhel9-web-server     │ production     │ Running     │ worker-01.example.com    │ Red Hat Enterprise 9  │ 1024.0 MiB   │
│ centos-stream-db     │ production     │ Running     │ worker-02.example.com    │ CentOS Stream 9       │ 2048.0 MiB   │
│ fedora-dev-box       │ development    │ Running     │ worker-01.example.com    │ Fedora 39             │ 512.0 MiB    │
│ win2022-legacy-app   │ legacy         │ Stopped     │ N/A                      │ Microsoft Windows     │ N/A          │
│ ubuntu-test          │ testing        │ Running     │ worker-03.example.com    │ Ubuntu 22.04          │ 768.0 MiB    │
└──────────────────────┴────────────────┴─────────────┴──────────────────────────┴───────────────────────┴──────────────┘
```

### KubeVirt Version

```
=== KUBEVIRT VERSION ===

┌──────────┬─────────────┐
│ PROPERTY │ VALUE       │
├──────────┼─────────────┤
│ Version  │ v1.3.1      │
│ Status   │ Deployed    │
└──────────┴─────────────┘
```

## Project Structure

```
vm-scanner/
├── cmd/
│   └── main.go                 # Application entry point with CLI flags
├── pkg/
│   ├── client/                 # Kubernetes/KubeVirt client management
│   ├── vm/                     # VirtualMachine operations and scanning
│   ├── hardware/               # Hardware detection (CPU, memory, network, storage)
│   ├── cluster/                # Cluster-wide information gathering
│   ├── config/                 # Configuration management
│   ├── output/                 # Output formatting (table, JSON, YAML, CSV, XLSX)
│   ├── utils/                  # Shared utility functions
│   └── error_handling/         # Centralized error handling
├── testdata/                   # Test data and fixtures
├── go.mod                      # Go module dependencies
├── Makefile                    # Build automation
├── DEV_README.md              # Developer setup guide
└── REFACTORING_GUIDE.md       # Contributor guidelines
```

## Documentation

| Document | Description |
|---|---|
| [Usage Guide](docs/USAGE.md) | Authentication, CLI reference, output formats, example output |
| [Architecture](docs/ARCHITECTURE.md) | Component flow, data pipeline, API endpoints, dependencies |
| [Code Flow](docs/CODEFLOW.md) | Detailed execution paths, data collection pipelines, mermaid diagrams |
| [Contributing](CONTRIBUTING.md) | Issue reporting, pull requests, build commands |
| [Developer Setup](vm-scanner/DEV_README.md) | Local development, running without compilation |
| [Refactoring Guide](vm-scanner/REFACTORING_GUIDE.md) | Code organization and architecture decisions |

## Development Status

Active development with core functionality implemented:

- VM scanning and information extraction
- Modular VM package architecture
- Hardware detection (CPU, memory, network, storage, baremetal)
- Multiple export formats (JSON, YAML, CSV, multi-CSV, Excel)
- Comprehensive reporting with Excel workbook generation (14 sheets)
- Cluster information collection (PVCs, NADs, DataVolumes, operators)
- Migration readiness assessment with blocker detection
- Capacity planning with overcommit ratio analysis
- Storage analysis with utilization flags

## License

This project is licensed under the [GNU Affero General Public License v3.0 (AGPL-3.0)](LICENSE).
