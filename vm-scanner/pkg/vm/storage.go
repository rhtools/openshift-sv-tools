package vm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"vm-scanner/pkg/client"
	"vm-scanner/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// I am creating a struct in order to unmarshal the dataVolumeTemplates for cleaner code
type DataVolumeTemplate struct {
	Metadata struct {
		Name string `json:"name" yaml:"name"`
	} `json:"metadata" yaml:"metadata"`
	Spec struct {
		Storage struct {
			Resources struct {
				Requests struct {
					Storage string `json:"storage" yaml:"storage"`
				} `json:"requests" yaml:"requests"`
			} `json:"resources" yaml:"resources"`
			StorageClassName string `json:"storageClassName,omitempty" yaml:"storageClassName,omitempty"`
		} `json:"storage" yaml:"storage"`
	} `json:"spec" yaml:"spec"`
}

type PVCVolume struct {
	Name                  string `json:"name" yaml:"name"`
	PersistentVolumeClaim struct {
		ClaimName string `json:"claimName" yaml:"claimName"`
	} `json:"persistentVolumeClaim" yaml:"persistentVolumeClaim"`
	DataVolume struct {
		Name string `json:"name" yaml:"name"`
	} `json:"dataVolume" yaml:"dataVolume"`
}

// GetVMStorageInfo extracts storage information for all volumes attached to a VM.
//
// This function uses spec.template.spec.volumes[] as the source of truth, which contains
// ALL volumes attached to the VM regardless of how they were created:
// - DataVolumes from dataVolumeTemplates (created at VM creation)
// - DataVolumes that were hotplugged later
// - External PersistentVolumeClaims
//
// For each volume, it queries the actual PVC to get size and StorageClass information.
func GetVMStorageInfo(vm *unstructured.Unstructured, k8sClient *client.ClusterClient, ctx context.Context) ([]StorageInfo, error) {
	vm_spec := vm.Object["spec"].(map[string]interface{})
	namespace := vm.GetNamespace()

	// Get all volumes from spec.template.spec.volumes[] (the source of truth)
	hasVolumes, allStorageInfo := hasExternalPVCs(vm_spec, namespace, k8sClient, ctx)
	
	if !hasVolumes {
		return nil, fmt.Errorf("no storage volumes found")
	}

	// Calculate total storage across all volumes
	totalStorage := sumTotalStorage(allStorageInfo)
	
	// Set the total on each volume
	for i := range allStorageInfo {
		allStorageInfo[i].TotalStorage = totalStorage
		allStorageInfo[i].TotalStorageHuman = fmt.Sprintf("%.2f", utils.BytesToGiB(totalStorage))
	}

	return allStorageInfo, nil
}

// hasDataVolumeTemplates checks if VM has dataVolumeTemplates AND parses the storage info.
// Returns: (hasTemplates bool, storageInfo []StorageInfo)
// This dual return value allows us to check AND extract data in one pass
func hasDataVolumeTemplates(vm_spec map[string]interface{}) (bool, []StorageInfo) {
	dataVolumeTemplates, err := json.Marshal(vm_spec["dataVolumeTemplates"].([]interface{}))
	if err != nil {
		return false, nil
	}
	var specStorage []DataVolumeTemplate
	errUnmarshal := json.Unmarshal(dataVolumeTemplates, &specStorage)
	if errUnmarshal != nil {
		log.Fatalf("Error unmarshalling dataVolumeTemplates: %v", errUnmarshal)
		return false, nil
	}

	var storageInfo []StorageInfo
	for _, storage := range specStorage {
		// this is a string like "30Gi"
		sizeBytes, err := utils.ParseQuantityToBytes(storage.Spec.Storage.Resources.Requests.Storage)
		if err != nil {
			return false, nil
		}
		sizeGiB := utils.BytesToGiB(sizeBytes)
		// Map the unmarshalled data to the StorageInfo struct
		storageInfo = append(storageInfo, StorageInfo{
			VolumeName:   storage.Metadata.Name,
			SizeBytes:    sizeBytes,
			SizeHuman:    fmt.Sprintf("%.2f", sizeGiB),
			VolumeType:   "dataVolumeTemplate",
			StorageClass: storage.Spec.Storage.StorageClassName,
		})
	}
	return true, storageInfo
}

func getPVCList(vm_spec map[string]interface{}) (bool, []string) {
	// before I unmarshall the json to make strong typing
	// I need to get the template spec and volumes
	template, ok := vm_spec["template"].(map[string]interface{})
	if !ok {
		return false, nil
	}
	pvcSpec, ok := template["spec"].(map[string]interface{})
	if !ok {
		return false, nil
	}
	volumes, ok := pvcSpec["volumes"].([]interface{})
	if !ok {
		return false, nil
	}
	// I marshall the volumes to get the json string to make it easier to extract
	// into the data types I want in the PVCVolume struct
	pvcVolumes, err := json.Marshal(volumes)
	if err != nil {
		return false, nil
	}

	var pvcVolumeList []PVCVolume
	errUnmarshal := json.Unmarshal(pvcVolumes, &pvcVolumeList)
	if errUnmarshal != nil {
		return false, nil
	}
	var pvcClaimNames []string
	for _, volume := range pvcVolumeList {
		// Check for persistentVolumeClaim type
		if volume.PersistentVolumeClaim.ClaimName != "" {
			pvcClaimNames = append(pvcClaimNames, volume.PersistentVolumeClaim.ClaimName)
		}
		// Check for dataVolume type (includes hotplugged disks)
		if volume.DataVolume.Name != "" {
			pvcClaimNames = append(pvcClaimNames, volume.DataVolume.Name)
		}
	}
	if len(pvcClaimNames) == 0 {
		return false, nil
	}

	return true, pvcClaimNames
}

func getPVCStorageInfo(pvcList []string, namespace string, k8sClient *client.ClusterClient, ctx context.Context) ([]StorageInfo, error) {
	// loop over the pvcList and get the storage info for each pvc
	var storageInfo []StorageInfo
	for _, pvcName := range pvcList {
		pvc, err := k8sClient.Typed.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvcName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// Storage() returns *resource.Quantity - no need to check type
		quantity := pvc.Spec.Resources.Requests.Storage()
		sizeBytes := quantity.Value()
		sizeGiB := utils.BytesToGiB(sizeBytes)
		storageClass := ""
		if pvc.Spec.StorageClassName != nil {
			storageClass = *pvc.Spec.StorageClassName
		}

		storageInfo = append(storageInfo, StorageInfo{
			VolumeName:   pvc.Name,
			SizeBytes:    sizeBytes,
			SizeHuman:    fmt.Sprintf("%.2f", sizeGiB),
			VolumeType:   "pvc",
			StorageClass: storageClass,
		})
	}
	return storageInfo, nil
}

func hasExternalPVCs(vm_spec map[string]interface{}, namespace string, k8sClient *client.ClusterClient, ctx context.Context) (bool, []StorageInfo) {
	_, pvcList := getPVCList(vm_spec)
	if len(pvcList) == 0 {
		return false, nil
	}
	pvcStorageInfo, err := getPVCStorageInfo(pvcList, namespace, k8sClient, ctx)
	if err != nil {
		return false, nil
	}
	return true, pvcStorageInfo
}

// need a function to sum up total storage for a vm
func sumTotalStorage(storageInfo []StorageInfo) int64 {
	totalStorage := int64(0)
	for _, storage := range storageInfo {
		totalStorage += storage.SizeBytes
	}
	return totalStorage
}

// guestAgentFSInfo is an intermediate struct to handle the nested structure
// returned by the KubeVirt guest agent API
type guestAgentFSInfo struct {
	FSInfo struct {
		Disks []struct {
			DiskName       string `json:"diskName"`
			FileSystemType string `json:"fileSystemType"`
			MountPoint     string `json:"mountPoint"`
			TotalBytes     int64  `json:"totalBytes"`
			UsedBytes      int64  `json:"usedBytes"`
		} `json:"disks"`
	} `json:"fsInfo"`
}

// parseStorageInfoFromGuestAgentInfo parses the guest agent API response
// and extracts filesystem information into VMDiskInfo structs
func parseStorageInfoFromGuestAgentInfo(guestAgentInfo string) ([]VMDiskInfo, error) {
	if guestAgentInfo == "" {
		return nil, nil
	}

	// First unmarshal into intermediate struct to handle nested fsInfo.disks
	var fsInfo guestAgentFSInfo
	err := json.Unmarshal([]byte(guestAgentInfo), &fsInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal guest agent fsInfo: %w", err)
	}

	// Map the intermediate struct to VMDiskInfo with correct field names
	diskInfo := make([]VMDiskInfo, 0, len(fsInfo.FSInfo.Disks))
	for _, disk := range fsInfo.FSInfo.Disks {
		diskInfo = append(diskInfo, VMDiskInfo{
			DiskName:   disk.DiskName,
			FsType:     disk.FileSystemType,
			MountPoint: disk.MountPoint,
			TotalBytes: disk.TotalBytes,
			UsedBytes:  disk.UsedBytes,
		})
	}

	return diskInfo, nil
}
