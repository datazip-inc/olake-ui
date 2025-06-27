package pods

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"olake-k8s-worker/logger"
	"olake-k8s-worker/shared"
	"olake-k8s-worker/utils"
)

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
	logger.Infof("Cleaning up Pod %s in namespace %s", podName, k.namespace)

	// Delete the pod only
	err := k.clientset.CoreV1().Pods(k.namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		// Enhanced error logging with more context
		logger.Errorf("Failed to delete pod %s in namespace %s: %v (error type: %T)",
			podName, k.namespace, err, err)
		return fmt.Errorf("failed to delete pod %s in namespace %s: %v", podName, k.namespace, err)
	}

	logger.Infof("Successfully cleaned up Pod %s in namespace %s (directory preserved for UI access)",
		podName, k.namespace)
	return nil
}