package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// GetPVCInventory lists all PVCs across all namespaces and returns inventory items.
// OwningVM is left empty here -- cross-referencing with VMs happens in the merge layer.
func GetPVCInventory(dynamicClient dynamic.Interface) ([]PVCInventoryItem, error) {
	ctx := context.Background()
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolumeclaims"}
	list, err := dynamicClient.Resource(gvr).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list PVCs: %w", err)
	}
	items := make([]PVCInventoryItem, 0, len(list.Items))
	for _, obj := range list.Items {
		item, err := parsePVCItem(obj)
		if err != nil {
			log.Printf("Warning: skip PVC %s/%s: %v", obj.GetNamespace(), obj.GetName(), err)
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func parsePVCItem(obj unstructured.Unstructured) (PVCInventoryItem, error) {
	item := PVCInventoryItem{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		CreatedAt: metav1.Time{Time: obj.GetCreationTimestamp().Time},
	}

	if phase, found, _ := unstructured.NestedString(obj.Object, "status", "phase"); found {
		item.Status = phase
	}

	if storageStr, found, _ := unstructured.NestedString(obj.Object, "spec", "resources", "requests", "storage"); found {
		qty, err := resource.ParseQuantity(storageStr)
		if err == nil {
			item.Capacity = qty.Value()
			item.CapacityHuman = storageStr
		}
	}
	// Also check status.capacity.storage (actual bound capacity)
	if storageStr, found, _ := unstructured.NestedString(obj.Object, "status", "capacity", "storage"); found {
		qty, err := resource.ParseQuantity(storageStr)
		if err == nil {
			item.Capacity = qty.Value()
			item.CapacityHuman = storageStr
		}
	}

	if modes, found, _ := unstructured.NestedStringSlice(obj.Object, "spec", "accessModes"); found {
		item.AccessModes = modes
	}

	if sc, found, _ := unstructured.NestedString(obj.Object, "spec", "storageClassName"); found {
		item.StorageClass = sc
	}

	if vm, found, _ := unstructured.NestedString(obj.Object, "spec", "volumeMode"); found {
		item.VolumeMode = vm
	} else {
		item.VolumeMode = "Filesystem"
	}

	if pvName, found, _ := unstructured.NestedString(obj.Object, "spec", "volumeName"); found {
		item.VolumeName = pvName
	}

	return item, nil
}

// GetNADInventory lists all NetworkAttachmentDefinitions across all namespaces.
// Returns empty slice if Multus/NAD CRD is not installed.
func GetNADInventory(dynamicClient dynamic.Interface) ([]NADInfo, error) {
	ctx := context.Background()
	gvr := schema.GroupVersionResource{Group: "k8s.cni.cncf.io", Version: "v1", Resource: "network-attachment-definitions"}
	list, err := dynamicClient.Resource(gvr).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		if isAPIMissing(err) {
			log.Printf("info: NAD API not available (Multus not installed): %v", err)
			return nil, nil
		}
		return nil, fmt.Errorf("list NADs: %w", err)
	}
	items := make([]NADInfo, 0, len(list.Items))
	for _, obj := range list.Items {
		item, err := parseNADItem(obj)
		if err != nil {
			log.Printf("Warning: skip NAD %s/%s: %v", obj.GetNamespace(), obj.GetName(), err)
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func parseNADItem(obj unstructured.Unstructured) (NADInfo, error) {
	item := NADInfo{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		CreatedAt: metav1.Time{Time: obj.GetCreationTimestamp().Time},
	}

	// Extract resource name from annotation (SR-IOV)
	annotations := obj.GetAnnotations()
	if rn, ok := annotations["k8s.v1.cni.cncf.io/resourceName"]; ok {
		item.ResourceName = rn
	}

	// Parse spec.config JSON to extract type, vlan, and resource name
	configStr, found, _ := unstructured.NestedString(obj.Object, "spec", "config")
	if found && configStr != "" {
		var cfg map[string]interface{}
		if err := json.Unmarshal([]byte(configStr), &cfg); err == nil {
			if t, ok := cfg["type"].(string); ok {
				item.Type = t
			}
			if v, ok := cfg["vlan"].(float64); ok {
				item.VLAN = fmt.Sprintf("%.0f", v)
			}
			if item.ResourceName == "" {
				if b, ok := cfg["bridge"].(string); ok && b != "" {
					item.ResourceName = b
				} else if m, ok := cfg["master"].(string); ok && m != "" {
					item.ResourceName = m
				}
			}
		}
		if len(configStr) > 200 {
			configStr = configStr[:200] + "..."
		}
		item.Config = configStr
	}

	return item, nil
}

// GetDataVolumeInventory lists all CDI DataVolumes across all namespaces.
// Returns empty slice if CDI is not installed.
func GetDataVolumeInventory(dynamicClient dynamic.Interface) ([]DataVolumeInfo, error) {
	ctx := context.Background()
	gvr := schema.GroupVersionResource{Group: "cdi.kubevirt.io", Version: "v1beta1", Resource: "datavolumes"}
	list, err := dynamicClient.Resource(gvr).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		if isAPIMissing(err) {
			log.Printf("info: DataVolume API not available (CDI not installed): %v", err)
			return nil, nil
		}
		return nil, fmt.Errorf("list DataVolumes: %w", err)
	}
	items := make([]DataVolumeInfo, 0, len(list.Items))
	for _, obj := range list.Items {
		item, err := parseDataVolumeItem(obj)
		if err != nil {
			log.Printf("Warning: skip DataVolume %s/%s: %v", obj.GetNamespace(), obj.GetName(), err)
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func parseDataVolumeItem(obj unstructured.Unstructured) (DataVolumeInfo, error) {
	item := DataVolumeInfo{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		CreatedAt: metav1.Time{Time: obj.GetCreationTimestamp().Time},
	}

	if phase, found, _ := unstructured.NestedString(obj.Object, "status", "phase"); found {
		item.Phase = phase
	}
	if progress, found, _ := unstructured.NestedString(obj.Object, "status", "progress"); found {
		item.Progress = progress
	} else {
		item.Progress = "N/A"
	}

	// Determine source type from spec.source
	item.SourceType = detectDVSourceType(obj)

	// Storage size and class: try spec.storage first (new CDI), then spec.pvc (legacy)
	item.StorageSize, item.StorageHuman, item.StorageClass = extractDVStorage(obj)

	// Owning VM from ownerReferences
	if refs, found, _ := unstructured.NestedSlice(obj.Object, "metadata", "ownerReferences"); found {
		for _, ref := range refs {
			if refMap, ok := ref.(map[string]interface{}); ok {
				if kind, _ := refMap["kind"].(string); kind == "VirtualMachine" {
					if name, _ := refMap["name"].(string); name != "" {
						item.OwningVM = name
					}
				}
			}
		}
	}

	return item, nil
}

func detectDVSourceType(obj unstructured.Unstructured) string {
	sourceTypes := []string{"http", "registry", "pvc", "blank", "upload", "s3", "gcs", "imageio", "vddk"}
	for _, st := range sourceTypes {
		if _, found, _ := unstructured.NestedMap(obj.Object, "spec", "source", st); found {
			return st
		}
	}
	return "unknown"
}

func extractDVStorage(obj unstructured.Unstructured) (int64, string, string) {
	// Try spec.storage (new CDI API)
	paths := [][]string{
		{"spec", "storage", "resources", "requests", "storage"},
		{"spec", "pvc", "resources", "requests", "storage"},
	}
	scPaths := [][]string{
		{"spec", "storage", "storageClassName"},
		{"spec", "pvc", "storageClassName"},
	}

	var sizeBytes int64
	var sizeHuman string
	for _, p := range paths {
		if storageStr, found, _ := unstructured.NestedString(obj.Object, p...); found {
			qty, err := resource.ParseQuantity(storageStr)
			if err == nil {
				sizeBytes = qty.Value()
				sizeHuman = storageStr
				break
			}
		}
	}

	var storageClass string
	for _, p := range scPaths {
		if sc, found, _ := unstructured.NestedString(obj.Object, p...); found {
			storageClass = sc
			break
		}
	}

	return sizeBytes, sizeHuman, storageClass
}
