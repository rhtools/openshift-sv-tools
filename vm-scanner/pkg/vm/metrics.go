package vm

// NodeMetrics represents node-level resource metrics

// func GetGuestOSInfo(k8sClient *client.ClusterClient, ctx context.Context, VMIInfo VMIInfo) ([]VMIStorageDeviceMetrics, error) {
// 	// In order to get the best metrics, query the /apis/subresources.kubevirt.io/v1/namespaces/{namespace}/virtualmachineinstances/{vmi-name}/guestosinfo
// 	metricsGVR := schema.GroupVersionResource{
// 		Group:    "subresources.kubevirt.io",
// 		Version:  "v1",
// 		Resource: "guestosinfo",
// 	}
// 	// Loop over the VMIInfo Struct and get the guestosinfo for each VMI
// 	for _, vmi := range VMIInfo {
// 		guestHardwareInfo, err := k8sClient.Dynamic.Resource(metricsGVR).Namespace().Get(ctx, vmiName, metav1.GetOptions{})
// 		if err != nil {
// 			return nil, err
// 		}
// 	return nil, nil
// }
