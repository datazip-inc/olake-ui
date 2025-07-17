package pods

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"

	"olake-ui/olake-workers/k8s/logger"
	"olake-ui/olake-workers/k8s/shared"
	"olake-ui/olake-workers/k8s/utils/filesystem"
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
func (k *K8sPodManager) getPodResults(ctx context.Context, podName string, operation shared.Command, workflowID string) (map[string]interface{}, error) {
	// For sync operations, prioritize reading the state.json file for results
	if operation == shared.Sync && workflowID != "" {
		fsHelper := filesystem.NewHelper()
		stateData, err := fsHelper.ReadAndValidateStateFile(workflowID)
		if err == nil {
			var result map[string]interface{}
			if unmarshalErr := json.Unmarshal(stateData, &result); unmarshalErr == nil {
				logger.Debugf("Successfully read state.json for sync pod %s", podName)
				return result, nil
			} else {
				// This case is unlikely if ReadAndValidateStateFile truly validates JSON, but it's safe to handle
				logger.Warnf("Failed to parse validated state.json for sync pod %s: %v, falling back to logs", podName, unmarshalErr)
			}
		} else {
			// Log the error from ReadAndValidateStateFile, which could be os.ErrNotExist or something else
			logger.Warnf("Failed to read state.json for sync pod %s: %v, falling back to logs", podName, err)
		}
	}

	// Fallback/default: Parse logs from the pod
	logger.Debugf("Parsing pod logs for pod %s", podName)
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
	return k.WaitForPodCompletion(ctx, pod.Name, req.Timeout, req.Operation, req.WorkflowID)
}
