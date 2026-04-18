package cluster

import (
	"context"
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// GetOperatorStatus collects operator health from two sources:
// 1. CVO-managed ClusterOperators (config.openshift.io/v1) — core platform operators
// 2. OLM-managed ClusterServiceVersions (operators.coreos.com/v1alpha1) — add-on operators
func GetOperatorStatus(dynamicClient dynamic.Interface) ([]OperatorStatus, error) {
	var out []OperatorStatus

	cvoOps, err := getClusterOperators(dynamicClient)
	if err != nil {
		log.Printf("warning: failed to list ClusterOperators: %v", err)
	} else {
		out = append(out, cvoOps...)
	}

	csvOps, err := getOLMOperators(dynamicClient)
	if err != nil {
		log.Printf("warning: failed to list ClusterServiceVersions: %v", err)
	} else {
		out = append(out, csvOps...)
	}

	return out, nil
}

func getClusterOperators(dynamicClient dynamic.Interface) ([]OperatorStatus, error) {
	ctx := context.Background()
	gvr := schema.GroupVersionResource{Group: "config.openshift.io", Version: "v1", Resource: "clusteroperators"}

	list, err := dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		if isAPIMissing(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list clusteroperators: %w", err)
	}

	out := make([]OperatorStatus, 0, len(list.Items))
	for i := range list.Items {
		out = append(out, parseClusterOperator(list.Items[i]))
	}
	return out, nil
}

func parseClusterOperator(obj unstructured.Unstructured) OperatorStatus {
	version := ""
	versions, _, _ := unstructured.NestedSlice(obj.Object, "status", "versions")
	for _, v := range versions {
		vm, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if name, _ := vm["name"].(string); name == "operator" {
			version, _ = vm["version"].(string)
			break
		}
	}

	available, degraded, progressing := false, false, false
	conditions, _, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
	for _, c := range conditions {
		cm, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		ctype, _ := cm["type"].(string)
		cstatus, _ := cm["status"].(string)
		switch ctype {
		case "Available":
			available = cstatus == "True"
		case "Degraded":
			degraded = cstatus == "True"
		case "Progressing":
			progressing = cstatus == "True"
		}
	}

	health := "Healthy"
	status := "Available"
	if degraded {
		health = "Degraded"
		status = "Degraded"
	} else if progressing {
		health = "Progressing"
		status = "Progressing"
	} else if !available {
		health = "Unavailable"
		status = "Unavailable"
	}

	return OperatorStatus{
		Name:      obj.GetName(),
		Source:    "CVO",
		Version:   version,
		Status:    status,
		Health:    health,
		Labels:    obj.GetLabels(),
		CreatedAt: metav1.Time{Time: obj.GetCreationTimestamp().Time},
	}
}

func getOLMOperators(dynamicClient dynamic.Interface) ([]OperatorStatus, error) {
	ctx := context.Background()
	gvr := schema.GroupVersionResource{Group: "operators.coreos.com", Version: "v1alpha1", Resource: "clusterserviceversions"}

	list, err := dynamicClient.Resource(gvr).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		if isAPIMissing(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list clusterserviceversions: %w", err)
	}
	out := make([]OperatorStatus, 0, len(list.Items))
	for i := range list.Items {
		out = append(out, parseCSVToOperatorStatus(list.Items[i]))
	}
	return out, nil
}

func parseCSVToOperatorStatus(obj unstructured.Unstructured) OperatorStatus {
	version, _, _ := unstructured.NestedString(obj.Object, "spec", "version")
	phase, _, _ := unstructured.NestedString(obj.Object, "status", "phase")
	health := "Degraded"
	if phase == "Succeeded" {
		health = "Healthy"
	}
	return OperatorStatus{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Source:    "OLM",
		Version:   version,
		Status:    phase,
		Health:    health,
		Labels:    obj.GetLabels(),
		CreatedAt: metav1.Time{Time: obj.GetCreationTimestamp().Time},
	}
}
