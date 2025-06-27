package activities

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"olake-k8s-worker/logger"
	"olake-k8s-worker/shared"
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

// PodSpec defines the specification for creating a Kubernetes Pod
type PodSpec struct {
	Name               string
	Image              string
	Command            []string
	Args               []string
	Operation          shared.Command
	OriginalWorkflowID string
}

// CreatePod creates a Kubernetes Pod for running sync operations
func (k *K8sPodManager) CreatePod(ctx context.Context, spec *PodSpec, configs []shared.JobConfig) (*corev1.Pod, error) {
	// Get workflow directory using filesystem helper
	workflowDir := k.filesystemHelper.GetWorkflowDirectory(spec.Operation, spec.OriginalWorkflowID)

	// Use filesystem helper to setup directory and write config files
	if err := k.filesystemHelper.SetupWorkDirectory(workflowDir); err != nil {
		return nil, fmt.Errorf("failed to setup work directory: %v", err)
	}

	if err := k.filesystemHelper.WriteConfigFiles(workflowDir, configs); err != nil {
		return nil, fmt.Errorf("failed to write config files: %v", err)
	}

	// Create Pod with PV mount
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.Name,
			Namespace: k.namespace,
			Labels: map[string]string{
				"app":                  "olake-connector",
				"type":                 "sync-pod",
				"operation":            string(spec.Operation),
				"olake.io/workflow-id": utils.SanitizeK8sName(spec.OriginalWorkflowID),
				"olake.io/autoscaling": "enabled",
			},
			Annotations: map[string]string{
				"olake.io/created-by":           "olake-k8s-worker",
				"olake.io/created-at":           time.Now().Format(time.RFC3339),
				"olake.io/original-workflow-id": spec.OriginalWorkflowID,
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:    "connector",
					Image:   spec.Image,
					Command: spec.Command,
					Args:    spec.Args,
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "job-storage",
							MountPath: "/mnt/config",
							SubPath:   workflowDir,
						},
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: utils.ParseQuantity("256Mi"),
							corev1.ResourceCPU:    utils.ParseQuantity("100m"),
						},
						// No limits for autoscaling flexibility
					},
					Env: []corev1.EnvVar{
						{
							Name:  "OLAKE_WORKFLOW_ID",
							Value: spec.OriginalWorkflowID,
						},
						{
							Name:  "OLAKE_OPERATION",
							Value: string(spec.Operation),
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "job-storage",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "olake-jobs-pvc",
						},
					},
				},
			},
		},
	}

	logger.Infof("Creating Pod %s with image %s", spec.Name, spec.Image)
	result, err := k.clientset.CoreV1().Pods(k.namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		logger.Errorf("Failed to create Pod %s: %v", spec.Name, err)
		return nil, err
	}

	logger.Infof("Successfully created Pod %s", spec.Name)
	return result, nil
}

// WaitForPodCompletion waits for a Pod to complete and returns the result
func (k *K8sPodManager) WaitForPodCompletion(ctx context.Context, podName string, timeout time.Duration) (map[string]interface{}, error) {
	logger.Infof("Waiting for Pod %s to complete (timeout: %v)", podName, timeout)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		pod, err := k.clientset.CoreV1().Pods(k.namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get pod status: %v", err)
		}

		// Check if pod completed successfully
		if pod.Status.Phase == corev1.PodSucceeded {
			logger.Infof("Pod %s completed successfully", podName)
			return k.getPodResults(ctx, podName)
		}

		// Check if pod failed
		if pod.Status.Phase == corev1.PodFailed {
			logger.Errorf("Pod %s failed", podName)
			logs, _ := k.getPodLogs(ctx, podName)
			return nil, fmt.Errorf("pod failed: %s", logs)
		}

		// Wait before checking again
		time.Sleep(5 * time.Second)
	}

	logger.Errorf("Pod %s timed out after %v", podName, timeout)
	return nil, fmt.Errorf("pod timed out after %v", timeout)
}

// CleanupPod removes a pod (but keeps work directory for UI access)
func (k *K8sPodManager) CleanupPod(ctx context.Context, podName string) error {
	logger.Infof("Cleaning up Pod %s", podName)

	// Delete the pod only
	err := k.clientset.CoreV1().Pods(k.namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		logger.Errorf("Failed to delete pod %s: %v", podName, err)
		return fmt.Errorf("failed to delete pod: %v", err)
	}

	logger.Infof("Successfully cleaned up Pod %s (directory preserved)", podName)
	return nil
}

// getPodLogs retrieves logs from a completed pod
func (k *K8sPodManager) getPodLogs(ctx context.Context, podName string) (string, error) {
	req := k.clientset.CoreV1().Pods(k.namespace).GetLogs(podName, &corev1.PodLogOptions{})
	logs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer logs.Close()

	buf := make([]byte, 4096)
	var result string
	for {
		n, err := logs.Read(buf)
		if n > 0 {
			result += string(buf[:n])
		}
		if err != nil {
			break
		}
	}

	return result, nil
}

// getPodResults extracts results from completed pod
func (k *K8sPodManager) getPodResults(ctx context.Context, podName string) (map[string]interface{}, error) {
	// Get the pod to find the workflow directory
	pod, err := k.clientset.CoreV1().Pods(k.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %v", err)
	}

	// Extract ORIGINAL workflow ID from pod annotations
	workflowID := pod.Annotations["olake.io/original-workflow-id"]
	if workflowID == "" {
		return nil, fmt.Errorf("original workflow ID not found in pod annotations")
	}

	// Determine the operation type
	operation := shared.Command(pod.Labels["operation"])

	if operation == shared.Discover {
		// For discover operations, read streams.json file (like Docker does)
		workflowDir := k.filesystemHelper.GetWorkflowDirectory(operation, workflowID)
		catalogPath := k.filesystemHelper.GetFilePath(workflowDir, "streams.json")
		return utils.ParseJSONFile(catalogPath)
	} else {
		// For other operations (check, sync), parse logs as before
		logs, err := k.getPodLogs(ctx, podName)
		if err != nil {
			return nil, fmt.Errorf("failed to get pod logs: %v", err)
		}
		logger.Debugf("Raw pod logs for pod %s:\n%s", podName, logs)
		return utils.ParseJobOutput(logs)
	}
}

// PodActivityRequest defines a request for executing a pod activity
type PodActivityRequest struct {
	WorkflowID string
	Operation  shared.Command
	Image      string
	Args       []string
	Configs    []shared.JobConfig
	Timeout    time.Duration
}

// ExecutePodActivity executes a pod activity with common workflow
func (k *K8sPodManager) ExecutePodActivity(ctx context.Context, req PodActivityRequest) (map[string]interface{}, error) {
	// Create Pod specification
	podSpec := &PodSpec{
		Name:               utils.SanitizeK8sName(req.WorkflowID),
		OriginalWorkflowID: req.WorkflowID,
		Image:              req.Image,
		Command:            []string{},
		Args:               req.Args,
		Operation:          req.Operation,
	}

	// Create Pod
	pod, err := k.CreatePod(ctx, podSpec, req.Configs)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}

	// Always cleanup pod when done
	defer func() {
		if err := k.CleanupPod(ctx, pod.Name); err != nil {
			logger.Errorf("Failed to cleanup pod %s: %v", pod.Name, err)
		}
	}()

	// Wait for pod completion
	return k.WaitForPodCompletion(ctx, pod.Name, req.Timeout)
}
