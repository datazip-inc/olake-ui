package activities

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"olake-k8s-worker/logger"
	"olake-k8s-worker/utils"
)

// K8sPodManager handles Kubernetes Pod operations only
type K8sPodManager struct {
	clientset        kubernetes.Interface
	namespace        string
	filesystemHelper *utils.FilesystemHelper
}

// NewK8sPodManager creates a new Kubernetes Pod manager
func NewK8sPodManager() (*K8sPodManager, error) {
	// Use in-cluster configuration
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Get namespace from environment or use default
	namespace := utils.GetEnv("WORKER_NAMESPACE", "default")

	logger.Infof("Initialized K8s pod manager for namespace: %s", namespace)

	return &K8sPodManager{
		clientset:        clientset,
		namespace:        namespace,
		filesystemHelper: utils.NewFilesystemHelper(),
	}, nil
}

// GetDockerImageName constructs a Docker image name based on source type and version
func (k *K8sPodManager) GetDockerImageName(sourceType, version string) string {
	if version == "" {
		version = "latest"
	}
	return fmt.Sprintf("olakego/source-%s:%s", sourceType, version)
}
