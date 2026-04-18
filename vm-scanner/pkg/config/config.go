package config

import (
	"flag"
	"vm-scanner/pkg/client"
)

type Config struct {
	// Embed existing AuthOptions - like inheritance in Python
	client.AuthOptions

	// Add additional CLI/API specific fields
	OutputFormat     string `json:"outputFormat" yaml:"outputFormat"`
	OutputFile       string `json:"outputFile" yaml:"outputFile"`
	TestConnection   bool   `json:"testConnection" yaml:"testConnection"`
	NodeHardwareInfo bool   `json:"nodeHardwareInfo" yaml:"nodeHardwareInfo"`
	KubevirtVersion  bool   `json:"kubevirtVersion" yaml:"kubevirtVersion"`
	StorageClasses   bool   `json:"storageClasses" yaml:"storageClasses"`
	StorageVolumes   bool   `json:"storageVolumes" yaml:"storageVolumes"`
	VMInfo           bool   `json:"vmInfo" yaml:"vmInfo"`
	TestAllOutputs   bool   `json:"testAllOutputs" yaml:"testAllOutputs"`
	VMInventory      bool   `json:"vmInventory" yaml:"vmInventory"`
	TestCurrentFeat  bool   `json:"testCurrentFeat" yaml:"testCurrentFeat"`
}

func FromCLIFlags() *Config {
	var (
		// Auth flags
		authMethod     = flag.String("auth-method", "kubeconfig", "Authentication method")
		kubeConfigFile = flag.String("kube-config", "", "Full path to kubeconfig")
		token          = flag.String("token", "", "Bearer token")
		apiURL         = flag.String("api-url", "", "API URL")

		// Operation flags
		outputFormat     = flag.String("output-format", "stdout", "Output format")
		outputFile       = flag.String("output-file", "", "Output file path (for xlsx, csv formats)")
		nodeHardwareInfo = flag.Bool("node-hardware-info", false, "Get node hardware info")
		kubevirtVersion  = flag.Bool("kubevirt-version", false, "Get KubeVirt version")
		storageClasses   = flag.Bool("storage-classes", false, "Get storage classes")
		storageVolumes   = flag.Bool("storage-volumes", false, "Get storage classes")
		testAllOutputs   = flag.Bool("test-all-outputs", false, "Test all output outputs")
		testConnection   = flag.Bool("test-connection", false, "Test connection")
		vmInventory      = flag.Bool("vm-inventory", false, "Get inventory of VMs")
		vmInfo           = flag.Bool("vm-info", false, "Get Information About VMs")
		testCurrentFeat  = flag.Bool("test-current-feat", false, "Test current feature in development")
	)
	flag.Parse()

	return &Config{
		AuthOptions: client.AuthOptions{
			Kubeconfig:   *kubeConfigFile,
			Token:        *token,
			APIURL:       *apiURL,
			UseInCluster: *authMethod == "in-cluster",
		},
		OutputFormat:     *outputFormat,
		OutputFile:       *outputFile,
		TestConnection:   *testConnection,
		NodeHardwareInfo: *nodeHardwareInfo,
		KubevirtVersion:  *kubevirtVersion,
		StorageClasses:   *storageClasses,
		StorageVolumes:   *storageVolumes,
		VMInfo:           *vmInfo,
		TestAllOutputs:   *testAllOutputs,
		VMInventory:      *vmInventory,
		TestCurrentFeat:  *testCurrentFeat,
	}
}
