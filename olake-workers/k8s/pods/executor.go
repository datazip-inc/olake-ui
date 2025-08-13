package pods

import (
	"context"
	"fmt"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"olake-ui/olake-workers/k8s/logger"
	"olake-ui/olake-workers/k8s/shared"
	"olake-ui/olake-workers/k8s/utils/k8s"
)

// PodSpec defines the specification for creating a Kubernetes Pod
type PodSpec struct {
	Name               string
	Image              string
	Command            []string
	Args               []string
	OriginalWorkflowID string
	JobID              int
	Operation          shared.Command
	ConnectorType      string
}

// CreatePod creates a Kubernetes Pod for running job operations
func (k *K8sPodManager) CreatePod(ctx context.Context, spec *PodSpec, configs []shared.JobConfig) (*corev1.Pod, error) {
	// Get workflow directory using filesystem helper with the actual operation type
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
				// Standard Kubernetes labels
				"app.kubernetes.io/name":       "olake",
				"app.kubernetes.io/component":  fmt.Sprintf("%s-%s", spec.ConnectorType, string(spec.Operation)),
				"app.kubernetes.io/managed-by": "olake-workers",

				// Custom Olake labels
				"olake.io/operation-type": string(spec.Operation),
				"olake.io/connector":      spec.ConnectorType,
				"olake.io/job-id":         strconv.Itoa(spec.JobID),
				"olake.io/workflow-id":    k8s.SanitizeName(spec.OriginalWorkflowID),
			},
			Annotations: map[string]string{
				"olake.io/created-by-pod": k8s.GenerateWorkerIdentity(),
				"olake.io/created-at":     time.Now().Format(time.RFC3339),
				"olake.io/workflow-id":    spec.OriginalWorkflowID,
				"olake.io/operation-type": string(spec.Operation),
				"olake.io/connector-type": spec.ConnectorType,
				"olake.io/job-id":         strconv.Itoa(spec.JobID),
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			NodeSelector:  k.getNodeSelectorForJob(spec.JobID),
			Tolerations:   k.getTolerationsForJob(spec.JobID),
			Affinity:      k.buildAffinityForJob(spec.JobID, spec.Operation),
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
							corev1.ResourceMemory: k8s.ParseQuantity("256Mi"),
							corev1.ResourceCPU:    k8s.ParseQuantity("100m"),
						},
						// No limits for flexibility
					},
					Env: []corev1.EnvVar{
						{
							Name:  "OLAKE_WORKFLOW_ID",
							Value: spec.OriginalWorkflowID,
						},
						{
							Name:  "OLAKE_SECRET_KEY",
							Value: k.config.Kubernetes.OLakeSecretKey,
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "job-storage",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: k.config.Kubernetes.PVCName,
						},
					},
				},
			},
		},
	}

	// Set ServiceAccountName only if configured (non-empty)
	// If empty, Kubernetes will use the namespace's default service account
	if k.config.Kubernetes.JobServiceAccount != "" && k.config.Kubernetes.JobServiceAccount != "default" {
		pod.Spec.ServiceAccountName = k.config.Kubernetes.JobServiceAccount
	}

	logger.Infof("Creating Pod %s with image %s", spec.Name, spec.Image)
	result, err := k.clientset.CoreV1().Pods(k.namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		logger.Errorf("Failed to create Pod %s: %v", spec.Name, err)
		return nil, err
	}

	logger.Debugf("Successfully created Pod %s", spec.Name)
	return result, nil
}

// WaitForPodCompletion waits for a Pod to complete and returns the result
func (k *K8sPodManager) WaitForPodCompletion(ctx context.Context, podName string, timeout time.Duration, operation shared.Command, workflowID string) (map[string]interface{}, error) {
	logger.Debugf("Waiting for Pod %s to complete (timeout: %v)", podName, timeout)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		pod, err := k.clientset.CoreV1().Pods(k.namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get pod status: %v", err)
		}

		// Check if pod completed successfully
		if pod.Status.Phase == corev1.PodSucceeded {
			logger.Infof("Pod %s completed successfully", podName)
			return k.getPodResults(ctx, podName, operation, workflowID)
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
	logger.Debugf("Cleaning up Pod %s in namespace %s", podName, k.namespace)

	// Delete the pod only
	err := k.clientset.CoreV1().Pods(k.namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		// Enhanced error logging with more context
		logger.Errorf("Failed to delete pod %s in namespace %s: %v (error type: %T)",
			podName, k.namespace, err, err)
		return fmt.Errorf("failed to delete pod %s in namespace %s: %v", podName, k.namespace, err)
	}

	logger.Debugf("Successfully cleaned up Pod %s in namespace %s",
		podName, k.namespace)
	return nil
}

// getNodeSelectorForJob returns node selector configuration for the given jobID
// Returns empty map if no mapping is found (graceful fallback)
func (k *K8sPodManager) getNodeSelectorForJob(jobID int) map[string]string {
	if mapping, exists := k.config.Kubernetes.JobMapping[jobID]; exists {
		logger.Infof("Found node mapping for JobID %d: %v", jobID, mapping)
		return mapping
	}
	logger.Debugf("No node mapping found for JobID %d, using default scheduling", jobID)
	return make(map[string]string)
}

// getTolerationsForJob returns tolerations configuration for the given jobID
// Returns empty slice if no mapping is found (graceful fallback)
func (k *K8sPodManager) getTolerationsForJob(jobID int) []corev1.Toleration {
	// TODO: Implement Helm ConfigMap lookup for jobID-based tolerations
	// For now, return empty slice as graceful fallback
	logger.Debugf("JobID-based tolerations lookup for job %d - returning empty (no mapping)", jobID)
	return []corev1.Toleration{}
}

// buildNodeAffinity creates node affinity from JobID mapping configuration
// Converts map[string]string to preferredDuringScheduling node affinity with proper edge case handling
func (k *K8sPodManager) buildNodeAffinity(nodeSelectorMap map[string]string) (*corev1.NodeAffinity, error) {
	if len(nodeSelectorMap) == 0 {
		return nil, nil // No mapping, no node affinity
	}

	var matchExpressions []corev1.NodeSelectorRequirement
	for key, value := range nodeSelectorMap {
		// Edge case: Ensure the key and value from config are not empty
		if key == "" || value == "" {
			logger.Warnf("Skipping invalid node mapping entry with empty key or value: key=%s, value=%s", key, value)
			continue
		}

		// The 'Values' field for the 'In' operator must be a slice of strings
		matchExpressions = append(matchExpressions, corev1.NodeSelectorRequirement{
			Key:      key,
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{value},
		})
	}

	// If all entries were invalid, there's nothing to do
	if len(matchExpressions) == 0 {
		return nil, fmt.Errorf("all node mapping entries were invalid")
	}

	return &corev1.NodeAffinity{
		PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
			{
				Weight: 100, // High preference for JobID-mapped nodes
				Preference: corev1.NodeSelectorTerm{
					MatchExpressions: matchExpressions,
				},
			},
		},
	}, nil
}

// buildAffinityForJob builds affinity rules for the given jobID and operation type
// Implements operation-based anti-affinity and JobID-based node affinity
func (k *K8sPodManager) buildAffinityForJob(jobID int, operation shared.Command) *corev1.Affinity {
	var affinity *corev1.Affinity

	// Apply anti-affinity rules for sync operations to spread pods across nodes
	if operation == shared.Sync {
		logger.Debugf("Applying sync operation anti-affinity rules for jobID %d", jobID)
		affinity = &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"olake.io/operation-type": "sync",
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		}
	} else {
		logger.Debugf("No anti-affinity rules applied for %s operation (jobID %d)", operation, jobID)
	}

	// Add JobID-based node affinity using jobMapping configuration
	nodeMapping := k.getNodeSelectorForJob(jobID)
	if len(nodeMapping) > 0 {
		nodeAffinity, err := k.buildNodeAffinity(nodeMapping)
		if err != nil {
			logger.Errorf("Failed to build node affinity for JobID %d: %v", jobID, err)
		} else if nodeAffinity != nil {
			if affinity == nil {
				affinity = &corev1.Affinity{}
			}
			affinity.NodeAffinity = nodeAffinity
			logger.Infof("Applied node affinity for JobID %d: preferring nodes with %v", jobID, nodeMapping)
		}
	} else {
		logger.Debugf("No JobID-based node affinity applied for JobID %d (no mapping found)", jobID)
	}

	return affinity
}
