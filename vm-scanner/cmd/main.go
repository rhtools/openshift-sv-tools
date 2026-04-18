package main

import (
	"context"
	"fmt"
	"log"

	"vm-scanner/pkg/client"
	"vm-scanner/pkg/cluster"
	"vm-scanner/pkg/config"
	"vm-scanner/pkg/hardware"
	"vm-scanner/pkg/output"
	"vm-scanner/pkg/vm"
)

// shouldPrintStatus determines if status messages should be printed based on output format
func shouldPrintStatus(format string) bool {
	return format == "table" || format == ""
}

// runTestConnection tests the connection to the cluster
func runTestConnection(ctx context.Context, clusterClient *client.ClusterClient) error {
	_, err := output.TestConnection(ctx, clusterClient)
	return err
}

// runKubevirtVersion fetches and displays KubeVirt version information
func runKubevirtVersion(ctx context.Context, clusterClient *client.ClusterClient, formatter *output.Formatter, cfg *config.Config) error {
	if shouldPrintStatus(cfg.OutputFormat) {
		fmt.Print("\n\nGetting KubeVirt version...\n\n")
	}
	kubevirtVersion, err := cluster.GetKubeVirtVersion(ctx, clusterClient)
	if err != nil {
		return fmt.Errorf("failed to get KubeVirt version: %w", err)
	}
	if err := formatter.Format(kubevirtVersion); err != nil {
		return fmt.Errorf("failed to format KubeVirt version: %w", err)
	}
	return nil
}

// runNodeHardwareInfo fetches and formats node hardware information
func runNodeHardwareInfo(ctx context.Context, clusterClient *client.ClusterClient, formatter *output.Formatter, cfg *config.Config) error {
	if shouldPrintStatus(cfg.OutputFormat) {
		fmt.Print("\n\nGetting node hardware info...\n\n")
	}
	nodeHardwareInfo, err := cluster.GetClusterNodeInfo(ctx, clusterClient)
	if err != nil {
		return fmt.Errorf("failed to get node hardware info: %w", err)
	}
	if err := formatter.Format(nodeHardwareInfo); err != nil {
		return fmt.Errorf("failed to format node hardware info: %w", err)
	}
	return nil
}

// runStorageClasses fetches and formats storage class information
func runStorageClasses(ctx context.Context, clusterClient *client.ClusterClient, formatter *output.Formatter, cfg *config.Config) error {
	if shouldPrintStatus(cfg.OutputFormat) {
		fmt.Print("\n\nGetting storage classes...\n\n")
	}
	storageClasses, err := hardware.GetStorageClasses(ctx, clusterClient)
	if err != nil {
		return fmt.Errorf("failed to get storage classes: %w", err)
	}
	if err := formatter.Format(storageClasses); err != nil {
		return fmt.Errorf("failed to format storage classes: %w", err)
	}
	return nil
}

// runVMInfo fetches and displays VM instance information
func runVMInfo(ctx context.Context, clusterClient *client.ClusterClient, formatter *output.Formatter, cfg *config.Config) error {
	if shouldPrintStatus(cfg.OutputFormat) {
		fmt.Print("\n\nGetting information about VMs...\n\n")
	}
	vmis, err := vm.GetAllVMIs(ctx, clusterClient)
	if err != nil {
		return fmt.Errorf("failed to get information about VMs: %w", err)
	}
	if err := formatter.Format(vmis); err != nil {
		return fmt.Errorf("failed to format VM info: %w", err)
	}
	return nil
}

// runStorageVolumes fetches and displays storage volume information for all VMs
func runStorageVolumes(ctx context.Context, clusterClient *client.ClusterClient, formatter *output.Formatter) error {
	allStorageVolumes, err := output.GenerateStorageVolumesReport(ctx, clusterClient)
	if err != nil {
		return err
	}
	if err := formatter.Format(allStorageVolumes); err != nil {
		return fmt.Errorf("failed to format storage volumes: %w", err)
	}
	return nil
}

// runVMInventory fetches and displays VM inventory
func runVMInventory(ctx context.Context, clusterClient *client.ClusterClient, formatter *output.Formatter, cfg *config.Config) error {
	if shouldPrintStatus(cfg.OutputFormat) {
		fmt.Print("\n\nGetting VM inventory...\n\n")
	}
	vmInventory, err := vm.BuildVMIInventory(ctx, clusterClient)
	if err != nil {
		return fmt.Errorf("failed to get VM inventory: %w", err)
	}
	if err := formatter.Format(vmInventory); err != nil {
		return fmt.Errorf("failed to format VM inventory: %w", err)
	}
	return nil
}

// runComprehensiveReport generates and formats a comprehensive cluster report
func runComprehensiveReport(ctx context.Context, clusterClient *client.ClusterClient, cfg *config.Config) error {
	if shouldPrintStatus(cfg.OutputFormat) {
		fmt.Print("\n\nGenerating comprehensive report...\n\n")
	}
	report, err := output.GenerateComprehensiveReport(ctx, clusterClient)
	if err != nil {
		return err
	}

	// Determine output file name if not provided
	outputFile := cfg.OutputFile
	if outputFile == "" {
		if cfg.OutputFormat == "xlsx" || cfg.OutputFormat == "excel" {
			outputFile = "comprehensive-report.xlsx"
		} else if cfg.OutputFormat == "csv" {
			outputFile = "comprehensive-report.csv"
		}
	}

	formatter := output.NewFormatter(cfg.OutputFormat, outputFile)

	// Use appropriate formatting based on output type
	switch cfg.OutputFormat {
	case "xlsx", "excel":
		return formatter.FormatXLSX(report)
	case "csv":
		return formatter.FormatMultiCSV(report)
	default:
		return formatter.Format(report)
	}
}

// testCurrentFeature is a dynamic function for testing whatever feature is currently in development
func testCurrentFeature(ctx context.Context, clusterClient *client.ClusterClient) error {
	fmt.Print("\n\n=== TESTING CURRENT FEATURE: GetGuestOSInfo ===\n\n")

	vmUnstructured, err := vm.GetAllVMIs(ctx, clusterClient)
	if err != nil {
		return fmt.Errorf("failed to get VMIs: %w", err)
	}
	for _, vmData := range vmUnstructured.Items {
		networkInterfaces, err := vm.GetNetworkInterfaces(vmData, true)
		if err != nil {
			return fmt.Errorf("failed to get network interfaces: %w", err)
		}
		fmt.Println(networkInterfaces)
	}
	return nil
}

// initializeClient creates and returns a ClusterClient based on configuration
func initializeClient(cfg *config.Config) (*client.ClusterClient, error) {
	var authOpts client.AuthOptions

	if cfg.AuthOptions.UseInCluster {
		authOpts = client.AuthOptions{UseInCluster: true}
	} else if cfg.AuthOptions.Kubeconfig != "" {
		authOpts = client.AuthOptions{Kubeconfig: cfg.AuthOptions.Kubeconfig}
	} else if cfg.AuthOptions.Token != "" {
		authOpts = client.AuthOptions{Token: cfg.AuthOptions.Token}
	} else if cfg.AuthOptions.APIURL != "" {
		authOpts = client.AuthOptions{APIURL: cfg.AuthOptions.APIURL}
	} else {
		return nil, fmt.Errorf("invalid authentication method")
	}

	return client.NewClusterClient(authOpts)
}

// executeCommand routes to the appropriate command handler based on configuration
func executeCommand(ctx context.Context, cfg *config.Config, clusterClient *client.ClusterClient) error {
	// Create formatter for commands that need it
	formatter := output.NewFormatter(cfg.OutputFormat, cfg.OutputFile)

	switch {
	case cfg.TestConnection:
		return runTestConnection(ctx, clusterClient)

	case cfg.TestAllOutputs:
		return runComprehensiveReport(ctx, clusterClient, cfg)

	case cfg.KubevirtVersion:
		return runKubevirtVersion(ctx, clusterClient, formatter, cfg)

	case cfg.NodeHardwareInfo:
		return runNodeHardwareInfo(ctx, clusterClient, formatter, cfg)

	case cfg.StorageClasses:
		return runStorageClasses(ctx, clusterClient, formatter, cfg)

	case cfg.VMInfo:
		return runVMInfo(ctx, clusterClient, formatter, cfg)

	case cfg.StorageVolumes:
		return runStorageVolumes(ctx, clusterClient, formatter)

	case cfg.VMInventory:
		return runVMInventory(ctx, clusterClient, formatter, cfg)

	case cfg.TestCurrentFeat:
		return testCurrentFeature(ctx, clusterClient)

	default:
		return fmt.Errorf("no operation specified - use --help for available options")
	}
}

func main() {
	cfg := config.FromCLIFlags()
	ctx := context.Background()

	clusterClient, err := initializeClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	if err := executeCommand(ctx, cfg, clusterClient); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
