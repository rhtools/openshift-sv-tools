package client

import (
	"context"
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// I expect the calling process to populate this struct with the appropriate values
// Structs are similar to the __init__ in python
type AuthOptions struct {
	Kubeconfig   string `json:"kubeconfig" yaml:"kubeconfig"`
	Token        string `json:"token" yaml:"token"`
	APIURL       string `json:"apiUrl" yaml:"apiUrl"`
	UseInCluster bool   `json:"useInCluster" yaml:"useInCluster"`
}

// ClusterClient provides both typed and dynamic clients for comprehensive Kubernetes access
type ClusterClient struct {
	// Typed client for built-in Kubernetes resources (Pods, Services, Nodes, etc.)
	Typed kubernetes.Interface `json:"-" yaml:"-"`
	// Dynamic client for CRDs and custom resources (KubeVirt VMs, etc.)
	Dynamic dynamic.Interface `json:"-" yaml:"-"`
	Config  *rest.Config      `json:"-" yaml:"-"`
}

// buildConfig creates a rest.Config from AuthOptions - shared helper function
func buildConfig(options AuthOptions) (*rest.Config, error) {
	var config *rest.Config
	var err error

	if options.UseInCluster {
		config, err = rest.InClusterConfig()
	} else if options.Kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", options.Kubeconfig)
	} else if options.Token != "" {
		config = &rest.Config{
			Host:            options.APIURL,
			BearerToken:     options.Token,
			TLSClientConfig: rest.TLSClientConfig{Insecure: false},
		}
	} else {
		err = errors.New("no valid authentication method provided")
	}

	return config, err
}

// NewClusterClient creates a new ClusterClient with both typed and dynamic clients
func NewClusterClient(options AuthOptions) (*ClusterClient, error) {
	config, err := buildConfig(options)
	if err != nil {
		return nil, err
	}

	// Create typed client for built-in Kubernetes resources
	typedClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create typed client: %w", err)
	}

	// Create dynamic client for CRDs and custom resources
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &ClusterClient{
		Typed:   typedClient,
		Dynamic: dynamicClient,
		Config:  config,
	}, nil
}

// TestClusterConnection tests both typed and dynamic client connections
func TestClusterConnection(client *ClusterClient) error {
	ctx := context.Background()

	// I am specifically accessing the typed and dynamic clients
	// As a learning exercise, I want to understand how manage the clients
	// Access the typed client
	typedClient := client.Typed
	// Access the dynamic client
	dynamicClient := client.Dynamic
	// Test the typed client connection via healthz endpoint
	err := typedClient.Discovery().RESTClient().Get().AbsPath("/healthz").Do(ctx).Error()
	if err != nil {
		return fmt.Errorf("healthz endpoint not reachable: %w", err)
	}

	// Test authentication by getting a list of namespaces using typed client
	namespaces, err := typedClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list namespaces with typed client: %w", err)
	}

	fmt.Printf("✓ Typed client connected successfully. Found %d namespaces:\n", len(namespaces.Items))
	for _, ns := range namespaces.Items {
		fmt.Printf("  - %s\n", ns.Name)
	}

	// Test dynamic client by listing namespaces as well
	_, err = dynamicClient.Resource(namespaceGVR()).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list namespaces with dynamic client: %w", err)
	}

	fmt.Println("✓ Dynamic client connected successfully")
	return nil
}

// namespaceGVR returns the GroupVersionResource for namespaces
func namespaceGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",   // Core API group is empty string
		Version:  "v1", // Core API version
		Resource: "namespaces",
	}
}
