package pods

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	appConfig "olake-ui/olake-workers/k8s/config"
	"olake-ui/olake-workers/k8s/logger"
	"olake-ui/olake-workers/k8s/utils/filesystem"
)

// K8sPodManager handles Kubernetes Pod operations only
// This is the central orchestrator for all pod-related operations in the K8s worker.
// It maintains the Kubernetes client connection, target namespace, filesystem utilities,
// and configuration needed to create, manage, and clean up activity pods.
type K8sPodManager struct {
	clientset        kubernetes.Interface // Kubernetes API client for pod operations
	namespace        string               // Target namespace where all activity pods are created
	filesystemHelper *filesystem.Helper   // Utility for managing shared storage and config files
	config           *appConfig.Config    // Application configuration including timeouts, storage, etc.
}

// NewK8sPodManager creates a new Kubernetes Pod manager
// This initializes the pod manager with in-cluster Kubernetes configuration,
// which allows the worker running inside a pod to communicate with the K8s API server.
// The manager will be responsible for creating, monitoring, and cleaning up activity pods.
func NewK8sPodManager(cfg *appConfig.Config) (*K8sPodManager, error) {
	// Use in-cluster configuration - this reads the service account token and CA cert
	// that Kubernetes automatically mounts into every pod at /var/run/secrets/kubernetes.io/serviceaccount/
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %v", err)
	}

	// Create the Kubernetes clientset using the in-cluster config
	// This clientset provides access to all Kubernetes API operations (pods, services, etc.)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Get the target namespace from environment variable, defaulting to "olake"
	// All activity pods will be created in this namespace
	namespace := appConfig.GetEnv("WORKER_NAMESPACE", "olake")

	logger.Infof("Initialized K8s pod manager for namespace: %s", namespace)

	// Initialize and return the pod manager with all required dependencies
	return &K8sPodManager{
		clientset:        clientset,              // K8s API client
		namespace:        namespace,              // Target namespace for pods
		filesystemHelper: filesystem.NewHelper(), // Shared storage utilities
		config:           cfg,                    // Application configuration
	}, nil
}

// GetDockerImageName constructs a Docker image name based on source type and version
// This maps abstract connector references (e.g., "mysql", "postgres") to concrete container images
// that contain the actual connector implementation. Each connector type has its own Docker image
// in the "olakego" registry with naming convention: olakego/source-{type}:{version}
func (k *K8sPodManager) GetDockerImageName(sourceType, version string) (string, error) {
	// Strict validation: version is required (no 'latest' fallback)
	if version == "" {
		return "", fmt.Errorf("version cannot be empty - no 'latest' tag exists for connector images")
	}

	// Construct the full image name using the olakego registry convention
	// Examples: olakego/source-mysql:v0.1.7, olakego/source-postgres:v1.2.3
	return fmt.Sprintf("olakego/source-%s:%s", sourceType, version), nil
}
