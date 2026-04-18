package rbac

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// createClusterInfoGatherRole creates a ClusterRole with comprehensive permissions
// for gathering cluster information including KubeVirt resources, core Kubernetes
// resources, application resources, storage, networking, and metrics.
//
// The ClusterRole includes read-only access to:
// - KubeVirt resources (virtualmachines, virtualmachineinstances, datavolumes)
// - Core Kubernetes resources (pods, services, PVCs, PVs, events, nodes, namespaces)
// - Application resources (deployments, replicasets)
// - Storage resources (storageclasses, volumeattachments)
// - Network resources (networkpolicies, ingresses)
// - Metrics resources (nodes, pods, custom metrics)
//
// If the ClusterRole already exists, it will be updated. If creation fails,
// the function will panic.
func createClusterInfoGatherRole(client *kubernetes.Clientset) {
	roleName := "cluster-openshift-sv-tools"
	rules := []rbac.PolicyRule{
		// KubeVirt resources
		{
			APIGroups: []string{"kubevirt.io"},
			Resources: []string{"virtualmachines", "virtualmachineinstances", "datavolumes", "virtualmachinesnapshots"},
			Verbs:     []string{"get", "list"},
		},
		// Instance type resources (CPU/memory specs for instancetype-based VMs)
		{
			APIGroups: []string{"instancetype.kubevirt.io"},
			Resources: []string{
				"virtualmachineclusterinstancetypes",
				"virtualmachineinstancetypes",
				"virtualmachineclusterpreferences",
				"virtualmachinepreferences",
			},
			Verbs: []string{"get", "list"},
		},
		// Network Attachment Definitions (Multus network inventory)
		{
			APIGroups: []string{"k8s.cni.cncf.io"},
			Resources: []string{"network-attachment-definitions"},
			Verbs:     []string{"get", "list"},
		},
		// CDI DataVolume resources (import/clone tracking)
		{
			APIGroups: []string{"cdi.kubevirt.io"},
			Resources: []string{"datavolumes"},
			Verbs:     []string{"get", "list"},
		},
		// Core Kubernetes resources
		{
			APIGroups: []string{""},
			Resources: []string{"pods", "services", "persistentvolumeclaims", "persistentvolumes", "events", "nodes", "namespaces"},
			Verbs:     []string{"get", "list"},
		},
		// Application resources
		{
			APIGroups: []string{"apps"},
			Resources: []string{"deployments", "replicasets"},
			Verbs:     []string{"get", "list"},
		},
		// Storage resources
		{
			APIGroups: []string{"storage.k8s.io"},
			Resources: []string{"storageclasses", "volumeattachments"},
			Verbs:     []string{"get", "list"},
		},
		// Network resources
		{
			APIGroups: []string{"networking.k8s.io"},
			Resources: []string{"networkpolicies", "ingresses"},
			Verbs:     []string{"get", "list"},
		},
		// NMState network state (NIC speed via NodeNetworkState)
		{
			APIGroups: []string{"nmstate.io"},
			Resources: []string{"nodenetworkstates"},
			Verbs:     []string{"get", "list"},
		},
		// BareMetalHost resources (NIC speed/model on bare-metal clusters)
		{
			APIGroups: []string{"metal3.io"},
			Resources: []string{"baremetalhosts"},
			Verbs:     []string{"get", "list"},
		},
		// SR-IOV network node state (NIC speed via SR-IOV operator)
		{
			APIGroups: []string{"sriovnetwork.openshift.io"},
			Resources: []string{"sriovnetworknodestates"},
			Verbs:     []string{"get", "list"},
		},
		// OLM ClusterServiceVersions (operator status tracking)
		{
			APIGroups: []string{"operators.coreos.com"},
			Resources: []string{"clusterserviceversions"},
			Verbs:     []string{"get", "list"},
		},
		// CVO ClusterOperators (core platform operator health)
		{
			APIGroups: []string{"config.openshift.io"},
			Resources: []string{"clusteroperators"},
			Verbs:     []string{"get", "list"},
		},
		// Metrics resources
		{
			APIGroups: []string{"metrics.k8s.io"},
			Resources: []string{"nodes", "pods"},
			Verbs:     []string{"get", "list"},
		},
		// Node information
		{
			APIGroups: []string{""},
			Resources: []string{"nodes/status"},
			Verbs:     []string{"get", "list"},
		},
	}

	clusterRole := &rbac.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleName,
		},
		Rules: rules,
	}

	_, err := client.RbacV1().ClusterRoles().Update(context.TODO(), clusterRole, metav1.UpdateOptions{})
	if err != nil {
		_, createErr := client.RbacV1().ClusterRoles().Create(context.TODO(), clusterRole, metav1.CreateOptions{})
		if createErr != nil {
			panic(createErr)
		}
	}
}

// createClusterInfoGatherRoleBinding creates a ClusterRoleBinding that binds
// the "cluster-openshift-sv-tools" ClusterRole to the specified ServiceAccount
// in the given namespace.
//
// The binding allows the ServiceAccount to assume the permissions defined in
// the ClusterRole, enabling cluster-wide read access for gathering information.
//
// Parameters:
//   - namespaceName: The namespace containing the ServiceAccount
//   - serviceAccountName: The name of the ServiceAccount to bind
//   - client: The Kubernetes client for API operations
//
// If the ClusterRoleBinding already exists, it will be updated. If creation
// fails, the function will panic.
func createClusterInfoGatherRoleBinding(namespaceName string, serviceAccountName string, client *kubernetes.Clientset) {
	bindingName := "cluster-openshift-sv-tools-binding"
	subjects := []rbac.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      serviceAccountName,
			Namespace: namespaceName,
		},
	}
	roleRef := rbac.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "cluster-openshift-sv-tools",
	}
	clusterRoleBinding := &rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: bindingName,
		},
		Subjects: subjects,
		RoleRef:  roleRef,
	}
	_, err := client.RbacV1().ClusterRoleBindings().Update(context.TODO(), clusterRoleBinding, metav1.UpdateOptions{})
	if err != nil {
		_, createErr := client.RbacV1().ClusterRoleBindings().Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{})
		if createErr != nil {
			panic(createErr)
		}
	}
}

// createServiceAccount creates a ServiceAccount in the specified namespace
// if it does not already exist.
//
// The ServiceAccount is used for authentication and authorization when the
// application needs to interact with the Kubernetes API. It will be bound
// to a ClusterRole through a ClusterRoleBinding to grant necessary permissions.
//
// Parameters:
//   - namespaceName: The namespace where the ServiceAccount will be created
//   - serviceAccountName: The name of the ServiceAccount to create
//   - client: The Kubernetes client for API operations
//
// If the ServiceAccount already exists, no action is taken. If creation
// fails, the function will panic.
func createServiceAccount(namespaceName string, serviceAccountName string, client *kubernetes.Clientset) {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespaceName,
		},
	}
	_, existErr := client.CoreV1().ServiceAccounts(namespaceName).Get(context.TODO(), serviceAccountName, metav1.GetOptions{})
	if existErr != nil {
		_, err := client.CoreV1().ServiceAccounts(namespaceName).Create(context.TODO(), serviceAccount, metav1.CreateOptions{})
		if err != nil {
			panic(err)
		}
	}
}
