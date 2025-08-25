package pods

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"

	"github.com/datazip-inc/olake-ui/olake-workers/k8s/logger"
	"github.com/datazip-inc/olake-ui/olake-workers/k8s/shared"
	"github.com/datazip-inc/olake-ui/olake-workers/k8s/utils/filesystem"
	"github.com/datazip-inc/olake-ui/olake-workers/k8s/utils/parser"
)

// getPodResults extracts results from completed pod
// This function implements operation-specific result retrieval strategies:
// - Sync operations: read state.json from shared filesystem
// - Discover operations: read streams.json from shared filesystem
// - Check operations: parse connection status from pod logs
// The strategy varies because different activities output results in different ways.
func (k *K8sPodManager) getPodResults(podName string, operation shared.Command, workflowID string) (map[string]interface{}, error) {
	// SYNC OPERATIONS: Extract results from state.json file on shared filesystem
	// Sync operations write their final state (including metrics, counts, etc.) to state.json
	// This file contains the complete sync results and is the authoritative source for sync status
	if operation == shared.Sync && workflowID != "" {
		if data, err := filesystem.NewHelper().ReadAndValidateStateFile(workflowID); err == nil {
			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err == nil {
				logger.Debugf("Successfully read state.json for sync pod %s", podName)
				return result, nil
			}
			logger.Errorf("Failed to parse validated state.json for sync pod %s: %v", podName, err)
		} else {
			logger.Warnf("Failed to read state.json for sync pod %s: %v", podName, err)
		}
	}

	// DISCOVER OPERATIONS: Extract results from streams.json file on shared filesystem
	// Discover operations scan the data source and write discovered streams/tables to streams.json
	// This file contains the catalog of available data streams and their schema information
	if operation == shared.Discover && workflowID != "" {
		if data, err := filesystem.NewHelper().ReadAndValidateStreamsFile(workflowID); err == nil {
			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err == nil {
				logger.Debugf("Successfully read streams.json for discover pod %s", podName)
				logger.Debugf("Discovered streams configuration: %s", string(data))
				return result, nil
			}
			logger.Errorf("Failed to parse validated streams.json for discover pod %s: %v", podName, err)
		} else {
			logger.Warnf("Failed to read streams.json for discover pod %s: %v", podName, err)
		}
	}

	// CHECK OPERATIONS: Extract results from pod logs using log parsing
	// Check operations test connectivity and write status messages to stdout
	// Unlike sync/discover, check operations don't write files - they only output to logs
	if operation == shared.Check {
		if logs, err := k.getPodLogs(context.Background(), podName); err == nil { // TODO: set context at parent i.e. context to be received from whoever calls getPodResults(...)
			if result, err := parser.ParseJobOutput(logs); err == nil {
				logger.Debugf("Successfully parsed connection status from logs for check pod %s", podName)
				return result, nil
			} else {
				logger.Errorf("Failed to parse connection status from logs for check pod %s: %v", podName, err)
				return nil, fmt.Errorf("failed to parse connection status from logs: %v", err)
			}
		} else {
			logger.Errorf("Failed to get logs for check pod %s: %v", podName, err)
			return nil, fmt.Errorf("failed to get logs for check pod %s: %v", podName, err)
		}
	}

	// No fallback available - return error indicating file-based results are required
	return nil, fmt.Errorf("failed to read results from filesystem for pod %s, operation %s", podName, operation)
}

// getPodLogs retrieves logs from a completed pod
func (k *K8sPodManager) getPodLogs(ctx context.Context, podName string) (string, error) {
	req := k.clientset.CoreV1().Pods(k.namespace).GetLogs(podName, &corev1.PodLogOptions{})
	logs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %v", err)
	}
	defer logs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, logs)
	if err != nil {
		return "", fmt.Errorf("failed to read pod logs: %v", err)
	}

	return buf.String(), nil
}
