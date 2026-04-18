package cluster

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"vm-scanner/pkg/client"
	"vm-scanner/pkg/error_handling"
	"vm-scanner/pkg/utils"

	"vm-scanner/pkg/hardware"
	"vm-scanner/pkg/vm"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

// ProtectedNamespaces defines protected/system namespaces in OpenShift
var ProtectedNamespaces = []string{
	"default",
	"kube-public",
	"kube-system",
	"kube-node-lease",
	"openshift",
	"openshift-infra",
	"openshift-node",
	"openshift-apiserver",
	"openshift-apiserver-operator",
	"openshift-authentication",
	"openshift-authentication-operator",
	"openshift-catalogd",
	"openshift-cloud-controller-manager",
	"openshift-cloud-controller-manager-operator",
	"openshift-cloud-credential-operator",
	"openshift-cloud-network-config-controller",
	"openshift-cloud-platform-infra",
	"openshift-cluster-csi-drivers",
	"openshift-cluster-machine-approver",
	"openshift-cluster-node-tuning-operator",
	"openshift-cluster-olm-operator",
	"openshift-cluster-samples-operator",
	"openshift-cluster-storage-operator",
	"openshift-cluster-version",
	"openshift-cnv",
	"openshift-config",
	"openshift-config-managed",
	"openshift-config-operator",
	"openshift-console",
	"openshift-console-operator",
	"openshift-console-user-settings",
	"openshift-controller-manager",
	"openshift-controller-manager-operator",
	"openshift-dns",
	"openshift-dns-operator",
	"openshift-etcd",
	"openshift-etcd-operator",
	"openshift-host-network",
	"openshift-image-registry",
	"openshift-ingress",
	"openshift-ingress-canary",
	"openshift-ingress-operator",
	"openshift-insights",
	"openshift-kni-infra",
	"openshift-kube-apiserver",
	"openshift-kube-apiserver-operator",
	"openshift-kube-controller-manager",
	"openshift-kube-controller-manager-operator",
	"openshift-kube-scheduler",
	"openshift-kube-scheduler-operator",
	"openshift-kube-storage-version-migrator",
	"openshift-kube-storage-version-migrator-operator",
	"openshift-machine-api",
	"openshift-machine-config-operator",
	"openshift-marketplace",
	"openshift-monitoring",
	"openshift-multus",
	"openshift-network-console",
	"openshift-network-diagnostics",
	"openshift-network-node-identity",
	"openshift-network-operator",
	"openshift-node",
	"openshift-nutanix-infra",
	"openshift-oauth-apiserver",
	"openshift-openstack-infra",
	"openshift-operator-controller",
	"openshift-operator-lifecycle-manager",
	"openshift-operators",
	"openshift-ovirt-infra",
	"openshift-ovn-kubernetes",
	"openshift-route-controller-manager",
	"openshift-service-ca",
	"openshift-service-ca-operator",
	"openshift-user-workload-monitoring",
	"openshift-virtualization-os-images",
	"openshift-vsphere-infra",
}

func GetClusterSummary(ctx context.Context, k8sClient *client.ClusterClient) (*ClusterSummary, error) {
	nodes, err := GetClusterNodeInfo(ctx, k8sClient)
	if err != nil {
		return nil, err
	}
	kubevirtVersion, err := GetKubeVirtVersion(ctx, k8sClient)
	if err != nil {
		return nil, err
	}
	kubernetesVersion, err := GetKubernetesVersion(ctx, k8sClient)
	if err != nil {
		return nil, err
	}

	infrastructureGVR := schema.GroupVersionResource{
		Group:    "config.openshift.io",
		Version:  "v1",
		Resource: "infrastructures",
	}
	clusterVersionGVR := schema.GroupVersionResource{
		Group:    "config.openshift.io",
		Version:  "v1",
		Resource: "clusterversions",
	}

	clusterNameObj, err := k8sClient.Dynamic.Resource(infrastructureGVR).Get(ctx, "cluster", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	clusterName, found, err := unstructured.NestedString(clusterNameObj.Object, "status", "infrastructureName")
	error_handling.GetRequiredString(err, found, "status.infrastructureName")

	clusterIDObj, err := k8sClient.Dynamic.Resource(clusterVersionGVR).Get(ctx, "version", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	clusterID, found, err := unstructured.NestedString(clusterIDObj.Object, "spec", "clusterID")
	error_handling.GetRequiredString(err, found, "spec.clusterID")

	clusterVersion, found, err := unstructured.NestedString(clusterIDObj.Object, "status", "desired", "version")
	error_handling.GetRequiredString(err, found, "status.desired.version")

	hasSchedulableControlPlane, schedulableControlPlaneCount, workerNodesCount, err := GetClusterTopology(nodes)
	if err != nil {
		return nil, err
	}
	protectedNamespaces, userNamespaces, totalNamespaces, err := GetAllNamespaces(ctx, k8sClient)
	if err != nil {
		return nil, err
	}
	return &ClusterSummary{
		ClusterName:                  clusterName,
		ClusterID:                    clusterID,
		ClusterVersion:               clusterVersion,
		HasSchedulableControlPlane:   hasSchedulableControlPlane,
		SchedulableControlPlaneCount: schedulableControlPlaneCount,
		KubernetesVersion:            kubernetesVersion,
		KubeVirtVersion:              kubevirtVersion,
		ProtectedNamespaces:          protectedNamespaces,
		UserNamespaces:               userNamespaces,
		TotalNamespaces:              totalNamespaces,
		WorkerNodesCount:             workerNodesCount,
	}, nil
}

// GetKubernetesVersion retrieves the Kubernetes version from the cluster
func GetKubernetesVersion(ctx context.Context, k8sClient *client.ClusterClient) (string, error) {
	typedClient := k8sClient.Typed
	versionInfo, err := typedClient.Discovery().ServerVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get Kubernetes version: %w", err)
	}
	return versionInfo.GitVersion, nil
}

// GetKubeVirtVersion retrieves KubeVirt version and configuration
func GetKubeVirtVersion(ctx context.Context, k8sClient *client.ClusterClient) (*KubeVirtVersion, error) {
	// maybe oc get kubevirt -n openshift-cnv -o yaml
	// specifically reference the dynamic client
	// The dynamic client requires you to manually set the groupversionresource
	// the resource is the plural version of the kind:
	dynamicClient := k8sClient.Dynamic
	kubevirtGroupVersionResource := schema.GroupVersionResource{
		Group:    "kubevirt.io",
		Version:  "v1",
		Resource: "kubevirts",
	}
	kubevirtCR, err := dynamicClient.Resource(kubevirtGroupVersionResource).Namespace("openshift-cnv").Get(ctx, "kubevirt-kubevirt-hyperconverged", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	productVersion, found, err := unstructured.NestedString(kubevirtCR.Object, "spec", "productVersion")
	error_handling.GetRequiredString(err, found, "spec.productVersion")

	kubevirtDeployed, found, err := unstructured.NestedString(kubevirtCR.Object, "status", "phase")
	error_handling.GetRequiredString(err, found, "status.phase")

	return &KubeVirtVersion{
		Version:  productVersion,
		Deployed: kubevirtDeployed,
	}, nil
}

// GetClusterNodeInfo retrieves node hardware information
func GetClusterNodeInfo(ctx context.Context, k8sClient *client.ClusterClient) ([]ClusterNodeInfo, error) {
	// maybe oc get node -o yaml
	// Nodes is a std object so used the typed client to get the list of nodes
	typedClient := k8sClient.Typed
	nodes, err := typedClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var allNodesInfo []ClusterNodeInfo

	for _, node := range nodes.Items {
		hostname := node.ObjectMeta.Labels["kubernetes.io/hostname"]

		rawFSData, err := client.GetNodeRawAPIData(ctx, k8sClient, hostname)
		if err != nil {
			return nil, err
		}
		fsAvailable, fsCapacity, fsUsed, fsUsagePercent, err := hardware.ParseFileSystemStats(rawFSData)
		if err != nil {
			return nil, err
		}

		// Fetch BMH once per node -- non-fatal if unavailable (VM-based clusters)
		bmh, err := hardware.GetBareMetalHost(ctx, k8sClient, node)
		if err != nil {
			log.Printf("Warning: BMH not available for node %s (expected on VM-based clusters): %v", hostname, err)
		}

		cpuInfo, err := hardware.GetCPUInfo(node, bmh)
		if err != nil {
			return nil, err
		}

		memoryInfo, err := hardware.GetMemoryInfo(ctx, k8sClient, node, rawFSData)
		if err != nil {
			return nil, err
		}
		hostCIDRs, err := hardware.GetHostCIDRs(node)
		if err != nil {
			return nil, err
		}
		podNetworkSubnets, err := hardware.GetPodNetworkSubnets(node)
		if err != nil {
			return nil, err
		}

		physicalNames, err := hardware.ParsePhysicalInterfaceNames(rawFSData)
		if err != nil {
			return nil, err
		}

		nicDetails := hardware.ResolveNICDetails(ctx, k8sClient, hostname, bmh, physicalNames)

		networkGatewayConfig, err := hardware.GetL3GatewayConfig(node)
		if err != nil {
			return nil, err
		}

		var roles []string
		for labelKey := range node.ObjectMeta.Labels {
			if strings.HasPrefix(labelKey, "node-role.kubernetes.io/") {
				role := strings.TrimPrefix(labelKey, "node-role.kubernetes.io/")
				roles = append(roles, role)
			}
		}
		sort.Strings(roles)

		singleNodeInfo := ClusterNodeInfo{
			ClusterNodeName: hostname,
			CoreOSVersion:   node.Status.NodeInfo.OSImage,
			CPU:             *cpuInfo,
			Filesystem: hardware.NodeFilesystemInfo{
				FilesystemAvailable:    fsAvailable,
				FilesystemCapacity:     fsCapacity,
				FilesystemUsed:         fsUsed,
				FilesystemUsagePercent: fsUsagePercent,
			},
			Memory: *memoryInfo,
			Network: hardware.NetworkInfo{
				HostCIDRs:         hostCIDRs.HostCIDRs,
				MACAddress:        networkGatewayConfig.MACAddress,
				NextHops:          networkGatewayConfig.NextHops,
				NetworkInterfaces: nicDetails,
				PodNetworkSubnet:  podNetworkSubnets.PodNetworkSubnet,
				BridgeID:          networkGatewayConfig.BridgeID,
				InterfaceID:       networkGatewayConfig.InterfaceID,
				IPAddresses:       networkGatewayConfig.IPAddresses,
				Mode:              networkGatewayConfig.Mode,
				NodePortEnable:    networkGatewayConfig.NodePortEnable,
				VLANID:            networkGatewayConfig.VLANID,
			},
			NodeKernelVersion: node.Status.NodeInfo.KernelVersion,
			NodePodLimits:     node.Status.Capacity.Pods().Value(),
			NodeRoles:         roles,
			NodeSchedulable:   node.ObjectMeta.Labels["kubevirt.io/schedulable"],
			StorageCapacity:   utils.BytesToGiB(node.Status.Capacity.Storage().Value()),
		}
		allNodesInfo = append(allNodesInfo, singleNodeInfo)
	}
	return allNodesInfo, nil
}

// GetCNIConfiguration retrieves CNI configuration
func GetCNIConfiguration(ctx context.Context, k8sClient kubernetes.Interface) (*CNIConfiguration, error) {
	// Implementation will be added here
	// maybe oc get networks.operator.openshift.io cluster -o yaml
	return nil, fmt.Errorf("not implemented")
}

// I need to parse out and track which nodes are control plane, and if they are schedulable or not.
// I am already tracking the node roles, so I just need to parse out the control plane nodes and check if they are schedulable or not.
func GetClusterTopology(nodes []ClusterNodeInfo) (bool, int, int, error) {
	var hasSchedulableControlPlane bool
	var schedulableControlPlaneCount int
	var totalWorkerNodes int
	for _, node := range nodes {
		isControlPlane := containsRole(node.NodeRoles, "control-plane") || containsRole(node.NodeRoles, "master")
		if isControlPlane && node.NodeSchedulable == "true" {
			hasSchedulableControlPlane = true
			schedulableControlPlaneCount++
		}
		if containsRole(node.NodeRoles, "worker") {
			totalWorkerNodes++
		}
	}
	if hasSchedulableControlPlane {
		return true, schedulableControlPlaneCount, totalWorkerNodes, nil
	}
	return false, 0, 0, nil
}

func containsRole(roles []string, targetRole string) bool {
	for _, role := range roles {
		if role == targetRole {
			return true
		}
	}
	return false
}

func CalculateClusterResources(nodes []ClusterNodeInfo) (*ClusterResources, error) {
	var totalCPU int64
	var totalMemoryGiB float64
	var totalStorageGiB float64
	var usedMemoryGiB float64
	var totalUsedStorageGiB float64

	for _, node := range nodes {
		totalCPU += node.CPU.CPUCores
		totalMemoryGiB += node.Memory.MemoryCapacityGiB
		usedMemoryGiB += node.Memory.MemoryUsedGiB
		totalStorageGiB += node.Filesystem.FilesystemCapacity
		totalUsedStorageGiB += node.Filesystem.FilesystemUsed
	}

	return &ClusterResources{
		TotalCPU:                         totalCPU,
		TotalMemory:                      utils.RoundToOneDecimal(totalMemoryGiB),
		TotalLocalStorage:                utils.RoundToOneDecimal(totalStorageGiB),
		TotalLocalStorageUsed:            utils.RoundToOneDecimal(totalUsedStorageGiB),
		TotalApplicationRequestedStorage: 0,
		TotalApplicationUsedStorage:      0,
		UsedMemory:                       utils.RoundToOneDecimal(usedMemoryGiB),
	}, nil
}

// CalculateVMStorageTotals calculates the total requested and used storage across all VMs
// Returns requestedGiB (allocated PVC storage) and usedGiB (actual filesystem usage from guest agent)
func CalculateVMStorageTotals(vms []vm.VMDetails) (requestedGiB int64, usedGiB int64) {
	var totalRequestedBytes int64
	var totalUsedBytes int64

	for _, vmDetail := range vms {
		for _, disk := range vmDetail.Disks {
			totalRequestedBytes += disk.SizeBytes
			totalUsedBytes += disk.TotalStorageInUse
		}
	}

	const bytesPerGiB = 1024 * 1024 * 1024
	requestedGiB = totalRequestedBytes / bytesPerGiB
	usedGiB = totalUsedBytes / bytesPerGiB

	return requestedGiB, usedGiB
}

func GetAllNamespaces(ctx context.Context, k8sClient *client.ClusterClient) (int, int, int, error) {
	namespaces, err := k8sClient.Typed.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, 0, 0, err
	}

	protectedNamespaces, userNamespaces, totalNamespaces, err := SortNamespaces(namespaces.Items)
	if err != nil {
		return 0, 0, 0, err
	}

	return protectedNamespaces, userNamespaces, totalNamespaces, nil
}

func SortNamespaces(namespaces []v1.Namespace) (int, int, int, error) {
	var protectedNamespaces int
	var userNamespaces int
	var totalNamespaces int

	for _, namespace := range namespaces {
		if containsNamespace(ProtectedNamespaces, namespace.Name) {
			protectedNamespaces++
		} else {
			userNamespaces++
		}
		totalNamespaces++
	}

	return protectedNamespaces, userNamespaces, totalNamespaces, nil
}

func containsNamespace(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
