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

// PodActivityRequest defines a request for executing a pod activity
// This struct encapsulates all the information needed to execute a Temporal activity
// as a Kubernetes pod, bridging the gap between Temporal's activity model and K8s execution.
type PodActivityRequest struct {
	WorkflowID    string             // Unique identifier for the Temporal workflow instance
	JobID         int                // Database job ID for labeling and resource mapping
	Operation     shared.Command     // Type of operation (sync, discover, check) - affects result retrieval
	ConnectorType string             // Source connector type (mysql, postgres, etc.) for labeling
	Image         string             // Full Docker image name for the connector container
	Args          []string           // Command-line arguments passed to the connector
	Configs       []shared.JobConfig // Configuration files to mount into the pod
	Timeout       time.Duration      // Maximum execution time before pod is considered failed
}

// PodSpec defines the specification for creating a Kubernetes Pod
// This is an internal representation used during pod creation, containing both
// the container specification and metadata needed for proper labeling and organization.
type PodSpec struct {
	Name               string         // Sanitized pod name derived from WorkflowID
	Image              string         // Docker image for the connector container
	Command            []string       // Container entrypoint command (usually empty, uses image default)
	Args               []string       // Arguments passed to the container command
	OriginalWorkflowID string         // Original unsanitized WorkflowID for directory naming
	JobID              int            // Job ID for resource mapping and labeling
	Operation          shared.Command // Operation type for labeling and result retrieval strategy
	ConnectorType      string         // Connector type for labeling and identification
}

// CreatePod creates a Kubernetes Pod for running job operations
// This function handles the complete pod creation process: preparing shared storage,
// writing configuration files, constructing the pod specification with proper labels/annotations,
// and submitting it to the Kubernetes API server for execution.
func (k *K8sPodManager) CreatePod(ctx context.Context, spec *PodSpec, configs []shared.JobConfig) (*corev1.Pod, error) {
	// Determine the workflow directory name based on operation type and workflow ID
	// Different operations use different directory naming strategies (hash vs direct)
	workflowDir := k.filesystemHelper.GetWorkflowDirectory(spec.Operation, spec.OriginalWorkflowID)

	// Prepare the shared storage directory for this workflow
	// This creates the directory structure on the NFS/shared volume
	if err := k.filesystemHelper.SetupWorkDirectory(workflowDir); err != nil {
		return nil, fmt.Errorf("failed to setup work directory: %v", err)
	}

	// Write all configuration files to the shared storage directory
	// These files (config.json, streams.json, etc.) will be mounted into the pod
	if err := k.filesystemHelper.WriteConfigFiles(workflowDir, configs); err != nil {
		return nil, fmt.Errorf("failed to write config files: %v", err)
	}

	// Construct the complete Kubernetes Pod specification
	// This includes metadata (labels/annotations), container spec, volumes, and scheduling preferences
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.Name,   // Sanitized name safe for Kubernetes
			Namespace: k.namespace, // Target namespace for pod creation

			// Labels are used for querying, filtering, and organizing pods
			Labels: map[string]string{
				// Standard Kubernetes labels for ecosystem compatibility
				"app.kubernetes.io/name":       "olake",                                                          // Application name
				"app.kubernetes.io/component":  fmt.Sprintf("%s-%s", spec.ConnectorType, string(spec.Operation)), // Component identifier
				"app.kubernetes.io/managed-by": "olake-workers",                                                  // Management tool

				// Custom Olake labels for internal operations and queries
				"olake.io/operation-type": string(spec.Operation),                    // sync, discover, or check
				"olake.io/connector":      spec.ConnectorType,                        // mysql, postgres, etc.
				"olake.io/job-id":         strconv.Itoa(spec.JobID),                  // Database job reference
				"olake.io/workflow-id":    k8s.SanitizeName(spec.OriginalWorkflowID), // Sanitized workflow ID
			},

			// Annotations store metadata that doesn't affect pod selection/scheduling
			Annotations: map[string]string{
				"olake.io/created-by-pod": fmt.Sprintf("olake.io/olake-workers/%s", k.config.Worker.WorkerIdentity), // Which worker pod created this
				"olake.io/created-at":     time.Now().Format(time.RFC3339),                                          // Creation timestamp
				"olake.io/workflow-id":    spec.OriginalWorkflowID,                                                  // Original unsanitized workflow ID
				"olake.io/operation-type": string(spec.Operation),                                                   // Operation type for reference
				"olake.io/connector-type": spec.ConnectorType,                                                       // Connector type for reference
				"olake.io/job-id":         strconv.Itoa(spec.JobID),                                                 // Job ID for reference
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			NodeSelector:  k.getNodeSelectorForJob(spec.JobID, spec.Operation),
			Tolerations:   []corev1.Toleration{}, // No tolerations supported yet
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
			return k.getPodResults(podName, operation, workflowID)
		}

		// Check if pod failed
		if pod.Status.Phase == corev1.PodFailed {
			logger.Errorf("Pod %s failed", podName)
			return nil, fmt.Errorf("pod %s failed with status: %s", podName, pod.Status.Phase)
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
		return fmt.Errorf("failed to delete pod %s in namespace %s: %v", podName, k.namespace, err)
	}

	logger.Debugf("Successfully cleaned up Pod %s in namespace %s",
		podName, k.namespace)
	return nil
}

// getNodeSelectorForJob returns node selector configuration for the given jobID
// Returns empty map if no mapping is found (graceful fallback)
// Only applies node mapping for sync operations
func (k *K8sPodManager) getNodeSelectorForJob(jobID int, operation shared.Command) map[string]string {
	// Only apply node mapping for sync operations
	if operation != shared.Sync {
		return make(map[string]string)
	}

	if mapping, exists := k.config.Kubernetes.JobMapping[jobID]; exists {
		logger.Infof("Found node mapping for JobID %d: %v", jobID, mapping)
		return mapping
	}
	logger.Debugf("No node mapping found for JobID %d, using default scheduling", jobID)
	return make(map[string]string)
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
	// Skip affinity rules for jobID 0 (test/discover operations)
	if jobID == 0 {
		return nil
	}

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
	nodeMapping := k.getNodeSelectorForJob(jobID, operation)
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

// ExecutePodActivity executes a pod activity with common workflow
// This is the main entry point for executing Temporal activities as Kubernetes pods.
// It orchestrates the complete lifecycle: pod creation, execution monitoring, result retrieval, and cleanup.
func (k *K8sPodManager) ExecutePodActivity(ctx context.Context, req PodActivityRequest) (map[string]interface{}, error) {
	// Transform the high-level activity request into a concrete Kubernetes pod specification
	// This bridges the gap between Temporal's activity model and Kubernetes execution
	podSpec := &PodSpec{
		Name:               k8s.SanitizeName(req.WorkflowID), // Safe Kubernetes pod name
		OriginalWorkflowID: req.WorkflowID,                   // Original ID for directory naming
		JobID:              req.JobID,                        // Database job reference
		Image:              req.Image,                        // Connector container image
		Command:            []string{},                       // Use image default entrypoint
		Args:               req.Args,                         // Connector-specific arguments
		Operation:          req.Operation,                    // Operation type (affects result retrieval)
		ConnectorType:      req.ConnectorType,                // Connector type for labeling
	}

	// Create the Kubernetes pod with all necessary configuration and volume mounts
	pod, err := k.CreatePod(ctx, podSpec, req.Configs)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}

	// Ensure pod cleanup happens regardless of success or failure
	// This prevents resource leaks and maintains cluster hygiene
	defer func() {
		if err := k.CleanupPod(ctx, pod.Name); err != nil {
			logger.Errorf("Failed to cleanup pod %s for %s operation (workflow: %s): %v",
				pod.Name, req.Operation, req.WorkflowID, err)
			// Note: We continue execution despite cleanup failure as the core operation may have succeeded
			// and cleanup failures shouldn't invalidate successful work results
		}
	}()

	// Wait for pod completion
	return k.WaitForPodCompletion(ctx, pod.Name, req.Timeout, req.Operation, req.WorkflowID)
}
