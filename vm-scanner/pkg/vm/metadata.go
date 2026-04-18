// capture power on state, creation timestamp, updated timestamp, labels, annotations
package vm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"vm-scanner/pkg/client"
	"vm-scanner/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// guestAgentOSInfo is an intermediate struct that matches the "os" nested object in Guest Agent API
type guestAgentOSInfo struct {
	ID            string `json:"id" yaml:"id"`
	KernelRelease string `json:"kernelRelease" yaml:"kernelRelease"`
	KernelVersion string `json:"kernelVersion" yaml:"kernelVersion"`
	Machine       string `json:"machine" yaml:"machine"`
	Name          string `json:"name" yaml:"name"`
	PrettyName    string `json:"prettyName" yaml:"prettyName"`
	Version       string `json:"version" yaml:"version"`
	VersionID     string `json:"versionId" yaml:"versionId"`
}

// guestAgentResponse is an intermediate struct that matches the Guest Agent API JSON structure
type guestAgentResponse struct {
	Hostname          string           `json:"hostname" yaml:"hostname"`
	GuestAgentVersion string           `json:"guestAgentVersion" yaml:"guestAgentVersion"`
	Timezone          string           `json:"timezone" yaml:"timezone"`
	OS                guestAgentOSInfo `json:"os" yaml:"os"`
}

// GetVirtLauncherPodName retrieves the name of the virt-launcher pod for a running VMI
func GetVirtLauncherPodName(k8sClient *client.ClusterClient, ctx context.Context, namespace string, vmiName string) (string, error) {
	// get the name of the virt-launcher pod that is running the VMI using the VMI name and namespace
	virtLauncherPods, err := k8sClient.Typed.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("vm.kubevirt.io/name=%s", vmiName),
	})
	if err != nil {
		return "", err
	}
	if len(virtLauncherPods.Items) == 0 {
		return "", fmt.Errorf("no virt-launcher pod found for VM %s", vmiName)
	}
	// combine the namespace with a slash and the virt-launcher pod name
	virtLauncherPodName := namespace + "/" + virtLauncherPods.Items[0].Name
	return virtLauncherPodName, nil
}

func ParseVMIMachineInfo(vmiUnstructured unstructured.Unstructured) (string, string, string, error) {
	prettyName, found, err := unstructured.NestedString(vmiUnstructured.Object, "status", "guestOSInfo", "prettyName")
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get guest OS prettyName: %w", err)
	}
	// get the nodename from the label kubevirt.io/nodeName
	runningOnNode, found, err := unstructured.NestedString(vmiUnstructured.Object, "metadata", "labels", "kubevirt.io/nodeName")
	if err != nil || !found {
		return "", "", "", fmt.Errorf("failed to get running on node: %w", err)
	}
	guestOSVersion, found, err := unstructured.NestedString(vmiUnstructured.Object, "status", "guestOSInfo", "versionId")

	if err != nil || !found {
		return "", "", "", fmt.Errorf("failed to get guest metadata: %w", err)
	}
	return prettyName, runningOnNode, guestOSVersion, nil
}

func ParseVMMachineInfo(vmUnstructured unstructured.Unstructured) (string, error) {
	// get os name from spec.template.metadata.annotations.vm.kubevirt.io/os
	// This annotation is optional, so return empty string if not found rather than error
	guestOSVersion, _, _ := unstructured.NestedString(vmUnstructured.Object, "spec", "template", "metadata", "annotations", "vm.kubevirt.io/os")
	return guestOSVersion, nil
}

func GetMachineMetadata(k8sClient *client.ClusterClient, ctx context.Context, objectUnstructured unstructured.Unstructured, vmi bool, prefix ...string) (string, string, metav1.Time, string, string, string, string, error) {
	machineType, uid, creationTimestamp, err := GetCommonGuestMetadata(k8sClient, ctx, objectUnstructured, prefix...)
	if err != nil {
		return "", "", metav1.Time{}, "", "", "", "", fmt.Errorf("failed to get common guest metadata: %w", err)
	}
	// if it's a vmi, i need to get the common metadata as well as parse the vmi machine info and get the launcher pod name
	// but I need to return both the common metadata and the vmi machine info
	if vmi {
		prettyName, runningOnNode, guestOSVersion, err := ParseVMIMachineInfo(objectUnstructured)
		if err != nil {
			return "", "", metav1.Time{}, "", "", "", "", fmt.Errorf("failed to get VMIMachineInfo: %w", err)
		}
		launcherPodName, err := GetVirtLauncherPodName(k8sClient, ctx, objectUnstructured.GetNamespace(), objectUnstructured.GetName())
		if err != nil {
			return "", "", metav1.Time{}, "", "", "", "", fmt.Errorf("failed to get launcher pod name: %w", err)
		}

		return machineType, uid, creationTimestamp, prettyName, runningOnNode, guestOSVersion, launcherPodName, nil
	}
	guestOSVersion, _ := ParseVMMachineInfo(objectUnstructured)
	return machineType, uid, creationTimestamp, "", "", guestOSVersion, "", nil
}

// GetCommonGuestMetadata retrieves guest agent metadata for a running VMI
func GetCommonGuestMetadata(k8sClient *client.ClusterClient, ctx context.Context, objectUnstructured unstructured.Unstructured, prefix ...string) (string, string, metav1.Time, error) {
	machineTypePath := utils.BuildPath(prefix, "domain", "machine", "type")
	uid := string(objectUnstructured.GetUID())
	creationTimestamp := objectUnstructured.GetCreationTimestamp()

	// Machine type is optional - return empty string if not found rather than error
	machineType, _, _ := unstructured.NestedString(objectUnstructured.Object, machineTypePath...)

	return machineType, uid, creationTimestamp, nil
}

// parseGuestMetadataFromGuestAgentInfo unmarshals Guest Agent JSON and returns populated GuestMetadata
// Uses intermediate struct to handle JSON structure, then maps to GuestMetadata from types.go
func parseGuestMetadataFromGuestAgentInfo(guestAgentInfo string) (*GuestMetadata, error) {
	var response guestAgentResponse
	err := json.Unmarshal([]byte(guestAgentInfo), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal guest agent response: %w", err)
	}

	timezoneCode := ""
	if response.Timezone != "" {
		timezoneParts := strings.Split(response.Timezone, ",")
		timezoneCode = strings.TrimSpace(timezoneParts[0])
	}

	metadata := &GuestMetadata{
		HostName:          response.Hostname,
		GuestAgentVersion: response.GuestAgentVersion,
		KernelVersion:     response.OS.KernelRelease,
		Timezone:          timezoneCode,
	}

	return metadata, nil
}
