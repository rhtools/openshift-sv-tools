package vm

import (
	"context"
	"fmt"
	"log"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// InstanceTypeSpecs holds CPU and memory from a VirtualMachine*Instancetype spec.
type InstanceTypeSpecs struct {
	CPUGuest    int64  // number of vCPUs (spec.cpu.guest)
	MemoryGuest string // memory quantity string (spec.memory.guest), e.g. "2Gi"
}

func isInstancetypeAPIMissing(err error) bool {
	if err == nil {
		return false
	}
	if apierrors.IsNotFound(err) {
		return true
	}
	return meta.IsNoMatchError(err)
}

func parseInstanceTypeSpecs(obj unstructured.Unstructured) (InstanceTypeSpecs, error) {
	cpu, found, err := unstructured.NestedInt64(obj.Object, "spec", "cpu", "guest")
	if err != nil {
		return InstanceTypeSpecs{}, fmt.Errorf("spec.cpu.guest: %w", err)
	}
	if !found {
		return InstanceTypeSpecs{}, fmt.Errorf("spec.cpu.guest not found")
	}
	mem, found, err := unstructured.NestedString(obj.Object, "spec", "memory", "guest")
	if err != nil {
		return InstanceTypeSpecs{}, fmt.Errorf("spec.memory.guest: %w", err)
	}
	if !found {
		return InstanceTypeSpecs{}, fmt.Errorf("spec.memory.guest not found")
	}
	return InstanceTypeSpecs{CPUGuest: cpu, MemoryGuest: mem}, nil
}

// BuildInstanceTypeMap lists cluster and namespaced instancetypes and returns a lookup map.
// Cluster-scoped keys are the resource name; namespaced keys are "namespace/name".
func BuildInstanceTypeMap(dynamicClient dynamic.Interface) (map[string]InstanceTypeSpecs, error) {
	out := make(map[string]InstanceTypeSpecs)
	ctx := context.Background()
	gvrCluster := schema.GroupVersionResource{Group: "instancetype.kubevirt.io", Version: "v1beta1", Resource: "virtualmachineclusterinstancetypes"}
	gvrNS := schema.GroupVersionResource{Group: "instancetype.kubevirt.io", Version: "v1beta1", Resource: "virtualmachineinstancetypes"}

	clusterList, err := dynamicClient.Resource(gvrCluster).List(ctx, metav1.ListOptions{})
	if err != nil {
		if !isInstancetypeAPIMissing(err) {
			return nil, fmt.Errorf("list VirtualMachineClusterInstancetype: %w", err)
		}
		log.Printf("info: instancetype API not available for VirtualMachineClusterInstancetype: %v", err)
		clusterList = &unstructured.UnstructuredList{}
	}
	for _, item := range clusterList.Items {
		specs, err := parseInstanceTypeSpecs(item)
		if err != nil {
			log.Printf("warning: skip cluster instancetype %q: %v", item.GetName(), err)
			continue
		}
		out[item.GetName()] = specs
	}

	nsList, err := dynamicClient.Resource(gvrNS).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		if !isInstancetypeAPIMissing(err) {
			return nil, fmt.Errorf("list VirtualMachineInstancetype: %w", err)
		}
		log.Printf("info: instancetype API not available for VirtualMachineInstancetype: %v", err)
		nsList = &unstructured.UnstructuredList{}
	}
	for _, item := range nsList.Items {
		key := item.GetNamespace() + "/" + item.GetName()
		specs, err := parseInstanceTypeSpecs(item)
		if err != nil {
			log.Printf("warning: skip instancetype %q: %v", key, err)
			continue
		}
		out[key] = specs
	}
	return out, nil
}
