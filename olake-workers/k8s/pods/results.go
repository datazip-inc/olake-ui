package pods

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"olake-ui/olake-workers/k8s/logger"
	"olake-ui/olake-workers/k8s/shared"
	"olake-ui/olake-workers/k8s/utils/json"
	"olake-ui/olake-workers/k8s/utils/k8s"
	"olake-ui/olake-workers/k8s/utils/parser"
)

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

	// Extract workflow ID and operation type from pod annotations
	workflowID := pod.Annotations["olake.io/original-workflow-id"]
	if workflowID == "" {
		return nil, fmt.Errorf("original workflow ID not found in pod annotations")
	}

	operationType := pod.Annotations["olake.io/operation-type"]
	if operationType == "" {
		logger.Warnf("Operation type not found in pod annotations for pod %s, falling back to log parsing", podName)
		return k.fallbackToLogParsing(ctx, podName, workflowID)
	}

	logger.Debugf("Processing results for pod %s with operation type: %s", podName, operationType)

	// Handle results based on explicit operation type from annotations
	switch shared.Command(operationType) {
	case shared.Discover:
		// For discover operations, look for streams.json catalog file
		workflowDir := k.filesystemHelper.GetWorkflowDirectory(shared.Discover, workflowID)
		catalogPath := k.filesystemHelper.GetFilePath(workflowDir, "streams.json")

		if result, err := json.ParseJSONFile(catalogPath); err == nil {
			logger.Debugf("Successfully parsed streams.json for discover operation")
			return result, nil
		} else {
			logger.Warnf("Failed to parse streams.json for discover operation: %v, falling back to logs", err)
			return k.fallbackToLogParsing(ctx, podName, workflowID)
		}

	case shared.Sync, shared.Check:
		// For sync and check operations, parse logs directly
		return k.fallbackToLogParsing(ctx, podName, workflowID)

	default:
		logger.Warnf("Unknown operation type '%s' for pod %s, falling back to log parsing", operationType, podName)
		return k.fallbackToLogParsing(ctx, podName, workflowID)
	}
}

// fallbackToLogParsing handles result extraction by parsing pod logs
// This is used as a fallback when annotation-based detection fails or for operations that rely on log parsing
func (k *K8sPodManager) fallbackToLogParsing(ctx context.Context, podName, workflowID string) (map[string]interface{}, error) {
	logs, err := k.getPodLogs(ctx, podName)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod logs: %v", err)
	}

	logger.Debugf("Raw pod logs for pod %s:\n%s", podName, logs)
	return parser.ParseJobOutput(logs)
}

// PodActivityRequest defines a request for executing a pod activity
type PodActivityRequest struct {
	WorkflowID    string
	JobID         int
	Operation     shared.Command
	ConnectorType string
	Image         string
	Args          []string
	Configs       []shared.JobConfig
	Timeout       time.Duration
}

// ExecutePodActivity executes a pod activity with common workflow
func (k *K8sPodManager) ExecutePodActivity(ctx context.Context, req PodActivityRequest) (map[string]interface{}, error) {
	// Create Pod specification
	podSpec := &PodSpec{
		Name:               k8s.SanitizeName(req.WorkflowID),
		OriginalWorkflowID: req.WorkflowID,
		JobID:              req.JobID,
		Image:              req.Image,
		Command:            []string{},
		Args:               req.Args,
		Operation:          req.Operation,
		ConnectorType:      req.ConnectorType,
	}

	// Create Pod
	pod, err := k.CreatePod(ctx, podSpec, req.Configs)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}

	// Always cleanup pod when done
	defer func() {
		if err := k.CleanupPod(ctx, pod.Name); err != nil {
			logger.Errorf("Failed to cleanup pod %s for %s operation (workflow: %s): %v",
				pod.Name, req.Operation, req.WorkflowID, err)
			// Note: We continue execution despite cleanup failure as the core operation may have succeeded
			// and cleanup failures shouldn't invalidate successful work results
		}
	}()

	// Wait for pod completion
	return k.WaitForPodCompletion(ctx, pod.Name, req.Timeout)
}
