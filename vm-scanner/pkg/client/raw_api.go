package client

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetNodeRawAPIData(ctx context.Context, k8sClient *ClusterClient, nodeName string) ([]byte, error) {
	statsURL := "/api/v1/nodes/" + nodeName + "/proxy/stats/summary"
	statsData, err := k8sClient.Typed.CoreV1().RESTClient().Get().AbsPath(statsURL).DoRaw(ctx)
	return statsData, err
}

func GetVirtLauncherPodRawAPIData(ctx context.Context, k8sClient *ClusterClient, VirtLauncherPodName string) ([]byte, error) {
	// I expect the VirtLauncherPodName to be in the format of namespace/pod-name
	namespace := strings.Split(VirtLauncherPodName, "/")[0]
	podName := strings.Split(VirtLauncherPodName, "/")[1]
	podUrl := "/apis/metrics.k8s.io/v1beta1/namespaces/" + namespace + "/pods/" + podName
	podData, err := k8sClient.Typed.CoreV1().RESTClient().Get().AbsPath(podUrl).DoRaw(ctx)
	return podData, err
}

func GetMonitoringRawAPIData(ctx context.Context, k8sClient *ClusterClient, namespace string, vmiName string, nodeName string) ([]byte, error) {
	// I need to get the raw data from the monitoring stack in order to have accurate RAM usage for individual vms
	virtHandlerPods, err := k8sClient.Typed.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "kubevirt.io=virt-handler",
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list virt-handler pods: %w", err)
	}
	proxyPath := fmt.Sprintf("/api/v1/namespaces/openshift-cnv/pods/https:%s:8443/proxy/metrics", virtHandlerPods.Items[0].Name)
	metricsData, err := k8sClient.Typed.CoreV1().RESTClient().Get().AbsPath(proxyPath).DoRaw(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics for VMI %s: %w", vmiName, err)
	}

	return metricsData, nil

}

func GetGuestAgentInfoRawAPIData(ctx context.Context, k8sClient *ClusterClient, namespace string, vmiName string) (string, error) {
	// I want to get the KVM Guest Agent Version from the guestosinfo resource
	// this has disperate data from disks, to OS info and guest agents
	guestOSInfoPath := fmt.Sprintf("/apis/subresources.kubevirt.io/v1/namespaces/%s/virtualmachineinstances/%s/guestosinfo", namespace, vmiName)
	guestOSInfoData, err := k8sClient.Typed.CoreV1().RESTClient().Get().AbsPath(guestOSInfoPath).DoRaw(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to query guest agent version for VMI %s: %w", vmiName, err)
	}
	return string(guestOSInfoData), nil
}
