package vm

import (
	"fmt"
	"vm-scanner/pkg/utils"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// interfaceTypeFromMap detects the interface type from the YAML key (bridge, masquerade, slirp, etc.).
func interfaceTypeFromMap(ifaceMap map[string]interface{}) string {
	for _, key := range []string{"bridge", "masquerade", "slirp", "sriov", "tap"} {
		if _, ok := ifaceMap[key]; ok {
			return key
		}
	}
	return ""
}

// modelFromInterface extracts the NIC model, handling both plain string and nested object formats.
func modelFromInterface(ifaceMap map[string]interface{}) string {
	modelVal, ok := ifaceMap["model"]
	if !ok || modelVal == nil {
		return ""
	}
	if s, ok := modelVal.(string); ok && s != "" {
		return s
	}
	modelObj, ok := modelVal.(map[string]interface{})
	if !ok {
		return ""
	}
	if t, ok := modelObj["type"].(string); ok {
		return t
	}
	return ""
}

// buildNetworkToNADMap builds a map from network name to NAD name from the networks list.
// Pod networks map to "None" since they are not backed by a NAD.
func buildNetworkToNADMap(networks []interface{}) map[string]string {
	result := make(map[string]string)
	for _, net := range networks {
		netMap, ok := net.(map[string]interface{})
		if !ok {
			continue
		}
		netName, _, _ := unstructured.NestedString(netMap, "name")
		if multusRaw, found, _ := unstructured.NestedFieldNoCopy(netMap, "multus"); found && multusRaw != nil {
			if multusMap, ok := multusRaw.(map[string]interface{}); ok {
				if nadName, _, _ := unstructured.NestedString(multusMap, "networkName"); nadName != "" {
					result[netName] = nadName
				}
			}
		}
		if podNetRaw, found, _ := unstructured.NestedFieldNoCopy(netMap, "pod"); found && podNetRaw != nil {
			result[netName] = "None"
		}
	}
	return result
}

// enrichInterfaceFromSpec extracts Type, Model, Network, and NAD from an interface
// spec map using the network-to-NAD lookup. Used by both ReturnNetworkInterfaceMap
// (for stopped VMs) and GetVMNetworkInterfaces (for running VMIs).
func enrichInterfaceFromSpec(ifaceMap map[string]interface{}, networkToNAD map[string]string) (ifaceType, model, netName, nad string) {
	ifaceType = interfaceTypeFromMap(ifaceMap)
	model = modelFromInterface(ifaceMap)
	netName, _, _ = unstructured.NestedString(ifaceMap, "network")
	if netName == "" {
		netName, _, _ = unstructured.NestedString(ifaceMap, "name")
	}
	nad = networkToNAD[netName]
	return
}

func ReturnNetworkInterfaceMap(interfaces []interface{}, networks []interface{}, vmi bool) (map[string]VMNetworkInfo, error) {
	interfaceInfo := make(map[string]VMNetworkInfo)
	var macAddressPath string
	if vmi {
		macAddressPath = "mac"
	} else {
		macAddressPath = "macAddress"
	}

	networkToNAD := buildNetworkToNADMap(networks)

	for _, iface := range interfaces {
		ifaceMap, ok := iface.(map[string]interface{})
		if !ok {
			continue
		}
		interfaceName, found, err := unstructured.NestedString(ifaceMap, "name")
		if err != nil {
			return nil, fmt.Errorf("failed to get interface name: %w", err)
		}
		if !found {
			return nil, fmt.Errorf("interface 'name' field not found in VM spec")
		}
		macAddress, found, err := unstructured.NestedString(ifaceMap, macAddressPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get macAddress: %w", err)
		}
		if !found {
			return nil, fmt.Errorf("'macAddress' field not found in VM spec for interface %s", interfaceName)
		}

		ifaceType, model, netName, nadName := enrichInterfaceFromSpec(ifaceMap, networkToNAD)

		var ipAddresses []string
		if vmi {
			ipAddressesRaw, found, _ := unstructured.NestedSlice(ifaceMap, "ipAddresses")
			if found {
				ipAddresses = utils.InterfaceSliceToStringSlice(ipAddressesRaw)
			}
		}

		interfaceInfo[interfaceName] = VMNetworkInfo{
			Name:                        interfaceName,
			MACAddress:                  macAddress,
			IPAddresses:                 ipAddresses,
			Type:                        ifaceType,
			Model:                       model,
			Network:                     netName,
			NetworkAttachmentDefinition: nadName,
		}
	}
	return interfaceInfo, nil
}

// GetVMNetworkInterfaces reads Type/Model/NAD from spec and MAC/IP from status for VMI.
func GetVMNetworkInterfaces(vmUnstructured unstructured.Unstructured) (map[string]VMNetworkInfo, error) {
	statusInterfaces, _, err := unstructured.NestedSlice(vmUnstructured.Object, "status", "interfaces")
	if err != nil {
		return nil, fmt.Errorf("failed to get status interfaces: %w", err)
	}

	interfaceInfo := make(map[string]VMNetworkInfo)
	for _, iface := range statusInterfaces {
		ifaceMap, ok := iface.(map[string]interface{})
		if !ok {
			continue
		}
		interfaceName, found, err := unstructured.NestedString(ifaceMap, "name")
		if err != nil || !found {
			continue
		}
		macAddress, _, _ := unstructured.NestedString(ifaceMap, "mac")
		ipAddressesRaw, _, _ := unstructured.NestedSlice(ifaceMap, "ipAddresses")
		ipAddresses := utils.InterfaceSliceToStringSlice(ipAddressesRaw)

		interfaceInfo[interfaceName] = VMNetworkInfo{
			Name:        interfaceName,
			MACAddress:  macAddress,
			IPAddresses: ipAddresses,
		}
	}

	specInterfaces, _, err := unstructured.NestedSlice(vmUnstructured.Object, "spec", "domain", "devices", "interfaces")
	if err != nil {
		return interfaceInfo, nil
	}
	specNetworks, _, err := unstructured.NestedSlice(vmUnstructured.Object, "spec", "networks")
	if err != nil {
		specNetworks = []interface{}{}
	}

	networkToNAD := buildNetworkToNADMap(specNetworks)

	for _, iface := range specInterfaces {
		ifaceMap, ok := iface.(map[string]interface{})
		if !ok {
			continue
		}
		ifaceName, found, err := unstructured.NestedString(ifaceMap, "name")
		if err != nil || !found {
			continue
		}
		ifaceType, model, netName, nadName := enrichInterfaceFromSpec(ifaceMap, networkToNAD)

		if info, exists := interfaceInfo[ifaceName]; exists {
			info.Type = ifaceType
			info.Model = model
			info.Network = netName
			info.NetworkAttachmentDefinition = nadName
			interfaceInfo[ifaceName] = info
		}
	}

	return interfaceInfo, nil
}

func GetNetworkInterfaces(vmiUnstructured unstructured.Unstructured, vmi bool) (map[string]VMNetworkInfo, error) {
	var interfaces []interface{}
	var networks []interface{}
	var err error

	if vmi {
		interfaces, _, err = unstructured.NestedSlice(vmiUnstructured.Object, "spec", "domain", "devices", "interfaces")
		if err != nil {
			interfaces = []interface{}{}
		}
		networks, _, err = unstructured.NestedSlice(vmiUnstructured.Object, "spec", "networks")
		if err != nil {
			networks = []interface{}{}
		}
	} else {
		interfaces, _, err = unstructured.NestedSlice(vmiUnstructured.Object, "spec", "template", "spec", "domain", "devices", "interfaces")
		if err != nil {
			interfaces = []interface{}{}
		}
		networks, _, err = unstructured.NestedSlice(vmiUnstructured.Object, "spec", "template", "spec", "networks")
		if err != nil {
			networks = []interface{}{}
		}
	}

	networkInterfaces, err := ReturnNetworkInterfaceMap(interfaces, networks, vmi)
	if err != nil {
		return nil, fmt.Errorf("failed to parse network interfaces: %w", err)
	}

	return networkInterfaces, nil
}
