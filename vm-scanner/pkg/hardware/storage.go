package hardware

import (
	"context"
	"encoding/json"
	"fmt"
	"vm-scanner/pkg/client"
	"vm-scanner/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetNodeFilesystemInfo retrieves filesystem information from a node
// func GetNodeFilesystemInfo(ctx context.Context, k8sClient *client.ClusterClient, nodeName string) ([]byte, error) {
// 	statsURL := "/api/v1/nodes/" + nodeName + "/proxy/stats/summary"
// 	statsData, err := k8sClient.Typed.CoreV1().RESTClient().Get().AbsPath(statsURL).DoRaw(ctx)
// 	return statsData, err
// }

// GetStorageClasses retrieves all storage classes in the cluster
func GetStorageClasses(ctx context.Context, k8sClient *client.ClusterClient) ([]StorageClassInfo, error) {
	typedClient := k8sClient.Typed
	storageClasses, err := typedClient.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var storageClassInfo []StorageClassInfo
	for _, storageClass := range storageClasses.Items {
		storageClassInfo = append(storageClassInfo, StorageClassInfo{
			Name:              storageClass.Name,
			Provisioner:       storageClass.Provisioner,
			Parameters:        storageClass.Parameters,
			ReclaimPolicy:     string(*storageClass.ReclaimPolicy),
			VolumeBindingMode: string(*storageClass.VolumeBindingMode),
			IsDefault:         storageClass.Annotations["storageclass.kubernetes.io/is-default-class"] == "true",
			CreatedAt:         storageClass.CreationTimestamp,
		})
	}
	return storageClassInfo, nil
}

// ParseFileSystemStats parses filesystem statistics from raw API data
func ParseFileSystemStats(rawData []byte) (float64, float64, float64, float64, error) {
	var filesystemInfo map[string]interface{}
	err := json.Unmarshal(rawData, &filesystemInfo)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	nodeFSData, ok := filesystemInfo["node"].(map[string]interface{})
	if !ok {
		return 0, 0, 0, 0, fmt.Errorf("node data not found")
	}
	fsData, ok := nodeFSData["fs"].(map[string]interface{})
	if !ok {
		return 0, 0, 0, 0, fmt.Errorf("filesystem data not found")
	}

	availableBytes := int64(fsData["availableBytes"].(float64))
	capacityBytes := int64(fsData["capacityBytes"].(float64))
	usedBytes := int64(fsData["usedBytes"].(float64))

	availableGiB := utils.BytesToGiB(availableBytes)
	capacityGiB := utils.BytesToGiB(capacityBytes)
	usedGiB := utils.BytesToGiB(usedBytes)
	usagePercent := float64(usedBytes) / float64(capacityBytes) * 100

	return availableGiB, capacityGiB, usedGiB, usagePercent, nil
}
