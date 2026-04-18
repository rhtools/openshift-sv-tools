package hardware

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"vm-scanner/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResolveNICDetails collects per-NIC speed and metadata for a node.
func ResolveNICDetails(
	ctx context.Context,
	k8sClient *client.ClusterClient,
	nodeName string,
	bmh *unstructured.Unstructured,
	physicalNames []string,
) []NICInfo {
	if nics, err := GetNICDetailsFromNMState(ctx, k8sClient, nodeName); err == nil && len(nics) > 0 {
		return nics
	}

	if bmh != nil {
		if nics, err := GetNICDetailsFromBMH(bmh); err == nil && len(nics) > 0 {
			return nics
		}
	}

	if nics, err := GetNICDetailsFromSRIOV(ctx, k8sClient, nodeName); err == nil && len(nics) > 0 {
		return nics
	}

	return buildFallbackNICs(physicalNames)
}

// GetNICDetailsFromNMState retrieves NIC details from the NMState NodeNetworkState CR.
func GetNICDetailsFromNMState(
	ctx context.Context,
	k8sClient *client.ClusterClient,
	nodeName string,
) ([]NICInfo, error) {
	nnsGVR := schema.GroupVersionResource{
		Group:    "nmstate.io",
		Version:  "v1",
		Resource: "nodenetworkstates",
	}
	nns, err := k8sClient.Dynamic.Resource(nnsGVR).Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("nmstate not available: %w", err)
	}

	currentState, ok := nns.Object["status"].(map[string]interface{})["currentState"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("status.currentState not found")
	}
	ifaceList, ok := currentState["interfaces"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("interfaces list not found")
	}

	var nics []NICInfo
	for _, raw := range ifaceList {
		iface, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := iface["name"].(string)
		ifaceType, _ := iface["type"].(string)
		if ifaceType != "ethernet" || !isPhysicalInterface(name) {
			continue
		}

		nic := NICInfo{
			Name:  name,
			State: stringField(iface, "state"),
		}
		if mac, ok := iface["mac-address"].(string); ok {
			nic.MACAddress = mac
		}
		if eth, ok := iface["ethernet"].(map[string]interface{}); ok {
			nic.SpeedMbps = intFromInterface(eth["speed"])
			nic.Duplex = stringField(eth, "duplex")
		}
		if ipv4, ok := iface["ipv4"].(map[string]interface{}); ok {
			if addrs, ok := ipv4["address"].([]interface{}); ok && len(addrs) > 0 {
				if first, ok := addrs[0].(map[string]interface{}); ok {
					nic.IPAddress, _ = first["ip"].(string)
				}
			}
		}
		nics = append(nics, nic)
	}
	return nics, nil
}

// GetNICDetailsFromBMH extracts NIC details from a BareMetalHost resource.
func GetNICDetailsFromBMH(bmh *unstructured.Unstructured) ([]NICInfo, error) {
	status, ok := bmh.Object["status"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("status not found")
	}
	hw, ok := status["hardware"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("status.hardware not found")
	}
	nicList, ok := hw["nics"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("status.hardware.nics not found")
	}

	var nics []NICInfo
	for _, raw := range nicList {
		entry, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := entry["name"].(string)
		if !isPhysicalInterface(name) {
			continue
		}
		nic := NICInfo{
			Name:       name,
			MACAddress: stringField(entry, "mac"),
			IPAddress:  stringField(entry, "ip"),
			Model:      stringField(entry, "model"),
			SpeedMbps:  intFromInterface(entry["speedGbps"]) * 1000,
			State:      "up",
			Duplex:     "unknown",
		}
		nics = append(nics, nic)
	}
	return nics, nil
}

// GetNICDetailsFromSRIOV retrieves NIC details from the SR-IOV operator.
func GetNICDetailsFromSRIOV(
	ctx context.Context,
	k8sClient *client.ClusterClient,
	nodeName string,
) ([]NICInfo, error) {
	sriovGVR := schema.GroupVersionResource{
		Group:    "sriovnetwork.openshift.io",
		Version:  "v1",
		Resource: "sriovnetworknodestates",
	}
	sriov, err := k8sClient.Dynamic.Resource(sriovGVR).Namespace("openshift-sriov-network-operator").Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("sriov not available: %w", err)
	}

	status, ok := sriov.Object["status"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("status not found")
	}
	ifaceList, ok := status["interfaces"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("status.interfaces not found")
	}

	var nics []NICInfo
	for _, raw := range ifaceList {
		entry, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := entry["name"].(string)
		if !isPhysicalInterface(name) {
			continue
		}
		nic := NICInfo{
			Name:       name,
			MACAddress: stringField(entry, "mac"),
			SpeedMbps:  parseLinkSpeedString(stringField(entry, "linkSpeed")),
			State:      "up",
			Duplex:     "unknown",
		}
		nics = append(nics, nic)
	}
	return nics, nil
}

func buildFallbackNICs(physicalNames []string) []NICInfo {
	nics := make([]NICInfo, 0, len(physicalNames))
	for _, name := range physicalNames {
		nics = append(nics, NICInfo{
			Name:   name,
			State:  "unknown",
			Duplex: "unknown",
		})
	}
	return nics
}

// parseLinkSpeedString converts SR-IOV linkSpeed strings like "25000 Mb/s" to Mbps int.
func parseLinkSpeedString(s string) int {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, " Mb/s")
	s = strings.TrimSuffix(s, "Mb/s")
	val, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return val
}

func stringField(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

func intFromInterface(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int64:
		return int(val)
	case int:
		return val
	case string:
		n, _ := strconv.Atoi(val)
		return n
	default:
		return 0
	}
}
