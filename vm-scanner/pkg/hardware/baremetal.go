package hardware

import (
	"context"
	"vm-scanner/pkg/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func GetBareMetalHost(ctx context.Context, k8sClient *client.ClusterClient, node corev1.Node) (*unstructured.Unstructured, error) {
	bareMetalHostGVR := schema.GroupVersionResource{
		Group:    "metal3.io",
		Version:  "v1alpha1",
		Resource: "baremetalhosts",
	}
	bmh, err := k8sClient.Dynamic.Resource(bareMetalHostGVR).Namespace("openshift-machine-api").Get(ctx, node.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return bmh, nil
}
