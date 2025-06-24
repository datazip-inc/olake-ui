package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/datazip/olake-frontend/server/internal/kms"
	"github.com/datazip/olake-frontend/server/internal/models"
	"github.com/datazip/olake-frontend/server/internal/temporal"
	"github.com/datazip/olake-frontend/server/utils"
	"go.temporal.io/api/workflowservice/v1"
)

// get id from path
func GetIDFromPath(c *web.Controller) int {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid id")
		return 0
	}
	return id
}

// setUsernames sets the created and updated usernames if available
func setUsernames(createdBy, updatedBy *string, creator, updater *models.User) {
	if creator != nil {
		*createdBy = creator.Username
	}
	if updater != nil {
		*updatedBy = updater.Username
	}
}

// buildJobDataItems creates job data items with workflow information
// Returns (jobItems, success). If success is false, an error occurred and the handler should return.
func buildJobDataItems(jobs []*models.Job, err error, projectIDStr, contextType string, tempClient *temporal.Client, controller *web.Controller) ([]models.JobDataItem, bool) {
	jobItems := make([]models.JobDataItem, 0)

	if err != nil {
		return jobItems, true // No jobs is OK, return empty slice
	}

	for _, job := range jobs {
		jobInfo := models.JobDataItem{
			Name:     job.Name,
			ID:       job.ID,
			Activate: job.Active,
		}

		// Set source/destination info based on context
		if contextType == "source" && job.DestID != nil {
			jobInfo.DestinationName = job.DestID.Name
			jobInfo.DestinationType = job.DestID.DestType
		} else if contextType == "destination" && job.SourceID != nil {
			jobInfo.SourceName = job.SourceID.Name
			jobInfo.SourceType = job.SourceID.Type
		}

		if !setJobWorkflowInfo(&jobInfo, job.ID, projectIDStr, tempClient, controller) {
			return nil, false // Error occurred, signal failure
		}
		jobItems = append(jobItems, jobInfo)
	}

	return jobItems, true
}

// setJobWorkflowInfo fetches and sets workflow execution information for a job
// Returns false if an error occurred that should stop processing
func setJobWorkflowInfo(jobInfo *models.JobDataItem, jobID int, projectIDStr string, tempClient *temporal.Client, controller *web.Controller) bool {
	query := fmt.Sprintf("WorkflowId between 'sync-%s-%d' and 'sync-%s-%d-~'", projectIDStr, jobID, projectIDStr, jobID)

	resp, err := tempClient.ListWorkflow(context.Background(), &workflowservice.ListWorkflowExecutionsRequest{
		Query:    query,
		PageSize: 1,
	})

	if err != nil {
		utils.ErrorResponse(controller, http.StatusInternalServerError, fmt.Sprintf("failed to list workflows: %v", err))
		return false
	}

	if len(resp.Executions) > 0 {
		jobInfo.LastRunTime = resp.Executions[0].StartTime.AsTime().Format(time.RFC3339)
		jobInfo.LastRunState = resp.Executions[0].Status.String()
	} else {
		jobInfo.LastRunTime = ""
		jobInfo.LastRunState = ""
	}
	return true
}

// EncryptJSONValues encrypts each value in the JSON object separately
func EncryptJSONValues(rawConfig string) (string, error) {
	// Unmarshal to map
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(rawConfig), &configMap); err != nil {
		return "", fmt.Errorf("invalid JSON: %v", err)
	}

	// Encrypt values
	encryptedMap := make(map[string]interface{})
	for k, v := range configMap {
		encryptedBytes, err := kms.Encrypt(fmt.Sprintf("%v", v))
		if err != nil {
			return "", fmt.Errorf("encryption failed for key '%s': %v", k, err)
		}
		// Base64 encode encrypted value to keep JSON valid
		encryptedMap[k] = base64.StdEncoding.EncodeToString(encryptedBytes)
	}

	// Marshal back to JSON
	encryptedJSON, err := json.Marshal(encryptedMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal encrypted JSON: %v", err)
	}

	return string(encryptedJSON), nil
}

// EncryptJSONString encrypts the entire JSON string as a single value
func EncryptJSONString(rawConfig string) (string, error) {
	return EncryptJSONAsObject(rawConfig)
}

// EncryptJSONAsObject encrypts the JSON string and returns it as a structured JSON object
// The output will be in the format: {"encrypted_data": "base64-encoded-encrypted-json"}
func EncryptJSONAsObject(rawConfig string) (string, error) {
	// Encrypt the entire config string
	encryptedBytes, err := kms.Encrypt(rawConfig)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %v", err)
	}

	// Create a structured object with the encrypted data
	encryptedObj := struct {
		EncryptedData string `json:"encrypted_data"`
	}{
		EncryptedData: base64.StdEncoding.EncodeToString(encryptedBytes),
	}

	// Marshal to JSON
	encryptedJSON, err := json.Marshal(encryptedObj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal encrypted data: %v", err)
	}

	return string(encryptedJSON), nil
}

// DecryptJSONObject decrypts a JSON object in the format {"encrypted_data": "base64-encoded-encrypted-json"}
// and returns the original JSON string
func DecryptJSONString(encryptedObjStr string) (string, error) {
	// Unmarshal the encrypted object
	var encryptedObj struct {
		EncryptedData string `json:"encrypted_data"`
	}

	if err := json.Unmarshal([]byte(encryptedObjStr), &encryptedObj); err != nil {
		// If it's not in the expected format, try to decrypt it directly
		// (for backward compatibility with old format)
		return decryptRawString(encryptedObjStr)
	}

	// Decode the base64-encoded encrypted data
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedObj.EncryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 data: %v", err)
	}

	// Decrypt the data
	decrypted, err := kms.Decrypt(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %v", err)
	}

	return string(decrypted), nil
}

// decryptRawString is a helper function to handle backward compatibility with old format
func decryptRawString(encryptedStr string) (string, error) {
	// Try to decode as base64 first
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedStr)
	if err != nil {
		// If not base64, try to use it as is
		encryptedData = []byte(encryptedStr)
	}

	// Try to decrypt
	decrypted, err := kms.Decrypt(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %v", err)
	}

	return string(decrypted), nil
}
