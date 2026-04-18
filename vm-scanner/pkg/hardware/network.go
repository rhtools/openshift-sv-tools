package hardware

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"vm-scanner/pkg/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GetPodNetworkSubnets retrieves pod network subnets from node annotations
func GetPodNetworkSubnets(node corev1.Node) (*NetworkInfo, error) {
	var nodeSubnets map[string][]string
	subnetStr := node.ObjectMeta.Annotations["k8s.ovn.org/node-subnets"]
	if subnetStr != "" {
		json.Unmarshal([]byte(subnetStr), &nodeSubnets)
	}
	return &NetworkInfo{
		PodNetworkSubnet: nodeSubnets,
	}, nil
}

// GetHostCIDRs retrieves host CIDRs from node annotations
func GetHostCIDRs(node corev1.Node) (*NetworkInfo, error) {
	var hostCIDRs []string
	cidrStr := node.ObjectMeta.Annotations["k8s.ovn.org/host-cidrs"]
	if cidrStr != "" {
		json.Unmarshal([]byte(cidrStr), &hostCIDRs)
	}
	return &NetworkInfo{
		HostCIDRs: hostCIDRs,
	}, nil
}

// GetL3GatewayConfig retrieves L3 gateway configuration from node annotations
func GetL3GatewayConfig(node corev1.Node) (*NetworkInfo, error) {
	var L3GatewayConfig L3GatewayAnnotation
	L3GatewayConfigStr := node.ObjectMeta.Annotations["k8s.ovn.org/l3-gateway-config"]
	if L3GatewayConfigStr != "" {
		json.Unmarshal([]byte(L3GatewayConfigStr), &L3GatewayConfig)
	}
	config := L3GatewayConfig.Default
	return &NetworkInfo{
		Mode:           config.Mode,
		BridgeID:       config.BridgeID,
		InterfaceID:    config.InterfaceID,
		MACAddress:     config.MACAddress,
		IPAddresses:    config.IPAddresses,
		NextHops:       config.NextHops,
		NodePortEnable: config.NodePortEnable,
		VLANID:         config.VLANID,
	}, nil
}

func filterPhysicalInterfaces(interfaces []string) []string {
	var physical []string

	for _, iface := range interfaces {
		if isPhysicalInterface(iface) {
			physical = append(physical, iface)
		}
	}

	return physical
}

// ParsePhysicalInterfaceNames parses physical interface names from raw node stats API response
func ParsePhysicalInterfaceNames(rawAPIData []byte) ([]string, error) {
	var interfaceNames NodeStatsResponse
	err := json.Unmarshal(rawAPIData, &interfaceNames)
	if err != nil {
		return nil, fmt.Errorf("failed to parse network stats: %w", err)
	}
	var allInterfaces []string
	for _, iface := range interfaceNames.Node.Network.Interfaces {
		allInterfaces = append(allInterfaces, iface.Name)
	}
	return filterPhysicalInterfaces(allInterfaces), nil
}

func isPhysicalInterface(name string) bool {
	// Physical interface patterns (predictable network interface naming)
	physicalPatterns := []string{
		"^en[pso]",     // enp1s0, eno1, ens3 (Ethernet)
		"^wl[pso]",     // wlp2s0, wlo1, wls1 (Wireless)
		"^eth[0-9]+$",  // eth0, eth1 (traditional)
		"^wlan[0-9]+$", // wlan0, wlan1 (wireless)
		"^br-",         // br-int, br-ex
	}

	// Known virtual/bridge interfaces to exclude
	virtualPatterns := []string{
		"^ovs-", // ovs-system
		"^ovn-", // ovn-k8s-mp0
		// "^br-",    // br-int, br-ex
		"^veth",   // veth pairs
		"^docker", // Docker interfaces
		"^cni",    // CNI interfaces
		"^lo$",    // loopback
	}

	// Exclude virtual interfaces
	for _, pattern := range virtualPatterns {
		if matched, _ := regexp.MatchString(pattern, name); matched {
			return false
		}
	}

	// Exclude random hex strings (container interfaces)
	if len(name) >= 10 && isHexString(name) {
		return false
	}

	// Include known physical patterns
	for _, pattern := range physicalPatterns {
		if matched, _ := regexp.MatchString(pattern, name); matched {
			return true
		}
	}

	return false
}

func isHexString(s string) bool {
	for _, char := range s {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}

// GetClusterNetworkConfig retrieves cluster-wide network configuration
func GetClusterNetworkConfig(ctx context.Context, k8sClient *client.ClusterClient) (*ClusterNetworkConfig, error) {
	networkGVR := schema.GroupVersionResource{
		Group:    "config.openshift.io",
		Version:  "v1",
		Resource: "networks",
	}
	networkObj, err := k8sClient.Dynamic.Resource(networkGVR).Get(ctx, "cluster", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	spec, ok := networkObj.Object["spec"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("spec not found")
	}
	serviceNetwork, ok := spec["serviceNetwork"].([]string)
	if !ok {
		return nil, fmt.Errorf("serviceNetwork not found")
	}
	clusterNetworkMTU, ok := spec["clusterNetworkMTU"].(int)
	if !ok {
		return nil, fmt.Errorf("clusterNetworkMTU not found")
	}
	clusterNetworkType, ok := spec["clusterNetworkType"].(string)
	if !ok {
		return nil, fmt.Errorf("clusterNetworkType not found")
	}
	externalIP, ok := spec["externalIP"].(string)
	if !ok {
		return nil, fmt.Errorf("externalIP not found")
	}

	return &ClusterNetworkConfig{
		ServiceNetwork:     serviceNetwork,
		ClusterNetworkMTU:  clusterNetworkMTU,
		ClusterNetworkType: clusterNetworkType,
		ExternalIP:         externalIP,
	}, nil
}

// GetVMNetworkInfo retrieves network information for a specific VM
func GetVMNetworkInfo(ctx context.Context, namespace, vmName string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetCNIConfiguration retrieves CNI configuration information
func GetCNIConfiguration(ctx context.Context) ([]string, error) {
	// Implementation will be added here
	return nil, fmt.Errorf("not implemented")
}
